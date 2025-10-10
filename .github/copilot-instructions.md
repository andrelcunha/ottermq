# OtterMQ AI Agent Instructions

## Project Overview
OtterMQ is a **high-performance message broker** implementing the **AMQP 0.9.1 protocol** in Go, designed for RabbitMQ compatibility. The project consists of a broker backend and a Vue/Quasar management UI.

## Architecture Components

### Core Structure
- **`internal/core/broker/`**: Main broker orchestration and connection handling
- **`internal/core/amqp/`**: AMQP 0.9.1 protocol implementation (framing, handshake, message parsing)
- **`internal/core/broker/vhost/`**: Virtual host management with exchanges, queues, and message routing
- **`web/`**: Fiber-based REST API with Swagger docs for management UI
- **`ottermq_ui/`**: Vue 3 + Quasar SPA frontend

### Key Patterns

#### AMQP State Management
The broker maintains **stateful AMQP connections** through `ChannelState` in `internal/core/amqp/`. Multi-frame messages (method + header + body) are assembled across multiple TCP packets before processing.

```go
// Pattern: Check for complete message state before processing
if currentState.MethodFrame.Content != nil && currentState.HeaderFrame != nil && 
   currentState.BodySize > 0 && currentState.Body != nil {
    // Process complete message
}
```

#### Graceful Shutdown Flow
Uses atomic flags and wait groups for coordinated shutdown:
1. Set `ShuttingDown.Store(true)` to reject new connections
2. `BroadcastConnectionClose()` to notify all clients
3. `ActiveConns.Wait()` for graceful cleanup

#### Persistence Interface
Implements swappable persistence via `internal/core/persistdb/persistence/`. Currently uses JSON files in `data/` directory. Future plans include:
- Refactoring to `pkg/persistence/` with pluggable backends
- **Memento WAL Engine**: Custom append-only transaction log inspired by RabbitMQ's Mnesia
- Event-driven persistence optimized for message broker workloads

## Development Workflows

### Build & Run
```bash
make build           # Builds broker with git version tag
make run            # Build and run broker only
make docs           # Generate Swagger API docs
make test           # Run all tests with verbose output
make lint           # Run golangci-lint for code quality
make install        # Install binary to GOPATH/bin
make clean          # Remove all build artifacts

# UI Integration
make ui-deps        # Install UI dependencies (npm install)
make ui-build       # Build UI and copy to ./ui directory
make build-all      # Build both UI and broker (production ready)
make run-dev        # Run broker only (UI runs separately on :9000)
```

### UI Integration Modes
1. **Development**: Use `make run-dev` for broker, run UI separately with `cd ottermq_ui && quasar dev`
2. **Production**: Use `make build-all` to build UI into `./ui` directory for embedded serving
3. **No symlinks required** - UI files are copied directly via makefile targets

### Configuration
Environment variables override `.env` file settings. Key configs:
- `OTTERMQ_BROKER_PORT=5672` (AMQP)
- `OTTERMQ_WEB_PORT=3000` (Management UI)
- `OTTERMQ_QUEUE_BUFFER_SIZE=100000` (Message buffering)

## Project-Specific Conventions

### Database Setup
On first run, creates SQLite database in `data/ottermq.db` with default admin user (guest/guest). Handles migration through `persistdb.InitDB()`.

### AMQP Frame Processing
Each connection runs its own goroutine calling `processRequest()` which routes by `ClassID` (CONNECTION, CHANNEL, EXCHANGE, QUEUE, BASIC). State is connection-specific and channel-scoped.

### Error Patterns
- Network errors during shutdown are expected (`net.ErrClosed`)
- Incomplete AMQP frames return `nil` to continue receiving
- Protocol violations return errors that close connections

### API Authentication
Uses JWT middleware for all `/api/*` routes except `/api/login`. Tokens include user role information for authorization.

## Integration Points

### Broker-Web Communication
Web server connects to broker via standard AMQP client (`rabbitmq/amqp091-go`) on localhost, enabling real-time management operations through the same protocol as external clients.

### Message Routing
VHost contains `MessageController` interface for exchange-to-queue routing. Default implementation supports direct, fanout, topic exchanges with binding key matching.

### Cross-Component Dependencies
- Web handlers inject broker instance for read operations (listings, stats)
- AMQP operations go through connected client for consistency
- Persistence layer abstracts storage from broker logic

### Persistence Architecture
Currently uses JSON file storage with plans for **Memento WAL Engine**:
- **Current**: JSON snapshots in `data/` directory via `DefaultPersistence`
- **Planned**: Event-sourced WAL similar to RabbitMQ's Mnesia approach
- **Architecture**: Swappable backends via persistence interface
- **Events vs State**: Memento will use append-only event log vs JSON's state snapshots

Focus on **protocol compliance**, **connection lifecycle management**, and **stateful message assembly** when working with AMQP components. UI changes require understanding the REST API contract defined in `web/handlers/api/`.