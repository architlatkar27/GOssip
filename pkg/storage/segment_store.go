package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SegmentStore interface {
	// Create a new segment
	CreateSegment(topic string, partition uint32, baseOffset uint32) (Segment, error)

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
	Partition    uint32    `json:"partition"`
	BaseOffset   uint32    `json:"base_offset"`
	LatestOffset uint32    `json:"latest_offset"`
	Size         uint32    `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type FileSegmentStore struct {
	// this should track different segments, how do we wanna track them though?
	// topic --> partition --> segments?
	// what is the use of segment info exactly?
	// at what level is  a segment store even working?
	ActiveSegments map[string]map[uint32]map[uint32]Segment // topic -> partition -> offset -> segment
	config         *StorageConfig
	basePath       string
	mu             sync.Mutex
}

func (store *FileSegmentStore) CreateSegment(topic string, partition uint32, baseOffset uint32) (Segment, error) {
	segment := FileSegment{}
	segment.baseOffset = baseOffset
	segment.nextOffset = baseOffset
	segment.maxSegmentSize = uint32(store.config.SegmentSize)
	segment.messageCount = 0
	segment.currentSize = 0
	segment.indexInterval = store.config.IndexInterval
	segment.indexMapping = make(map[uint32]uint32)
	segment.mu = sync.RWMutex{}
	segment.topic = topic
	segment.partition = partition
	segment.closed = false

	dir := filepath.Dir(segment.logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	segment.logFilePath = filepath.Join(store.config.DataDir,
		topic,
		fmt.Sprintf("%d", partition),
		fmt.Sprintf("%010d.log", baseOffset)) // <base>/topic/partition/offset.idx or offset.log

	segment.indexFilePath = filepath.Join(store.config.DataDir,
		topic,
		fmt.Sprintf("%d", partition),
		fmt.Sprintf("%010d.idx", baseOffset))

	var err error
	segment.logFile, err = os.OpenFile(segment.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %v", err)
	}

	segment.indexFile, err = os.OpenFile(segment.indexFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening index file: %v", err)
	}

	segment.logFileWriter = bufio.NewWriter(segment.logFile)
	segment.indexFileWriter = bufio.NewWriter(segment.indexFile)

	segment.createdAt = time.Now()

	store.mu.Lock()
	if store.ActiveSegments[topic] == nil {
		store.ActiveSegments[topic] = make(map[uint32]map[uint32]Segment)
	}
	if store.ActiveSegments[topic][partition] == nil {
		store.ActiveSegments[topic][partition] = make(map[uint32]Segment)
	}
	store.ActiveSegments[topic][partition][baseOffset] = &segment
	store.mu.Unlock()

	return &segment, nil
}
