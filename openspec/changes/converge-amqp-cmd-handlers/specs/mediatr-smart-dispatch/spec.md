## ADDED Requirements

### Requirement: Smart mediatr dispatch handles any without caller generics
The system SHALL expose `Dispatch(ctx context.Context, req any)(any,error)` that routes `any` typed requests through the `mediatr` pipeline, with per-requestType dispatch closures registered via `mediatr.RegisterRequestHandler` at `handler.Init`.

#### Scenario: Any request dispatched via pipeline
- **WHEN** `Dispatch(ctx, *PingRequest)` or `Dispatch(ctx, *OHLCQuery)` is called from either `cmd` (after flag->JSON->factory) or `amqp` (after `handleMessage` factory)
- **THEN** it invokes the stored closure `mediatr.Send[*TReq,*TResp](ctx, req.(*TReq))` so the `ValidationBehavior` pipeline runs before `Handle`

#### Scenario: Single bootstrap for both transports
- **WHEN** `handler.Init(dataDir)` is called (from `RootCmd.PersistentPreRunE` and `cmd/serve.go`)
- **THEN** it registers `ValidationBehavior` once via `mediatr.RegisterRequestPipelineBehaviors` and iterates the unified registry to `mediatr.RegisterRequestHandler` for each handler, with `Deps{Executor}` injected only for SQL handlers (`Query`, `OHLC`, `Volatility`, `OptionsStat`, `ExpectedMove`) and zero-arg for `Ping`

#### Scenario: Remove duplicate dispatch switch
- **WHEN** the change is applied
- **THEN** `amqp/consumer.go:186` `dispatchMediatr` type-switch is deleted and replaced by `registry.Dispatch`; `cmd/*RunE` may keep `mediatr.Send` or also call `registry.Dispatch` — both hit same pipeline

