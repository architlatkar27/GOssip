package replication

import (
	"context"
	"time"

	"github.com/gossip-broker/gossip/pkg/types"
)

// ReplicationManager manages partition replication
type ReplicationManager interface {
	// Start replication for a partition
	StartReplication(ctx context.Context, topic string, partition int32) error

	// Stop replication for a partition
	StopReplication(ctx context.Context, topic string, partition int32) error

	// Get replication status for a partition
	GetReplicationStatus(ctx context.Context, topic string, partition int32) (*ReplicationStatus, error)

	// List all replicated partitions
	ListReplications(ctx context.Context) ([]*ReplicationInfo, error)

	// Add a replica to a partition
	AddReplica(ctx context.Context, topic string, partition int32, nodeID types.NodeID) error

	// Remove a replica from a partition
	RemoveReplica(ctx context.Context, topic string, partition int32, nodeID types.NodeID) error

	// Reassign partition replicas
	ReassignReplicas(ctx context.Context, assignments map[string]map[int32][]types.NodeID) error

	// Get ISR (In-Sync Replicas) for a partition
	GetISR(ctx context.Context, topic string, partition int32) ([]types.NodeID, error)

	// Update ISR for a partition
	UpdateISR(ctx context.Context, topic string, partition int32, isr []types.NodeID) error

	// Get replication lag for all replicas
	GetReplicationLag(ctx context.Context) (map[string]map[int32]map[types.NodeID]int64, error)

	// Close the replication manager
	Close() error
}

// ReplicationStatus represents the replication status of a partition
type ReplicationStatus struct {
	Topic             string                        `json:"topic"`
	Partition         int32                         `json:"partition"`
	Leader            types.NodeID                  `json:"leader"`
	Replicas          []types.NodeID                `json:"replicas"`
	ISR               []types.NodeID                `json:"isr"`
	LastISRUpdate     time.Time                     `json:"last_isr_update"`
	ReplicationLag    map[types.NodeID]int64        `json:"replication_lag"`
	ReplicaStates     map[types.NodeID]ReplicaState `json:"replica_states"`
	IsUnderReplicated bool                          `json:"is_under_replicated"`
	PreferredLeader   types.NodeID                  `json:"preferred_leader"`
}

// ReplicationInfo contains basic replication information
type ReplicationInfo struct {
	Topic     string         `json:"topic"`
	Partition int32          `json:"partition"`
	Leader    types.NodeID   `json:"leader"`
	Replicas  []types.NodeID `json:"replicas"`
	ISR       []types.NodeID `json:"isr"`
}

// ReplicaState represents the state of a replica
type ReplicaState string

const (
	ReplicaStateOnline     ReplicaState = "ONLINE"
	ReplicaStateOffline    ReplicaState = "OFFLINE"
	ReplicaStateCatchingUp ReplicaState = "CATCHING_UP"
	ReplicaStateInSync     ReplicaState = "IN_SYNC"
	ReplicaStateOutOfSync  ReplicaState = "OUT_OF_SYNC"
)

// PartitionLeader manages partition leadership
type PartitionLeader interface {
	// Start leading a partition
	StartLeading(ctx context.Context, topic string, partition int32) error

	// Stop leading a partition
	StopLeading(ctx context.Context, topic string, partition int32) error

	// Check if this node is the leader for a partition
	IsLeader(ctx context.Context, topic string, partition int32) bool

	// Get current leader for a partition
	GetLeader(ctx context.Context, topic string, partition int32) (types.NodeID, error)

	// Append messages to the partition log
	AppendMessages(ctx context.Context, topic string, partition int32, messages []*types.Message) error

	// Replicate messages to followers
	ReplicateToFollowers(ctx context.Context, topic string, partition int32, messages []*types.Message) error

	// Handle follower fetch requests
	HandleFollowerFetch(ctx context.Context, req *FollowerFetchRequest) (*FollowerFetchResponse, error)

	// Get high water mark for a partition
	GetHighWaterMark(ctx context.Context, topic string, partition int32) (int64, error)

	// Get log end offset for a partition
	GetLogEndOffset(ctx context.Context, topic string, partition int32) (int64, error)
}

// PartitionFollower manages partition following
type PartitionFollower interface {
	// Start following a partition
	StartFollowing(ctx context.Context, topic string, partition int32, leader types.NodeID) error

	// Stop following a partition
	StopFollowing(ctx context.Context, topic string, partition int32) error

	// Check if this node is following a partition
	IsFollowing(ctx context.Context, topic string, partition int32) bool

	// Fetch messages from the leader
	FetchFromLeader(ctx context.Context, topic string, partition int32, offset int64) ([]*types.Message, error)

	// Send fetch request to leader
	SendFetchRequest(ctx context.Context, leader types.NodeID, req *FollowerFetchRequest) (*FollowerFetchResponse, error)

	// Get current fetch offset
	GetFetchOffset(ctx context.Context, topic string, partition int32) (int64, error)

	// Update fetch offset
	UpdateFetchOffset(ctx context.Context, topic string, partition int32, offset int64) error
}

// FollowerFetchRequest represents a fetch request from a follower
type FollowerFetchRequest struct {
	FollowerID  types.NodeID      `json:"follower_id"`
	MaxWaitTime time.Duration     `json:"max_wait_time"`
	MinBytes    int32             `json:"min_bytes"`
	MaxBytes    int32             `json:"max_bytes"`
	Partitions  []*FetchPartition `json:"partitions"`
}

// FetchPartition represents a partition in a fetch request
type FetchPartition struct {
	Topic          string `json:"topic"`
	Partition      int32  `json:"partition"`
	FetchOffset    int64  `json:"fetch_offset"`
	LogStartOffset int64  `json:"log_start_offset"`
	MaxBytes       int32  `json:"max_bytes"`
}

// FollowerFetchResponse represents a fetch response to a follower
type FollowerFetchResponse struct {
	ThrottleTimeMs int32                     `json:"throttle_time_ms"`
	Partitions     []*FetchPartitionResponse `json:"partitions"`
}

// FetchPartitionResponse represents a partition response in a fetch response
type FetchPartitionResponse struct {
	Topic            string           `json:"topic"`
	Partition        int32            `json:"partition"`
	ErrorCode        types.ErrorCode  `json:"error_code"`
	HighWaterMark    int64            `json:"high_water_mark"`
	LastStableOffset int64            `json:"last_stable_offset"`
	LogStartOffset   int64            `json:"log_start_offset"`
	Messages         []*types.Message `json:"messages"`
}

// LeaderElector handles partition leader election
type LeaderElector interface {
	// Elect a leader for a partition
	ElectLeader(ctx context.Context, topic string, partition int32, candidates []types.NodeID) (types.NodeID, error)

	// Handle leader failure
	HandleLeaderFailure(ctx context.Context, topic string, partition int32, failedLeader types.NodeID) error

	// Get preferred leader for a partition
	GetPreferredLeader(ctx context.Context, topic string, partition int32) (types.NodeID, error)

	// Trigger preferred leader election
	TriggerPreferredLeaderElection(ctx context.Context, topic string, partition int32) error

	// Register leader change callback
	OnLeaderChange(callback func(topic string, partition int32, oldLeader, newLeader types.NodeID))
}

// ISRManager manages In-Sync Replica sets
type ISRManager interface {
	// Add replica to ISR
	AddToISR(ctx context.Context, topic string, partition int32, replica types.NodeID) error

	// Remove replica from ISR
	RemoveFromISR(ctx context.Context, topic string, partition int32, replica types.NodeID) error

	// Check if replica is in ISR
	IsInISR(ctx context.Context, topic string, partition int32, replica types.NodeID) (bool, error)

	// Get ISR for a partition
	GetISR(ctx context.Context, topic string, partition int32) ([]types.NodeID, error)

	// Update ISR
	UpdateISR(ctx context.Context, topic string, partition int32, isr []types.NodeID) error

	// Check ISR health
	CheckISRHealth(ctx context.Context) error

	// Get ISR shrink/expand events
	GetISREvents(ctx context.Context, since time.Time) ([]*ISREvent, error)
}

// ISREvent represents an ISR change event
type ISREvent struct {
	Topic     string         `json:"topic"`
	Partition int32          `json:"partition"`
	Type      ISREventType   `json:"type"`
	Replica   types.NodeID   `json:"replica"`
	OldISR    []types.NodeID `json:"old_isr"`
	NewISR    []types.NodeID `json:"new_isr"`
	Timestamp time.Time      `json:"timestamp"`
	Reason    string         `json:"reason"`
}

// ISREventType represents the type of ISR event
type ISREventType string

const (
	ISREventTypeExpand ISREventType = "EXPAND"
	ISREventTypeShrink ISREventType = "SHRINK"
)

// ReplicationMonitor monitors replication health
type ReplicationMonitor interface {
	// Start monitoring
	Start(ctx context.Context) error

	// Stop monitoring
	Stop(ctx context.Context) error

	// Get replication metrics
	GetMetrics(ctx context.Context) (*ReplicationMetrics, error)

	// Check replication health
	CheckHealth(ctx context.Context) (*ReplicationHealth, error)

	// Get under-replicated partitions
	GetUnderReplicatedPartitions(ctx context.Context) ([]*ReplicationInfo, error)

	// Get offline partitions
	GetOfflinePartitions(ctx context.Context) ([]*ReplicationInfo, error)

	// Register health change callback
	OnHealthChange(callback func(health *ReplicationHealth))
}

// ReplicationMetrics contains replication metrics
type ReplicationMetrics struct {
	TotalPartitions           int64                                       `json:"total_partitions"`
	UnderReplicatedPartitions int64                                       `json:"under_replicated_partitions"`
	OfflinePartitions         int64                                       `json:"offline_partitions"`
	PreferredLeaderImbalance  float64                                     `json:"preferred_leader_imbalance"`
	AverageReplicationLag     float64                                     `json:"average_replication_lag"`
	MaxReplicationLag         int64                                       `json:"max_replication_lag"`
	ISRShrinks                int64                                       `json:"isr_shrinks"`
	ISRExpands                int64                                       `json:"isr_expands"`
	LeaderElections           int64                                       `json:"leader_elections"`
	PartitionCounts           map[types.NodeID]int64                      `json:"partition_counts"`
	ReplicationLags           map[string]map[int32]map[types.NodeID]int64 `json:"replication_lags"`
}

// ReplicationHealth represents overall replication health
type ReplicationHealth struct {
	Status               HealthStatus  `json:"status"`
	UnderReplicatedCount int64         `json:"under_replicated_count"`
	OfflineCount         int64         `json:"offline_count"`
	Issues               []HealthIssue `json:"issues"`
	LastChecked          time.Time     `json:"last_checked"`
}

// HealthStatus represents health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "HEALTHY"
	HealthStatusDegraded  HealthStatus = "DEGRADED"
	HealthStatusUnhealthy HealthStatus = "UNHEALTHY"
)

// HealthIssue represents a replication health issue
type HealthIssue struct {
	Type        IssueType     `json:"type"`
	Severity    IssueSeverity `json:"severity"`
	Topic       string        `json:"topic"`
	Partition   int32         `json:"partition"`
	Description string        `json:"description"`
	Timestamp   time.Time     `json:"timestamp"`
}

// IssueType represents the type of health issue
type IssueType string

const (
	IssueTypeUnderReplicated IssueType = "UNDER_REPLICATED"
	IssueTypeOffline         IssueType = "OFFLINE"
	IssueTypeHighLag         IssueType = "HIGH_LAG"
	IssueTypeLeaderImbalance IssueType = "LEADER_IMBALANCE"
	IssueTypeISRFlapping     IssueType = "ISR_FLAPPING"
)

// IssueSeverity represents the severity of a health issue
type IssueSeverity string

const (
	IssueSeverityLow      IssueSeverity = "LOW"
	IssueSeverityMedium   IssueSeverity = "MEDIUM"
	IssueSeverityHigh     IssueSeverity = "HIGH"
	IssueSeverityCritical IssueSeverity = "CRITICAL"
)

// ReplicationConfig represents replication configuration
type ReplicationConfig struct {
	// Default replication factor for new topics
	DefaultReplicationFactor int32 `json:"default_replication_factor"`

	// Minimum in-sync replicas
	MinISR int32 `json:"min_isr"`

	// ISR configuration
	ReplicaLagTimeMaxMs    int64 `json:"replica_lag_time_max_ms"`
	ReplicaLagMaxMessages  int64 `json:"replica_lag_max_messages"`
	ReplicaSocketTimeoutMs int32 `json:"replica_socket_timeout_ms"`
	ReplicaFetchMaxBytes   int32 `json:"replica_fetch_max_bytes"`
	ReplicaFetchWaitMaxMs  int32 `json:"replica_fetch_wait_max_ms"`
	ReplicaFetchMinBytes   int32 `json:"replica_fetch_min_bytes"`

	// Leader election
	UncleanLeaderElectionEnable         bool  `json:"unclean_leader_election_enable"`
	LeaderImbalancePerBrokerPercentage  int32 `json:"leader_imbalance_per_broker_percentage"`
	LeaderImbalanceCheckIntervalSeconds int64 `json:"leader_imbalance_check_interval_seconds"`

	// Monitoring
	MonitoringIntervalMs       int64 `json:"monitoring_interval_ms"`
	HealthCheckIntervalMs      int64 `json:"health_check_interval_ms"`
	MetricsReportingIntervalMs int64 `json:"metrics_reporting_interval_ms"`

	// Performance tuning
	NumReplicaFetchers                       int32 `json:"num_replica_fetchers"`
	ReplicaFetchBackoffMs                    int32 `json:"replica_fetch_backoff_ms"`
	ReplicaHighWatermarkCheckpointIntervalMs int64 `json:"replica_high_watermark_checkpoint_interval_ms"`
}

// ReplicationFactory creates replication components
type ReplicationFactory interface {
	// Create a new replication manager
	NewReplicationManager(config *ReplicationConfig) (ReplicationManager, error)

	// Create a new partition leader
	NewPartitionLeader(config *ReplicationConfig) (PartitionLeader, error)

	// Create a new partition follower
	NewPartitionFollower(config *ReplicationConfig) (PartitionFollower, error)

	// Create a new leader elector
	NewLeaderElector(config *ReplicationConfig) (LeaderElector, error)

	// Create a new ISR manager
	NewISRManager(config *ReplicationConfig) (ISRManager, error)

	// Create a new replication monitor
	NewReplicationMonitor(config *ReplicationConfig) (ReplicationMonitor, error)
}
