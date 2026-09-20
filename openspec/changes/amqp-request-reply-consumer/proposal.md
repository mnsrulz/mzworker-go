## Why

The project currently only exposes DuckDB query execution via gRPC. Adding an AMQP request-reply consumer enables the same query capability over a message broker, allowing decoupled, asynchronous clients (services, workers, scripts) to execute queries without maintaining persistent gRPC connections. This is useful for environments where AMQP/RabbitMQ is the primary messaging infrastructure.

## What Changes

- Add an AMQP consumer that listens on a request queue, processes `ExecuteQueryRequest` payloads, and returns `ExecuteQueryResponse` on the reply queue using correlation ID matching.
- Introduce AMQP connection management (connect, declare queues, handle reconnection).
- Wire the consumer into the existing server startup/shutdown lifecycle alongside the gRPC server.
- Add `github.com/rabbitmq/amqp091-go` as a new dependency.

## Capabilities

### New Capabilities

- `amqp-request-reply`: AMQP-based request-reply consumer for executing DuckDB queries. Covers connection management, queue declaration, message consumption, request deserialization, query execution delegation, response serialization, and reply publishing.

### Modified Capabilities

- (none)

## Impact

- **New dependency**: `github.com/rabbitmq/amqp091-go` added to `go.mod`.
- **New files**: AMQP consumer implementation (connection, consumer, handler).
- **Modified files**: `main.go` to start/stop AMQP consumer alongside gRPC server; environment variables for AMQP configuration (`AMQP_URL`, `AMQP_REQUEST_QUEUE`).
- **No breaking changes** to existing gRPC interface or query execution logic.
