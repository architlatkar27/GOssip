package storage

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// implements WAL interface
// should this be a singleton?
type FileBasedWALWriter struct {
	// maybe this should contain the base location where logs are
	// along with that the index etc?

	// storage configuration
	config *StorageConfig

	mu sync.RWMutex
	file *os.File
	writer *bufio.Writer
	filePath string

	// State tracking
    currentIndex int64
    lastSync     time.Time
    
    // Background sync - flush the file to disk
    syncTicker *time.Ticker // This keeps performing flush every X seconds
    stopCh     chan struct{} // A message on this channel will stop the above ticker
    wg         sync.WaitGroup // 
    
    // Metrics
    entriesWritten int64
    bytesWritten   int64
    syncCount      int64
}

// WALEntry represents a single WAL entry
type WALEntry struct {
	Index     int64     `json:"index"`
	Term      int64     `json:"term"`
	Type      WALType   `json:"type"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Checksum  uint32    `json:"checksum"`
}

var walCreatorLock = &sync.RWMutex{}

var instance *FileBasedWALWriter

func getOrCreateWALWriter(config *StorageConfig) (*FileBasedWALWriter, error) {
	walCreatorLock.RLock()
	defer walCreatorLock.RUnlock()
	if instance == nil {
		walCreatorLock.Lock()
		defer walCreatorLock.Unlock()
		if instance == nil {
			log.Printf("creating WAL Writer")
			return newFileBasedWALWriter(config)
		} else {
			return instance, nil
		}
	} else {
		return instance, nil
	}
}

func newFileBasedWALWriter(config *StorageConfig) (*FileBasedWALWriter, error) {
	instance := FileBasedWALWriter{}
	instance.mu = sync.RWMutex{}
	instance.config = config

	// file assignment
	if err := os.MkdirAll(config.WALDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to cerate WAL directory: %v", err)
	}
	filePath := filepath.Join(config.WALDir, "wal.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not create/open WAL file: %v", err)
	}

	instance.file = file
	instance.filePath = filePath

	// Recover current index from existing file
	currentIndex, err := instance.recoverIndex()
    if err != nil {
        file.Close()
        return nil, fmt.Errorf("failed to recover WAL index: %w", err)
    }
	instance.currentIndex = currentIndex
    
    // Start background sync routine
    instance.backgroundFlushRoutine()
    
    return &instance, nil
}


func(w *FileBasedWALWriter) shouldSync() bool {
	// if time has passed since last sync then return true
	return time.Since(w.lastSync) > w.config.WALSyncInterval
}

func(w *FileBasedWALWriter) flushFileToDisk() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}

	if err := w.file.Sync(); err != nil {
		return err
	}

	w.lastSync = time.Now()
	w.syncCount++
	return nil
}

func(w *FileBasedWALWriter) backgroundFlushRoutine() {
	w.syncTicker = time.NewTicker(w.config.WALSyncInterval)
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		defer w.syncTicker.Stop()

		for {
			select {
			case <-w.syncTicker.C:
				w.mu.Lock()
				w.flushFileToDisk()
				w.mu.Unlock()
			case <-w.stopCh:
				return
			}
		}
	}()
}

func(w *FileBasedWALWriter) serializeEntry(entry *WALEntry) ([]byte, error) {
	// convert entry into bytes, calculate its checksum then 
	// add that back into the entry and calculate checksum 
	entryData, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	checksum := crc32.ChecksumIEEE(entryData)
	entry.Checksum = checksum
	entryData, err = json.Marshal(entry)

	frame := make([]byte, 4+len(entryData))
	binary.BigEndian.PutUint32(frame[0:4], uint32(len(entryData)))
	copy(frame[4:], entryData)
	return frame, nil
}

func(w *FileBasedWALWriter) deserializeEntry(reader *bufio.Reader) (*WALEntry, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(reader, lengthBytes); err != nil {
		return nil, fmt.Errorf("error while reading length: %v", err)
	}

	length := binary.BigEndian.Uint32(lengthBytes)

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("error while reading data: %v", err)
	}

	var entry WALEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("error unmarshalling data: %v", err)
	}

	originalChecksum := entry.Checksum
	entry.Checksum = 0
	expectedData, err := json.Marshal(entry) 
	if err != nil {
		return nil, err
	}
	checksum := crc32.ChecksumIEEE(expectedData)
	if checksum != originalChecksum {
		return nil, fmt.Errorf("checksum does not match")
	}

	return &entry, nil
}
// index recovery

func(w *FileBasedWALWriter) recoverIndex() (int64, error) {
	file, err := os.Open(w.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return -1, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	var lastIndex int64 = 0

	for {
		entry, err := w.deserializeEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return -1, fmt.Errorf("corrupted WAL entry: %v", err)
		}
		lastIndex = entry.Index
	}
	return lastIndex, nil
}


// Start implementing the interface

func(w *FileBasedWALWriter) Append(ctx context.Context, entry *WALEntry) error {
	// add metadata
	// serialize the wal entry 
	// increment the index
	// if force sync then run the sync once
	// do all this while acquiring lock
	w.mu.Lock()
	defer w.mu.Unlock()

	w.currentIndex++
	entry.Index = w.currentIndex
	entry.Timestamp = time.Now()

	data, err := w.serializeEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to serialize wal entry: %v", err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write WAL entry: %v", err)
	}

	w.entriesWritten++
	w.bytesWritten += int64(len(data))

	if w.shouldSync() {
		if err := w.flushFileToDisk(); err != nil {
			return fmt.Errorf("failed to flush WAL to disk: %v", err)
		}
	}
	return nil
}

func(w *FileBasedWALWriter) Read(ctx context.Context, startIndex int64, maxEntries int32) ([]*WALEntry, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	file, err := os.Open(w.filePath)
	if err != nil {
		return nil, fmt.Errorf("error while reading WAL file: %v", err)
	}
	defer file.Close()

	var entryList []*WALEntry
	reader := bufio.NewReader(file)

	for len(entryList) < int(maxEntries) {
		entry, err := w.deserializeEntry(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read WAL entry: %v", err)
		}
		if entry.Index < startIndex {
			continue
		}
		entryList = append(entryList, entry)
	}
	return entryList, nil
}

func(w *FileBasedWALWriter) LatestIndex(ctx context.Context) (int64, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentIndex, nil
}

func(w *FileBasedWALWriter) Sync(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushFileToDisk()
}

func(w *FileBasedWALWriter) Close() error {
	close(w.stopCh)
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.flushFileToDisk(); err != nil {
		return err
	}
	return w.file.Close()
}