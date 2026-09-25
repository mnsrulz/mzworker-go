## 1. Registry & Smart Dispatch

- [x] 1.1 Extend `handler/registry.go` to store `registryEntry{RequestFactory, dispatchAny func(ctx context.Context, req any)(any,error)}` keyed by `requestType`; add `Register(requestType string, factory RequestFactory, handlerFactory func(Deps) any)` that creates handler, calls `mediatr.RegisterRequestHandler`, and stores `dispatchAny` closure capturing typed `mediatr.Send[TReq,TResp]`
- [x] 1.2 Add `handler/mediatr_bridge.go` with `Dispatch(ctx context.Context, req any)(any,error)` and `DispatchByType(ctx, requestType string, payload []byte)(any,error)` helpers delegating to registry `dispatchAny` (used by `amqp` and optionally `cmd`)
- [x] 1.3 Add `handler/bootstrap.go` `Deps{Executor QueryExecutor}` and `Init(dataDir string) error` that registers `ValidationBehavior` once and iterates registry to instantiate handlers (SQL handlers with `Deps`, `Ping` zero-arg), plus `newInternalQueryExecutor` migration from `cmd/root.go:137`

## 2. Validation

- [x] 2.1 Enhance `handler/validation.go` `ValidationBehavior.Handle` to assert `Validatable` and call `Validate()` before `next(ctx)`; add `Validatable` interface
- [x] 2.2 Add `Validate() error` on each request type in `handler/types.go` and `handler/*_handler.go` delegating to its `zog` schema (`DynamicSQLQuerySchema`, `OHLCQuerySchema`, `VolatilityQuerySchema`, `OptionsStatQuerySchema`, `ExpectedMoveQuerySchema`); `PingRequest` returns nil

## 3. Handlers

- [x] 3.1 Update each `handler/*_handler.go` `init()` to `Register(requestType, jsonRequestFactory[T], handlerFactory)` (5 SQL handlers with `func(Deps)any{NewXHandler(deps.Executor)}`, ping with zero-arg); remove now-duplicate per-handler `Handle` validation after pipeline covers it (or keep as defense per design open question)
- [x] 3.2 Ensure `handler/query_handler.go:10` `QueryExecutor` type shared with `handler/bootstrap.go`

## 4. Transports

- [x] 4.1 Replace `cmd/root.go:112` `registerMediatr` per-handler `mediatr.RegisterRequestHandler` calls with `handler.Init`; wire `RootCmd.PersistentPreRunE` and `cmd/serve.go:33` to single bootstrap
- [x] 4.2 Replace `amqp/consumer.go:186` `dispatchMediatr` type-switch with `handler.Dispatch`/`registry.DispatchByType`; keep `toAmqpResponse:205` and `publishError:227` behavior (validation error -> `{error:...}` + `Ack`)

## 5. Verification

- [x] 5.1 `go build ./...` and `go vet ./...` with reverted baseline + new registry
- [x] 5.2 `go test ./...` including `amqp/consumer_test.go:10` ping case; add `handler/validation_test.go` for valid/invalid `VolatilityQuery` via `Dispatch`
- [x] 5.3 Manual: `go run . --help`, `go run .query --help` style CLI, and AMQP round-trip `{"requestType":"ping"}` and `{"requestType":"volatility-query", Symbol:""}` returns validation error

