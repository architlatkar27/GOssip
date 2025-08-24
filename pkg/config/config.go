package config

import (
	"time"

	"github.com/gossip-broker/gossip/pkg/consensus"
	"github.com/gossip-broker/gossip/pkg/network"
	"github.com/gossip-broker/gossip/pkg/partition"
	"github.com/gossip-broker/gossip/pkg/replication"
	"github.com/gossip-broker/gossip/pkg/storage"
	"github.com/gossip-broker/gossip/pkg/types"
)

// Config represents the complete GOssip broker configuration
type Config struct {
	// Node configuration
	Node NodeConfig `yaml:"node" json:"node"`

	// Cluster configuration
	Cluster ClusterConfig `yaml:"cluster" json:"cluster"`

	// Storage configuration
	Storage storage.StorageConfig `yaml:"storage" json:"storage"`

	// Network configuration
	Network network.NetworkConfig `yaml:"network" json:"network"`

	// Raft configuration
	Raft consensus.RaftConfig `yaml:"raft" json:"raft"`

	// Replication configuration
	Replication replication.ReplicationConfig `yaml:"replication" json:"replication"`

	// Partition configuration
	Partition partition.PartitionConfig `yaml:"partition" json:"partition"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`

	// Security configuration
	Security SecurityConfig `yaml:"security" json:"security"`
}

// NodeConfig represents node-specific configuration
type NodeConfig struct {
	ID               types.NodeID `yaml:"id" json:"id"`
	DataDir          string       `yaml:"data_dir" json:"data_dir"`
	BindAddress      string       `yaml:"bind_address" json:"bind_address"`
	AdvertiseAddress string       `yaml:"advertise_address" json:"advertise_address"`
	Rack             string       `yaml:"rack" json:"rack,omitempty"`
	Zone             string       `yaml:"zone" json:"zone,omitempty"`
	Region           string       `yaml:"region" json:"region,omitempty"`
}

// ClusterConfig represents cluster-wide configuration
type ClusterConfig struct {
	Name                   string        `yaml:"name" json:"name"`
	InitialNodes           []string      `yaml:"initial_nodes" json:"initial_nodes"`
	BootstrapExpectedNodes int           `yaml:"bootstrap_expected_nodes" json:"bootstrap_expected_nodes"`
	JoinTimeout            time.Duration `yaml:"join_timeout" json:"join_timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`             // debug, info, warn, error
	Format     string `yaml:"format" json:"format"`           // json, text
	Output     string `yaml:"output" json:"output"`           // stdout, stderr, file
	File       string `yaml:"file" json:"file,omitempty"`     // log file path
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // max size in MB
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // max backup files
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // max age in days
	Compress   bool   `yaml:"compress" json:"compress"`       // compress old files
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Port            int           `yaml:"port" json:"port"`
	Path            string        `yaml:"path" json:"path"`
	Namespace       string        `yaml:"namespace" json:"namespace"`
	Subsystem       string        `yaml:"subsystem" json:"subsystem"`
	ReportInterval  time.Duration `yaml:"report_interval" json:"report_interval"`
	RetentionPeriod time.Duration `yaml:"retention_period" json:"retention_period"`

	// Prometheus configuration
	Prometheus PrometheusConfig `yaml:"prometheus" json:"prometheus"`

	// Custom metrics
	CustomMetrics []CustomMetric `yaml:"custom_metrics" json:"custom_metrics"`
}

// PrometheusConfig represents Prometheus-specific configuration
type PrometheusConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Registry string            `yaml:"registry" json:"registry"`
	Gatherer string            `yaml:"gatherer" json:"gatherer"`
	Labels   map[string]string `yaml:"labels" json:"labels"`
}

// CustomMetric represents a custom metric definition
type CustomMetric struct {
	Name        string    `yaml:"name" json:"name"`
	Type        string    `yaml:"type" json:"type"` // counter, gauge, histogram
	Description string    `yaml:"description" json:"description"`
	Labels      []string  `yaml:"labels" json:"labels"`
	Buckets     []float64 `yaml:"buckets" json:"buckets,omitempty"` // for histograms
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	// Authentication
	Authentication AuthenticationConfig `yaml:"authentication" json:"authentication"`

	// Authorization
	Authorization AuthorizationConfig `yaml:"authorization" json:"authorization"`

	// TLS configuration
	TLS TLSConfig `yaml:"tls" json:"tls"`

	// SASL configuration
	SASL SASLConfig `yaml:"sasl" json:"sasl"`
}

// AuthenticationConfig represents authentication configuration
type AuthenticationConfig struct {
	Enabled   bool           `yaml:"enabled" json:"enabled"`
	Method    string         `yaml:"method" json:"method"` // sasl, tls, jwt
	Providers []AuthProvider `yaml:"providers" json:"providers"`
}

// AuthProvider represents an authentication provider
type AuthProvider struct {
	Name     string            `yaml:"name" json:"name"`
	Type     string            `yaml:"type" json:"type"`
	Config   map[string]string `yaml:"config" json:"config"`
	Priority int               `yaml:"priority" json:"priority"`
}

// AuthorizationConfig represents authorization configuration
type AuthorizationConfig struct {
	Enabled       bool      `yaml:"enabled" json:"enabled"`
	DefaultPolicy string    `yaml:"default_policy" json:"default_policy"` // allow, deny
	ACLs          []ACLRule `yaml:"acls" json:"acls"`
	Roles         []Role    `yaml:"roles" json:"roles"`
}

// ACLRule represents an access control rule
type ACLRule struct {
	Principal  string   `yaml:"principal" json:"principal"`
	Operation  string   `yaml:"operation" json:"operation"`
	Resource   string   `yaml:"resource" json:"resource"`
	Permission string   `yaml:"permission" json:"permission"` // allow, deny
	Conditions []string `yaml:"conditions" json:"conditions,omitempty"`
}

// Role represents a security role
type Role struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Permissions []string `yaml:"permissions" json:"permissions"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled            bool     `yaml:"enabled" json:"enabled"`
	CertFile           string   `yaml:"cert_file" json:"cert_file"`
	KeyFile            string   `yaml:"key_file" json:"key_file"`
	CAFile             string   `yaml:"ca_file" json:"ca_file"`
	ClientAuth         string   `yaml:"client_auth" json:"client_auth"` // none, request, require, verify
	CipherSuites       []string `yaml:"cipher_suites" json:"cipher_suites"`
	MinVersion         string   `yaml:"min_version" json:"min_version"`
	MaxVersion         string   `yaml:"max_version" json:"max_version"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// SASLConfig represents SASL configuration
type SASLConfig struct {
	Enabled    bool       `yaml:"enabled" json:"enabled"`
	Mechanisms []string   `yaml:"mechanisms" json:"mechanisms"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	Realm      string     `yaml:"realm" json:"realm"`
	Users      []SASLUser `yaml:"users" json:"users"`
}

// SASLUser represents a SASL user
type SASLUser struct {
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	Roles    []string `yaml:"roles" json:"roles"`
	Enabled  bool     `yaml:"enabled" json:"enabled"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			ID:               "gossip-node-1",
			DataDir:          "/var/lib/gossip",
			BindAddress:      "0.0.0.0:9092",
			AdvertiseAddress: "localhost:9092",
		},
		Cluster: ClusterConfig{
			Name:                   "gossip-cluster",
			InitialNodes:           []string{"localhost:9093"},
			BootstrapExpectedNodes: 1,
			JoinTimeout:            30 * time.Second,
		},
		Storage: storage.StorageConfig{
			DataDir:          "/var/lib/gossip",
			SegmentSize:      1024 * 1024 * 1024, // 1GB
			SegmentMaxAge:    24 * time.Hour,
			IndexInterval:    4096,
			RetentionTime:    7 * 24 * time.Hour, // 7 days
			RetentionBytes:   -1,                 // unlimited
			CleanupInterval:  5 * time.Minute,
			FlushInterval:    10 * time.Second,
			FlushMessages:    10000,
			FlushBytes:       1024 * 1024, // 1MB
			CompressionType:  "snappy",
			CompressionLevel: 6,
			WALDir:           "/var/lib/gossip/wal",
			WALSegmentSize:   64 * 1024 * 1024, // 64MB
			WALSyncInterval:  1 * time.Second,
			SnapshotDir:      "/var/lib/gossip/snapshots",
			SnapshotInterval: 10000,
			SnapshotRetain:   3,
		},
		Network: network.NetworkConfig{
			BindAddress:             "0.0.0.0",
			Port:                    9092,
			MaxConnections:          1000,
			ConnectionTimeout:       30 * time.Second,
			Protocol:                "tcp",
			EnableTLS:               false,
			ReadBufferSize:          32 * 1024, // 32KB
			WriteBufferSize:         32 * 1024, // 32KB
			ReadTimeout:             30 * time.Second,
			WriteTimeout:            30 * time.Second,
			IdleTimeout:             5 * time.Minute,
			LoadBalancer:            "round_robin",
			HealthCheckInterval:     10 * time.Second,
			RateLimit:               1000,
			RateLimitWindow:         1 * time.Second,
			CircuitBreakerEnabled:   true,
			CircuitBreakerThreshold: 50,
			CircuitBreakerTimeout:   60 * time.Second,
		},
		Raft: consensus.RaftConfig{
			NodeID:                     types.NodeID("gossip-node-1"),
			BindAddress:                "0.0.0.0:9093",
			DataDir:                    "/var/lib/gossip/raft",
			HeartbeatTimeout:           1 * time.Second,
			ElectionTimeout:            1 * time.Second,
			CommitTimeout:              50 * time.Millisecond,
			LeaderLeaseTimeout:         500 * time.Millisecond,
			MaxAppendEntries:           64,
			BatchApplyCh:               true,
			LogLevel:                   "INFO",
			SnapshotInterval:           120 * time.Second,
			SnapshotThreshold:          8192,
			TrailingLogs:               10240,
			SnapshotRetention:          2,
			TransportMaxPool:           3,
			TransportTimeout:           10 * time.Second,
			DisableBootstrapAfterElect: false,
			NoSnapshotRestoreOnStart:   false,
			SkipStartup:                false,
		},
		Replication: replication.ReplicationConfig{
			DefaultReplicationFactor:                 3,
			MinISR:                                   2,
			ReplicaLagTimeMaxMs:                      30000,
			ReplicaLagMaxMessages:                    4000,
			ReplicaSocketTimeoutMs:                   30000,
			ReplicaFetchMaxBytes:                     1024 * 1024, // 1MB
			ReplicaFetchWaitMaxMs:                    500,
			ReplicaFetchMinBytes:                     1,
			UncleanLeaderElectionEnable:              false,
			LeaderImbalancePerBrokerPercentage:       10,
			LeaderImbalanceCheckIntervalSeconds:      300,
			MonitoringIntervalMs:                     10000,
			HealthCheckIntervalMs:                    30000,
			MetricsReportingIntervalMs:               60000,
			NumReplicaFetchers:                       1,
			ReplicaFetchBackoffMs:                    1000,
			ReplicaHighWatermarkCheckpointIntervalMs: 5000,
		},
		Partition: partition.PartitionConfig{
			DefaultPartitions:        3,
			MaxPartitionsPerTopic:    1000,
			AssignmentStrategy:       "round_robin",
			RebalanceEnabled:         true,
			RebalanceIntervalMs:      300000, // 5 minutes
			RebalanceThresholdPct:    10.0,
			RebalanceMaxDataMovement: 1024 * 1024 * 1024, // 1GB
			MonitoringEnabled:        true,
			MonitoringIntervalMs:     10000,
			HealthCheckIntervalMs:    30000,
			MetricsRetentionHours:    24,
			HighLagThresholdMs:       30000,
			HighErrorRateThreshold:   0.05,
			LowThroughputThreshold:   10.0,
			DefaultRoutingStrategy:   "hash",
			StickyPartitioning:       true,
			PartitionKeyHashFunction: "murmur3",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100, // 100MB
			MaxBackups: 3,
			MaxAge:     28, // 28 days
			Compress:   true,
		},
		Metrics: MetricsConfig{
			Enabled:         true,
			Port:            8080,
			Path:            "/metrics",
			Namespace:       "gossip",
			Subsystem:       "broker",
			ReportInterval:  10 * time.Second,
			RetentionPeriod: 24 * time.Hour,
			Prometheus: PrometheusConfig{
				Enabled:  true,
				Registry: "default",
				Gatherer: "default",
				Labels: map[string]string{
					"service": "gossip",
					"version": "1.0.0",
				},
			},
		},
		Security: SecurityConfig{
			Authentication: AuthenticationConfig{
				Enabled: false,
				Method:  "sasl",
			},
			Authorization: AuthorizationConfig{
				Enabled:       false,
				DefaultPolicy: "allow",
			},
			TLS: TLSConfig{
				Enabled:            false,
				ClientAuth:         "none",
				MinVersion:         "1.2",
				MaxVersion:         "1.3",
				InsecureSkipVerify: false,
			},
			SASL: SASLConfig{
				Enabled:    false,
				Mechanisms: []string{"PLAIN", "SCRAM-SHA-256"},
				Realm:      "gossip",
			},
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here
	// Check required fields, valid ranges, etc.
	return nil
}

// LoadFromFile loads configuration from a file
func LoadFromFile(filename string) (*Config, error) {
	// Implementation would load from YAML/JSON file
	// For now, return default config
	return DefaultConfig(), nil
}

// SaveToFile saves configuration to a file
func (c *Config) SaveToFile(filename string) error {
	// Implementation would save to YAML/JSON file
	return nil
}
