# pipeline-timing-log

## Purpose
End-to-end request duration logging via a mediatr pipeline behavior: one timing line per request covering validation and the handler, identical for the AMQP and CLI transports.

## Requirements

### Requirement: Request timing log
The mediatr request pipeline SHALL include a timing behavior that measures each request end-to-end — including validation and the handler body — and logs exactly one line with the request's Go type and the elapsed duration. The behavior SHALL pass the response or error through unchanged.

#### Scenario: Successful request
- **WHEN** a request completes successfully through the pipeline
- **THEN** one line `Pipeline: <request type> completed in <duration>` is logged (e.g. `Pipeline: *handler.ExpectedMoveQuery completed in 12.4ms`) and the response passes through unchanged

#### Scenario: Failed request
- **WHEN** the pipeline returns an error (from validation or from the handler)
- **THEN** one line `Pipeline: <request type> failed after <duration>: <error>` is logged and the error passes through unchanged

#### Scenario: Timing wraps validation
- **WHEN** the pipeline behaviors are registered during `handler.Init`
- **THEN** the timing behavior is registered before the validation behavior, so the measured duration includes request validation

#### Scenario: Both transports
- **WHEN** a request is dispatched via the AMQP consumer or executed via the CLI
- **THEN** the same timing line is logged for both, since both go through `mediatr.Send`
