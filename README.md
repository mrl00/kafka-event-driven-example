# Kafka Event-Driven Example

A Go-based event-driven architecture example demonstrating Apache Kafka integration with order processing. This project showcases a producer-consumer pattern using a 3-node Kafka cluster with proper replication and fault tolerance.

## 🏗️ Architecture

This project implements a microservices architecture with the following components:

- **Order Producer**: HTTP service that generates and publishes order events to Kafka
- **Order Consumer**: Service that consumes order events from Kafka and processes them
- **Kafka Cluster**: 3-node Kafka cluster with KRaft mode (no Zookeeper)
- **AKHQ**: Web UI for Kafka cluster management and monitoring

## 📋 Features

- **Event-Driven Architecture**: Asynchronous communication using Kafka
- **High Availability**: 3-node Kafka cluster with replication factor of 3
- **Health Checks**: HTTP endpoints for service health monitoring
- **Docker Support**: Complete containerization with Docker Compose
- **Kafka Management**: Built-in AKHQ for cluster monitoring
- **Structured Logging**: Using Go's structured logging with slog
- **Graceful Shutdown**: Proper context handling and resource cleanup

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.25.1+ (for local development)

### Running with Docker Compose

1. **Clone the repository**
   ```bash
   git clone https://github.com/mrl00/kafka-event-driven-example.git
   cd kafka-event-driven-example
   ```

2. **Start the entire stack**
   ```bash
   docker-compose up -d
   ```

3. **Verify services are running**
   ```bash
   docker-compose ps
   ```

4. **Check logs**
   ```bash
   # View all logs
   docker-compose logs -f
   
   # View specific service logs
   docker-compose logs -f order-producer
   docker-compose logs -f order-consumer
   ```

### Service Endpoints

- **Order Producer**: http://localhost:4000
  - Health Check: http://localhost:4000/health
- **Order Consumer**: http://localhost:4001
  - Health Check: http://localhost:4001/health
- **AKHQ (Kafka UI)**: http://localhost:9090

### Kafka Brokers

- **Kafka1**: localhost:29092
- **Kafka2**: localhost:39092
- **Kafka3**: localhost:49092

## 🏃‍♂️ Local Development

### Prerequisites

- Go 1.25.1+
- Apache Kafka (or use Docker for Kafka)

### Running Locally

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Start Kafka cluster** (using Docker)
   ```bash
   docker-compose up kafka1 kafka2 kafka3 -d
   ```

3. **Run the producer**
   ```bash
   go run cmd/order-producer/main.go
   ```

4. **Run the consumer** (in another terminal)
   ```bash
   go run cmd/order-consumer/main.go
   ```

## 📊 Project Structure

```
kafka-event-driven-example/
├── cmd/
│   ├── order-producer/          # Order producer service
│   │   └── main.go
│   └── order-consumer/          # Order consumer service
│       └── main.go
├── docker/
│   ├── Dockerfile.producer      # Producer Docker image
│   └── Dockerfile.consumer      # Consumer Docker image
├── internal/
│   ├── handler/                 # HTTP handlers
│   │   ├── hc.go               # Health check handler
│   │   └── hc_test.go          # Health check tests
│   ├── kafka/                  # Kafka client implementation
│   │   ├── kafka.go           # Core Kafka functionality
│   │   └── kafka_test.go      # Kafka tests
│   └── router/                 # HTTP router
│       └── router.go
├── docker-compose.yaml         # Multi-service orchestration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── README.md                   # This file
```

## 🔧 Configuration

### Kafka Configuration

The Kafka cluster is configured with:
- **3 brokers** with KRaft mode (no Zookeeper)
- **3 partitions** per topic
- **Replication factor of 3** for high availability
- **Topic**: `orders`
- **Consumer Group**: `order-consumer-group`

### Order Event Schema

```go
type OrderEvent struct {
    OrderID    string    `json:"order_id"`
    CustomerID string    `json:"customer_id"`
    Amount     float64   `json:"amount"`
    CreatedAt  time.Time `json:"created_at"`
}
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/handler
go test ./internal/kafka
```

### Test Coverage

The project includes unit tests for:
- Health check handlers
- Kafka producer/consumer functionality
- Error handling scenarios

## 📈 Monitoring

### AKHQ Dashboard

Access the AKHQ dashboard at http://localhost:9090 to:
- Monitor Kafka cluster health
- View topic details and message flow
- Check consumer group status
- Browse messages in topics

### Health Checks

Both services expose health check endpoints:
- Producer: `GET /health`
- Consumer: `GET /health`

## 🐳 Docker Details

### Multi-stage Builds

Both services use multi-stage Docker builds:
1. **Builder stage**: Compiles the Go application
2. **Runner stage**: Creates minimal runtime image

### Dependencies

- **librdkafka**: C library for Kafka client
- **dockerize**: For service dependency waiting
- **ca-certificates**: For HTTPS connections

## 🔄 Event Flow

1. **Order Producer** generates sample order events
2. Events are published to the `orders` topic
3. **Order Consumer** subscribes to the topic
4. Consumer processes events asynchronously
5. Both services expose health check endpoints

## 🛠️ Development

### Adding New Event Types

1. Define new event struct in `internal/kafka/kafka.go`
2. Add producer/consumer methods for the new event type
3. Update the main functions to handle the new events

### Extending the API

1. Add new handlers in `internal/handler/`
2. Register routes in `internal/router/router.go`
3. Add corresponding tests

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📚 Resources

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Kafka Go Client](https://github.com/confluentinc/confluent-kafka-go)
- [AKHQ Documentation](https://akhq.io/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)