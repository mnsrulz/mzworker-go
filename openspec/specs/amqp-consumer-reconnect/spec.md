# amqp-consumer-reconnect

## Purpose
Connection lifecycle resilience for the AMQP consumer: library-built-in runtime recovery with unlimited attempts, fail-fast initial connection, terminal connection death fails the process, recovery observability, reply-publish failure requeue, session-scoped handler publishing, and graceful bounded-drain shutdown.

## Requirements

### Requirement: Built-in automatic recovery
The consumer SHALL enable the amqp091-go library's built-in connection recovery (`Config.Recovery`) so that, after an established connection drops, the library re-dials the broker, reopens channels, redeclares tracked topology, and re-subscribes the consumer on the original delivery channel without restarting the process.

#### Scenario: Broker restarts at runtime
- **WHEN** an established connection is closed by the broker or the network
- **THEN** the library reconnects automatically, reopens the channel, redeclares the durable queue, re-subscribes the consumer, and deliveries resume on the same delivery channel without process restart

#### Scenario: Queue missing during recovery
- **WHEN** the request queue no longer exists while the connection is being recovered
- **THEN** topology recovery redeclares it as durable before the consumer is re-subscribed

#### Scenario: Broker unreachable at process start
- **WHEN** the process starts while the broker is unreachable
- **THEN** the initial dial error propagates from the consumer's constructor and the application fails to boot with a non-zero exit (no in-process dial retry)

### Requirement: Unlimited runtime recovery attempts
The library recovery policy SHALL be configured with a fixed retry interval of 5 seconds (plus the library's built-in jitter) and a maximum retry count large enough to be effectively infinite, so that no single outage causes recovery to be abandoned.

#### Scenario: Long broker outage
- **WHEN** the broker remains unreachable for far longer than the default retry budget
- **THEN** the library keeps retrying the reconnection at the configured interval instead of exhausting retries and closing the connection permanently

### Requirement: Fail-fast initial connection
The consumer SHALL establish its only connection during construction. There SHALL be no dial retry loop in the application: if the initial dial fails, the constructor returns the error and the application fails to boot.

#### Scenario: Broker reachable at startup
- **WHEN** the process starts while the broker is reachable
- **THEN** the consumer dials once, opens the channel, and declares the durable queue during construction

#### Scenario: Broker unreachable at startup
- **WHEN** the process starts while the broker is unreachable
- **THEN** the constructor returns the dial error, the application does not boot, and it exits non-zero so an external supervisor can act

### Requirement: Recovery observability
The consumer SHALL route the library's logger to the process log so the full recovery lifecycle (connection loss, recovery-in-progress, each recovery attempt, successful recovery including any skipped topology entities, and recovery exhaustion) is observable, and SHALL itself log terminal connection death before the process exits.

#### Scenario: Connection drops and recovery starts
- **WHEN** an established connection is closed unexpectedly while recovery is enabled
- **THEN** the library's logged output shows the close reason and recovery start, and the consumer does NOT tear down or re-dial the connection itself

#### Scenario: Recovery succeeds
- **WHEN** the library recovers the connection
- **THEN** the logged output shows the successful recovery, including any topology entities that were skipped

#### Scenario: Recovery ends terminally
- **WHEN** the connection reaches the terminal closed state carrying a fatal error
- **THEN** the consumer logs that error as terminal connection death

### Requirement: Terminal connection death fails the process
When the connection reaches the terminal closed state (recovery exhausted or non-recoverable error) or the delivery channel closes, the consumer SHALL drain in-flight handlers (bounded), close the connection, and terminate the process with a non-zero exit. The application SHALL NOT attempt its own redial.

#### Scenario: Library recovery gives up
- **WHEN** the library finishes recovery attempts with a fatal error and transitions to terminal closed
- **THEN** the consumer logs the fatal error, drains handlers (bounded), closes the connection, and the process exits non-zero

#### Scenario: Delivery channel closes terminally
- **WHEN** the delivery channel is closed (final teardown of the connection)
- **THEN** the process fails fast as above rather than continuing to run without a connection

#### Scenario: Non-terminal notifications do not trigger teardown
- **WHEN** a non-terminal state change (e.g. reconnecting) is delivered while the library is still recovering
- **THEN** the consumer does not close the connection, exit, or otherwise interfere with the library's recovery

### Requirement: Reply publish failure requeues the request
When publishing a reply fails, the consumer SHALL negatively acknowledge the request message with requeue enabled instead of acknowledging it.

#### Scenario: Reply publish fails
- **WHEN** `Publish` of the response to `reply_to` returns an error (including during a recovery window)
- **THEN** the consumer logs the error and nacks the request with `requeue=true` so it is redelivered

#### Scenario: Reply publish succeeds
- **WHEN** the response is published without error
- **THEN** the consumer acknowledges the request message

### Requirement: Session-scoped handler access
Message handlers SHALL receive the session that delivered their message as an explicit parameter and publish replies through it, rather than reading shared mutable connection state from the consumer.

#### Scenario: Handler publishes after its connection died
- **WHEN** a handler reaches its reply publish step after the connection that delivered the message has been torn down
- **THEN** the publish targets that (dead) session, returns an error, and the request is requeued per the reply-publish requirement — the handler never touches a different connection's channel

#### Scenario: No shared connection state
- **WHEN** handlers run concurrently with connection teardown
- **THEN** there is no shared mutable session field on the consumer, so no locking is required for handler publishing

### Requirement: Graceful shutdown
`Stop` SHALL cancel the consuming goroutine, wait for it to drain in-flight handlers (bounded by a drain timeout) and close the connection before returning. The same bounded drain SHALL apply before closing a terminally failed connection.

#### Scenario: Stop signal with idle consumer
- **WHEN** `Stop` is called with no messages in flight
- **THEN** the consuming goroutine exits and the connection is closed cleanly

#### Scenario: Stop signal with handlers in flight
- **WHEN** `Stop` is called while handlers are processing
- **THEN** the consumer waits up to the drain timeout for handlers to finish before closing the connection

#### Scenario: Terminal teardown with handlers in flight
- **WHEN** a terminally failed connection is torn down while handlers are processing
- **THEN** the consumer drains them (bounded) before closing; unacknowledged messages are redelivered by the broker
