# Project Backlog

A prioritized, actionable improvement backlog for the Kafka Event-Driven Example project.

---

## 🛠️ Configuration Management

- [ ] Switch broker, topic, group, and HTTP port definitions to environment variables
- [ ] Add .env file support (with dotenv loading during development)
- [ ] Document all configuration options in the README

## 📈 Observability & Monitoring

- [ ] Implement readiness/liveness/health endpoints (Kubernetes compatible)
- [ ] Integrate metrics for producer and consumer (suggest: Prometheus client)
- [ ] Enhance structured logging (add event/correlation/request IDs and log-level flags)

## 🚦 Error Handling & Resilience

- [ ] Add retry with backoff for failed event publishing/consumption
- [ ] Add dead-letter queue (DLQ) or at least file log for poison messages
- [ ] Classify/log errors better (avoid just log.Fatal; bubble critical errors up)

## 🧪 Testing & Quality

- [ ] Increase unit test coverage on all core modules
- [ ] Add integration tests using a local Kafka instance (suggest: testcontainers-go)
- [ ] Set up linting and static analysis (e.g., golangci-lint)

## 🚀 CI/CD & Automation

- [ ] Add GitHub Actions for build, lint, and test on PRs
- [ ] Add Docker image build & push steps to CI

## 🌐 API & Service Design

- [ ] Add OpenAPI/Swagger documentation for HTTP endpoints
- [ ] Implement HTTP request validation & structured error responses

## 📚 Documentation

- [ ] Expand README: architecture diagrams, message/event flow, troubleshooting
- [ ] Add detailed development and contribution instructions
- [ ] Document AKHQ setup and usage

## 🔐 Security

- [ ] Use non-root Docker containers; prefer multi-stage builds
- [ ] Verify secrets are gitignored; move to environment/configuration
- [ ] Add Kafka authentication/encryption notes in docs for users

## 💻 Deployment/Flexibility

- [ ] Provide Kubernetes manifests for all services
- [ ] Add config option for HTTP port

---

> This backlog is intended to guide ongoing improvement and best practice adoption for a robust, production-ready event-driven Go/Kafka microservices stack.