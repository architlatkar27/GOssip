# GOssip - A Modern Message Broker in Go

GOssip is a high-performance, cloud-native message broker built in Go, designed to address the limitations of Apache Kafka while maintaining compatibility with its core concepts.

## 🎯 Design Goals

- **Zero External Dependencies**: No ZooKeeper, no JVM - just a single Go binary
- **Developer Experience First**: Simple APIs, great tooling, clear documentation
- **Cloud Native**: Kubernetes-friendly, observability built-in
- **High Performance**: Sub-millisecond p99 latency, high throughput
- **Operational Simplicity**: Self-healing, auto-scaling, minimal configuration

## 🏗️ Architecture Overview

GOssip is built with a modular architecture using clean interfaces:

```
┌─────────────────────────────────────────────────────────────┐
│                    GOssip Broker                            │
├─────────────────────────────────────────────────────────────┤
│  API Layer (gRPC, HTTP, Binary Protocol)                    │
├─────────────────────────────────────────────────────────────┤
│  Core Components                                            │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │ Partition   │ Replication │ Consensus   │ Network     │  │
│  │ Manager     │ Manager     │ (Raft)      │ Layer       │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  Storage Layer                                              │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │ Log Store   │ Metadata    │ Index Store │ WAL         │  │
│  │             │ Store       │             │             │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Core Components

### 1. Broker Interface (`pkg/interfaces/broker.go`)
The main entry point providing high-level operations:
- Producer/Consumer operations
- Topic and partition management
- Consumer group coordination
- Cluster information and health checks

### 2. Storage Layer (`pkg/interfaces/storage.go`)
Handles persistent storage with multiple specialized stores:
- **LogStore**: Append-only message storage with segments
- **MetadataStore**: Cluster metadata and configuration
- **IndexStore**: Fast offset-to-position lookups
- **OffsetStore**: Consumer offset management
- **WAL**: Write-ahead logging for durability

### 3. Consensus Layer (`pkg/interfaces/consensus.go`)
Embedded Raft consensus for metadata management:
- Leader election and cluster coordination
- Metadata replication across nodes
- State machine for applying changes
- No external ZooKeeper dependency

### 4. Network Layer (`pkg/interfaces/network.go`)
Multi-protocol network support:
- TCP binary protocol (high performance)
- gRPC API (type-safe, streaming)
- HTTP/REST API (easy integration)
- Load balancing and circuit breakers

### 5. Replication Layer (`pkg/interfaces/replication.go`)
Manages partition replication:
- Leader/follower replication
- ISR (In-Sync Replicas) management
- Replication monitoring and health checks
- Automatic failover and recovery

### 6. Partition Layer (`pkg/interfaces/partition.go`)
Handles partition lifecycle and assignment:
- Partition creation and deletion
- Consumer group partition assignment
- Partition balancing across nodes
- Health monitoring and metrics

## 📋 Key Features

### Phase 1 (Core Foundation)
- [x] **Interface Design**: Clean, modular interfaces
- [x] **Configuration System**: Comprehensive config management
- [ ] **Basic Pub/Sub**: Topics, partitions, producers, consumers
- [ ] **Embedded Raft**: Consensus without external dependencies
- [ ] **Local Storage**: Log-structured storage engine
- [ ] **Binary Protocol**: High-performance network protocol

### Phase 2 (Advanced Features)
- [ ] **Consumer Groups**: Automatic partition assignment and rebalancing
- [ ] **Replication**: Multi-replica fault tolerance
- [ ] **Schema Management**: Built-in schema registry
- [ ] **REST API**: HTTP/JSON API for easy integration
- [ ] **CLI Tools**: Command-line interface for administration

### Phase 3 (Enhanced Capabilities)
- [ ] **Stream Processing**: Built-in stream processing engine
- [ ] **Query Engine**: SQL-like queries on message streams
- [ ] **Multi-tenancy**: Built-in tenant isolation
- [ ] **Advanced Monitoring**: Comprehensive observability

### Phase 4 (Production Ready)
- [ ] **Performance Optimization**: Sub-millisecond latency
- [ ] **Cloud Integration**: Kubernetes operators, cloud storage
- [ ] **Migration Tools**: Import/export from Kafka
- [ ] **Enterprise Features**: Advanced security, compliance

## 🚀 Quick Start

### 1. Install Dependencies
```bash
go mod download
```

### 2. Run the Example
```bash
go run examples/simple_broker/main.go
```

### 3. Basic Usage
```go
package main

import (
    "context"
    "github.com/gossip-broker/gossip/pkg/interfaces"
    "github.com/gossip-broker/gossip/pkg/types"
)

func main() {
    // Create broker (implementation would be provided by factory)
    var broker interfaces.Broker
    
    // Create a topic
    topicConfig := &types.TopicConfig{
        Name:              "my-topic",
        Partitions:        3,
        ReplicationFactor: 1,
    }
    broker.CreateTopic(context.Background(), topicConfig)
    
    // Produce a message
    req := &types.ProduceRequest{
        Topic: "my-topic",
        Key:   []byte("key1"),
        Value: []byte("Hello, GOssip!"),
    }
    resp, err := broker.Produce(context.Background(), req)
    
    // Consume messages
    consumeReq := &types.ConsumeRequest{
        Topic:     "my-topic",
        Partition: 0,
        Offset:    0,
    }
    messages, err := broker.Consume(context.Background(), consumeReq)
}
```

## 🔧 Configuration

GOssip uses a comprehensive configuration system with sensible defaults:

```yaml
# gossip.yaml
node:
  id: "gossip-node-1"
  data_dir: "/var/lib/gossip"
  bind_address: "0.0.0.0:9092"

cluster:
  name: "gossip-cluster"
  initial_nodes:
    - "node1:9093"
    - "node2:9093"
    - "node3:9093"

storage:
  segment_size: "1GB"
  retention_time: "7d"
  compression: "snappy"

network:
  protocol: "tcp"
  max_connections: 1000
  read_timeout: "30s"

raft:
  election_timeout: "1s"
  heartbeat_interval: "100ms"
```

## 🏗️ Development

### Project Structure
```
GOssip/
├── pkg/
│   ├── interfaces/     # Core interfaces
│   ├── types/         # Common types and data structures
│   ├── config/        # Configuration management
│   ├── storage/       # Storage implementations (TBD)
│   ├── consensus/     # Raft implementation (TBD)
│   ├── network/       # Network protocols (TBD)
│   └── replication/   # Replication logic (TBD)
├── examples/
│   └── simple_broker/ # Example broker implementation
├── cmd/
│   ├── gossip/        # Main broker binary (TBD)
│   └── gossip-cli/    # CLI tools (TBD)
└── docs/              # Documentation (TBD)
```

### Interface-First Development
GOssip follows an interface-first development approach:

1. **Interfaces Define Contracts**: All major components are defined as interfaces first
2. **Mock-Friendly**: Easy to test with mock implementations
3. **Pluggable Architecture**: Swap implementations without changing consumers
4. **Clean Dependencies**: No circular dependencies, clear separation of concerns

### Key Interfaces

| Interface | Purpose | Location |
|-----------|---------|----------|
| `Broker` | Main broker operations | `pkg/interfaces/broker.go` |
| `LogStore` | Message persistence | `pkg/interfaces/storage.go` |
| `RaftEngine` | Consensus protocol | `pkg/interfaces/consensus.go` |
| `Server` | Network communication | `pkg/interfaces/network.go` |
| `ReplicationManager` | Partition replication | `pkg/interfaces/replication.go` |
| `PartitionManager` | Partition lifecycle | `pkg/interfaces/partition.go` |

## 🎯 Kafka Limitations Addressed

### 1. **ZooKeeper Dependency**
- **Problem**: Complex external dependency, operational overhead
- **Solution**: Embedded Raft consensus, single binary deployment

### 2. **Operational Complexity** 
- **Problem**: 200+ configuration parameters, difficult tuning
- **Solution**: Smart defaults, auto-tuning, minimal configuration

### 3. **JVM Memory Management**
- **Problem**: GC pauses, heap tuning complexity
- **Solution**: Go's efficient GC, predictable memory usage

### 4. **Limited Query Capabilities**
- **Problem**: Only sequential access, no indexing
- **Solution**: Built-in indexing, SQL-like queries (Phase 3)

### 5. **Schema Evolution**
- **Problem**: Requires external schema registry
- **Solution**: Built-in schema management (Phase 2)

### 6. **Rebalancing Pain**
- **Problem**: Stop-the-world rebalancing
- **Solution**: Incremental rebalancing with minimal disruption

## 📊 Performance Targets

| Metric | Target | Kafka Baseline |
|--------|--------|----------------|
| P99 Latency | < 1ms | 3-10ms |
| Throughput | 10M+ msg/sec | 7M+ msg/sec |
| Memory Usage | < 512MB base | 1GB+ base |
| Startup Time | < 5 seconds | 30+ seconds |
| Partition Limit | 1M+ partitions | 200K partitions |

## 🤝 Contributing

GOssip is in early development. We welcome contributions!

1. **Phase 1 Implementation**: Help implement the core interfaces
2. **Testing**: Write comprehensive tests for interfaces
3. **Documentation**: Improve documentation and examples
4. **Performance**: Benchmark and optimize implementations

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Apache Kafka**: Inspiration for the messaging model
- **HashiCorp Raft**: Reference for consensus implementation
- **etcd**: Inspiration for embedded consensus
- **NATS**: Inspiration for simplicity and performance
