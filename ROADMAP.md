# OtterMQ Development Roadmap

## Overview

OtterMQ aims to be a fully AMQP 0.9.1 compliant message broker with RabbitMQ compatibility. This roadmap tracks our progress toward complete protocol implementation and production readiness.

## AMQP 0.9.1 Implementation Status

### ✅ **Implemented Features**

- [x] **Connection Management**
  - [x] Protocol handshake and negotiation
  - [x] SASL PLAIN authentication
  - [x] Connection heartbeat handling
  - [x] Graceful connection close
- [x] **Channel Operations**
  - [x] Channel open/close lifecycle
  - [x] Multi-channel support per connection
- [x] **Exchange Management**
  - [x] Exchange declare/delete
  - [x] Direct, fanout, topic exchange types
  - [x] Mandatory exchanges (default, amq.*)
- [x] **Queue Management**
  - [x] Queue declare/delete with properties
  - [x] Queue binding to exchanges
  - [x] Message buffering and storage
- [x] **Basic Publishing**
  - [x] `BASIC_PUBLISH` with routing
  - [x] Multi-frame message assembly (method + header + body)
  - [x] Message properties and headers
- [x] **Basic Getting**
  - [x] `BASIC_GET` for pull-based consumption
  - [x] Message count reporting

### � **Recently Completed**

- [x] Consumer management system refactoring
- [x] **`BASIC_CONSUME`** - Start consuming messages from queue
  - [x] Consumer tag generation and management
  - [x] Consumer registration per channel
  - [x] Integration with queue message delivery
- [x] **`BASIC_DELIVER`** - Server-initiated message delivery
  - [x] Push-based message delivery to consumers
  - [x] Delivery tag generation and tracking
- [x] **`BASIC_CANCEL`** - Cancel consumer subscription
  - [x] Clean consumer shutdown
  - [x] Resource cleanup on cancellation
- [x] **`BASIC_CONSUME_OK`** / **`BASIC_CANCEL_OK`** - Response frames

### ❌ **Missing Critical Features**


#### **Phase 2: Message Acknowledgments (High Priority)**

- [ ] **`BASIC_ACK`** - Acknowledge message delivery
  - [ ] Single and multiple message acknowledgment
  - [ ] Integration with delivery tracking
- [ ] **`BASIC_REJECT`** - Reject single message
  - [ ] Requeue option support
  - [ ] Dead letter handling (future)
- [ ] **`BASIC_RECOVER`** - Redeliver unacknowledged messages
  - [ ] Async and sync variants
  - [ ] Channel-wide message recovery

#### **Phase 3: Quality of Service (Medium Priority)**
- [ ] **`BASIC_QOS`** - Control message prefetching
  - [ ] Per-channel prefetch limits
  - [ ] Per-consumer prefetch limits
  - [ ] Global QoS settings
- [ ] **Flow control integration**
  - [ ] Backpressure handling
  - [ ] Channel flow control (`CHANNEL_FLOW`)

#### **Phase 4: Transaction Support (Medium Priority)**

- [ ] **`TX_SELECT`** - Enter transaction mode
- [ ] **`TX_COMMIT`** - Commit transaction
- [ ] **`TX_ROLLBACK`** - Rollback transaction
- [ ] **Transactional publishing/consuming**

#### **Phase 5: Advanced Features (Low Priority)**

- [ ] **`QUEUE_UNBIND`** - Remove queue bindings
- [ ] **`QUEUE_PURGE`** - Clear queue contents
- [ ] **Message TTL and expiration**
- [ ] **Dead letter exchanges**
- [ ] **Priority queues**
- [ ] **Cluster support**

## Architecture Improvements

### **Persistence Layer**

- [ ] **Swappable Persistence Architecture** - Move to pluggable persistence backends
  - [ ] Refactor current JSON implementation to `pkg/persistence/implementations/json/`
  - [ ] Create abstract persistence interface for multiple backends
  - [ ] Configuration-based persistence selection
- [ ] **Memento WAL Engine** - Custom append-only transaction log (Long-term)
  - [ ] WAL-based persistence inspired by RabbitMQ's Mnesia approach
  - [ ] Message event streaming (publish/ack/reject)
  - [ ] Crash recovery via log replay
  - [ ] Periodic state snapshots for performance
  - [ ] Foundation for future clustering capabilities
- [ ] **Recovery system** - Restore state after restart
  - [ ] Durable queues and exchanges
  - [ ] Persistent message recovery

### **Performance & Scalability**

- [ ] **Connection pooling** optimizations
- [ ] **Memory management** for high-throughput scenarios
- [ ] **Metrics and monitoring** integration
- [ ] **Load testing** and benchmarking

### **Management & Observability**

- [ ] **Enhanced Web UI** features
  - [ ] Real-time connection monitoring
  - [ ] Message flow visualization
  - [ ] Performance metrics dashboard
- [ ] **REST API** completeness
  - [ ] Full queue/exchange management
  - [ ] Consumer monitoring endpoints
- [ ] **Logging improvements**
  - [ ] Structured logging with correlation IDs
  - [ ] Debug modes for protocol tracing

## Development Phases

### **Phase 1: Core Messaging (Weeks 1-2)**

**Goal**: Enable push-based message consumption

**Tasks**:

1. Implement `BASIC_CONSUME` parsing and handler
2. Add consumer lifecycle management to VHost
3. Implement `BASIC_DELIVER` for message pushing
4. Add consumer cleanup with `BASIC_CANCEL`

**Files to modify**:

- `internal/core/amqp/basic.go`
- `internal/core/broker/broker.go`
- `internal/core/broker/vhost/vhost.go`
- `internal/core/broker/vhost/message.go`

### **Phase 2: Reliable Delivery (Weeks 3-4)**

**Goal**: Message acknowledgment and recovery

**Tasks**:

1. Implement `BASIC_ACK` with delivery tracking
2. Add `BASIC_REJECT` with requeue support
3. Implement `BASIC_RECOVER` for unacked messages
4. Add delivery tag management

### **Phase 3: Performance Tuning (Week 5)**

**Goal**: QoS and flow control

**Tasks**:

1. Implement `BASIC_QOS` prefetch limits
2. Add channel flow control
3. Integrate QoS with message delivery loops

### **Phase 4: Transactions (Week 6)**

**Goal**: ACID message operations

**Tasks**:

1. Implement transaction state tracking
2. Add transactional publishing
3. Implement commit/rollback logic

## Testing Strategy

### **Compatibility Testing**

- [ ] Test with official RabbitMQ clients
  - [ ] `rabbitmq/amqp091-go` (already working)
  - [ ] RabbitMQ .NET Client
  - [ ] Python `pika` library
  - [ ] Node.js `amqplib`

### **Integration Testing**

- [ ] End-to-end message flow tests
- [ ] Multi-consumer scenarios
- [ ] High-throughput stress testing
- [ ] Failure recovery testing

### **Performance Benchmarks**

- [ ] Message throughput comparison with RabbitMQ
- [ ] Memory usage under load
- [ ] Connection handling capacity

## Contributing

### **Current Priority**

The highest priority is **Phase 1: Core Messaging**. Contributors should focus on:

1. `BASIC_CONSUME` implementation
2. Consumer management system
3. Push-based message delivery

### **Getting Started**

1. Review `internal/core/amqp/basic.go` for commented-out parsers
2. Check `internal/core/broker/broker.go` for empty switch cases
3. See `.github/copilot-instructions.md` for architecture patterns

### **Code Guidelines**

- Follow existing AMQP frame processing patterns
- Add comprehensive error handling
- Include unit tests for new parsers
- Test with RabbitMQ clients for compatibility

---

## Progress Tracking

**Last Updated**: October 2025  
**Current Focus**: Phase 2 - Reliable Delivery (Message Acknowledgments)  
**Next Milestone**: `BASIC_ACK` and delivery tracking implementation

For detailed implementation tasks, see GitHub Issues tagged with the respective phase labels.