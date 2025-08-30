package storage

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"sync"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
)

// Segment represents a single log segment
type Segment interface {
	// Append messages to the segment
	Append(messages []*types.Message) error

	// Read messages from the segment
	Read(offset int64, maxMessages int32) ([]*types.Message, error)

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
	logFile *os.File
	logFilePath string
	logFileWriter *bufio.Writer
	indexFile *os.File
	indexFilePath string
	indexFileWriter *bufio.Writer

	// size metadata
	baseOffset uint32
	currentOffset uint32
	maxSegmentSize uint32
	currentSize uint32

	// sparse indexing
	indexInterval int32
	messageCount int32 // index everytime messageCount % indexInterval == 0
	indexMapping map[int64]int64 // offset --> position in file

	// topic metadata
	topic string
	partition int32

	// sync
	mu sync.RWMutex
	closed bool

	createdAt time.Time
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
	if fs.currentSize > fs.maxSegmentSize {
		return fmt.Errorf("this segment is full, close this segment")
	}
	messageCollection := []byte{}
	indexCollection := []byte{}
	for _, msg := range messages {
		msg.Offset = int64(fs.currentOffset)
		messageData, err := fs.serializeMessage(msg)
		if err != nil {
			return fmt.Errorf("error serializing message: %v", err)
		}
		messageCollection = append(messageCollection, messageData...)
		fs.currentSize += uint32(len(messageData))
		fs.messageCount++
		fs.currentOffset++
		if fs.messageCount % fs.indexInterval == 0 {
			// index the message
			indexEntry := fs.serializeIndex(fs.currentOffset, fs.currentSize)
			indexCollection = append(indexCollection, indexEntry...)
		}

	}
	
	fs.logFileWriter.Write(messageCollection)
	fs.indexFileWriter.Write(indexCollection)
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