O ficheiro **README.md** atualizado foi gerado e está pronto para ser utilizado. Como o conteúdo é extenso e o ambiente de chat por vezes dificulta a cópia integral de blocos de texto muito grandes, compilei a versão final abaixo.

Podes copiar o conteúdo diretamente para o teu ficheiro `README.md` no projeto:

```markdown
# Kafka Event-Driven Example (Production Ready)

[![CI](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/ci.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/ci.yml)
[![Check Pull Request Source](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/branch-check.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/branch-check.yml)
[![Docker Publish](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/mrl00/kafka-event-driven-example/actions/workflows/docker-publish.yml)

Este projeto é um exemplo de arquitetura event-driven em Go, demonstrando a integração com Apache Kafka para processamento de pedidos. Foi refatorado para demonstrar padrões de produção focados em resiliência, tratamento de erros e encerramento gracioso (Graceful Shutdown).

## 🏗️ Arquitetura e Resiliência

O projeto implementa uma arquitetura de microsserviços com foco em robustez:

- **Order Producer**: Serviço HTTP que publica eventos de pedidos. Utiliza **idempotência** e confirmações de entrega assíncronas via `deliveryChan` para garantir que nenhuma mensagem seja perdida ou duplicada.
- **Order Consumer**: Consome e processa pedidos. Implementa proteção contra **Poison Pills** (mensagens corrompidas) para garantir que o serviço não pare em caso de falha de parsing, além de respeitar o cancelamento de contexto.
- **Graceful Shutdown**: Gerencia sinais do SO (SIGINT/SIGTERM) para encerrar recursos de forma limpa, garantindo flush de buffers no Kafka e saída coordenada do grupo de consumo.

## 📋 Funcionalidades de Produção

- **Encerramento Gracioso**: Gerenciamento centralizado via pacote `internal/lifecycle`.
- **Alta Disponibilidade**: Cluster Kafka de 3 nós com fator de replicação 3 em modo KRaft.
- **Segurança de Dados**: Produtor configurado com `acks=all` e `enable.idempotence=true`.
- **Logs Estruturados**: Uso de `slog` com propagação de contexto para melhor rastreabilidade.
- **Health Checks**: Endpoints HTTP para monitoramento de integridade e prontidão.

## 🚀 Início Rápido

### Pré-requisitos

- Docker e Docker Compose
- Go 1.25.1+ (para desenvolvimento local)

### Executando com Docker Compose

```bash
docker-compose up -d
```

## 🔧 Configuração

Configurações geridas via variáveis de ambiente. Para uso local, crie um ficheiro `.env` baseado no `.env.example`.

| Variável | Descrição | Padrão |
| :--- | :--- | :--- |
| `KAFKA_BROKERS` | Lista de brokers Kafka (obrigatório) | - |
| `KAFKA_TOPIC` | Tópico para os eventos de pedidos | `orders` |
| `KAFKA_GROUP_ID` | ID do grupo de consumo (Consumer) | `order-consumer-group` |
| `SHUTDOWN_TIMEOUT` | Tempo limite para encerramento gracioso | `30s` |
| `HTTP_PORT` | Porta do servidor HTTP | `4000` |

## 📊 Estrutura do Projeto

```text
kafka-event-driven-example/
├── cmd/
│   ├── order-producer/      # Serviço de produção de pedidos
│   └── order-consumer/      # Serviço de consumo e processamento
├── internal/
│   ├── kafka/              # Implementação resiliente do cliente Kafka
│   ├── lifecycle/          # Gerenciador de sinais e encerramento (Shutdown)
│   ├── server/             # Fábrica de servidores HTTP controlados
│   ├── handler/            # Handlers HTTP
│   └── router/             # Roteamento centralizado
└── ...
```

## 🔄 Fluxo de Encerramento (Graceful Shutdown)

Ao receber um sinal de paragem, o pacote `lifecycle` executa as tarefas de limpeza em ordem reversa (**LIFO**):

1. **HTTP Server**: Encerra o servidor via `srv.Shutdown(ctx)`, interrompendo a aceitação de novas conexões.
2. **Kafka Client**: 
   - **Producer**: Executa `producer.Flush(timeout)` para enviar mensagens pendentes antes de fechar.
   - **Consumer**: Executa `consumer.Close()`, notificando o cluster para rebalanceamento imediato.
3. **Contexto**: Cancela o `context.Context` global, sinalizando a paragem imediata de goroutines.

## 🧪 Testes

O projeto inclui testes unitários para a lógica de Kafka e handlers:

```bash
go test ./...
```

## 📝 Licença

Este projeto está licenciado sob a MIT License.
```