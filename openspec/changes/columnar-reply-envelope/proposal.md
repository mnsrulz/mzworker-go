## Why

The AMQP reply format doesn't match what the client expects. mzworker-go currently replies `{"data": {"columns": [...], "rows": [{values: [...]}]}}` (row-oriented, no `requestId`, no `hasError`), while the consumer (mztrading-data) works with `{"requestId", "hasError", "value"}` where `value` is columnar — column name → array of values (same shape its own ingest worker already publishes). The client is the sole consumer and the worker is not in production, so changing the reply contract now is cheap; doing it later is not.

## What Changes

- Reply envelope becomes `{"requestId": ..., "hasError": false, "value": ...}` — **BREAKING** vs the current `{"data": ..., "error": ...}`.
- `requestId` is read from the incoming request envelope and echoed back in the reply (currently ignored).
- Query results are serialized columnar: `value` is an object mapping each column name to an array of that column's values, preserving original column order — **BREAKING** vs `columns` + `rows`.
- SQL `NULL` becomes JSON `null` in the column array.
- Errors reply `{"requestId": ..., "hasError": true, "value": {}}` — the exact shape mztrading-data's `publish(requestId, true, {})` emits; the failure message is logged, not replied.
- Non-query responses (ping) put their payload directly under `value`.
- Existing ack/nack semantics (nack+requeue on reply publish failure) are unchanged.

## Capabilities

### New Capabilities
- `columnar-reply-envelope`: The AMQP request/reply wire contract — envelope fields (`requestId`, `hasError`, `value`), columnar query result shape, null handling, and the empty-value error convention.

### Modified Capabilities

(None — `openspec/specs/` contains no synced specs yet. Note: the unarchived `amqp-consumer-reconnect` change owns connection-lifecycle requirements only and is unaffected.)

## Impact

- `amqp/dispatch.go` — envelope struct, `requestId` extraction, columnar conversion of `handler.QueryResponse`.
- `amqp/consumer_test.go` — reply-format assertions updated.
- Sole consumer (mztrading-data) must adopt the new format; its ingest worker already emits this exact shape, so the contract converges.
- No changes to request schemas, queue names, ack/nack behavior, or connection handling.
