package types

import (
	"context"
	"time"
)

// Message represents a single message in the broker
type Message struct {
	ID        MessageID `json:"id"`
	Key       []byte    `json:"key,omitempty"`
	Value     []byte    `json:"value"`
	Headers   Headers   `json:"headers,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Offset    int64     `json:"offset"`
	Partition int32     `json:"partition"`
	Topic     string    `json:"topic"`
}

// MessageID uniquely identifies a message
type MessageID struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}

// Headers represents message headers
type Headers map[string][]byte

// TopicConfig represents topic configuration
type TopicConfig struct {
	Name              string        `json:"name"`
	Partitions        int32         `json:"partitions"`
	ReplicationFactor int32         `json:"replication_factor"`
	RetentionTime     time.Duration `json:"retention_time"`
	RetentionBytes    int64         `json:"retention_bytes"`
	SegmentSize       int64         `json:"segment_size"`
	CompressionType   string        `json:"compression_type"`
	CleanupPolicy     string        `json:"cleanup_policy"` // "delete" or "compact"
}

// PartitionInfo represents partition metadata
type PartitionInfo struct {
	Topic     string   `json:"topic"`
	Partition int32    `json:"partition"`
	Leader    NodeID   `json:"leader"`
	Replicas  []NodeID `json:"replicas"`
	ISR       []NodeID `json:"isr"` // In-Sync Replicas
}

// NodeID represents a unique node identifier
type NodeID string

// NodeInfo represents node metadata
type NodeInfo struct {
	ID      NodeID `json:"id"`
	Address string `json:"address"`
	Rack    string `json:"rack,omitempty"`
	IsAlive bool   `json:"is_alive"`
}

// ClusterState represents the current state of the cluster
type ClusterState struct {
	Nodes      map[NodeID]*NodeInfo    `json:"nodes"`
	Topics     map[string]*TopicConfig `json:"topics"`
	Partitions []*PartitionInfo        `json:"partitions"`
	Controller NodeID                  `json:"controller"`
	Epoch      int64                   `json:"epoch"`
}

// ConsumerGroup represents a consumer group
type ConsumerGroup struct {
	ID       string             `json:"id"`
	Members  []ConsumerMember   `json:"members"`
	Offsets  map[string]int64   `json:"offsets"` // partition -> offset
	State    ConsumerGroupState `json:"state"`
	Protocol string             `json:"protocol"`
}

// ConsumerMember represents a member of a consumer group
type ConsumerMember struct {
	ID       string   `json:"id"`
	ClientID string   `json:"client_id"`
	Host     string   `json:"host"`
	Topics   []string `json:"topics"`
}

// ConsumerGroupState represents the state of a consumer group
type ConsumerGroupState string

const (
	ConsumerGroupStateEmpty       ConsumerGroupState = "Empty"
	ConsumerGroupStateStable      ConsumerGroupState = "Stable"
	ConsumerGroupStateRebalancing ConsumerGroupState = "Rebalancing"
)

// Offset represents a consumer offset
type Offset struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Metadata  string `json:"metadata,omitempty"`
}

// ProduceRequest represents a produce request
type ProduceRequest struct {
	Topic     string    `json:"topic"`
	Partition int32     `json:"partition,omitempty"` // -1 for auto-assignment
	Key       []byte    `json:"key,omitempty"`
	Value     []byte    `json:"value"`
	Headers   Headers   `json:"headers,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ProduceResponse represents a produce response
type ProduceResponse struct {
	Topic     string    `json:"topic"`
	Partition int32     `json:"partition"`
	Offset    int64     `json:"offset"`
	Timestamp time.Time `json:"timestamp"`
	Error     error     `json:"error,omitempty"`
}

// ConsumeRequest represents a consume request
type ConsumeRequest struct {
	Topic       string `json:"topic"`
	Partition   int32  `json:"partition"`
	Offset      int64  `json:"offset"`
	MaxMessages int32  `json:"max_messages"`
	MaxBytes    int32  `json:"max_bytes"`
	TimeoutMs   int32  `json:"timeout_ms"`
}

// ConsumeResponse represents a consume response
type ConsumeResponse struct {
	Messages []*Message `json:"messages"`
	Error    error      `json:"error,omitempty"`
}

// Error types
type ErrorCode int32

const (
	ErrorCodeNone ErrorCode = iota
	ErrorCodeUnknownTopic
	ErrorCodeUnknownPartition
	ErrorCodeInvalidMessage
	ErrorCodeNotLeader
	ErrorCodeTimeout
	ErrorCodeOffsetOutOfRange
	ErrorCodeBrokerNotAvailable
	ErrorCodeReplicaNotAvailable
	ErrorCodeMessageTooLarge
	ErrorCodeInvalidRequest
	ErrorCodeUnsupportedVersion
	ErrorCodeTopicAlreadyExists
)

// GossipError represents a broker error
type GossipError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *GossipError) Error() string {
	return e.Message
}

// Context keys for request metadata
type contextKey string

const (
	ContextKeyRequestID  contextKey = "request_id"
	ContextKeyClientID   contextKey = "client_id"
	ContextKeyUserAgent  contextKey = "user_agent"
	ContextKeyRemoteAddr contextKey = "remote_addr"
)

// RequestMetadata contains metadata about a request
type RequestMetadata struct {
	RequestID  string
	ClientID   string
	UserAgent  string
	RemoteAddr string
	StartTime  time.Time
}

// GetRequestMetadata extracts request metadata from context
func GetRequestMetadata(ctx context.Context) *RequestMetadata {
	return &RequestMetadata{
		RequestID:  getStringFromContext(ctx, ContextKeyRequestID),
		ClientID:   getStringFromContext(ctx, ContextKeyClientID),
		UserAgent:  getStringFromContext(ctx, ContextKeyUserAgent),
		RemoteAddr: getStringFromContext(ctx, ContextKeyRemoteAddr),
		StartTime:  time.Now(),
	}
}

func getStringFromContext(ctx context.Context, key contextKey) string {
	if val := ctx.Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
