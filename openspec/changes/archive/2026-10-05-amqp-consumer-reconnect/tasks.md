## 1. Fail-fast construction

- [x] 1.1 `NewConsumer` validates inputs, routes the library logger (`amqp.SetLogger(log.Default())`), then dials **once** via `dialSession`; dial failure returns the error (app fails to boot) — config-only intermediate version is discarded
- [x] 1.2 Store the session write-once on `Consumer`; delete the `dial` field and any config-only construction paths
- [x] 1.3 Delete all retry machinery: `supervise` loop, `retryDelay`, `sleepCtx`, dial/redial helpers

## 2. Built-in recovery (library-only)

- [x] 2.1 `dialSession()`: `amqp.DialConfig` with `Config.Recovery` enabled (`RetryInterval: 5s`, `MaxRetryCount: math.MaxInt32`) → `Channel` → `QueueDeclare`, with partial-cleanup on failure
- [x] 2.2 Session loop selects only on `ctx.Done()`, terminal `NotifyStateChange` (`StateClosed` / closed state channel), and delivery-channel closure; no own `NotifyClose` listener
- [x] 2.3 Non-terminal state changes: ignored — no teardown, no exit, no redial (library recovers underneath)
- [x] 2.4 Terminal connection death: log cause, bounded drain, close connection, `log.Fatalf` (guard against firing during a concurrent graceful stop)
- [x] 2.5 Restore fail-fast `Start` wiring: `Consume` failure in the session loop is terminal (process exits), not a reconnect trigger

## 3. Handler safety

- [x] 3.1 Pass the delivering session into `handleMessage(sess, msg)` explicitly; no session field accessors, no mutex
- [x] 3.2 Wrap handler execution with `WaitGroup.Add`/`Done` and semaphore
- [x] 3.3 Change reply publish failure path from `Ack` to `Nack(false, true)` with a log line (in `handleMessage` only; `publishError` keeps request-side ack semantics)

## 4. Observability and shutdown

- [x] 4.1 Recovery observability: route library logger via `amqp.SetLogger(log.Default())`; consumer logs terminal connection death itself
- [x] 4.2 `Stop()`: cancel session-loop ctx, wait for exit and bounded handler drain, then close the connection
- [x] 4.3 Verify `cmd/serve.go` still compiles unchanged (signal ownership stays there; `NewConsumer` error already fails the command)

## 5. Tests and verification

- [x] 5.1 Delete supervisor/redial/dial-retry tests (`TestSupervisor*`, initial-dial-retry, non-terminal-redial) — no retry code remains
- [x] 5.2 Add `TestNewConsumerFailsFastWhenBrokerUnreachable` (constructor returns error; no goroutines started) and graceful `Start`/`Stop` test (cancel → drain → connection closed)
- [x] 5.3 Unit tests: publish failure nacks with requeue; publish success acks
- [x] 5.4 Run `go build ./...`, `go vet ./...`, and `go test ./... -race` after the rework
- [ ] 5.5 ~~Manual smoke test against a local RabbitMQ: start `serve`, kill the broker container, restart it, confirm messages resume consuming and reconnect is logged~~ **CANCELLED** — no Docker/local broker in this dev environment; run manually against a real broker before release
