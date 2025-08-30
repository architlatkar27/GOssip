package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/gossip-broker/gossip/pkg/types"
)

type FileLogStore struct {
	segmentStore SegmentStore
	config       *StorageConfig
    
    // State management
    mu               sync.RWMutex
    activeSegments   map[string]map[int32]Segment  // topic -> partition -> active segment
    partitionOffsets map[string]map[int32]int64    // topic -> partition -> next offset
    
    // Directory structure
    dataDir string
    
    // Lifecycle
    closed bool
}

func NewFileLogStore(config *StorageConfig, segmentStore SegmentStore) *FileLogStore {
	return &FileLogStore{
		segmentStore: segmentStore,
		config: config,
		activeSegments:   make(map[string]map[int32]Segment),
        partitionOffsets: make(map[string]map[int32]int64),
        dataDir:         config.DataDir,
        closed:          false,
	}
}

func(fls *FileLogStore) Append(ctx context.Context, topic string, partition int32, messages []*types.Message) error {
	fls.mu.Lock()
	defer fls.mu.Unlock()
	if fls.closed{
		return fmt.Errorf("log store is closed")
	}

	return nil
}

func(fls *FileLogStore) getOrCreateActiveSegment(topic string, partition int32) (Segment, error) {
	if topicSegments, exists := fls.activeSegments[topic]; exists {
		if segment, exists := topicSegments[partition]; exists {
			return segment, nil
		}
	}

	
}