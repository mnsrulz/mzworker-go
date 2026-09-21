## 1. Docker Build Files

- [ ] 1.1 Create `.dockerignore` excluding `.git/`, `openspec/`, `*.exe`, and build artifacts
- [ ] 1.2 Create `Dockerfile` — distroless base, copy pre-built `mzworker-linux-amd64` binary to `/usr/local/bin/mzworker`, set entrypoint
- [ ] 1.3 Create `Dockerfile.rclone` — rclone base image, copy pre-built binary and entrypoint script, install fuse3 if needed
- [ ] 1.4 Create `entrypoint.sh` — mkdir `/data`, `/cache`, `/rcloneconfig`; rclone mount with `RCLONE_REMOTE` env var; wait for mount; run `mzworker serve`
- [ ] 1.5 Make `entrypoint.sh` executable

## 2. CI Workflow

- [ ] 2.1 Add `packages: write` to top-level permissions in `.github/workflows/ci.yml`
- [ ] 2.2 Add `docker-publish` job that depends on `version` and `build`
- [ ] 2.3 Add GHCR login step using `docker/login-action`
- [ ] 2.4 Add build-and-push step for Go-only image using `docker/build-push-action` with version + latest tags
- [ ] 2.5 Add build-and-push step for rclone image using `docker/build-push-action` with version + latest tags
- [ ] 2.6 Restrict docker-publish to `push` events only (skip PRs)

## 3. Verification

- [ ] 3.1 Verify Dockerfiles build locally with a dummy binary
- [ ] 3.2 Verify entrypoint.sh logic (rclone mount + wait + serve)
- [ ] 3.3 Review CI workflow YAML for correct job dependencies and permissions
