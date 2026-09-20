## Context

The mzworker-go binary currently runs as a gRPC server with an optional AMQP consumer. The gRPC transport requires protoc, code generators, and proto definitions — tooling that adds build complexity for a transport that isn't used in the target use case. The project needs to be callable from opencode MCP/skills via shell invocation.

Current state:
- `main.go`: starts gRPC server + optionally AMQP consumer
- `server.go`: gRPC service adapter (converts proto ↔ handler types)
- `amqp.go`: AMQP consumer with request-reply pattern
- `handler/`: transport-agnostic handlers, types, registry
- `query.go`: DuckDB execution engine

## Goals / Non-Goals

**Goals:**
- Drop gRPC server and all proto/protobuf dependencies
- CLI accepts raw SQL via argument or stdin, executes against DuckDB, outputs results
- Three output formats: `json` (flat objects), `table` (ASCII), `raw` (nested for AMQP compat)
- Subcommand routing: bare SQL → CLI, `serve` → AMQP daemon
- Binary is self-contained — no external tools needed to run queries

**Non-Goals:**
- AMQP consumer changes (stays as-is)
- Handler package changes (stays transport-agnostic)
- Query construction helpers (out of scope, future work)
- MCP server implementation (this binary is called BY MCP, not an MCP server itself)

## Decisions

### 1. urfave/cli v3 for CLI framework

**Decision**: Use `github.com/urfave/cli/v3` for subcommand routing, flag parsing, and help generation.

**Rationale**: Zero external dependencies (stdlib only), clean declarative API, auto-generated help and shell completions. Handles our simple case (1 subcommand, 2 flags) without overkill.

**Alternative considered**: stdlib `flag` — works but requires manual subcommand routing and help text. Cobra — heavier (pflag dependency) and overkill for this complexity.

### 2. Subcommand routing with bare SQL default

**Decision**: `mzworker-go <SQL>` works the same as `mzworker-go query <SQL>`. The `serve` subcommand starts the AMQP consumer.

**Rationale**: Minimizes typing for the common case. urfave/cli handles this via a default action that checks if the first arg is a known subcommand.

### 3. Flat JSON output as default

**Decision**: Default `--format json` outputs `[{"col": "val"}, ...]` instead of `{"columns": [...], "rows": [[...]]}`.

**Rationale**: Flat objects are natural for JSON parsing in any language. MCP/skills can directly use the output without restructuring. The `raw` format preserves backward compat.

### 4. New `cli.go` file, not modifying `amqp.go`

**Decision**: CLI logic lives in a new `cli.go` file. `amqp.go` is untouched.

**Rationale**: Separation of concerns. CLI has different I/O patterns (stdout, flag parsing, stdin) than AMQP (queue consume, publish).

### 5. Read SQL from stdin when no argument given

**Decision**: `echo "SELECT 1" | mzworker-go query` reads from stdin.

**Rationale**: Enables piping, which is essential for MCP/skills integration where the SQL might be dynamically constructed.

## Risks / Trade-offs

- **[Breaking change]** Any existing gRPC clients will break. → Mitigation: gRPC wasn't used in production; this is a development/internal tool.
- **[Backward compat]** AMQP wire format changes. → Mitigation: AMQP path is unchanged; `raw` format preserves the nested structure.
- **[Stdin blocking]** If no SQL arg and no stdin, program hangs. → Mitigation: detect if stdin is a terminal using `golang.org/x/term` or `os.Stdin.Stat()` and print usage.
