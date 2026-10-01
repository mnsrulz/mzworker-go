## Context

`amqp/dispatch.go` currently converts `handler.QueryResponse{Columns, Rows [][]any}` into `{"data": {"columns": [...], "rows": [{values: [...]}]}}` via `amqpValue` wrappers (`string_val`/`double_val`/...). The request envelope carries a `requestId` that is never read; the reply carries no correlation of request to reply except AMQP `correlationId`. The consumer (mztrading-data) expects the contract its own ingest worker already publishes: `{requestId, hasError, value}` with columnar `value` (see `api/worker.ts:579-587`). The worker is not in production and the consumer is a single known client — breaking the reply format now costs nothing.

## Goals / Non-Goals

**Goals:**
- One reply envelope: `requestId` echo, `hasError` flag, `value` always present (empty object on error).
- Columnar query serialization preserving column order, with SQL NULL → JSON `null`.
- Keep ack/nack, session, and connection behavior untouched.

**Non-Goals:**
- Response compression, paging, or schema versioning fields.
- Changing request parsing/validation or `handler` package types.
- The unarchived `amqp-consumer-reconnect` change (connection lifecycle) — unaffected.

## Decisions

### D1: One envelope struct, mirroring the reference contract
A single `replyEnvelope` struct: `RequestId string`, `HasError bool`, `Value any` — no `error` field. Go structs marshal in field order, so output is deterministic (`requestId, hasError, value`) without a custom marshaler. Success sets `Value` to the payload; every failure path (envelope parse, unknown type, factory, dispatch, conversion) sets `HasError: true` with `Value` = empty object `{}` — exactly matching mztrading-data's `publish(requestId, true, {})`. The failure message goes to the log only, as in the reference implementation.

**Alternative considered (rejected):** keeping `data`/`error` and adding `hasError` — two overlapping error signals and a renamed `value` anyway; since this is breaking regardless, do it once.

### D2: Columnar conversion with order preservation via a dedicated type
`columnarValue` holds `keys []string` (query column order) and `cols map[string][]any`, with a hand-written `MarshalJSON` that emits keys in slice order. Plain `map[string][]any` was rejected because `encoding/json` sorts map keys alphabetically, which would silently reorder columns relative to the query and the reference format.

Row `i` of `Rows` appends to `cols[Columns[j]]`. Cell values pass through as-is (`nil` → JSON `null`; `string`/`float64`/`int64`/`bool` keep their natural JSON form). The `amqpValue` wrapper types (`string_val` etc.) are deleted — they existed only for the row-oriented shape.

**Accepted limitation:** duplicate column names collide in the map (last wins). The previous row-oriented shape tolerated duplicates; DuckDB result sets in practice have unique names, and the sole consumer doesn't rely on duplicates.

### D3: `requestId` read during envelope parsing
`handleMessage` already unmarshal-parses the request for `requestType`; extend that anonymous struct with `RequestId string \`json:"requestId"\`` and thread it through `toAmqpResponse(requestId, resp)` and the error path. No second parse.

### D4: Non-query payloads go under `value` unchanged
`PingResponse` (and any future non-query type) is assigned directly to `Value`. The old `amqpResponse[T]` generic disappears.

## Risks / Trade-offs

- [Breaking reply format for the existing consumer] → Sole consumer, not in production (user decision); its ingest worker already emits the target shape, so both sides converge on one contract.
- [Duplicate column names lose data] → Accepted per D2; revisit only if a query actually returns duplicates.
- [Hand-written `MarshalJSON` can drift from expectations] → Covered by unit tests asserting exact output, including key order and nulls.

## Migration Plan

Ship the worker; the consumer switches its parsing in the same window (single known client, no production traffic). Rollback = redeploy previous worker image; old client code that expects `data`/`rows` is retained nowhere, so rollback pairs with a client rollback if needed.

## Open Questions

- None blocking.
