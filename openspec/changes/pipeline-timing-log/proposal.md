## Why

There is no visibility into how long a request takes end-to-end. The current logs only show event timestamps (`Pipeline: validating request ...` then `AMQP: response sent ...`), so the actual processing duration — validation plus the DuckDB query — has to be eyeballed from two lines. A mediatr pipeline behavior gives one duration line per request for free, covering both transports.

## What Changes

- Add a `TimingBehavior` to the mediatr request pipeline that measures each request end-to-end (validation + handler) and logs `Pipeline: <request type> completed in <duration>` on success or `Pipeline: <request type> failed after <duration>: <error>` on failure.
- Register it **before** `ValidationBehavior` in `handler.Init`, so timing wraps validation (go-mediatr executes behaviors in registration order — first registered is outermost).
- Request/response pass through unchanged; no new configuration, dependencies, or metrics backends.

## Capabilities

### New Capabilities
- `pipeline-timing-log`: End-to-end request duration logging via a mediatr pipeline behavior, for both the AMQP and CLI transports.

### Modified Capabilities

(None — no existing requirement changes; the validation behavior keeps its own log line.)

## Impact

- `handler/timing.go` (new) — `TimingBehavior`.
- `handler/bootstrap.go` — registration order in `Init`.
- `handler/timing_test.go` (new) — passthrough + log-line assertions.
- No changes to AMQP dispatch, CLI commands, or handlers.
