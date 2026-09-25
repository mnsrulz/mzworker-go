## Context

`amqp/consumer.go:112` `handleMessage` parses `{requestType}` envelope, looks up `handler.GetRequestFactory:125` (`handler/registry.go:15`), calls `factory(msg.Body):132` to get `any` typed request, then dispatches via `dispatchMediatr:142` type-switch to `mediatr.Send[TReq,TResp]`. `cmd/root.go:112` registers handlers explicitly via `mediatr.RegisterRequestHandler` per handler (`handler.NewQueryHandler(newInternalQueryExecutor)` etc.), with `ValidationBehavior:113` that only logs (`handler/validation.go:12`). Each `handler/*_handler.go:Handle` then re-validates with `zog` (`query_handler.go:21`, `ohlc_handler.go`, `volatility_handler.go:28`, `options_stat_handler.go`, `expected_move_handler.go`). Adding a handler touches both branches.

Stakeholders: CLI (cobra `cmd/*.go`) and AMQP daemon (`cmd/serve.go:33`). Both must share one dispatch and one validation.

## Goals / Non-Goals

**Goals:**
- Single registry `requestType -> {RequestFactory, dispatchAny}` where `dispatchAny func(ctx, any)(any,error)` captures typed `mediatr.Send` at registration, so receive-time code deduces handler from request type + data as `any` and pipeline handles it.
- Handlers remain source of truth (`Handle` + `zog` schemas in `handler/types.go:45`).
- Shared `ValidationBehavior` fail-closed for both transports (cmd stderr exit 1, amqp `{error:...}` + Ack).
- Executor opt-in: `PingHandler` (`handler/ping_handler.go`) no dep, 5 SQL handlers need `QueryExecutor` (`handler/query_handler.go:10`).

**Non-Goals:**
- Change external APIs (AMQP JSON envelope shape, cobra flags, `amqpValue` wire `consumer.go:36`).
- Change DuckDB CTE scaffolding (`query/internal.go:26`).
- Introduce new query types or new deps beyond `QueryExecutor`.

## Decisions

### 1. Capture typed `mediatr.Send` in closure at registration vs reflect call at dispatch
*Rationale:* `mediatr.Send` is generic; receive-time only has `any`. Storing `func(ctx,any)(any,error){return mediatr.Send[*Q,*Resp](ctx, req.(*Q))}` at `Register` in `handler/registry.go` avoids runtime reflect and keeps type-safety. Alternative reflect `MethodByName` over mediatr internals is slower and brittle against `go-mediatr` version bump. Dispatch map keyed by `requestType` string, not `reflect.Type`, so lookup is exact same envelope value.

### 2. Validation in pipeline, not per-handler
*Rationale:* Pipeline is the only shared seam both branches pass through. Add `Validatable` interface (`Validate() error`) on each request type delegating to its `zog` schema (`DynamicSQLQuerySchema`, `VolatilityQuerySchema` etc.). `ValidationBehavior.Handle` type-asserts `Validatable` and returns error before `next(ctx)`. Per-handler `Handle` validation is removed to avoid duplication; fallback type-switch for legacy handlers without interface. Alternative keeping per-handler validation defeats sharing goal.

### 3. Unified bootstrap `handler.Init(dataDir)` vs per-command `registerMediatr`
*Rationale:* `cmd/root.go:112` called per `*RunE` and `serve.go:33`; amqp path also needs it. Single `Init` registers pipeline behavior once + iterates registry to create handlers (factories with `Deps{Executor}` for SQL handlers, zero-arg for ping) + `mediatr.RegisterRequestHandler`. Uses existing `newInternalQueryExecutor(dataDir)` (`cmd/root.go:137`).

### 4. Executor opt-in via `Deps` struct
*Rationale:* Not all handlers need `QueryExecutor`. `Deps{Executor QueryExecutor}` passed only to factories that declare `func(Deps)any`. `PingHandler` factory ignores it. Keeps testing seam (mock `QueryExecutor` func) without global `handlers.go` mutexes.

### 5. Keep `amqpValue` wire, merge dispatch only
*Rationale:* Changing AMQP response envelope (`amqp/consumer.go:22-41` `toAmqpResponse`) is breaking. Convergence touches only dispatch path, not wire format.

## Risks / Trade-offs

- `mediatr` uses package-level state → `handler.Init` must run before any `Dispatch` (both `cmd` and `serve`). Mitigate by calling in `RootCmd.PersistentPreRunE`.
- `dispatchAny` closure must be added per handler at registration → missed entry = unknown `requestType` -> `publishError` + `Ack` (existing behavior `consumer.go:128`) unchanged.
- `zog` schemas use PascalCase keys (`DynamicSQLQuerySchema` `handler/types.go:45`) vs json snake_case; `Validatable` must unwrap with same naming. Mitigate by reusing existing schemas.
- `dispatchMediatr` type-switch (`consumer.go:186`) removal is breaking if external code imports it → not exported today, safe.

## Migration Plan

1. Enhance `handler/registry.go` and `handler/validation.go`, add `handler/mediatr_bridge.go` + `handler/bootstrap.go` (no callers yet).
2. Update each `handler/*_handler.go` `init()` to `Register`.
3. Switch `amqp/consumer.go:142` to `registry.Dispatch`, `cmd/root.go` to `handler.Init`.
4. Delete per-handler `Handle` validation after pipeline covers it; verify `go build ./...`/`go test ./...` and `amqp/consumer_test.go:10` ping case.
5. Rollback: revert `registry.go` and re-add `dispatchMediatr` switch; `mediatr` package state resets on restart.

## Open Questions

- Keep per-handler `Validate` calls as defense-in-depth or delete after pipeline is sole validator?
- Should `handler.Init` take `context.Context` for future scoped mediator?

