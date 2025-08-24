package partition

import (
	"context"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
	"github.com/gossip-broker/gossip/pkg/replication"
)

// PartitionManager manages topic partitions
type PartitionManager interface {
	// Create a new partition
	CreatePartition(ctx context.Context, topic string, partition int32, replicas []types.NodeID) error

	// Delete a partition
	DeletePartition(ctx context.Context, topic string, partition int32) error

	// Get partition information
	GetPartition(ctx context.Context, topic string, partition int32) (*types.PartitionInfo, error)

	// List all partitions for a topic
	ListPartitions(ctx context.Context, topic string) ([]*types.PartitionInfo, error)

	// List all partitions in the cluster
	ListAllPartitions(ctx context.Context) ([]*types.PartitionInfo, error)

	// Assign partitions to nodes
	AssignPartitions(ctx context.Context, assignments map[string]map[int32][]types.NodeID) error

	// Rebalance partitions across nodes
	RebalancePartitions(ctx context.Context, topic string) error

	// Get partition assignment for a topic
	GetPartitionAssignment(ctx context.Context, topic string) (map[int32][]types.NodeID, error)

	// Add partitions to an existing topic
	AddPartitions(ctx context.Context, topic string, count int32) error

	// Get partition leader
	GetPartitionLeader(ctx context.Context, topic string, partition int32) (types.NodeID, error)

	// Get partition replicas
	GetPartitionReplicas(ctx context.Context, topic string, partition int32) ([]types.NodeID, error)

	// Check if partition exists
	PartitionExists(ctx context.Context, topic string, partition int32) (bool, error)

	// Get partition statistics
	GetPartitionStats(ctx context.Context, topic string, partition int32) (*PartitionStats, error)

	// Close the partition manager
	Close() error
}

// PartitionStats contains partition statistics
type PartitionStats struct {
	Topic          string         `json:"topic"`
	Partition      int32          `json:"partition"`
	Leader         types.NodeID   `json:"leader"`
	Replicas       []types.NodeID `json:"replicas"`
	ISR            []types.NodeID `json:"isr"`
	LogSize        int64          `json:"log_size"`
	LogStartOffset int64          `json:"log_start_offset"`
	LogEndOffset   int64          `json:"log_end_offset"`
	MessageCount   int64          `json:"message_count"`
	CreatedAt      time.Time      `json:"created_at"`
	LastModified   time.Time      `json:"last_modified"`
}

// Partition represents a single topic partition
type Partition interface {
	// Get partition information
	Info() *types.PartitionInfo

	// Append messages to the partition
	Append(ctx context.Context, messages []*types.Message) error

	// Read messages from the partition
	Read(ctx context.Context, offset int64, maxMessages int32) ([]*types.Message, error)

	// Get the latest offset
	LatestOffset(ctx context.Context) (int64, error)

	// Get the earliest offset
	EarliestOffset(ctx context.Context) (int64, error)

	// Get partition size
	Size(ctx context.Context) (int64, error)

	// Truncate partition to offset
	Truncate(ctx context.Context, offset int64) error

	// Compact the partition (for log compaction)
	Compact(ctx context.Context) error

	// Check if this node is the leader
	IsLeader() bool

	// Check if this node is a replica
	IsReplica() bool

	// Get replication status
	ReplicationStatus() *replication.ReplicationStatus

	// Start partition (begin accepting reads/writes)
	Start(ctx context.Context) error

	// Stop partition (stop accepting reads/writes)
	Stop(ctx context.Context) error

	// Close partition resources
	Close() error
}

// PartitionAssigner assigns partitions to consumers
type PartitionAssigner interface {
	// Assign partitions to consumers in a group
	Assign(ctx context.Context, groupID string, members []*GroupMember, topics []string) (*AssignmentResult, error)

	// Get assignment strategy name
	Name() string

	// Check if assignment is balanced
	IsBalanced(assignment *AssignmentResult) bool

	// Rebalance assignments
	Rebalance(ctx context.Context, groupID string, currentAssignment *AssignmentResult, members []*GroupMember) (*AssignmentResult, error)
}

// GroupMember represents a member of a consumer group
type GroupMember struct {
	ID       string   `json:"id"`
	ClientID string   `json:"client_id"`
	Host     string   `json:"host"`
	Topics   []string `json:"topics"`
	UserData []byte   `json:"user_data,omitempty"`
}

// AssignmentResult contains the result of partition assignment
type AssignmentResult struct {
	GroupID     string                       `json:"group_id"`
	Assignments map[string]*MemberAssignment `json:"assignments"` // memberID -> assignment
	Metadata    map[string]interface{}       `json:"metadata"`
	Version     int32                        `json:"version"`
}

// MemberAssignment contains partition assignments for a group member
type MemberAssignment struct {
	MemberID   string             `json:"member_id"`
	Partitions map[string][]int32 `json:"partitions"` // topic -> partitions
	UserData   []byte             `json:"user_data,omitempty"`
}

// PartitionRouter routes messages to appropriate partitions
type PartitionRouter interface {
	// Route a message to a partition
	Route(ctx context.Context, topic string, key []byte, value []byte) (int32, error)

	// Get partition count for a topic
	GetPartitionCount(ctx context.Context, topic string) (int32, error)

	// Get routing strategy
	Strategy() RoutingStrategy
}

// RoutingStrategy represents partition routing strategies
type RoutingStrategy string

const (
	RoutingStrategyRoundRobin RoutingStrategy = "round_robin"
	RoutingStrategyHash       RoutingStrategy = "hash"
	RoutingStrategySticky     RoutingStrategy = "sticky"
	RoutingStrategyRandom     RoutingStrategy = "random"
	RoutingStrategyCustom     RoutingStrategy = "custom"
)

// PartitionBalancer balances partitions across brokers
type PartitionBalancer interface {
	// Generate a balanced partition assignment
	Balance(ctx context.Context, topics []string, brokers []types.NodeID) (map[string]map[int32][]types.NodeID, error)

	// Check if current assignment is balanced
	IsBalanced(ctx context.Context, assignment map[string]map[int32][]types.NodeID) (bool, error)

	// Get rebalance recommendations
	GetRebalanceRecommendations(ctx context.Context) ([]*RebalanceRecommendation, error)

	// Get balancing strategy
	Strategy() BalancingStrategy
}

// BalancingStrategy represents partition balancing strategies
type BalancingStrategy string

const (
	BalancingStrategyRoundRobin    BalancingStrategy = "round_robin"
	BalancingStrategyRackAware     BalancingStrategy = "rack_aware"
	BalancingStrategyLoadBased     BalancingStrategy = "load_based"
	BalancingStrategyCapacityBased BalancingStrategy = "capacity_based"
)

// RebalanceRecommendation represents a partition rebalance recommendation
type RebalanceRecommendation struct {
	Topic                 string         `json:"topic"`
	Partition             int32          `json:"partition"`
	CurrentReplicas       []types.NodeID `json:"current_replicas"`
	RecommendedReplicas   []types.NodeID `json:"recommended_replicas"`
	Reason                string         `json:"reason"`
	Priority              Priority       `json:"priority"`
	EstimatedDataMovement int64          `json:"estimated_data_movement"`
}

// Priority represents the priority of a recommendation
type Priority string

const (
	PriorityLow      Priority = "LOW"
	PriorityMedium   Priority = "MEDIUM"
	PriorityHigh     Priority = "HIGH"
	PriorityCritical Priority = "CRITICAL"
)

// PartitionMonitor monitors partition health and performance
type PartitionMonitor interface {
	// Start monitoring
	Start(ctx context.Context) error

	// Stop monitoring
	Stop(ctx context.Context) error

	// Get partition metrics
	GetMetrics(ctx context.Context, topic string, partition int32) (*PartitionMetrics, error)

	// Get all partition metrics
	GetAllMetrics(ctx context.Context) (map[string]map[int32]*PartitionMetrics, error)

	// Check partition health
	CheckHealth(ctx context.Context, topic string, partition int32) (*PartitionHealth, error)

	// Get unhealthy partitions
	GetUnhealthyPartitions(ctx context.Context) ([]*PartitionHealth, error)

	// Register health change callback
	OnHealthChange(callback func(topic string, partition int32, health *PartitionHealth))

	// Register metrics callback
	OnMetricsUpdate(callback func(topic string, partition int32, metrics *PartitionMetrics))
}

// PartitionMetrics contains partition performance metrics
type PartitionMetrics struct {
	Topic             string    `json:"topic"`
	Partition         int32     `json:"partition"`
	MessagesPerSecond float64   `json:"messages_per_second"`
	BytesPerSecond    float64   `json:"bytes_per_second"`
	ProduceRate       float64   `json:"produce_rate"`
	ConsumeRate       float64   `json:"consume_rate"`
	LogSize           int64     `json:"log_size"`
	LogGrowthRate     float64   `json:"log_growth_rate"`
	ReplicationLag    int64     `json:"replication_lag"`
	ConsumerLag       int64     `json:"consumer_lag"`
	ErrorRate         float64   `json:"error_rate"`
	LastUpdated       time.Time `json:"last_updated"`
}

// PartitionHealth represents partition health status
type PartitionHealth struct {
	Topic             string                `json:"topic"`
	Partition         int32                 `json:"partition"`
	Status            PartitionHealthStatus `json:"status"`
	Issues            []PartitionIssue      `json:"issues"`
	LastChecked       time.Time             `json:"last_checked"`
	Leader            types.NodeID          `json:"leader"`
	IsUnderReplicated bool                  `json:"is_under_replicated"`
	IsOffline         bool                  `json:"is_offline"`
	ReplicationLag    int64                 `json:"replication_lag"`
	ConsumerLag       int64                 `json:"consumer_lag"`
}

// PartitionHealthStatus represents partition health status
type PartitionHealthStatus string

const (
	PartitionHealthStatusHealthy   PartitionHealthStatus = "HEALTHY"
	PartitionHealthStatusDegraded  PartitionHealthStatus = "DEGRADED"
	PartitionHealthStatusUnhealthy PartitionHealthStatus = "UNHEALTHY"
	PartitionHealthStatusOffline   PartitionHealthStatus = "OFFLINE"
)

// PartitionIssue represents a partition health issue
type PartitionIssue struct {
	Type        PartitionIssueType     `json:"type"`
	Severity    replication.IssueSeverity          `json:"severity"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PartitionIssueType represents the type of partition issue
type PartitionIssueType string

const (
	PartitionIssueTypeUnderReplicated PartitionIssueType = "UNDER_REPLICATED"
	PartitionIssueTypeOffline         PartitionIssueType = "OFFLINE"
	PartitionIssueTypeHighLag         PartitionIssueType = "HIGH_LAG"
	PartitionIssueTypeNoLeader        PartitionIssueType = "NO_LEADER"
	PartitionIssueTypeISRShrink       PartitionIssueType = "ISR_SHRINK"
	PartitionIssueTypeHighErrorRate   PartitionIssueType = "HIGH_ERROR_RATE"
	PartitionIssueTypeDiskFull        PartitionIssueType = "DISK_FULL"
)

// PartitionConfig represents partition configuration
type PartitionConfig struct {
	// Default number of partitions for new topics
	DefaultPartitions int32 `json:"default_partitions"`

	// Maximum number of partitions per topic
	MaxPartitionsPerTopic int32 `json:"max_partitions_per_topic"`

	// Partition assignment strategy
	AssignmentStrategy string `json:"assignment_strategy"`

	// Rebalancing configuration
	RebalanceEnabled         bool    `json:"rebalance_enabled"`
	RebalanceIntervalMs      int64   `json:"rebalance_interval_ms"`
	RebalanceThresholdPct    float64 `json:"rebalance_threshold_pct"`
	RebalanceMaxDataMovement int64   `json:"rebalance_max_data_movement"`

	// Monitoring configuration
	MonitoringEnabled     bool  `json:"monitoring_enabled"`
	MonitoringIntervalMs  int64 `json:"monitoring_interval_ms"`
	HealthCheckIntervalMs int64 `json:"health_check_interval_ms"`
	MetricsRetentionHours int32 `json:"metrics_retention_hours"`

	// Performance thresholds
	HighLagThresholdMs     int64   `json:"high_lag_threshold_ms"`
	HighErrorRateThreshold float64 `json:"high_error_rate_threshold"`
	LowThroughputThreshold float64 `json:"low_throughput_threshold"`

	// Routing configuration
	DefaultRoutingStrategy   RoutingStrategy `json:"default_routing_strategy"`
	StickyPartitioning       bool            `json:"sticky_partitioning"`
	PartitionKeyHashFunction string          `json:"partition_key_hash_function"`
}

// PartitionFactory creates partition components
type PartitionFactory interface {
	// Create a new partition manager
	NewPartitionManager(config *PartitionConfig) (PartitionManager, error)

	// Create a new partition
	NewPartition(info *types.PartitionInfo) (Partition, error)

	// Create a new partition assigner
	NewPartitionAssigner(strategy string) (PartitionAssigner, error)

	// Create a new partition router
	NewPartitionRouter(strategy RoutingStrategy) (PartitionRouter, error)

	// Create a new partition balancer
	NewPartitionBalancer(strategy BalancingStrategy) (PartitionBalancer, error)

	// Create a new partition monitor
	NewPartitionMonitor(config *PartitionConfig) (PartitionMonitor, error)
}

// PartitionEvent represents partition lifecycle events
type PartitionEvent struct {
	Type      PartitionEventType     `json:"type"`
	Topic     string                 `json:"topic"`
	Partition int32                  `json:"partition"`
	NodeID    types.NodeID           `json:"node_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PartitionEventType represents the type of partition event
type PartitionEventType string

const (
	PartitionEventTypeCreated        PartitionEventType = "CREATED"
	PartitionEventTypeDeleted        PartitionEventType = "DELETED"
	PartitionEventTypeLeaderChanged  PartitionEventType = "LEADER_CHANGED"
	PartitionEventTypeISRChanged     PartitionEventType = "ISR_CHANGED"
	PartitionEventTypeReplicaAdded   PartitionEventType = "REPLICA_ADDED"
	PartitionEventTypeReplicaRemoved PartitionEventType = "REPLICA_REMOVED"
	PartitionEventTypeRebalanced     PartitionEventType = "REBALANCED"
)

// PartitionEventListener listens for partition events
type PartitionEventListener interface {
	// Handle partition events
	OnPartitionEvent(event *PartitionEvent)
}

// PartitionEventBus manages partition event distribution
type PartitionEventBus interface {
	// Publish a partition event
	Publish(ctx context.Context, event *PartitionEvent) error

	// Subscribe to partition events
	Subscribe(ctx context.Context, listener PartitionEventListener) error

	// Unsubscribe from partition events
	Unsubscribe(ctx context.Context, listener PartitionEventListener) error

	// Close the event bus
	Close() error
}
