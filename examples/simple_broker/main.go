package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gossip-broker/gossip/pkg/config"
	"github.com/gossip-broker/gossip/pkg/types"
	"github.com/gossip-broker/gossip/pkg/consensus"
	"github.com/gossip-broker/gossip/pkg/network"
	"github.com/gossip-broker/gossip/pkg/broker"
	"github.com/gossip-broker/gossip/pkg/storage"
)

// SimpleBroker demonstrates how the interfaces work together
// This is a mock implementation for demonstration purposes
type SimpleBroker struct {
	config   *config.Config
	storage  storage.LogStore
	metadata storage.MetadataStore
	raft     consensus.RaftEngine
	network  network.Server
	running  bool
}

// NewSimpleBroker creates a new simple broker instance
func NewSimpleBroker(cfg *config.Config) *SimpleBroker {
	return &SimpleBroker{
		config: cfg,
		// In a real implementation, these would be created by factories
		// storage:  storageFactory.NewLogStore(&cfg.Storage),
		// metadata: storageFactory.NewMetadataStore(&cfg.Storage),
		// raft:     consensusFactory.NewRaftEngine(&cfg.Raft, transport, store),
		// network:  networkFactory.NewTCPServer(&cfg.Network),
	}
}

// Start implements the Broker interface
func (b *SimpleBroker) Start(ctx context.Context) error {
	log.Printf("Starting GOssip broker on %s", b.config.Network.BindAddress)

	// 1. Start storage layer
	if b.storage != nil {
		log.Println("Starting storage layer...")
		// storage would be started here
	}

	// 2. Start metadata store
	if b.metadata != nil {
		log.Println("Starting metadata store...")
		// metadata store would be started here
	}

	// 3. Start Raft consensus
	if b.raft != nil {
		log.Println("Starting Raft consensus...")
		if err := b.raft.Start(ctx); err != nil {
			return fmt.Errorf("failed to start Raft: %w", err)
		}

		// Wait for leadership
		leader, err := b.raft.WaitForLeader(ctx, 30*time.Second)
		if err != nil {
			return fmt.Errorf("failed to establish leadership: %w", err)
		}
		log.Printf("Raft leader established: %s", leader)
	}

	// 4. Start network server
	if b.network != nil {
		log.Println("Starting network server...")
		if err := b.network.Start(ctx); err != nil {
			return fmt.Errorf("failed to start network server: %w", err)
		}
	}

	b.running = true
	log.Println("GOssip broker started successfully!")
	return nil
}

// Stop implements the Broker interface
func (b *SimpleBroker) Stop(ctx context.Context) error {
	log.Println("Stopping GOssip broker...")

	b.running = false

	// Stop in reverse order
	if b.network != nil {
		if err := b.network.Stop(ctx); err != nil {
			log.Printf("Error stopping network server: %v", err)
		}
	}

	if b.raft != nil {
		if err := b.raft.Stop(ctx); err != nil {
			log.Printf("Error stopping Raft: %v", err)
		}
	}

	if b.metadata != nil {
		if err := b.metadata.Close(); err != nil {
			log.Printf("Error closing metadata store: %v", err)
		}
	}

	if b.storage != nil {
		if err := b.storage.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}

	log.Println("GOssip broker stopped")
	return nil
}

// IsRunning implements the Broker interface
func (b *SimpleBroker) IsRunning() bool {
	return b.running
}

// Produce implements the Broker interface (mock implementation)
func (b *SimpleBroker) Produce(ctx context.Context, req *types.ProduceRequest) (*types.ProduceResponse, error) {
	log.Printf("Producing message to topic: %s, key: %s", req.Topic, string(req.Key))

	// Mock response
	return &types.ProduceResponse{
		Topic:     req.Topic,
		Partition: 0,
		Offset:    time.Now().UnixNano(), // Mock offset
		Timestamp: time.Now(),
	}, nil
}

// Consume implements the Broker interface (mock implementation)
func (b *SimpleBroker) Consume(ctx context.Context, req *types.ConsumeRequest) (*types.ConsumeResponse, error) {
	log.Printf("Consuming from topic: %s, partition: %d, offset: %d", req.Topic, req.Partition, req.Offset)

	// Mock response with sample messages
	messages := []*types.Message{
		{
			ID: types.MessageID{
				Topic:     req.Topic,
				Partition: req.Partition,
				Offset:    req.Offset,
			},
			Key:       []byte("sample-key"),
			Value:     []byte("sample-value"),
			Timestamp: time.Now(),
			Offset:    req.Offset,
			Partition: req.Partition,
			Topic:     req.Topic,
		},
	}

	return &types.ConsumeResponse{
		Messages: messages,
	}, nil
}

// CreateTopic implements the Broker interface (mock implementation)
func (b *SimpleBroker) CreateTopic(ctx context.Context, config *types.TopicConfig) error {
	log.Printf("Creating topic: %s with %d partitions", config.Name, config.Partitions)

	// In a real implementation, this would:
	// 1. Validate the topic configuration
	// 2. Propose the change through Raft
	// 3. Create partitions and assign replicas
	// 4. Update metadata store

	return nil
}

// Health implements the Broker interface
func (b *SimpleBroker) Health(ctx context.Context) error {
	if !b.running {
		return fmt.Errorf("broker is not running")
	}

	// Check component health
	if b.raft != nil && !b.raft.IsRunning() {
		return fmt.Errorf("raft consensus is not running")
	}

	return nil
}

// Implement other Broker interface methods with mock implementations
func (b *SimpleBroker) ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) ([]*types.ProduceResponse, error) {
	responses := make([]*types.ProduceResponse, len(reqs))
	for i, req := range reqs {
		resp, err := b.Produce(ctx, req)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}
	return responses, nil
}

func (b *SimpleBroker) Subscribe(ctx context.Context, topics []string, groupID string) (interfaces.ConsumerSubscription, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (b *SimpleBroker) DeleteTopic(ctx context.Context, topic string) error {
	log.Printf("Deleting topic: %s", topic)
	return nil
}

func (b *SimpleBroker) ListTopics(ctx context.Context) ([]string, error) {
	return []string{"sample-topic"}, nil
}

func (b *SimpleBroker) GetTopicConfig(ctx context.Context, topic string) (*types.TopicConfig, error) {
	return &types.TopicConfig{
		Name:              topic,
		Partitions:        3,
		ReplicationFactor: 1,
		RetentionTime:     7 * 24 * time.Hour,
	}, nil
}

func (b *SimpleBroker) UpdateTopicConfig(ctx context.Context, topic string, config *types.TopicConfig) error {
	return nil
}

func (b *SimpleBroker) GetPartitionInfo(ctx context.Context, topic string, partition int32) (*types.PartitionInfo, error) {
	return &types.PartitionInfo{
		Topic:     topic,
		Partition: partition,
		Leader:    types.NodeID(b.config.Node.ID),
		Replicas:  []types.NodeID{types.NodeID(b.config.Node.ID)},
		ISR:       []types.NodeID{types.NodeID(b.config.Node.ID)},
	}, nil
}

func (b *SimpleBroker) ListPartitions(ctx context.Context, topic string) ([]*types.PartitionInfo, error) {
	return []*types.PartitionInfo{}, nil
}

func (b *SimpleBroker) CreateConsumerGroup(ctx context.Context, groupID string) error {
	return nil
}

func (b *SimpleBroker) DeleteConsumerGroup(ctx context.Context, groupID string) error {
	return nil
}

func (b *SimpleBroker) ListConsumerGroups(ctx context.Context) ([]*types.ConsumerGroup, error) {
	return []*types.ConsumerGroup{}, nil
}

func (b *SimpleBroker) GetConsumerGroupInfo(ctx context.Context, groupID string) (*types.ConsumerGroup, error) {
	return nil, fmt.Errorf("not found")
}

func (b *SimpleBroker) CommitOffset(ctx context.Context, groupID string, offsets []*types.Offset) error {
	return nil
}

func (b *SimpleBroker) GetOffsets(ctx context.Context, groupID string, topics []string) ([]*types.Offset, error) {
	return []*types.Offset{}, nil
}

func (b *SimpleBroker) GetClusterInfo(ctx context.Context) (*types.ClusterState, error) {
	return &types.ClusterState{
		Nodes: map[types.NodeID]*types.NodeInfo{
			types.NodeID(b.config.Node.ID): {
				ID:      types.NodeID(b.config.Node.ID),
				Address: b.config.Node.AdvertiseAddress,
				IsAlive: true,
			},
		},
		Topics:     make(map[string]*types.TopicConfig),
		Partitions: []*types.PartitionInfo{},
		Controller: types.NodeID(b.config.Node.ID),
		Epoch:      1,
	}, nil
}

func (b *SimpleBroker) GetNodeInfo(ctx context.Context, nodeID types.NodeID) (*types.NodeInfo, error) {
	return &types.NodeInfo{
		ID:      nodeID,
		Address: b.config.Node.AdvertiseAddress,
		IsAlive: true,
	}, nil
}

func (b *SimpleBroker) Metrics(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"uptime":             time.Since(time.Now()),
		"topics":             1,
		"partitions":         3,
		"active_connections": 0,
	}, nil
}

// demonstrateUsage shows how to use the broker
func demonstrateUsage(broker interfaces.Broker) {
	ctx := context.Background()

	// Create a topic
	topicConfig := &types.TopicConfig{
		Name:              "demo-topic",
		Partitions:        3,
		ReplicationFactor: 1,
		RetentionTime:     24 * time.Hour,
	}

	if err := broker.CreateTopic(ctx, topicConfig); err != nil {
		log.Printf("Error creating topic: %v", err)
	}

	// Produce some messages
	for i := 0; i < 5; i++ {
		req := &types.ProduceRequest{
			Topic: "demo-topic",
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("message-%d", i)),
		}

		resp, err := broker.Produce(ctx, req)
		if err != nil {
			log.Printf("Error producing message: %v", err)
			continue
		}

		log.Printf("Produced message: offset=%d, partition=%d", resp.Offset, resp.Partition)
	}

	// Consume messages
	consumeReq := &types.ConsumeRequest{
		Topic:       "demo-topic",
		Partition:   0,
		Offset:      0,
		MaxMessages: 10,
	}

	consumeResp, err := broker.Consume(ctx, consumeReq)
	if err != nil {
		log.Printf("Error consuming messages: %v", err)
	} else {
		log.Printf("Consumed %d messages", len(consumeResp.Messages))
		for _, msg := range consumeResp.Messages {
			log.Printf("Message: key=%s, value=%s, offset=%d",
				string(msg.Key), string(msg.Value), msg.Offset)
		}
	}

	// Get cluster info
	clusterInfo, err := broker.GetClusterInfo(ctx)
	if err != nil {
		log.Printf("Error getting cluster info: %v", err)
	} else {
		log.Printf("Cluster: %d nodes, %d topics",
			len(clusterInfo.Nodes), len(clusterInfo.Topics))
	}
}

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Create broker
	broker := NewSimpleBroker(cfg)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start broker
	if err := broker.Start(ctx); err != nil {
		log.Fatalf("Failed to start broker: %v", err)
	}

	// Demonstrate usage
	go demonstrateUsage(broker)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal")

	// Stop broker gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := broker.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("GOssip broker shutdown complete")
}
