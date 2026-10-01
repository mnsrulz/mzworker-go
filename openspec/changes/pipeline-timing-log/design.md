## Context

`handler/validation.go` already establishes the pattern: a mediatr `PipelineBehavior` (`ValidationBehavior`) registered in `handler.Init`. go-mediatr v1.4.0 documents that behaviors execute in registration order (first registered runs first / is outermost) and rejects duplicate behavior types. Both transports (AMQP `handleMessage` → `handler.Dispatch`, and `cmd` → `handler.Dispatch`) funnel through `mediatr.Send`.

## Goals / Non-Goals

**Goals:**
- One duration log line per request, covering validation + handler, identical for AMQP and CLI.
- Zero configuration and zero new dependencies.

**Non-Goals:**
- Metrics backends, histograms, or per-phase (validation vs query) breakdowns.
- Structured/JSON logging or log sampling.
- Replacing the existing `Pipeline: validating request` or `AMQP: response sent` lines.

## Decisions

### D1: `TimingBehavior` in the `handler` package
A `TimingBehavior` with the same shape as `ValidationBehavior` (`Handle(ctx, request, next)`): record `time.Now()`, call `next`, log success (`completed in %s`) or failure (`failed after %s: %v`), return the result untouched. Placement in `handler` keeps all pipeline wiring next to its only sibling.

**Alternative considered (rejected):** timing inside the AMQP layer (`handleMessage`) — misses the CLI transport and measures dispatch overhead rather than the pipeline.

### D2: Registered first, in the same call as validation
`Init` calls `mediatr.RegisterRequestPipelineBehaviors(&TimingBehavior{}, &ValidationBehavior{})` — one call, timing first, so it is outermost and its duration includes validation. The existing duplicate-type guard in go-mediatr plus the `initialized` flag in `Init` prevent double registration.

### D3: Log format mirrors the existing `Pipeline:` prefix
`Pipeline: <type> completed in <duration>` / `Pipeline: <type> failed after <duration>: <err>` — greppable alongside the validation line, using the same `%T` request-type rendering.

## Risks / Trade-offs

- [One extra log line per request] → Acceptable at current volume; same order of magnitude as existing per-request lines.
- [Duration includes validation, not AMQP publish/ack overhead] → Intentional: the line measures what the caller waits for inside the pipeline; transport overhead stays in the AMQP lines.

## Migration Plan

No migration — additive logging only. Rollback = revert the registration line.

## Open Questions

- None.
