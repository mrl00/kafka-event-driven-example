# Project Backlog

A prioritized, actionable improvement backlog for the Kafka Event-Driven Example project.
Each backlog file below is a self-contained prompt that can be used directly with an AI coding assistant.

---

## 📋 Backlog Index

| # | Area | File | Priority | Impacto no Portfólio |
|---|------|------|----------|---------------------|
| 1 | 🔐 Segurança | [backlog-security.md](./backlog-security.md) | 🔴 Alta | Kafka TLS/SASL, container hardening, HTTP middleware, input validation |
| 2 | 📈 Observabilidade | [backlog-observability.md](./backlog-observability.md) | 🔴 Alta | OpenTelemetry tracing, Prometheus metrics, Grafana dashboards |
| 3 | 🚦 Resiliência | [backlog-resilience.md](./backlog-resilience.md) | 🔴 Alta | Graceful shutdown, retry backoff, DLQ, bug fixes, HTTP API |
| 4 | 🧪 Testes | [backlog-testing.md](./backlog-testing.md) | 🟡 Média | Interfaces/mocks, unit tests, integration (testcontainers), linting, coverage |
| 5 | 🏗️ Infraestrutura | [backlog-infra.md](./backlog-infra.md) | 🟡 Média | K8s manifests, Helm chart, CI/CD security scanning, ADRs |

---

## ✅ Already Done

- [x] Switch broker, topic, group, and HTTP port definitions to environment variables
- [x] Add .env file support (with dotenv loading during development)
- [x] Document all configuration options in the README
- [x] Add GitHub Actions for build, lint, and test on PRs
- [x] Add Docker image build & push steps to CI
- [x] Docker image signing with cosign
- [x] Branch protection (PRs to main only from dev)

---

## 🗺️ Suggested Implementation Order

```
Phase 1 — Foundation (Resiliência)
├── Bug fix: NewConsumer error handling
├── Graceful shutdown
├── HTTP server timeouts
└── Context keys tipadas

Phase 2 — Quality (Testes + Segurança básica)
├── Interfaces e testabilidade (refactor)
├── Unit tests abrangentes
├── golangci-lint
├── Input validation
└── HTTP security middleware

Phase 3 — Observabilidade
├── Structured logging consistente
├── Métricas Prometheus
├── Health check aprimorado (readiness/liveness)
├── Tracing OpenTelemetry
└── Stack Docker Compose (Prometheus + Grafana + Jaeger)

Phase 4 — Resiliência Avançada
├── Retry com exponential backoff
├── Dead Letter Queue (DLQ)
└── HTTP Order endpoint (POST /orders)

Phase 5 — Infraestrutura & Deploy
├── Kubernetes manifests (Kustomize)
├── Helm chart
├── CI/CD aprimorado (security scanning, multi-arch, SBOM)
└── Documentação de arquitetura (ADRs)

Phase 6 — Security Avançada
├── Container hardening (non-root, distroless)
├── Kafka TLS/SASL
└── Secrets management documentation
```

---

> This backlog is intended to guide ongoing improvement and best practice adoption for a robust, production-ready event-driven Go/Kafka microservices stack.