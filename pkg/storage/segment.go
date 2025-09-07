package storage

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"maps"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
)

// Segment represents a single log segment
type Segment interface {
	// Append messages to the segment
	Append(messages []*types.Message) error

	// Read messages from the segment
	Read(offset uint32, maxMessages int32) ([]*types.Message, error)

	// Get the base offset of the segment
	BaseOffset() int64

	// Get the latest offset in the segment
	LatestOffset() int64

	// Get the size of the segment
	Size() int64

	// Check if the segment is full
	IsFull() bool

	// Sync the segment to disk
	Sync() error

	// Close the segment
	Close() error
}

// SegmentStore manages individual log segments
type SegmentStore interface {
	// Create a new segment
	CreateSegment(topic string, partition int32, baseOffset int64) (Segment, error)

	// Open an existing segment
	OpenSegment(topic string, partition int32, baseOffset int64) (Segment, error)

	// List all segments for a topic partition
	ListSegments(topic string, partition int32) ([]SegmentInfo, error)

	// Delete a segment
	DeleteSegment(topic string, partition int32, baseOffset int64) error

	// Close all segments
	Close() error
}

// SegmentInfo contains metadata about a segment
type SegmentInfo struct {
	Topic        string    `json:"topic"`
	Partition    int32     `json:"partition"`
	BaseOffset   int64     `json:"base_offset"`
	LatestOffset int64     `json:"latest_offset"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

// implements segment interface
type FileSegment struct {
	// file details
	// filepath /data/topic/partition/log_base_offset.log(index_base_offset.log)
	logFile         *os.File
	logFilePath     string
	logFileWriter   *bufio.Writer
	indexFile       *os.File
	indexFilePath   string
	indexFileWriter *bufio.Writer

	// size metadata
	baseOffset     uint32
	nextOffset     uint32
	maxSegmentSize uint32
	currentSize    uint32

	// sparse indexing
	indexInterval int32
	messageCount  int32           // index everytime messageCount % indexInterval == 0
	indexMapping  map[uint32]uint32 // offset --> position in file

	// topic metadata
	topic     string
	partition int32

	// sync
	mu     sync.RWMutex
	closed bool

	createdAt     time.Time
	lastUpdatedAt time.Time
}

func (fs *FileSegment) Append(messages []*types.Message) error {
	// check for the size constraints
	// assign offset
	// generate checksum
	// serialize the message
	// write it to file
	// increment message count,
	// index it if needed and update the in memory index
	fs.mu.Lock()
	defer fs.mu.Unlock()

	shouldFlush := false
	if fs.currentSize > fs.maxSegmentSize {
		return fmt.Errorf("this segment is full, close this segment")
	}
	messageCollection := []byte{}
	indexCollection := []byte{}
	for _, msg := range messages {
		positionBeforeWrite := fs.currentSize
		msg.Offset = int64(fs.nextOffset)
		fs.nextOffset++
		messageData, err := fs.serializeMessage(msg)
		if err != nil {
			return fmt.Errorf("error serializing message: %v", err)
		}
		messageCollection = append(messageCollection, messageData...)
		fs.currentSize += uint32(len(messageData))
		fs.messageCount++
		if fs.messageCount%fs.indexInterval == 0 {
			// index the message
			fs.indexMapping[fs.nextOffset] = positionBeforeWrite
			indexEntry := fs.serializeIndex(fs.nextOffset, positionBeforeWrite)
			indexCollection = append(indexCollection, indexEntry...)
			shouldFlush = true
		}
	}
	if _, err := fs.logFileWriter.Write(messageCollection); err != nil {
		return fmt.Errorf("error while writing messages to buffer: %v", err)	
	}
	if _, err := fs.indexFileWriter.Write(indexCollection); err != nil {
		return fmt.Errorf("error while writing index to buffer: %v", err)	
	}

	if shouldFlush {
		if err := fs.logFileWriter.Flush(); err != nil {
			return fmt.Errorf("error while flushing message buffer to disk: %v", err)
		}
		if err := fs.logFileWriter.Flush(); err != nil {
			return fmt.Errorf("error while flushing message buffer to disk: %v", err)
		}
	}

	return nil
}

func (fs *FileSegment) serializeIndex(offset uint32, position uint32) []byte {
	indexData := make([]byte, 8)
	binary.BigEndian.PutUint32(indexData[0:4], offset)
	binary.BigEndian.PutUint32(indexData[4:], position)
	return indexData
}

func (fs *FileSegment) serializeMessage(message *types.Message) ([]byte, error) {
	messageData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("error while json-encoding data: %v", err)
	}

	checksum := crc32.ChecksumIEEE(messageData)
	length := uint32(4 + len(messageData))
	frame := make([]byte, 4+4+len(messageData)) // 4 for length, 4 for checksum and rest for message
	binary.BigEndian.PutUint32(frame[0:4], length)
	binary.BigEndian.PutUint32(frame[4:8], checksum)
	copy(frame[8:], messageData)
	return frame, nil
}

func (fs *FileSegment) deserializeMessage(data []byte) (*types.Message, error) {
	// read the first 4 bytes, convert to checksum
	// read the rest of the bytes, 
	// compare the checksum
	// return the result

	var message types.Message
	var checksum uint32

	checksum = binary.BigEndian.Uint32(data[0:4])

	if err := json.Unmarshal(data[4:], &message); err != nil {
		return nil, fmt.Errorf("error while unmarshaling data: %v", err)
	}

	expectedChecksum := crc32.ChecksumIEEE(data[4:])	
	if checksum != expectedChecksum {
		// what do we wanna do about the checksum not matching? 
		return nil, fmt.Errorf("checksum does not match")
	}

	return &message, nil
}

func (fs *FileSegment) Read(offset uint32, maxMessages int32) ([]*types.Message, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// if the offset is less than baseoffset then return error
	if offset < fs.baseOffset || offset > fs.nextOffset {
		return nil, fmt.Errorf("offset not present in this segment")
	}
	// search for the message with offset
	offsetList := slices.Collect(maps.Keys(fs.indexMapping))
	offsetInIndex := fs.findOffsetInIndex(offset, offsetList)

	basePosition, ok := fs.indexMapping[offsetInIndex]
	if !ok {
		return nil, fmt.Errorf("unable to find base position to search in file")
	}


	// what if we have more to read than this buffer? 
	// we will have to track what position we are on and again do a lot of 
	// custom read?
	// there should be a better way to handle this. 

	// the better way - first use readAt to read the number of bytes 
	// then make a buffer of that size to read the message and the checksum
	// put this in loop
	var messages []*types.Message
	currentPostion := basePosition
	for i := range(maxMessages) {
		sizeBuffer := make([]byte, 4)
		_, err := fs.logFile.ReadAt(sizeBuffer, int64(currentPostion))
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading size of message: %v", err)
		}
		currentPostion += 4
		// no need to handle EOF ig

		msgSize := binary.BigEndian.Uint32(sizeBuffer)
		buffer := make([]byte, msgSize)
		n, err := fs.logFile.ReadAt(buffer, int64(currentPostion))
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading message from segment: %v", err)
		}
		currentPostion += uint32(n)
		// desrialize the messgae 
		message, err := fs.deserializeMessage(buffer)
		if err != nil {
			return messages, fmt.Errorf("error while desriazlizing message %d: %v", i, err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (fs *FileSegment) findOffsetInIndex(key uint32, offsetList []uint32) uint32 {
	// find the last offset less than or equal to key
	var offsetValue uint32
	beg := 0
	end := len(offsetList) - 1

	for beg <= end {
		mid := (beg + end) / 2
		if offsetList[mid] > key {
			end = mid - 1
		} else {
			offsetValue = offsetList[mid]
			beg = mid + 1
		}
	}
	return offsetValue
}
