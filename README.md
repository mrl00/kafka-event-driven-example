# Kafka Event-Driven Example

[![CI](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/ci.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/ci.yml)
[![Check Pull Request Source](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/branch-check.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/branch-check.yml)
[![Docker Publish](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/docker-publish.yml)

A production-focused, event-driven architecture example in Go demonstrating Apache Kafka integration for order processing. This project showcases resilient producer-consumer patterns using a 3-node Kafka cluster with proper replication and fault tolerance.

## 🏗️ Architecture & Resilience

This project implements a microservices architecture with a focus on robustness:

- **Order Producer**: HTTP service that publishes order events to Kafka. Uses **idempotent** production and asynchronous delivery confirmations via dedicated `deliveryChan` to guarantee no messages are lost or duplicated.
- **Order Consumer**: Consumes and processes order events. Implements **Poison Pill** protection (corrupted/invalid messages are routed to a Dead Letter Queue) ensuring the service never stops due to a single bad message.
- **Graceful Shutdown**: OS signal handling (SIGINT/SIGTERM) for clean resource teardown — Kafka buffer flush, consumer group exit, and HTTP server drain.

## 📋 Production Features

- **Graceful Shutdown**: Centralized signal handling via `internal/lifecycle` with LIFO cleanup ordering
- **High Availability**: 3-node Kafka cluster with replication factor 3 in KRaft mode (no Zookeeper)
- **Data Safety**: Producer configured with `acks=all` and `enable.idempotence=true`
- **Retry Pattern**: Generic exponential backoff with jitter for all critical Kafka operations (connection, topic creation, message production)
- **Dead Letter Queue**: Invalid/unparseable messages are routed to a `.dlq` topic with rich metadata headers
- **Structured Logging**: `log/slog` with context propagation for traceability
- **HTTP Server Hardening**: Read, Write, Idle, and ReadHeader timeouts configured to prevent resource exhaustion
- **Health Checks**: HTTP endpoints for service health monitoring
- **CI/CD**: GitHub Actions for build, lint, test, Docker image publish, and cosign image signing

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.25.1+ (for local development)

### Running with Docker Compose

```bash
# Start the entire stack
docker-compose up -d

# Verify services are running
docker-compose ps

# View logs
docker-compose logs -f order-producer
docker-compose logs -f order-consumer
```

### Running Locally

```bash
# Start only the Kafka cluster
docker-compose up kafka1 kafka2 kafka3 -d

# Copy and edit environment
cp .env.example .env

# Run the producer
go run ./cmd/order-producer

# Run the consumer (in another terminal)
go run ./cmd/order-consumer
```

### Service Endpoints

| Service | URL |
| :--- | :--- |
| Order Producer | http://localhost:4000 |
| Producer Health Check | http://localhost:4000/health |
| Order Consumer | http://localhost:4001 |
| Consumer Health Check | http://localhost:4001/health |

### Kafka Brokers

| Broker | External Address |
| :--- | :--- |
| Kafka 1 | localhost:29092 |
| Kafka 2 | localhost:39092 |
| Kafka 3 | localhost:49092 |

## 🔧 Configuration

Configuration is managed via environment variables. For local development, create a `.env` file based on `.env.example`.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `KAFKA_BROKERS` | Comma-separated Kafka broker list (required) | — |
| `KAFKA_TOPIC` | Kafka topic for order events | `orders` |
| `KAFKA_GROUP_ID` | Consumer group ID (consumer only) | `order-consumer-group` |
| `KAFKA_NUM_PARTITIONS` | Number of topic partitions | `3` |
| `KAFKA_REPLICATION_FACTOR` | Replication factor for topics | `3` |
| `HTTP_PORT` | HTTP server port | `4000` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout (Go duration) | `5s` |

## 📊 Project Structure

```
kafka-event-driven-example/
├── cmd/
│   ├── order-producer/          # Order producer service entrypoint
│   └── order-consumer/          # Order consumer service entrypoint
├── internal/
│   ├── config/                  # Environment-based configuration loader
│   ├── handler/                 # HTTP handlers (health check)
│   ├── lifecycle/               # Graceful shutdown manager (signal handling, LIFO cleanup)
│   ├── mykafka/                 # Kafka client: producer, consumer, DLQ, topic management
│   ├── retry/                   # Generic retry engine with exponential backoff and jitter
│   ├── router/                  # HTTP routing
│   └── server/                  # HTTP server factory with production timeouts
├── build/
│   ├── Dockerfile.producer      # Multi-stage Docker build for producer
│   └── Dockerfile.consumer      # Multi-stage Docker build for consumer
├── docs/                        # Backlogs and documentation
├── docker-compose.yaml          # Full stack orchestration (Kafka cluster + services)
└── .github/workflows/           # CI/CD pipelines
```

## 🔄 Event Flow

```
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│   Producer   │────▶│  Kafka Topic  │────▶│   Consumer   │
│  (HTTP API)  │     │   "orders"    │     │  (Processor) │
└──────────────┘     │  3 partitions │     └──────┬───────┘
                     │  RF = 3       │            │
                     └───────────────┘            │ invalid?
                                                  ▼
                                          ┌──────────────┐
                                          │  DLQ Topic   │
                                          │ "orders.dlq" │
                                          └──────────────┘
```

## 🔄 Graceful Shutdown Flow

When a shutdown signal is received, the `lifecycle` package executes cleanup tasks in **LIFO** (reverse) order:

1. **HTTP Server**: Calls `srv.Shutdown(ctx)` to stop accepting new connections and drain in-flight requests
2. **Kafka Client**:
   - **Producer**: Executes `producer.Flush(timeout)` to deliver buffered messages, then `producer.Close()`
   - **Consumer**: Calls `consumer.Close()`, notifying the cluster for immediate group rebalancing
3. **DLQ Producer**: Closes the DLQ producer connection
4. **Context**: The parent context is cancelled, signaling all goroutines to stop

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./internal/...

# Run tests with coverage
go test -cover ./...
```

### Test Coverage

The project includes unit tests for:
- Retry engine (success after retries, fatal error abort, context cancellation)
- Kafka config and broker list building
- Producer context cancellation handling
- DLQ topic naming
- Health check HTTP handler

## 🐳 Docker Details

### Multi-stage Builds

Both services use multi-stage Docker builds for minimal runtime images:
1. **Builder stage**: Compiles the Go application with CGO (required for librdkafka)
2. **Runner stage**: Debian bookworm-slim with only `librdkafka1` and `ca-certificates`

### Docker Images

Images are published to Docker Hub on every push to `main`:
- `docker.io/<user>/kafka-event-driven-example-producer`
- `docker.io/<user>/kafka-event-driven-example-consumer`

Each image is signed with [cosign](https://docs.sigstore.dev/cosign/overview/) for supply chain security.

## 📝 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
