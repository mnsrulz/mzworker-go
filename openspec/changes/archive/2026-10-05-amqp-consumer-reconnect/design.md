## Context

`amqp/consumer.go` originally owned a single `*amqp.Connection` / `*amqp.Channel` pair created once in `NewConsumer` with no `NotifyClose` subscription, no retry, and no backoff. When the broker dropped the connection, the `msgs` delivery channel closed, the delivery goroutine exited silently (`consumer.go:89-91` in the old code), and `cmd/serve.go:55` kept blocking on the signal channel — the process looked healthy while consuming nothing. A failed startup `Dial` propagated to `log.Fatal`.

**Library facts that drive this design (verified in amqp091-go v1.15.0):**

- Since v1.12.0 the library has **built-in automatic recovery**, opt-in via `Config.Recovery` (`amqp.Dial` leaves it off). It re-dials, reopens channels, redeclares tracked topology, and **re-subscribes consumers onto the original delivery channel**. It is documented as *Experimental*.
- Recovery policy: `ReconnectionConfig{RetryInterval, MaxRetryCount}` — fixed interval + 0–500ms jitter, `for i := 0; i < MaxRetryCount; i++`. **No infinite mode**; default `MaxRetryCount` is 5 (≈25s outage budget). The counter is per-`Reconnect()` call, so it resets on each new drop.
- On every drop, `Connection.shutdown()` pushes an `*amqp.Error` to **all registered `NotifyClose` listeners before recovery starts**; listener channels are only *closed* on final teardown (`closeResources`). The delivery channel is likewise only closed on final teardown (`consumers.close()` ← `closeResources`).
- `NotifyStateChange` reports `StateOpen → StateReconnecting → StateOpen` per successful cycle, and a terminal `StateClosed` carrying the fatal `Err` when recovery is exhausted or the error is non-recoverable.
- RabbitMQ's Production Deployment Guidelines: *"If the client provides automatic connection recovery, it is recommended to use it instead of developing your own recovery mechanism."* The docs' "Go has no built-in recovery" caveat predates v1.12.0. Docs also cite ~5s as the common client recovery delay, and confirm unacked messages are automatically requeued on connection loss.

**Implication:** any application-level retry loop would either race the library's recovery (double recovery) or duplicate a mechanism the library already provides. The application should retry **nothing**: dial once at boot, let the library recover forever at runtime, and fail fast when the library gives up.

## Goals / Non-Goals

**Goals:**
- Runtime recovery handled entirely by the library's built-in mechanism (per official guidance): connection, queue declaration, and consumer subscription recovered in place.
- **Zero retry code in the application** — no dial-retry loop, no backoff, no redial loop. Fail fast at boot; the library retries at runtime; terminal death fails the process.
- Recovery lifecycle is observable in logs (drop, recovering, recovered, skipped topology, terminal death).
- Reply publish failures during any window requeue the request instead of dropping it.
- Graceful shutdown drains in-flight handlers (bounded).

**Non-Goals:**
- Publisher confirms, dead-letter queues, QoS/prefetch tuning.
- Changing the request/reply schema or handler dispatch.
- Process-level restart policies (Docker `restart:`, systemd/K8s) — recommended follow-up, but the app must be correct without them.
- Mitigating stale delivery-tag acks after recovery (library/protocol-level; see risks).
- Fixing `handleMessage`'s `context.Background()` (handlers don't observe cancellation) — follow-up.

## Decisions

### D1: Library owns all runtime recovery; the application dials exactly once
`dialSession` uses `amqp.DialConfig` with `Config.Recovery` enabled: `RetryInterval: 5s` (RabbitMQ's recommended recovery delay; library adds jitter), `MaxRetryCount: math.MaxInt32` (no infinite mode exists — this makes a single outage's retry budget effectively unbounded). The **only** application dial is the one in `NewConsumer`, which runs before any goroutine starts: failure returns an error, `cmd/serve.go` fails the command, and the process never boots. After that single dial, every subsequent reconnection attempt is issued by the library.

**Alternative considered (rejected):** thin supervisor for startup + terminal redial (the previous iteration of this change) — works, but it is retry code the user does not want, and the startup half re-implements exactly what fail-fast boot + external supervision already covers.

**Alternative considered (rejected):** rely on the library alone with default config — default `MaxRetryCount: 5` permanently kills the connection after ~25s of outage; we need the config change even though we write no loops.

### D2: Terminal detection via `NotifyStateChange` + delivery-channel closure only
The session loop selects on exactly three things: `ctx.Done()`, `NotifyStateChange` (`StateClosed` or a closed state channel ⇒ terminal), and the delivery channel closing ⇒ terminal. The consumer registers **no `NotifyClose` listener of its own** — the library's internal `watchConnection` already has one, and reacting to drop notifications would race the library's recovery. Recovery lifecycle logging comes free via `amqp.SetLogger` (the library's default logger is `NullLogger`). A non-terminal state change (reconnecting/recovered) is ignored.

**Rationale:** only terminal signals warrant action, and that action is now *fail the process*, not *retry*.

### D3: Fail fast — no retry machinery anywhere in the application
There is no backoff type, no retry-delay field, no dial loop, no redial loop. Three outcomes, one behavior each:

1. **Boot:** initial dial fails ⇒ `NewConsumer` returns error ⇒ app does not boot (non-zero exit). Same philosophy as the original code, preserved deliberately.
2. **Runtime outage:** library retries at its configured 5s flat interval (with jitter) — the application does nothing.
3. **Terminal death:** library gives up or the delivery channel closes terminally ⇒ bounded drain, close connection, log the cause, `log.Fatalf` (non-zero exit). The process dies *visibly* instead of stalling invisibly; an external supervisor decides whether to restart.

**Rationale:** a flat 5s interval is already the library's policy (and RabbitMQ's cited client recovery delay); mirroring it in our code adds no value, only surface. Failing the process on terminal death is the same trust-the-external-supervisor posture as fail-fast boot.

### D4: Session passed explicitly to handlers — no shared mutable state
`NewConsumer` stores the single `sess` once (write-once before any goroutine starts); `sessionLoop` passes it into `handleMessage(sess, msg)` and replies publish through that same session. There is no mutex, no accessor, no swapping — the only session that ever exists is the one `NewConsumer` dialed, and a handler can never publish on a connection other than the one that delivered its message. The `session` interface (`Consume`/`Publish`/`StateChanges`/`Close`) keeps tests free of a real broker.

### D5: No teardown during recovery; bounded drain only at shutdown or terminal teardown
Because the connection and channel objects survive recovery in place, there is nothing to drain on a normal drop. Handlers that run during the recovery window get a `Publish` error → `Nack(requeue)` → the broker redelivers (unacked messages are requeued automatically on connection loss anyway). The 30s bounded drain applies when: `Stop()` is called, or the process is about to exit for terminal death.

**Alternative considered (rejected):** drain before every drop — impossible now (we don't initiate the teardown; the library recovers underneath us) and unnecessary.

### D6: `Nack(requeue=true)` on reply publish failure (unchanged)
`Publish` failure → log + `Nack(false, true)` instead of `Ack`. **BREAKING** vs the original "always ack" behavior. Risk of redelivery loops for permanently unsendable replies is accepted (publish only fails when the channel is dead or closed); log clearly.

### D7: `Start()`/`Stop()` lifecycle (unchanged shape, fail-fast terminal)
`Start(ctx)` spawns one goroutine running the session loop; `Stop()` cancels the context, waits for that goroutine to drain (bounded) and close the connection. Terminal death inside the goroutine exits the process (guarded against firing during a concurrent graceful stop). `cmd/serve.go` remains the signal owner and compiles unchanged.

## Risks / Trade-offs

- [Experimental API — `Config.Recovery` may change] → Confined to `dialSession` config; tests pin the detection contract (terminal vs non-terminal).
- [No in-process boot retry: a broker blip during deploy means the container fails to boot] → Deliberate (user decision); needs an external restart policy (Docker image currently has none — recommended follow-up).
- [Terminal death kills the process] → Deliberate fail-fast; replaces the old silent stall with a visible non-zero exit. External supervisor restarts if configured.
- [Stale delivery-tag ack after recovery: a handler acking a pre-drop tag on the recovered channel gets a broker 406, closing the channel] → Library's own channel recovery reopens and re-subscribes; at-least-once preserved (message was requeued anyway). No mitigation in this change; noted as follow-up (epoch/tag validation).
- [Redelivery loops on permanently failing replies] → Accept per D6; log every nack; revisit with dead-letter if observed.
- [Spec conflict with unarchived `amqp-request-reply-consumer`] → Its "unreachable URL → exit non-zero" requirement is **restored** by fail-fast boot; its "always ack" requirement is superseded here. Reconcile when that change is archived.

## Migration Plan

No data migration or config change — same env vars (`AMQP_URL`, `AMQP_REQUEST_QUEUE`). Deploy = ship the new binary; rollback = redeploy previous image. Clients may see late replies via redelivery and duplicate `correlation_id` responses (request/reply correlation already permits this).

## Open Questions

- None blocking. Follow-ups: handler `ctx` cancellation; stale-ack epoch guard; container restart policy for boot/terminal failures.
