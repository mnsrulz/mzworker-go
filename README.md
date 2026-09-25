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

## Install

### Shell (macOS + Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/mnsrulz/mzworker-go/main/install.sh | bash
# specific version
curl -fsSL https://raw.githubusercontent.com/mnsrulz/mzworker-go/main/install.sh | bash -s -- --version v0.3.0
# custom prefix (no sudo)
./install.sh --version v0.3.0 --prefix ~/.local/bin
```

Downloads `mzworker-go-<os>-<arch>` from Releases (`linux-amd64`, `linux-arm64`, `macos-arm64` Apple Silicon) and installs to `/usr/local/bin/mzworker-go` (compat symlink `mzworker`).

### Go

```bash
go install github.com/mnsrulz/mzworker-go@latest
```

### Docker

```bash
docker pull ghcr.io/mnsrulz/mzworker-go:latest
docker pull ghcr.io/mnsrulz/mzworker-go-rclone:latest
```

## Build

```bash
make build
# Binary at bin/mzworker-go (compat symlink bin/mzworker)
```

## Development

```bash
make test
make lint
```

## Environment Variables

- `AMQP_URL` — AMQP broker URL (for `serve` mode)
- `AMQP_REQUEST_QUEUE` — AMQP queue name (for `serve` mode)
