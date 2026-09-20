## Context

The mzworker-go project is a single Go binary that currently runs a gRPC server (`OptionsQueryService`) for executing DuckDB queries against local Parquet data. It has no message queue integration. The goal is to add an AMQP request-reply consumer so that clients can submit queries via a RabbitMQ/CloudAMQP broker and receive responses on a reply queue, using the standard AMQP request-reply pattern (correlation ID + reply_to).

## Goals / Non-Goals

**Goals:**
- Expose the same `ExecuteQuery` capability over AMQP using request-reply semantics.
- Reuse the existing query execution logic (`validateQuery`, `enforceLimit`, `executeQuery`).
- Manage AMQP connection lifecycle alongside the existing gRPC server (start, graceful shutdown).
- Support configurable AMQP endpoint and queue name via environment variables.

**Non-Goals:**
- Authentication/authorization on the AMQP consumer (trust the broker configuration).
- Dead-letter queue or retry logic in this iteration.
- Fan-out, pub/sub, or routing patterns — only point-to-point request-reply.
- Modifying the existing gRPC server behavior.

## Decisions

### 1. Use `github.com/rabbitmq/amqp091-go` as the AMQP client library

**Rationale**: This is the most widely used, actively maintained AMQP 0-9-1 client for Go. It's battle-tested with CloudAMQP and RabbitMQ. Alternatives like `github.com/wagslane/go-rabbitmq` add unnecessary abstraction layers.

### 2. JSON message format (not protobuf over AMQP)

**Rationale**: AMQP clients may not have protobuf. JSON is universally interoperable. The request/response payloads are small enough that JSON overhead is negligible. The proto message definitions serve as the schema source of truth, but messages on the wire are JSON-encoded.

### 3. Dedicated consumer struct with Start/Stop methods

**Rationale**: A `Consumer` struct encapsulates AMQP connection, channel, and queue state. `Start(ctx)` runs the consume loop in a goroutine. `Stop()` closes the connection gracefully. This mirrors the gRPC server lifecycle pattern already in `main.go`.

### 4. Single request queue, dynamic reply queue per message

**Rationale**: Clients declare their own exclusive reply queue and set `reply_to` on each request. The consumer publishes the response to `msg.ReplyTo` with `correlation_id` matching `msg.CorrelationId`. This is the canonical AMQP request-reply pattern and avoids the broker needing to manage reply routing.

### 5. Reuse existing `executeQuery` + `enforceLimit` directly

**Rationale**: The query logic in `query.go` is already decoupled from gRPC (it takes a context, data dir, SQL string, and limit). The AMQP handler simply deserializes the JSON request, calls these functions, and serializes the response. No refactoring needed.

## Risks / Trade-offs

- **AMQP connection drops** → The consumer will log the error and exit. A future iteration can add automatic reconnection with backoff. For now, process-level restart (via container orchestrator) is the expected recovery.
- **No message TTL or expiration** → Messages could sit indefinitely if the consumer is down. Mitigated by broker-side TTL configuration.
- **JSON schema drift** → If the proto messages change, the JSON structs must be updated manually. Mitigated by keeping the JSON structs simple and aligned with the proto fields.
