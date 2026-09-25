## Why

`amqp/` and `cmd/` evolved as two independent dispatch branches to the same DuckDB handlers. `cmd/*` uses `mediatr.Send[TReq,TResp]` with per-handler registration in `cmd/root.go:112`, while `amqp/consumer.go:186` duplicates dispatch via a hand-rolled `dispatchMediatr` type-switch and `handler/registry.go:15` factory map. This creates drift: adding a handler requires 2 registrations, validation lives per-handler (`handler/query_handler.go:21`) while `handler/validation.go:10` pipeline is log-only and only `cmd` hits it. The goal is a single handler-owned dispatch where both transports know `requestType + payload` and the `mediatr` pipeline is smart enough to handle `any`.

## What Changes

- Unify `handler/registry.go` into a single `requestType -> {RequestFactory, dispatchAny}` registry where `dispatchAny func(ctx, any)(any,error)` is a closure capturing the typed `mediatr.Send[TReq,TResp]` at registration time.
- Make handlers the source of truth: each `handler/*_handler.go` registers via one `Register(requestType, factory, handlerFactory)` that wires both the JSON factory and the typed `mediatr` handler plus its `dispatchAny`.
- Enhance `handler/validation.go` `ValidationBehavior` from log-only to real `zog` validation shared by both branches (fail-closed, pipeline before `Handle`).
- Add `handler/mediatr_bridge.go` smart dispatch so `amqp/consumer.go:142` can call `Dispatch(ctx, any)` via `mediatr` instead of a separate `handler.Dispatch` reflect map; `cmd/*RunE` either keeps `mediatr.Send` or also goes through the registry dispatch.
- Extract unified bootstrap `handler.Init(dataDir)` replacing `cmd/root.go:112` `registerMediatr` duplication; supports executor opt-in (`PingHandler` no dep, 5 SQL handlers need `QueryExecutor`).
- Remove duplicate `amqp/consumer.go:186` `dispatchMediatr` switch and `handler/*_handler.go` local `Validate` duplication (keep schemas, move check to pipeline).

## Capabilities

### New Capabilities
- `unified-handler-registry`: Single registry mapping `requestType` string to `RequestFactory` and typed `mediatr` dispatch closure, deduces handler from request type + data as `any` at receive time for both `cmd` and `amqp`.
- `mediatr-smart-dispatch`: Smart `mediatr` pipeline that handles `any` requests via `dispatchAny` without caller knowing generics, plus shared bootstrap and executor opt-in.

### Modified Capabilities
- `zog-validation`: Move validation from per-handler `Handle` into shared `ValidationBehavior` pipeline so `cmd` and `amqp` both fail-closed on validation (return error to `cobra` as `stderr exit 1`, to `amqp` as `{error:...}` envelope with `Ack`).

## Impact

- Affected code: `handler/registry.go`, `handler/validation.go`, new `handler/mediatr_bridge.go` / `handler/bootstrap.go`, `handler/*_handler.go` (5 SQL + ping) init registration, `cmd/root.go` + `cmd/*.go` RunE, `amqp/consumer.go` `handleMessage`/`dispatchMediatr`.
- Dependencies: no new deps, uses existing `github.com/mehdihadeli/go-mediatr` and `github.com/Oudwins/zog`.
- External APIs unchanged: CLI flags and AMQP JSON envelope `requestType`/`requestId`/`replyTo` stay same; behavior change: `amqp` now gets pipeline validation errors as `publishError` + `Ack`, `cmd` keeps same exit-code but validation now in pipeline instead of per-handler.
