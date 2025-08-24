package storage

import (
	"bufio"
	"fmt"
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
    if err := instance.recoverIndex(); err != nil {
        file.Close()
        return nil, fmt.Errorf("failed to recover WAL index: %w", err)
    }
    
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
// WALEntry represents a single WAL entry
type WALEntry struct {
	Index     int64     `json:"index"`
	Term      int64     `json:"term"`
	Type      WALType   `json:"type"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Checksum  uint32    `json:"checksum"`
}