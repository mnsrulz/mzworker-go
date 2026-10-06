## Why

The AMQP consumer silently stops receiving messages after the broker connection drops. There was no reconnect logic anywhere: the delivery goroutine saw the closed `msgs` channel and exited without logging, nothing subscribed to close notifications, and the process kept blocking on the signal channel in `cmd/serve.go:55`, so it looked healthy while consuming nothing. Separately, a broker outage at startup failed the single `amqp.Dial` attempt and exited the process, with no in-process retry (the Docker image has no restart policy).

Since amqp091-go v1.12.0 the official RabbitMQ Go client has **built-in automatic recovery** (connection, channels, queue declaration, consumer re-subscription), and RabbitMQ's production guidelines recommend using the client's recovery rather than hand-rolling one — so this change enables it and removes **all** retry code from the application.

## What Changes

- Enable the library's built-in recovery (`Config.Recovery`) with a fixed 5s retry interval (RabbitMQ's recommended recovery delay, plus built-in jitter) and an effectively unlimited retry count, so runtime drops are recovered in place: connection reopened, durable queue redeclared, consumer re-subscribed onto the same delivery channel — no process restart, no supervisor teardown.
- **Zero retry code in the application.** The consumer dials exactly once, during `NewConsumer`; if that dial fails the constructor returns the error and the app fails to boot (non-zero exit — preserved fail-fast behavior, external supervision decides what happens next). Runtime retries are 100% the library's.
- **Terminal connection death fails the process**: when the library's recovery is exhausted (or the delivery channel closes terminally), the consumer drains in-flight handlers (bounded, 30s), closes the connection, logs the cause, and exits non-zero. No hand-rolled redial.
- Recovery lifecycle is logged via `amqp.SetLogger`: drop, recovering, recovered (including skipped topology entities), terminal death.
- Handlers receive the delivering session explicitly (`handleMessage(sess, msg)`); there is no shared session state on the consumer — replies always publish on the connection that delivered the request.
- A reply publish failure mid-handler changes from "log + Ack" to "log + Nack(requeue=true)" so the request is redelivered and the client can still receive a reply. **BREAKING** to the prior "always ack on processing failure" behavior.
- `Start()` stays non-fatal for ordinary startup wiring; `log.Fatalf` reappears only on terminal connection death (the fail-fast exit path).

## Capabilities

### New Capabilities
- `amqp-consumer-reconnect`: Connection lifecycle resilience for the AMQP consumer — library-built-in runtime recovery, fail-fast initial dial, terminal death fails the process, recovery observability, reply-failure requeue, and graceful shutdown.

### Modified Capabilities

(None — `openspec/specs/` contains no synced specs yet. Note: the unarchived `amqp-request-reply-consumer` change contains an `AMQP connection management` requirement ("unreachable URL → exit non-zero") — this change's fail-fast boot **restores** alignment with it — and an "always ack" requirement that this change supersedes; reconcile when that change is archived or synced.)

## Impact

- `amqp/consumer.go` — `NewConsumer` validates + dials once (error ⇒ boot fails), session loop keyed on terminal signals only, bounded drain, terminal death ⇒ `log.Fatalf`.
- `amqp/session.go` — `DialConfig` with `Recovery` (unchanged in this revision).
- `cmd/serve.go` — unchanged lifecycle ownership (still owns signals; `NewConsumer` error already fails the command).
- `amqp/consumer_test.go` — fail-fast constructor test, graceful start/stop, nack/ack tests; retry/redial tests deleted (no retry code left).
- Runtime: broker outages are recovered by the library without restarting the process; an unrecoverable connection kills the process visibly instead of stalling silently; un-acked messages continue to be redelivered by the broker (at-least-once preserved).
- Dependency: relies on the library's *Experimental* `Config.Recovery` API (confined to one function).
