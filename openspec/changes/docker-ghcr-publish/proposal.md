## Why

The CI pipeline builds native binaries for multiple platforms and publishes GitHub Releases, but there is no containerized distribution. Docker images on GHCR enable simpler deployment, especially for server environments where the `serve` command runs as a daemon. Two image variants are needed: a minimal Go-only image and a full image with rclone for mounting remote storage.

## What Changes

- New `Dockerfile` — minimal distroless image containing only the pre-built linux-amd64 Go binary
- New `Dockerfile.rclone` — rclone base image with the Go binary and an entrypoint that mounts remote storage via rclone before starting the worker
- New `entrypoint.sh` — mounts rclone remote to `/data`, then runs `mzworker serve`
- New `.dockerignore` — excludes build artifacts, .git, and openspec from Docker context
- Modified `.github/workflows/ci.yml` — new `docker-publish` job builds and pushes both images to GHCR on main branch pushes

## Capabilities

### New Capabilities

- `docker-build`: Dockerfiles for two image variants (Go-only and Go+rclone), built from pre-built linux-amd64 binary
- `ghcr-publish`: GitHub Actions workflow job that logs into GHCR, builds images, and pushes with version tags

### Modified Capabilities

_None. No existing spec-level behavior changes._

## Impact

- New files: `Dockerfile`, `Dockerfile.rclone`, `entrypoint.sh`, `.dockerignore`
- Modified file: `.github/workflows/ci.yml` (new job + `packages: write` permission)
- GHCR packages created: `ghcr.io/mnsrulz/mzworker-go` and `ghcr.io/mnsrulz/mzworker-go-rclone`
- No changes to Go code or existing CI behavior
