## Context

The mzworker-go project is a single Go binary that exposes DuckDB query execution via gRPC and AMQP. Currently, both the gRPC server (`server.go`) and AMQP consumer (`amqp.go`) directly call `executeQuery` from `query.go`. There is no consistent dispatch pattern, no validation layer, and adding a new query type requires modifying `server.go`, `amqp.go`, and potentially `main.go`. The mztrading-data repo (TypeScript/Deno) uses a `handlers` map with Zod validation — this refactor brings the same architectural patterns to the Go codebase using `go-mediatr` and `zog`.

## Goals / Non-Goals

**Goals:**
- Introduce mediator-based request dispatch via `go-mediatr` so both gRPC and AMQP route through a single dispatch path.
- Add schema validation via `zog` at handler entry, matching the mztrading-data Zod pattern.
- Create a `handler/` package that encapsulates request types, response types, schemas, and handler logic.
- Establish a factory registry for mapping AMQP `requestType` strings to typed request structs.
- Add a `ValidationBehavior` pipeline for future cross-cutting concerns (logging, metrics, retry).

**Non-Goals:**
- Changing the external API contracts (gRPC proto, AMQP message format).
- Adding new query types in this change (only refactoring the existing dynamic SQL handler).
- Implementing authentication or authorization.
- Modifying `query.go` (shared DuckDB execution logic stays as-is).

## Decisions

### 1. Use `go-mediatr` for request dispatch

**Rationale**: go-mediatr provides typed request/response dispatch with pipeline behaviors out of the box. It matches the C# MediatR pattern the user is familiar with. Alternatives considered: custom dispatcher (less code but no middleware), godispatch (reflection-based, less control).

### 2. Use `zog` for schema validation

**Rationale**: Zog is a Zod-like validation library for Go with zero dependencies, `zjson` support, and method chaining API. It matches the Zod pattern used in mztrading-data. Alternatives considered: go-playground/validator (struct tags only, no JSON schema), custom validation (more code).

### 3. Zog schema keys use PascalCase struct field names

**Rationale**: Zog's parsing priority is `json` tag -> `zog` tag -> schema field name. Since we use `json` tags on structs for JSON mapping, schema keys must match struct field names (PascalCase), not JSON field names (snake_case).

### 4. Factory registry for AMQP dispatch

**Rationale**: The AMQP consumer receives raw JSON bytes + a `requestType` string. We need a registry mapping `requestType` -> factory function that creates the typed request from JSON. This keeps the type switch in one place and makes adding new request types a single registration call.

### 5. Validation inside each handler (not in pipeline)

**Rationale**: Each handler validates its own request type with its own Zog schema at the top of `Handle()`. The `ValidationBehavior` pipeline is a pass-through for now, ready for cross-cutting concerns later. This matches the mztrading-data pattern where Zod `.parse()` is called at the top of each handler.

### 6. Shared `QueryResponse` type

**Rationale**: All handlers return the same response shape (columns + rows). A single `QueryResponse` type avoids duplication and simplifies the mediatr generic parameters.

## Risks / Trade-offs

- **go-mediatr generics require Go 1.24+** → Project uses Go 1.27.1, so this is fine.
- **AMQP consumer needs type switch for factory registry** → The `dispatchMediatr` function must type-switch to call the correct generic `mediatr.Send`. This is unavoidable when bridging untyped JSON to typed generics. Mitigated by keeping the switch in one function.
- **go-mediatr uses package-level state** → All handler registrations happen at startup. This is fine for a single-binary application. If the project ever needs per-request scoped mediators, this would need revisiting.
- **Zog v0.x (not v1.0 yet)** → API is considered stable by the author. Minor breaking changes possible on minor version bumps.
