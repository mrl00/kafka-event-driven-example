Kafka Order Processing Service
Overview
This is a Go-based microservice for processing order events in an event-driven e-commerce system. It consists of two separate services: a producer (order-producer) that sends order events to the Kafka topic orders, and a consumer (order-consumer) that reads and processes these events. The services use the confluent-kafka-go library for Kafka integration, support environment profiles (local, dev, prod), and include a /health endpoint to verify Kafka connectivity. The services are deployed via Docker Compose, integrating with a three-broker Kafka cluster in KRaft mode (no ZooKeeper) and AKHQ for monitoring.
The project demonstrates a modular architecture with a shared Kafka connection package and environment-specific configurations, aligning with modern microservices practices, DevOps tools, and event-driven design.
Features

Producer Service: Sends order events (e.g., {"orderId":"ORD001","customerId":"CUST001","amount":199.99}) to the orders topic.
Consumer Service: Subscribes to the orders topic and processes events in real-time.
Environment Profiles: Supports local, dev, and prod profiles via the APP_PROFILE environment variable.
Health Check: Both services expose a /health endpoint to verify Kafka connectivity.
Kafka Integration: Connects to a three-broker Kafka cluster (KRaft mode) using confluent-kafka-go.
Docker Support: Fully containerized with Docker Compose for easy deployment.
Modular Design: Shared Kafka and configuration logic in internal/kafka and internal/config.
Event-Driven: Compatible with Spring Boot or other microservices for e-commerce workflows.

Project Structure
kafka-order-service/
├── cmd/
│   ├── order-producer/
│   │   └── main.go           # Producer service with profile support
│   └── order-consumer/
│       └── main.go           # Consumer service with profile support
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management for profiles
│   └── kafka/
│       └── kafka.go          # Kafka connection logic using confluent-kafka-go
├── docker/
│   ├── Dockerfile.producer   # Dockerfile for producer
│   └── Dockerfile.consumer   # Dockerfile for consumer
├── kafka1_data/              # Volume for kafka1 data
├── kafka2_data/              # Volume for kafka2 data
├── kafka3_data/              # Volume for kafka3 data
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
├── docker-compose.yml        # Kafka cluster, AKHQ, producer, and consumer configuration
└── README.md                 # This file

Prerequisites

Go: Version 1.20 or higher (go version to check).
Docker: Version 20.10 or higher (docker --version to check).
Docker Compose: Version 2.0 or higher (docker compose version to check).
Kafka Cluster: Configured via docker-compose.yml (included in the project).
librdkafka: Required for confluent-kafka-go, installed via apk add librdkafka in Dockerfiles.

Setup

Clone the Repository:
git clone <repository-url>
cd kafka-order-service


Set Up Kafka Cluster:

Create directories for Kafka data:mkdir kafka1_data kafka2_data kafka3_data


Start the Kafka cluster, AKHQ, producer, and consumer:docker compose up -d




Install Go Dependencies (if running locally):
go mod tidy



Running the Services
The services are containerized and run via Docker Compose. By default, the docker-compose.yml uses the local profile. To start:
docker compose up -d


Producer: Automatically creates the orders topic and sends sample order events. Health check at http://localhost:8080/health.
Consumer: Reads events from the orders topic and logs them. Health check at http://localhost:8081/health.
AKHQ: Monitor topics at http://localhost:8082.

To run with a different profile (dev or prod):

Update the APP_PROFILE in docker-compose.yml:environment:
  - APP_PROFILE=dev


Ensure the Kafka brokers for dev or prod are accessible (update internal/config/config.go with real broker addresses).

To run locally (without Docker):
# Run producer with local profile
APP_PROFILE=local go run ./cmd/order-producer/main.go

# Run consumer with local profile in a separate terminal
APP_PROFILE=local go run ./cmd/order-consumer/main.go

Health Check
Both services expose a /health endpoint to verify connectivity to the Kafka cluster and the topic configured for the profile:

Producer: http://localhost:8080/health
Consumer: http://localhost:8081/health

Response:

200 OK with body OK if the service is healthy.
503 Service Unavailable with an error message if the Kafka connection fails.

Test the health check:
curl http://localhost:8080/health
curl http://localhost:8081/health

Environment Profiles
The services support three profiles, configured via the APP_PROFILE environment variable:

local: Uses the local Kafka cluster (kafka1:19092,kafka2:19092,kafka3:19092, topic orders).
dev: Uses a development Kafka cluster (e.g., kafka-dev:9092, topic orders-dev).
prod: Uses a production Kafka cluster (e.g., kafka-prod:9092, topic orders-prod).

To customize dev or prod profiles, update internal/config/config.go with the appropriate Kafka broker addresses and topic names.
Docker Compose Configuration
The docker-compose.yml configures:

Kafka Cluster: Three brokers (kafka1, kafka2, kafka3) using confluentinc/cp-kafka:8.0.1.
AKHQ: UI for monitoring at http://localhost:8082.
Producer Service: Sends events to the configured topic, exposes /health on port 8080.
Consumer Service: Reads events from the configured topic, exposes /health on port 8081.
Network: kafka-cluster-net for internal communication.

Ensure the CLUSTER_ID matches across brokers. To generate a new ID:
docker run --rm confluentinc/cp-kafka:8.0.1 kafka-storage random-uuid

Testing

Verify Topic Creation (for local profile):

The order-producer and order-consumer automatically create the orders topic.
Confirm with:docker exec -it kafka1 kafka-topics.sh --list --bootstrap-server kafka1:19092




Check Health Endpoints:
curl http://localhost:8080/health
curl http://localhost:8081/health


Monitor with AKHQ:

Open http://localhost:8082.
Select the local-kraft-cluster cluster and check the orders topic for messages.


Integration with Spring Boot (optional):

Configure a Spring Boot service with:spring:
  kafka:
    bootstrap-servers: localhost:29092,localhost:39092,localhost:49092


Send events via:curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"orderId":"ORD001","customerId":"CUST001","amount":199.99}'




Unit Tests:

Add tests to validate JSON serialization/deserialization and health check:go test ./internal/kafka
go test ./internal/config





Monitoring and Debugging

AKHQ: Monitor topics, offsets, and consumer groups at http://localhost:8082.
Logs: Check container logs for errors:docker compose logs order-producer
docker compose logs order-consumer
docker compose logs kafka1


Prometheus: Add Prometheus for metrics (optional, requires setup).

Future Improvements

Kafka Connect: Integrate with PostgreSQL using a JDBC Sink Connector to persist order events.
Security: Implement SASL_SSL with Keycloak for OAuth2 authentication.
CI/CD: Use GitHub Actions or Jenkins for automated builds and deployment.
Monitoring: Add Prometheus and Grafana for real-time metrics.

Contributing
Contributions are welcome! Submit a pull request or open an issue for bugs, features, or improvements.
License
This project is licensed under the MIT License.