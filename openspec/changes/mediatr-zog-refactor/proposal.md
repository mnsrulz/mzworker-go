## Why

The project currently has ad-hoc request handling: the gRPC server directly calls `executeQuery` and the AMQP consumer does the same with inline logic. There is no consistent dispatch pattern, no validation layer, and adding new query types requires modifying multiple files. This refactor introduces `go-mediatr` for centralized request dispatch and `zog` for schema validation, establishing a CQRS-ready architecture that scales cleanly as new handlers are added.

## What Changes

- Add `go-mediatr` for mediator-based request dispatch (request/response pattern with pipeline behaviors).
- Add `zog` for schema validation (Zod-like API, zero dependencies, `zjson` for JSON parsing).
- Create a `handler/` package with request types, response types, Zog schemas, handler implementations, and a factory registry for AMQP JSON-to-typed-request mapping.
- Refactor `amqp.go` to dispatch through mediatr instead of calling `executeQuery` directly.
- Refactor `server.go` to dispatch through mediatr instead of calling `executeQuery` directly.
- Add a `ValidationBehavior` pipeline for cross-cutting validation concerns.
- Wire handler registration and pipeline in `main.go`.

## Capabilities

### New Capabilities

- `mediator-dispatch`: Mediator-based request dispatch using go-mediatr. Covers handler registration, request routing, pipeline behaviors, and typed request/response flow.
- `zog-validation`: Schema validation using zog. Covers request validation at handler entry, error formatting, and integration with the mediator pipeline.

### Modified Capabilities

- (none)

## Impact

- **New dependencies**: `github.com/mehdihadeli/go-mediatr`, `github.com/Oudwins/zog`
- **New package**: `handler/` with types, registry, handler, and validation files
- **Modified files**: `amqp.go` (dispatch through mediatr), `server.go` (dispatch through mediatr), `main.go` (register handlers + pipeline)
- **Removed**: Inline `executeQuery` calls from `amqp.go` and `server.go` (moved into handler)
- **No breaking changes** to external APIs (gRPC proto, AMQP message format)
