# Kafka Order Processing Service

## Overview

This is a Go-based microservice for processing order events in an event-driven e-commerce system. It uses Apache Kafka to produce and consume order events to/from the `orders` topic. The service demonstrates a producer that sends order events (e.g., `{"orderId":"ORD001","customerId":"CUST001","amount":199.99}`) and a consumer that reads and processes these events, suitable for integration with other microservices (e.g., payment or inventory).

The project leverages the `kafka-go` library and runs alongside a Kafka cluster configured with KRaft (no ZooKeeper) using Docker Compose. It aligns with modern microservices architecture and DevOps practices, including Docker, CI/CD, and monitoring.

## Features

- **Producer**: Sends order events to the Kafka topic `orders` in JSON format.
- **Consumer**: Subscribes to the `orders` topic and processes events in real-time.
- **Kafka Integration**: Connects to a three-broker Kafka cluster (KRaft mode) for high availability.
- **Docker Support**: Integrates with a Docker Compose setup for easy deployment and testing.
- **Event-Driven**: Designed for scalability in event-driven architectures, compatible with Spring Boot or other services.

## Prerequisites

- **Go**: Version 1.20 or higher (`go version` to check).
- **Docker**: Version 20.10 or higher (`docker --version` to check).
- **Docker Compose**: Version 2.0 or higher (`docker compose version` to check).
- **Kafka Cluster**: A running Kafka cluster configured via `docker-compose.yml` (provided in the project root).

## Setup

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd kafka-order-service
   ```

2. **Set Up Kafka Cluster**:
   - Ensure the `docker-compose.yml` is in the project root (see below for reference).
   - Create directories for Kafka data:
     ```bash
     mkdir kafka1_data kafka2_data kafka3_data
     ```
   - Start the Kafka cluster and AKHQ (UI for monitoring):
     ```bash
     docker compose up -d
     ```

3. **Install Dependencies**:
   - Install the `kafka-go` library:
     ```bash
     go get github.com/segmentio/kafka-go
     ```

4. **Create the `orders` Topic**:
   ```bash
   docker exec -it kafka1 kafka-topics.sh --create --topic orders --bootstrap-server localhost:29092 --partitions 3 --replication-factor 3
   ```

## Running the Service

1. **Run the Go Program**:
   ```bash
   go run main.go
   ```
   - The program starts a producer that sends sample order events and a consumer that reads from the `orders` topic.

2. **Monitor with AKHQ**:
   - Open `http://localhost:8081` in your browser.
   - Select the `local-kraft-cluster` cluster and navigate to the `orders` topic to view messages, offsets, and consumer groups.

3. **Test with HTTP (Optional)**:
   - If integrated with a Spring Boot service, send events via:
     ```bash
     curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"orderId":"ORD001","customerId":"CUST001","amount":199.99}'
     ```

## Project Structure

```
kafka-order-service/
├── main.go              # Main Go application (producer and consumer)
├── docker-compose.yml   # Kafka cluster configuration with three brokers and AKHQ
├── kafka1_data/         # Volume for kafka1 data
├── kafka2_data/         # Volume for kafka2 data
├── kafka3_data/         # Volume for kafka3 data
└── README.md            # This file
```

## Docker Compose Configuration

The `docker-compose.yml` sets up a Kafka cluster with three brokers in KRaft mode and AKHQ for monitoring. Key configurations:
- **Image**: `confluentinc/cp-kafka:8.0.1`
- **Brokers**: `kafka1`, `kafka2`, `kafka3` with ports `29092`, `39092`, `49092`
- **AKHQ**: Accessible at `http://localhost:8081`
- **Network**: `kafka-cluster-net` for internal communication

Ensure the `CLUSTER_ID` matches across all brokers. To generate a new ID:
```bash
docker run --rm confluentinc/cp-kafka:8.0.1 kafka-storage random-uuid
```

## Testing

- **Unit Tests**: Add tests using Go's `testing` package to validate JSON serialization/deserialization.
  ```bash
  go test ./...
  ```
- **Integration Tests**: Use JUnit (if integrating with Spring Boot) or Go tests to verify producer/consumer behavior.
- **Manual Testing**: Send events via `curl` and check AKHQ for message delivery.

## Monitoring and Debugging

- **AKHQ**: Monitor topics, offsets, and consumer groups at `http://localhost:8081`.
- **Logs**: Check container logs for errors:
  ```bash
  docker compose logs kafka1
  ```
- **Prometheus**: Add Prometheus for metrics (optional, requires additional setup).

## Future Improvements

- **Kafka Connect**: Integrate with PostgreSQL using a JDBC Sink Connector to persist order events.
- **Security**: Implement `SASL_SSL` with Keycloak for OAuth2 authentication.
- **CI/CD**: Set up GitHub Actions or Jenkins for automated builds and deployment.
- **Monitoring**: Add Prometheus and Grafana for real-time metrics.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for bugs, features, or improvements.

## License

This project is licensed under the MIT License.
