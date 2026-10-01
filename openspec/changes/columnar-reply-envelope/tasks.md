## 1. Envelope and requestId

- [x] 1.1 Replace `amqpResponse[T]` with `replyEnvelope{requestId, hasError, value}` (`value` always present, `{}` on error — no `omitempty`, no `error` field); delete `amqpQueryResponse`/`amqpRow`/`amqpValue` wrapper types
- [x] 1.2 Extract `requestId` in `handleMessage`'s envelope parse; thread it into `toAmqpResponse` and the error path (`publishError` → `hasError: true` + `value: {}`, message to log only)
- [x] 1.3 Route every failure branch in `handleMessage` (envelope parse, unknown type, factory, dispatch, conversion, marshal) through the `hasError: true` envelope

## 2. Columnar result conversion

- [x] 2.1 Implement `columnarValue` (`keys` in query column order + `map[string][]any`) with `MarshalJSON` emitting keys in order; SQL NULL → JSON `null`, native Go types pass through
- [x] 2.2 Convert `*handler.QueryResponse` into `columnarValue`; ping and other non-query responses assigned to `value` as-is

## 3. Tests and verification

- [x] 3.1 Unit test exact success reply JSON for a multi-row query: envelope fields, column key order, value types, null elements, array lengths
- [x] 3.2 Unit test error reply (`hasError: true`, `value: {}`) and `requestId` echo (present and missing cases)
- [x] 3.3 Update existing nack/ack tests for the new envelope; run `go build ./...`, `go vet ./...`, `go test ./... -count=1 -race`
