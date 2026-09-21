## Context

The project has a CI pipeline (`ci.yml`) that builds linux-amd64 binaries with CGO enabled (required by DuckDB). The binary is self-contained (DuckDB statically linked). The existing versioning scheme computes a version string from git tags and uses it for release artifacts and GitHub Releases.

The deno-rclone pattern from `mediacatalog-infra` demonstrates mounting remote storage via rclone before running an application. This same pattern is needed for the mzworker-go rclone variant.

## Goals / Non-Goals

**Goals:**
- Publish two Docker images to GHCR: a minimal Go-only image and an rclone-enabled image
- Reuse the pre-built linux-amd64 binary from CI (no multi-stage build)
- Use the existing versioning scheme for image tags
- Keep the docker-publish job integrated in the existing CI workflow
- Follow the rclone entrypoint pattern from mediacatalog-infra/deno-rclone

**Non-Goals:**
- Multi-arch builds (amd64 only for now)
- Building the Go binary inside Docker (use pre-built artifact)
- Modifying existing CI jobs or Go code

## Decisions

### Use pre-built binary instead of multi-stage build

The CI already builds a linux-amd64 binary with CGO_ENABLED=1. Copying this binary into a slim runtime image avoids rebuilding inside Docker, resulting in faster builds and smaller images. The trade-off is that `docker build` requires the binary to exist locally, but this is acceptable since the CI workflow handles it.

### Distroless base for Go-only image

`gcr.io/distroless/static-debian12` is the smallest possible runtime image with no package manager, shell, or unnecessary utilities. Since the mzworker binary is statically linked, this works without any additional dependencies.

### Rclone base image for rclone variant

`rclone/rclone:v1.72-stable` already includes rclone and its dependencies (fuse3, ca-certificates). The Dockerfile only needs to add the Go binary and entrypoint on top.

### Entrypoint pattern from deno-rclone

The entrypoint follows the same structure: create required directories, start rclone mount in background, wait for mount point, then run the application. The rclone remote name is configurable via `RCLONE_REMOTE` env var (unlike the hardcoded name in deno-rclone).

### Two separate Dockerfiles

Using `Dockerfile` and `Dockerfile.rclone` (with `docker build -f`) is simpler than a single parameterized Dockerfile. Each file is self-contained and easy to understand.

## Risks / Trade-offs

- **Pre-built binary required** → Mitigated by CI producing the artifact before the docker-publish job runs
- **No local `docker build` without binary** → Developers can run `go build` locally first; documented in tasks
- **amd64 only** → Can add arm64 later by extending the CI matrix and using buildx
