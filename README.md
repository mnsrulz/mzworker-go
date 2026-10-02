# mzworker-go

CLI tool for querying local options/OHLC market data with DuckDB, plus an AMQP request-reply worker. Callable from opencode MCP/skills.

## Usage

```bash
# Show available tables and columns
mzworker-go --schema

# Version
mzworker-go --version

# SQL query (--symbol and --data-dir are required)
mzworker-go query "SELECT * FROM dataset LIMIT 10" --symbol AAPL --data-dir /path/to/ODATA

# Output format: json (default), table, raw
mzworker-go query "SELECT * FROM T" -s AAPL --data-dir /path/to/ODATA --format table

# Custom row limit (default 1000)
mzworker-go query "SELECT * FROM dataset" -s AAPL --data-dir /path/to/ODATA --limit 500

# Pipe SQL from stdin (into the query subcommand)
echo "SELECT 42 AS answer" | mzworker-go query -s AAPL --data-dir /path/to/ODATA

# Market-data shortcuts (--symbol and --data-dir required on all)
mzworker-go ohlc          -s AAPL --data-dir /path/to/ODATA --lookback 30
mzworker-go options-stat  -s AAPL --data-dir /path/to/ODATA --lookback 30
mzworker-go volatility    -s AAPL --data-dir /path/to/ODATA --mode atm
mzworker-go expected-move -s AAPL --data-dir /path/to/ODATA --expiry-mode weekly

# AMQP daemon mode (requires env vars below)
mzworker-go serve --concurrency 5
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
# Binary at bin/mzworker-go, with version, git commit, and build time embedded
```

## Development

```bash
make test
make lint
```

## Environment Variables

Used by `serve` mode:

- `DATA_DIR` — data directory (required)
- `AMQP_URL` — AMQP broker URL (required)
- `AMQP_REQUEST_QUEUE` — AMQP queue name (required)

Interactive commands take the data directory via the `--data-dir` flag instead.
