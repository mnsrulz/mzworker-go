## 1. Dependencies & Setup

- [x] 1.1 Add `github.com/rabbitmq/amqp091-go` dependency via `go get`
- [x] 1.2 Create `amqp.go` file for the consumer implementation

## 2. AMQP Consumer Implementation

- [x] 2.1 Implement `Consumer` struct with connection, channel, queue name, and dataDir fields
- [x] 2.2 Implement `NewConsumer(amqpURL, queueName, dataDir string)` constructor that connects to the broker and declares the queue
- [x] 2.3 Implement `Start(ctx context.Context)` method with consume loop, JSON deserialization, query execution, and reply publishing
- [x] 2.4 Implement `Stop()` method to close channel and connection gracefully

## 3. Request-Reply Handler

- [x] 3.1 Define `amqpRequest` and `amqpResponse` JSON structs matching the proto message fields
- [x] 3.2 Implement handler that deserializes request, calls `executeQuery`/`enforceLimit`, serializes response, and publishes to `reply_to` with matching `correlation_id`
- [x] 3.3 Handle error cases: invalid JSON, query errors, missing `reply_to` — publish error response or log warning

## 4. Integration & Lifecycle

- [x] 4.1 Read `AMQP_URL` and `AMQP_REQUEST_QUEUE` env vars in `main.go`
- [x] 4.2 Start the AMQP consumer in `main.go` alongside the gRPC server (skip if env vars not set)
- [x] 4.3 Wire AMQP consumer shutdown into the existing signal handler alongside gRPC graceful stop

## 5. Verification

- [x] 5.1 Run `go build` to verify compilation
- [x] 5.2 Run `go vet` and existing tests to verify no regressions
