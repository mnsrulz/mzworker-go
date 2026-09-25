## ADDED Requirements

### Requirement: Single registry deduces handler from request type and data as any
The system SHALL maintain a single registry mapping `requestType` string to `{RequestFactory func(payload []byte)(any,error), dispatchAny func(ctx context.Context, req any)(any,error)}` where `dispatchAny` is a closure capturing the typed `mediatr.Send[TReq,TResp]` at registration time, so callers that only have `requestType` + `payload` as `any` can deduce the handler without knowing generics.

#### Scenario: Handler deduced from request type at receive time
- **WHEN** an AMQP message with `{"requestType":"volatility-query", ...}` arrives or a CLI command parses flags into JSON payload with `requestType "volatility-query"`
- **THEN** the registry `GetRequestFactory` creates `*VolatilityQuery` via `jsonRequestFactory` and `Dispatch(ctx, any)` routes to `VolatilityHandler` via the stored `dispatchAny` closure without caller specifying generic types

#### Scenario: Unknown request type handled uniformly
- **WHEN** `GetRequestFactory("unknown-type")` is called from either `cmd` or `amqp`
- **THEN** it returns `ok=false` and the caller logs and `Ack`s (AMQP: `amqp/consumer.go:128` style) or returns `fmt.Errorf` to `cobra` (CLI), with no handler invoked

#### Scenario: Registry owns both JSON and dispatch paths
- **WHEN** a new handler is added via `Register(requestType, factory, handlerFactory)` in its `handler/*_handler.go` init
- **THEN** both `GetRequestFactory` and `Dispatch` become available for that `requestType` without touching `cmd/root.go` or `amqp/consumer.go`
