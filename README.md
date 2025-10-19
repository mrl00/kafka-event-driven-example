```markdown
# Kafka Order Service

A microservice-based application demonstrating an event-driven architecture using Kafka, written in Go. It consists of two services: `order-producer` and `order-consumer`, which produce and consume order events to/from a Kafka topic (`orders`). The project uses the `confluent-kafka-go` library and runs a Kafka cluster in KRaft mode using Docker Compose.

## Features
- **Event-Driven Architecture**: Publishes and consumes order events using Kafka.
- **Go 1.25.1**: Built with the latest Go version for robust performance.
- **Kafka in KRaft Mode**: Runs a 3-broker Kafka cluster without ZooKeeper.
- **Dockerized Services**: Services are built and run using Docker and Docker Buildx.
- **Profile Support**: Configurable for `local`, `dev`, and `prod` environments via `APP_PROFILE`.
- **Health Checks**: HTTP endpoints (`/health`) for monitoring service status.
- **AKHQ**: Web UI for managing and inspecting Kafka topics at `http://localhost:8082`.

## Prerequisites
- **Go**: Version 1.25.1.
- **Docker**: Latest version with Docker Compose and Buildx support.
- **Arch Linux** (or compatible OS): For local development.
- **Dependencies**:
  - `librdkafka` for `confluent-kafka-go`.
  - Install on Arch Linux:
    ```bash
    sudo pacman -S git gcc pkgconf librdkafka
    ```

## Project Structure
```
kafka-order-service/
├── cmd/
│   ├── order-producer/
│   │   └── main.go
│   └── order-consumer/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── kafka/
│       └── kafka.go
├── docker/
│   ├── Dockerfile.producer
│   └── Dockerfile.consumer
├── go.mod
├── go.sum
├── docker-compose.yml
└── README.md
```

## Setup and Installation

### 1. Clone the Repository
```bash
git clone https://github.com/mrl00/kafka-event-driven-example.git
cd kafka-order-service
```

### 2. Build Docker Images
Use Docker Buildx to build the `order-producer` and `order-consumer` images.

#### Build `order-producer`
```bash
docker buildx build --file docker/Dockerfile.producer -t kafka-order-service-producer:latest .
```

#### Build `order-consumer`
```bash
docker buildx build --file docker/Dockerfile.consumer -t kafka-order-service-consumer:latest .
```

**Optional - Multi-Architecture Build**:
```bash
docker buildx build --platform linux/amd64,linux/arm64 --file docker/Dockerfile.producer -t kafka-order-service-producer:latest .
docker buildx build --platform linux/amd64,linux/arm64 --file docker/Dockerfile.consumer -t kafka-order-service-consumer:latest .
```

### 3. Run the Application
Create directories for Kafka data persistence:
```bash
mkdir -p kafka1_data kafka2_data kafka3_data
```

Start the Kafka cluster and services:
```bash
docker compose up -d
```

### 4. Verify Services
- **Health Checks**:
  ```bash
  curl http://localhost:8080/health  # order-producer
  curl http://localhost:8081/health  # order-consumer
  ```
  Expected output: `OK` (HTTP 200).

- **Kafka Topic**:
  Check logs to confirm the `orders` topic creation:
  ```bash
  docker compose logs order-producer | grep "Topic orders"
  ```
  Expected output: `Topic orders created successfully` or `Topic orders already exists`.

- **AKHQ UI**:
  Access the Kafka management UI at `http://localhost:8082`.

### 5. Stop the Application
```bash
docker compose down -v
```

## Configuration
The application supports three profiles: `local`, `dev`, and `prod`. Set the profile via the `APP_PROFILE` environment variable in `docker-compose.yml` or when running containers.

### Example Configuration (`internal/config/config.go`)
```go
type Config struct {
	Profile      string
	HTTPPort     string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
}

func LoadConfig() (Config, error) {
	profile := os.Getenv("APP_PROFILE")
	if profile == "" {
		profile = "local"
	}
	switch profile {
	case "local":
		return Config{
			Profile:      profile,
			HTTPPort:     "8080",
			KafkaBrokers: []string{"kafka1:19092", "kafka2:19092", "kafka3:19092"},
			KafkaTopic:   "orders",
			KafkaGroupID: "order-consumer-group",
		}, nil
	case "dev":
		return Config{
			Profile:      profile,
			HTTPPort:     "8080",
			KafkaBrokers: []string{"dev-kafka:9092"},
			KafkaTopic:   "orders",
			KafkaGroupID: "order-consumer-group",
		}, nil
	case "prod":
		return Config{
			Profile:      profile,
			HTTPPort:     "8080",
			KafkaBrokers: []string{"prod-kafka:9092"},
			KafkaTopic:   "orders",
			KafkaGroupID: "order-consumer-group",
		}, nil
	default:
		return Config{}, fmt.Errorf("unknown profile: %s", profile)
	}
}
```

## Troubleshooting
- **Error: `exec ./order-producer: no such file or directory`**:
  - Ensure `librdkafka1` is installed in the Docker image (`RUN apt-get install -y librdkafka1` in `Dockerfile.producer`).
  - Verify binary permissions:
    ```bash
    docker run --rm -it kafka-order-service-producer:latest /bin/bash
    ls -l /app/order-producer
    chmod +x /app/order-producer
    ```
  - Check dynamic libraries:
    ```bash
    ldd /app/order-producer
    ```

- **Error: `Unknown Topic Or Partition`**:
  - Verify that `EnsureTopic` is creating the topic:
    ```bash
    docker compose logs order-producer | grep "Topic orders"
    ```
  - List topics:
    ```bash
    docker exec -it kafka1 kafka-topics.sh --list --bootstrap-server kafka1:19092
    ```

## CI/CD
To integrate with GitHub Actions for automated builds:

```yaml
name: Build Producer
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    - name: Build and push
      run: |
        docker buildx build --platform linux/amd64,linux/arm64 \
          --file docker/Dockerfile.producer \
          -t mrl00/kafka-order-service-producer:latest \
          --push .
```

## Contributing
- Fork the repository.
- Create a feature branch (`git checkout -b feature/your-feature`).
- Commit changes (`git commit -m 'Add feature'`).
- Push to the branch (`git push origin feature/your-feature`).
- Open a Pull Request.

## License
MIT License. See [LICENSE](LICENSE) for details.
```