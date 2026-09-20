# mzworker-go

CLI tool for querying DuckDB with SQL. Callable from opencode MCP/skills.

## Usage

```bash
# Execute a query (default JSON output)
mzworker-go "SELECT * FROM trades LIMIT 10"

# Explicit query subcommand
mzworker-go query "SELECT * FROM trades LIMIT 10"

# Table output
mzworker-go "SELECT * FROM trades LIMIT 10" --format table

# Raw output (nested columns/rows)
mzworker-go "SELECT * FROM trades LIMIT 10" --format raw

# Custom row limit
mzworker-go "SELECT * FROM trades" --limit 500

# Pipe SQL from stdin
echo "SELECT 42 AS answer" | mzworker-go

# AMQP daemon mode
mzworker-go serve
```

## Build

```bash
make build
# Binary at bin/mzworker
```

## Development

```bash
make test
make lint
```

## Environment Variables

- `AMQP_URL` — AMQP broker URL (for `serve` mode)
- `AMQP_REQUEST_QUEUE` — AMQP queue name (for `serve` mode)
