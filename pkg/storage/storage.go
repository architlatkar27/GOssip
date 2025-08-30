package storage

import (
	"context"
	"io"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
)

// LogStore manages the persistent log storage for messages
type LogStore interface {
	// Append messages to a topic partition
	Append(ctx context.Context, topic string, partition int32, messages []*types.Message) error

	// Read messages from a topic partition starting at offset
	Read(ctx context.Context, topic string, partition int32, offset int64, maxMessages int32) ([]*types.Message, error)

	// Get the latest offset for a topic partition
	LatestOffset(ctx context.Context, topic string, partition int32) (int64, error)

	// Get the earliest offset for a topic partition
	EarliestOffset(ctx context.Context, topic string, partition int32) (int64, error)

	// Create a new topic partition
	CreatePartition(ctx context.Context, topic string, partition int32) error

	// Delete a topic partition
	DeletePartition(ctx context.Context, topic string, partition int32) error

	// List all topic partitions
	ListPartitions(ctx context.Context) (map[string][]int32, error)

	// Truncate log to a specific offset
	Truncate(ctx context.Context, topic string, partition int32, offset int64) error

	// Compact the log (remove duplicate keys, keep latest)
	Compact(ctx context.Context, topic string, partition int32) error

	// Get storage size for a topic partition
	Size(ctx context.Context, topic string, partition int32) (int64, error)

	// Close the log store
	Close() error
}

// IndexStore manages offset indexes for fast lookups
type IndexStore interface {
	// Add an entry to the index
	AddEntry(topic string, partition int32, offset int64, position int64) error

	// Find the position for a given offset
	FindPosition(topic string, partition int32, offset int64) (int64, error)

	// Get the latest indexed offset
	LatestOffset(topic string, partition int32) (int64, error)

	// Truncate index to a specific offset
	Truncate(topic string, partition int32, offset int64) error

	// Close the index store
	Close() error
}

// MetadataStore manages cluster metadata
type MetadataStore interface {
	// Topic operations
	CreateTopic(ctx context.Context, config *types.TopicConfig) error
	GetTopic(ctx context.Context, name string) (*types.TopicConfig, error)
	UpdateTopic(ctx context.Context, config *types.TopicConfig) error
	DeleteTopic(ctx context.Context, name string) error
	ListTopics(ctx context.Context) ([]*types.TopicConfig, error)

	// Partition operations
	CreatePartition(ctx context.Context, info *types.PartitionInfo) error
	GetPartition(ctx context.Context, topic string, partition int32) (*types.PartitionInfo, error)
	UpdatePartition(ctx context.Context, info *types.PartitionInfo) error
	DeletePartition(ctx context.Context, topic string, partition int32) error
	ListPartitions(ctx context.Context, topic string) ([]*types.PartitionInfo, error)

	// Node operations
	RegisterNode(ctx context.Context, info *types.NodeInfo) error
	GetNode(ctx context.Context, id types.NodeID) (*types.NodeInfo, error)
	UpdateNode(ctx context.Context, info *types.NodeInfo) error
	UnregisterNode(ctx context.Context, id types.NodeID) error
	ListNodes(ctx context.Context) ([]*types.NodeInfo, error)

	// Consumer group operations
	CreateConsumerGroup(ctx context.Context, group *types.ConsumerGroup) error
	GetConsumerGroup(ctx context.Context, id string) (*types.ConsumerGroup, error)
	UpdateConsumerGroup(ctx context.Context, group *types.ConsumerGroup) error
	DeleteConsumerGroup(ctx context.Context, id string) error
	ListConsumerGroups(ctx context.Context) ([]*types.ConsumerGroup, error)

	// Offset operations
	CommitOffsets(ctx context.Context, groupID string, offsets []*types.Offset) error
	GetOffsets(ctx context.Context, groupID string, topics []string) ([]*types.Offset, error)

	// Cluster state
	GetClusterState(ctx context.Context) (*types.ClusterState, error)
	UpdateClusterState(ctx context.Context, state *types.ClusterState) error

	// Watch for changes
	Watch(ctx context.Context, prefix string) (<-chan MetadataEvent, error)

	// Close the metadata store
	Close() error
}

// MetadataEvent represents a metadata change event
type MetadataEvent struct {
	Type     MetadataEventType `json:"type"`
	Key      string            `json:"key"`
	Value    []byte            `json:"value,omitempty"`
	OldValue []byte            `json:"old_value,omitempty"`
}

// MetadataEventType represents the type of metadata event
type MetadataEventType string

const (
	MetadataEventTypeCreate MetadataEventType = "CREATE"
	MetadataEventTypeUpdate MetadataEventType = "UPDATE"
	MetadataEventTypeDelete MetadataEventType = "DELETE"
)

// OffsetStore manages consumer offsets
type OffsetStore interface {
	// Commit offsets for a consumer group
	CommitOffsets(ctx context.Context, groupID string, offsets []*types.Offset) error

	// Get offsets for a consumer group
	GetOffsets(ctx context.Context, groupID string, topics []string) ([]*types.Offset, error)

	// Get offset for a specific topic partition
	GetOffset(ctx context.Context, groupID, topic string, partition int32) (int64, error)

	// Delete offsets for a consumer group
	DeleteOffsets(ctx context.Context, groupID string) error

	// List all consumer groups
	ListConsumerGroups(ctx context.Context) ([]string, error)

	// Close the offset store
	Close() error
}

// WAL (Write-Ahead Log) ensures durability
type WAL interface {
	// Append an entry to the WAL
	Append(ctx context.Context, entry *WALEntry) error

	// Read entries from the WAL starting at index
	Read(ctx context.Context, startIndex int64, maxEntries int32) ([]*WALEntry, error)

	// Get the latest index
	LatestIndex(ctx context.Context) (int64, error)

	// Sync WAL to disk
	Sync(ctx context.Context) error

	// Close the WAL
	Close() error
}

// WALType represents the type of WAL entry
type WALType string

const (
	WALTypeMessage  WALType = "MESSAGE"
	WALTypeMetadata WALType = "METADATA"
	WALTypeSnapshot WALType = "SNAPSHOT"
)

// SnapshotStore manages cluster snapshots
type SnapshotStore interface {
	// Create a new snapshot
	CreateSnapshot(ctx context.Context, index int64, term int64) (SnapshotWriter, error)

	// Open an existing snapshot
	OpenSnapshot(ctx context.Context, index int64, term int64) (SnapshotReader, error)

	// List available snapshots
	ListSnapshots(ctx context.Context) ([]*SnapshotInfo, error)

	// Delete a snapshot
	DeleteSnapshot(ctx context.Context, index int64, term int64) error

	// Close the snapshot store
	Close() error
}

// SnapshotWriter writes snapshot data
type SnapshotWriter interface {
	io.WriteCloser

	// Finalize the snapshot
	Finalize() error

	// Cancel the snapshot
	Cancel() error
}

// SnapshotReader reads snapshot data
type SnapshotReader interface {
	io.ReadCloser

	// Get snapshot info
	Info() *SnapshotInfo
}

// SnapshotInfo contains metadata about a snapshot
type SnapshotInfo struct {
	Index     int64     `json:"index"`
	Term      int64     `json:"term"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Checksum  uint32    `json:"checksum"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	// Base directory for all storage
	DataDir string `json:"data_dir"`

	// Segment configuration
	SegmentSize   int64         `json:"segment_size"`    // Max size per segment
	SegmentMaxAge time.Duration `json:"segment_max_age"` // Max age per segment
	IndexInterval int32         `json:"index_interval"`  // Messages between index entries

	// Retention configuration
	RetentionTime   time.Duration `json:"retention_time"`   // Time-based retention
	RetentionBytes  int64         `json:"retention_bytes"`  // Size-based retention
	CleanupInterval time.Duration `json:"cleanup_interval"` // How often to run cleanup

	// Performance configuration
	FlushInterval time.Duration `json:"flush_interval"` // How often to flush to disk
	FlushMessages int32         `json:"flush_messages"` // Messages before flush
	FlushBytes    int64         `json:"flush_bytes"`    // Bytes before flush

	// Compression
	CompressionType  string `json:"compression_type"`  // none, gzip, snappy, lz4
	CompressionLevel int    `json:"compression_level"` // Compression level

	// WAL configuration
	WALDir          string        `json:"wal_dir"`           // WAL directory
	WALSegmentSize  int64         `json:"wal_segment_size"`  // WAL segment size
	WALSyncInterval time.Duration `json:"wal_sync_interval"` // WAL sync interval

	// Snapshot configuration
	SnapshotDir      string `json:"snapshot_dir"`      // Snapshot directory
	SnapshotInterval int64  `json:"snapshot_interval"` // Entries between snapshots
	SnapshotRetain   int32  `json:"snapshot_retain"`   // Number of snapshots to retain
}

// StorageFactory creates storage components
type StorageFactory interface {
	// Create a new log store
	NewLogStore(config *StorageConfig) (LogStore, error)

	// Create a new metadata store
	NewMetadataStore(config *StorageConfig) (MetadataStore, error)

	// Create a new offset store
	NewOffsetStore(config *StorageConfig) (OffsetStore, error)

	// Create a new WAL
	NewWAL(config *StorageConfig) (WAL, error)

	// Create a new snapshot store
	NewSnapshotStore(config *StorageConfig) (SnapshotStore, error)
}
