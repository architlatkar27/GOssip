package consensus

import (
	"context"
	"time"

	storage "github.com/gossip-broker/gossip/pkg/storage"
	"github.com/gossip-broker/gossip/pkg/types"
)

// RaftEngine manages the Raft consensus protocol for metadata
type RaftEngine interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Leadership
	IsLeader() bool
	GetLeader() types.NodeID
	WaitForLeader(ctx context.Context, timeout time.Duration) (types.NodeID, error)

	// Membership management
	AddNode(ctx context.Context, nodeID types.NodeID, address string) error
	RemoveNode(ctx context.Context, nodeID types.NodeID) error
	GetClusterMembers() []ClusterMember

	// Log operations (only on leader)
	Propose(ctx context.Context, data []byte) error
	ProposeMetadataChange(ctx context.Context, change *MetadataChange) error

	// State machine
	RegisterStateMachine(sm StateMachine)
	GetStateMachine() StateMachine

	// Snapshots
	CreateSnapshot(ctx context.Context) error
	RestoreSnapshot(ctx context.Context, snapshot []byte) error

	// Metrics and status
	GetStats() RaftStats
	GetCurrentTerm() int64
	GetLastLogIndex() int64

	// Event subscription
	Subscribe(listener RaftEventListener)
	Unsubscribe(listener RaftEventListener)
}

// ClusterMember represents a member of the Raft cluster
type ClusterMember struct {
	ID      types.NodeID `json:"id"`
	Address string       `json:"address"`
	State   MemberState  `json:"state"`
	IsVoter bool         `json:"is_voter"`
}

// MemberState represents the state of a cluster member
type MemberState string

const (
	MemberStateFollower  MemberState = "Follower"
	MemberStateCandidate MemberState = "Candidate"
	MemberStateLeader    MemberState = "Leader"
	MemberStateNonVoter  MemberState = "NonVoter"
)

// StateMachine processes committed log entries
type StateMachine interface {
	// Apply a committed log entry
	Apply(ctx context.Context, entry *LogEntry) error

	// Create a snapshot of the current state
	Snapshot(ctx context.Context) ([]byte, error)

	// Restore state from a snapshot
	Restore(ctx context.Context, snapshot []byte) error

	// Get the last applied index
	LastAppliedIndex() int64
}

// LogEntry represents a Raft log entry
type LogEntry struct {
	Index     int64     `json:"index"`
	Term      int64     `json:"term"`
	Type      EntryType `json:"type"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// EntryType represents the type of log entry
type EntryType string

const (
	EntryTypeNoop          EntryType = "NOOP"
	EntryTypeMetadata      EntryType = "METADATA"
	EntryTypeConfiguration EntryType = "CONFIGURATION"
	EntryTypeSnapshot      EntryType = "SNAPSHOT"
)

// MetadataChange represents a change to cluster metadata
type MetadataChange struct {
	Type   MetadataChangeType `json:"type"`
	Topic  string             `json:"topic,omitempty"`
	Data   []byte             `json:"data"`
	NodeID types.NodeID       `json:"node_id,omitempty"`
}

// MetadataChangeType represents the type of metadata change
type MetadataChangeType string

const (
	MetadataChangeTypeCreateTopic     MetadataChangeType = "CREATE_TOPIC"
	MetadataChangeTypeDeleteTopic     MetadataChangeType = "DELETE_TOPIC"
	MetadataChangeTypeUpdateTopic     MetadataChangeType = "UPDATE_TOPIC"
	MetadataChangeTypeCreatePartition MetadataChangeType = "CREATE_PARTITION"
	MetadataChangeTypeUpdatePartition MetadataChangeType = "UPDATE_PARTITION"
	MetadataChangeTypeRegisterNode    MetadataChangeType = "REGISTER_NODE"
	MetadataChangeTypeUnregisterNode  MetadataChangeType = "UNREGISTER_NODE"
	MetadataChangeTypeUpdateNode      MetadataChangeType = "UPDATE_NODE"
)

// RaftEventListener receives Raft events
type RaftEventListener interface {
	// Called when leadership changes
	OnLeadershipChange(isLeader bool, leader types.NodeID)

	// Called when cluster membership changes
	OnMembershipChange(members []ClusterMember)

	// Called when a log entry is applied
	OnLogApplied(entry *LogEntry)

	// Called when a snapshot is created or restored
	OnSnapshot(index int64, term int64)
}

// RaftStats contains Raft statistics
type RaftStats struct {
	State              MemberState  `json:"state"`
	CurrentTerm        int64        `json:"current_term"`
	LastLogIndex       int64        `json:"last_log_index"`
	LastLogTerm        int64        `json:"last_log_term"`
	CommitIndex        int64        `json:"commit_index"`
	LastApplied        int64        `json:"last_applied"`
	Leader             types.NodeID `json:"leader"`
	VotedFor           types.NodeID `json:"voted_for"`
	NumNodes           int          `json:"num_nodes"`
	NumVoters          int          `json:"num_voters"`
	LastContact        time.Time    `json:"last_contact"`
	LastSnapshotIndex  int64        `json:"last_snapshot_index"`
	LastSnapshotTerm   int64        `json:"last_snapshot_term"`
	ProtocolVersion    int          `json:"protocol_version"`
	ProtocolVersionMin int          `json:"protocol_version_min"`
	ProtocolVersionMax int          `json:"protocol_version_max"`
	SnapshotVersionMin int          `json:"snapshot_version_min"`
	SnapshotVersionMax int          `json:"snapshot_version_max"`
}

// Transport handles Raft network communication
type Transport interface {
	// Consumer returns a channel for consuming RPC requests
	Consumer() <-chan RPC

	// LocalAddr returns the local address
	LocalAddr() string

	// AppendEntries sends an AppendEntries RPC to a target node
	AppendEntries(ctx context.Context, target string, req *AppendEntriesRequest) (*AppendEntriesResponse, error)

	// RequestVote sends a RequestVote RPC to a target node
	RequestVote(ctx context.Context, target string, req *RequestVoteRequest) (*RequestVoteResponse, error)

	// InstallSnapshot sends an InstallSnapshot RPC to a target node
	InstallSnapshot(ctx context.Context, target string, req *InstallSnapshotRequest) (*InstallSnapshotResponse, error)

	// Close the transport
	Close() error
}

// RPC represents a Raft RPC request/response
type RPC struct {
	Command  interface{}
	RespChan chan<- RPCResponse
}

// RPCResponse represents a Raft RPC response
type RPCResponse struct {
	Response interface{}
	Error    error
}

// AppendEntriesRequest represents an AppendEntries RPC request
type AppendEntriesRequest struct {
	Term         int64        `json:"term"`
	Leader       types.NodeID `json:"leader"`
	PrevLogIndex int64        `json:"prev_log_index"`
	PrevLogTerm  int64        `json:"prev_log_term"`
	Entries      []*LogEntry  `json:"entries"`
	LeaderCommit int64        `json:"leader_commit"`
}

// AppendEntriesResponse represents an AppendEntries RPC response
type AppendEntriesResponse struct {
	Term         int64 `json:"term"`
	Success      bool  `json:"success"`
	LastLogIndex int64 `json:"last_log_index"`
}

// RequestVoteRequest represents a RequestVote RPC request
type RequestVoteRequest struct {
	Term         int64        `json:"term"`
	Candidate    types.NodeID `json:"candidate"`
	LastLogIndex int64        `json:"last_log_index"`
	LastLogTerm  int64        `json:"last_log_term"`
}

// RequestVoteResponse represents a RequestVote RPC response
type RequestVoteResponse struct {
	Term        int64 `json:"term"`
	VoteGranted bool  `json:"vote_granted"`
}

// InstallSnapshotRequest represents an InstallSnapshot RPC request
type InstallSnapshotRequest struct {
	Term              int64        `json:"term"`
	Leader            types.NodeID `json:"leader"`
	LastIncludedIndex int64        `json:"last_included_index"`
	LastIncludedTerm  int64        `json:"last_included_term"`
	Offset            int64        `json:"offset"`
	Data              []byte       `json:"data"`
	Done              bool         `json:"done"`
}

// InstallSnapshotResponse represents an InstallSnapshot RPC response
type InstallSnapshotResponse struct {
	Term int64 `json:"term"`
}

// RaftConfig represents Raft configuration
type RaftConfig struct {
	// Node configuration
	NodeID      types.NodeID `json:"node_id"`
	BindAddress string       `json:"bind_address"`
	DataDir     string       `json:"data_dir"`

	// Timing configuration
	HeartbeatTimeout   time.Duration `json:"heartbeat_timeout"`
	ElectionTimeout    time.Duration `json:"election_timeout"`
	CommitTimeout      time.Duration `json:"commit_timeout"`
	LeaderLeaseTimeout time.Duration `json:"leader_lease_timeout"`

	// Log configuration
	MaxAppendEntries int    `json:"max_append_entries"`
	BatchApplyCh     bool   `json:"batch_apply_ch"`
	LogLevel         string `json:"log_level"`

	// Snapshot configuration
	SnapshotInterval  time.Duration `json:"snapshot_interval"`
	SnapshotThreshold int64         `json:"snapshot_threshold"`
	TrailingLogs      int64         `json:"trailing_logs"`
	SnapshotRetention int           `json:"snapshot_retention"`

	// Network configuration
	TransportMaxPool           int           `json:"transport_max_pool"`
	TransportTimeout           time.Duration `json:"transport_timeout"`
	DisableBootstrapAfterElect bool          `json:"disable_bootstrap_after_elect"`

	// Performance tuning
	NoSnapshotRestoreOnStart bool `json:"no_snapshot_restore_on_start"`
	SkipStartup              bool `json:"skip_startup"`
}

// ConsensusFactory creates consensus components
type ConsensusFactory interface {
	// Create a new Raft engine
	NewRaftEngine(config *RaftConfig, transport Transport, store storage.LogStore) (RaftEngine, error)

	// Create a new transport
	NewTransport(config *RaftConfig) (Transport, error)

	// Create a new state machine
	NewStateMachine() StateMachine
}

// LeaderElection handles leader election for non-Raft components
type LeaderElection interface {
	// Start the election process
	Start(ctx context.Context) error

	// Stop the election process
	Stop(ctx context.Context) error

	// Check if this node is the leader
	IsLeader() bool

	// Get the current leader
	GetLeader() types.NodeID

	// Register a leadership change callback
	OnLeadershipChange(callback func(isLeader bool, leader types.NodeID))
}

// CoordinationService provides coordination primitives
type CoordinationService interface {
	// Distributed locks
	Lock(ctx context.Context, key string, ttl time.Duration) (Lock, error)

	// Distributed barriers
	Barrier(ctx context.Context, key string, parties int) (Barrier, error)

	// Distributed semaphores
	Semaphore(ctx context.Context, key string, permits int) (Semaphore, error)

	// Leader election
	Election(ctx context.Context, key string, nodeID types.NodeID) (LeaderElection, error)
}

// Lock represents a distributed lock
type Lock interface {
	// Release the lock
	Release(ctx context.Context) error

	// Renew the lock TTL
	Renew(ctx context.Context, ttl time.Duration) error

	// Check if the lock is still held
	IsHeld(ctx context.Context) (bool, error)
}

// Barrier represents a distributed barrier
type Barrier interface {
	// Wait for all parties to reach the barrier
	Wait(ctx context.Context) error

	// Check how many parties have reached the barrier
	Count(ctx context.Context) (int, error)

	// Remove the barrier
	Remove(ctx context.Context) error
}

// Semaphore represents a distributed semaphore
type Semaphore interface {
	// Acquire a permit
	Acquire(ctx context.Context) error

	// Try to acquire a permit without blocking
	TryAcquire(ctx context.Context) (bool, error)

	// Release a permit
	Release(ctx context.Context) error

	// Get available permits
	AvailablePermits(ctx context.Context) (int, error)
}
