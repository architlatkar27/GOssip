package interfaces

import (
	"context"
	"io"

	"github.com/gossip-broker/gossip/pkg/types"
)

// Broker represents the main message broker interface
type Broker interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Producer operations
	Produce(ctx context.Context, req *types.ProduceRequest) (*types.ProduceResponse, error)
	ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) ([]*types.ProduceResponse, error)

	// Consumer operations
	Consume(ctx context.Context, req *types.ConsumeRequest) (*types.ConsumeResponse, error)
	Subscribe(ctx context.Context, topics []string, groupID string) (ConsumerSubscription, error)

	// Topic management
	CreateTopic(ctx context.Context, config *types.TopicConfig) error
	DeleteTopic(ctx context.Context, topic string) error
	ListTopics(ctx context.Context) ([]string, error)
	GetTopicConfig(ctx context.Context, topic string) (*types.TopicConfig, error)
	UpdateTopicConfig(ctx context.Context, topic string, config *types.TopicConfig) error

	// Partition management
	GetPartitionInfo(ctx context.Context, topic string, partition int32) (*types.PartitionInfo, error)
	ListPartitions(ctx context.Context, topic string) ([]*types.PartitionInfo, error)

	// Consumer group management
	CreateConsumerGroup(ctx context.Context, groupID string) error
	DeleteConsumerGroup(ctx context.Context, groupID string) error
	ListConsumerGroups(ctx context.Context) ([]*types.ConsumerGroup, error)
	GetConsumerGroupInfo(ctx context.Context, groupID string) (*types.ConsumerGroup, error)

	// Offset management
	CommitOffset(ctx context.Context, groupID string, offsets []*types.Offset) error
	GetOffsets(ctx context.Context, groupID string, topics []string) ([]*types.Offset, error)

	// Cluster information
	GetClusterInfo(ctx context.Context) (*types.ClusterState, error)
	GetNodeInfo(ctx context.Context, nodeID types.NodeID) (*types.NodeInfo, error)

	// Health and metrics
	Health(ctx context.Context) error
	Metrics(ctx context.Context) (map[string]interface{}, error)
}

// ConsumerSubscription represents an active consumer subscription
type ConsumerSubscription interface {
	// Receive messages from subscribed topics
	Receive(ctx context.Context) (*types.Message, error)

	// Commit the offset for a message
	Commit(ctx context.Context, msg *types.Message) error

	// Get current assignment
	Assignment() map[string][]int32 // topic -> partitions

	// Close the subscription
	Close() error
}

// Producer provides high-level producer operations
type Producer interface {
	// Send a single message
	Send(ctx context.Context, topic string, key, value []byte) (*types.ProduceResponse, error)

	// Send a message with headers
	SendWithHeaders(ctx context.Context, topic string, key, value []byte, headers types.Headers) (*types.ProduceResponse, error)

	// Send to a specific partition
	SendToPartition(ctx context.Context, topic string, partition int32, key, value []byte) (*types.ProduceResponse, error)

	// Send multiple messages in batch
	SendBatch(ctx context.Context, messages []*types.ProduceRequest) ([]*types.ProduceResponse, error)

	// Flush pending messages
	Flush(ctx context.Context) error

	// Close producer
	Close() error
}

// Consumer provides high-level consumer operations
type Consumer interface {
	// Subscribe to topics
	Subscribe(ctx context.Context, topics []string) error

	// Unsubscribe from topics
	Unsubscribe(ctx context.Context, topics []string) error

	// Poll for messages
	Poll(ctx context.Context, timeoutMs int32) ([]*types.Message, error)

	// Commit offsets
	CommitSync(ctx context.Context) error
	CommitAsync(ctx context.Context) error

	// Seek to specific offset
	Seek(ctx context.Context, topic string, partition int32, offset int64) error

	// Get current assignment
	Assignment() map[string][]int32

	// Close consumer
	Close() error
}

// AdminClient provides administrative operations
type AdminClient interface {
	// Topic operations
	CreateTopic(ctx context.Context, config *types.TopicConfig) error
	DeleteTopic(ctx context.Context, topic string) error
	ListTopics(ctx context.Context) ([]string, error)
	DescribeTopic(ctx context.Context, topic string) (*types.TopicConfig, error)

	// Partition operations
	AddPartitions(ctx context.Context, topic string, count int32) error
	ReassignPartitions(ctx context.Context, reassignments map[string]map[int32][]types.NodeID) error

	// Consumer group operations
	ListConsumerGroups(ctx context.Context) ([]*types.ConsumerGroup, error)
	DescribeConsumerGroup(ctx context.Context, groupID string) (*types.ConsumerGroup, error)
	DeleteConsumerGroup(ctx context.Context, groupID string) error

	// Cluster operations
	ListNodes(ctx context.Context) ([]*types.NodeInfo, error)
	DescribeCluster(ctx context.Context) (*types.ClusterState, error)

	// Configuration
	DescribeConfigs(ctx context.Context, resources []string) (map[string]map[string]string, error)
	AlterConfigs(ctx context.Context, configs map[string]map[string]string) error

	// Close client
	Close() error
}

// ClientConfig represents client configuration
type ClientConfig struct {
	BootstrapServers []string          `json:"bootstrap_servers"`
	ClientID         string            `json:"client_id"`
	RequestTimeout   int32             `json:"request_timeout_ms"`
	RetryBackoff     int32             `json:"retry_backoff_ms"`
	MaxRetries       int32             `json:"max_retries"`
	Compression      string            `json:"compression"`
	SecurityProtocol string            `json:"security_protocol"`
	SASLMechanism    string            `json:"sasl_mechanism"`
	SASLUsername     string            `json:"sasl_username"`
	SASLPassword     string            `json:"sasl_password"`
	Properties       map[string]string `json:"properties"`
}

// ProducerConfig represents producer-specific configuration
type ProducerConfig struct {
	ClientConfig
	BatchSize         int32  `json:"batch_size"`
	LingerMs          int32  `json:"linger_ms"`
	BufferMemory      int64  `json:"buffer_memory"`
	MaxBlockMs        int32  `json:"max_block_ms"`
	Acks              string `json:"acks"` // "0", "1", "all"
	Retries           int32  `json:"retries"`
	EnableIdempotence bool   `json:"enable_idempotence"`
	MaxInFlightReqs   int32  `json:"max_in_flight_requests"`
}

// ConsumerConfig represents consumer-specific configuration
type ConsumerConfig struct {
	ClientConfig
	GroupID              string `json:"group_id"`
	AutoOffsetReset      string `json:"auto_offset_reset"` // "earliest", "latest", "none"
	EnableAutoCommit     bool   `json:"enable_auto_commit"`
	AutoCommitIntervalMs int32  `json:"auto_commit_interval_ms"`
	MaxPollRecords       int32  `json:"max_poll_records"`
	MaxPollIntervalMs    int32  `json:"max_poll_interval_ms"`
	SessionTimeoutMs     int32  `json:"session_timeout_ms"`
	HeartbeatIntervalMs  int32  `json:"heartbeat_interval_ms"`
	FetchMinBytes        int32  `json:"fetch_min_bytes"`
	FetchMaxWaitMs       int32  `json:"fetch_max_wait_ms"`
}

// ClientFactory creates broker clients
type ClientFactory interface {
	// Create a new producer
	NewProducer(config *ProducerConfig) (Producer, error)

	// Create a new consumer
	NewConsumer(config *ConsumerConfig) (Consumer, error)

	// Create an admin client
	NewAdminClient(config *ClientConfig) (AdminClient, error)

	// Create a broker client
	NewBroker(config *ClientConfig) (Broker, error)
}

// StreamProcessor represents a stream processing interface
type StreamProcessor interface {
	// Process messages from input topics and produce to output topics
	Process(ctx context.Context, input <-chan *types.Message, output chan<- *types.Message) error

	// Get processor topology
	Topology() ProcessorTopology

	// Start processing
	Start(ctx context.Context) error

	// Stop processing
	Stop(ctx context.Context) error
}

// ProcessorTopology represents the processing topology
type ProcessorTopology interface {
	// Add a source (input topic)
	Source(name string, topics ...string) ProcessorNode

	// Add a processor node
	Process(name string, processor ProcessorFunc) ProcessorNode

	// Add a sink (output topic)
	Sink(name string, topic string) ProcessorNode

	// Build the topology
	Build() error
}

// ProcessorNode represents a node in the processing topology
type ProcessorNode interface {
	// Connect to another node
	To(node ProcessorNode) ProcessorNode

	// Add a child processor
	Process(name string, processor ProcessorFunc) ProcessorNode

	// Add a sink
	Sink(name string, topic string) ProcessorNode
}

// ProcessorFunc represents a message processing function
type ProcessorFunc func(ctx context.Context, msg *types.Message) ([]*types.Message, error)

// Serializer handles message serialization
type Serializer interface {
	Serialize(data interface{}) ([]byte, error)
}

// Deserializer handles message deserialization
type Deserializer interface {
	Deserialize(data []byte, target interface{}) error
}

// Codec combines serialization and deserialization
type Codec interface {
	Serializer
	Deserializer
}

// Compressor handles message compression
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	Type() string
}

// MetricsReporter handles metrics reporting
type MetricsReporter interface {
	// Record a counter metric
	Counter(name string, value int64, tags map[string]string)

	// Record a gauge metric
	Gauge(name string, value float64, tags map[string]string)

	// Record a histogram metric
	Histogram(name string, value float64, tags map[string]string)

	// Record a timer metric
	Timer(name string, duration int64, tags map[string]string)
}

// Logger provides logging interface
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})

	With(fields ...interface{}) Logger
}

// Closer represents resources that need to be closed
type Closer interface {
	Close() error
}

// MultiCloser closes multiple resources
type MultiCloser []Closer

func (mc MultiCloser) Close() error {
	var firstErr error
	for _, closer := range mc {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ReadWriteCloser combines io interfaces
type ReadWriteCloser interface {
	io.Reader
	io.Writer
	io.Closer
}
