## ADDED Requirements

### Requirement: Go-only Docker image

The system SHALL provide a `Dockerfile` that produces a minimal container image containing only the pre-built linux-amd64 mzworker binary on a distroless base.

#### Scenario: Build Go-only image from pre-built binary

- **WHEN** a user runs `docker build -t mzworker .` with the `mzworker-linux-amd64` binary in the build context
- **THEN** the image SHALL contain the binary at `/usr/local/bin/mzworker` and the entrypoint SHALL be `mzworker`

#### Scenario: Run Go-only image

- **WHEN** a user runs the image with `DATA_DIR`, `AMQP_URL`, and `AMQP_REQUEST_QUEUE` environment variables set
- **THEN** the container SHALL start the AMQP consumer daemon

### Requirement: Rclone Docker image

The system SHALL provide a `Dockerfile.rclone` that produces a container image with the mzworker binary and rclone pre-installed on the rclone base image.

#### Scenario: Build rclone image from pre-built binary

- **WHEN** a user runs `docker build -t mzworker-rclone -f Dockerfile.rclone .` with the `mzworker-linux-amd64` binary in the build context
- **THEN** the image SHALL contain rclone, the mzworker binary at `/usr/local/bin/mzworker`, and the entrypoint script at `/entrypoint.sh`

#### Scenario: Run rclone image

- **WHEN** a user runs the image with `RCLONE_REMOTE`, `RCLONE_CONFIG`, `AMQP_URL`, and `AMQP_REQUEST_QUEUE` environment variables set and a valid rclone config mounted at `/rcloneconfig/rclone.conf`
- **THEN** the container SHALL mount the rclone remote to `/data` and start the AMQP consumer daemon with `DATA_DIR=/data`

### Requirement: Entrypoint script

The entrypoint script SHALL mount a rclone remote to `/data` using configurable env vars, wait for the mount to be ready, then execute `mzworker serve`.

#### Scenario: Rclone mount with env var configuration

- **WHEN** the entrypoint starts with `RCLONE_REMOTE=myremote` and `RCLONE_CONFIG=/rcloneconfig/rclone.conf`
- **THEN** it SHALL run `rclone mount myremote: /data --config /rcloneconfig/rclone.conf` in the background with cache and VFS options

#### Scenario: Wait for mount readiness

- **WHEN** the rclone mount is initiated
- **THEN** the entrypoint SHALL poll `/data` with `mountpoint -q` until the mount is ready before starting the worker

#### Scenario: Start worker after mount

- **WHEN** the rclone mount is confirmed ready
- **THEN** the entrypoint SHALL execute `mzworker serve`

### Requirement: Docker ignore file

A `.dockerignore` file SHALL exclude unnecessary files from the Docker build context.

#### Scenario: Build context exclusion

- **WHEN** a Docker build runs with the `.dockerignore` in place
- **THEN** `.git/`, `openspec/`, `*.exe`, and build artifacts SHALL be excluded from the build context
