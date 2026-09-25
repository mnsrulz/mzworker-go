## MODIFIED Requirements

### Requirement: Request validation at handler entry
The system SHALL validate incoming requests using Zog via a shared `ValidationBehavior` pipeline that runs before any `Handle` for both `cmd` and `amqp`, instead of duplicated per-handler `Handle` validation.

#### Scenario: Valid request passes pipeline
- **WHEN** a `*VolatilityQuery` with `Symbol:"AAPL" LookbackDays:30 Mode:"atm"` is dispatched via `Dispatch(ctx, any)` from either transport
- **THEN** `ValidationBehavior.Handle` asserts `Validatable` and calls `Validate()` (delegating to `VolatilityQuerySchema`), returns `next(ctx)`, and the handler executes business logic

#### Scenario: Invalid request fails in pipeline before handler
- **WHEN** a request violates its Zog schema (e.g., `Symbol=""` or missing required `Mode`)
- **THEN** `ValidationBehavior` returns `fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))` without calling `Handle`; `cmd` `*RunE` returns `fmt.Errorf("... failed: %w", err)` to `cobra` (stderr exit 1, no `Output`), `amqp` `handleMessage` calls `publishError` + `Ack` with `{error: "..."}` envelope

#### Scenario: Stateless handler bypasses validation
- **WHEN** `*PingRequest{}` (no fields) is dispatched
- **THEN** its `Validate() error` returns nil and pipeline proceeds to `PingHandler.Handle` which returns `pong`

