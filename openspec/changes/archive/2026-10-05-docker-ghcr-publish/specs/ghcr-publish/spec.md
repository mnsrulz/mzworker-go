## ADDED Requirements

### Requirement: GHCR login

The docker-publish job SHALL authenticate to GitHub Container Registry using `GITHUB_TOKEN`.

#### Scenario: Login to GHCR

- **WHEN** the docker-publish job runs
- **THEN** it SHALL use `docker/login-action` with `registry: ghcr.io`, `username: ${{ github.actor }}`, and `password: ${{ secrets.GITHUB_TOKEN }}`

### Requirement: Build and push Go-only image

The job SHALL build the Go-only Docker image and push it to GHCR with version tags.

#### Scenario: Push image on main branch

- **WHEN** code is pushed to the `main` branch
- **THEN** the job SHALL build `Dockerfile` and push to `ghcr.io/${{ github.repository }}` with tags `latest` and the computed version string

#### Scenario: Skip on pull requests

- **WHEN** the workflow is triggered by a pull request
- **THEN** the docker-publish job SHALL not run

### Requirement: Build and push rclone image

The job SHALL build the rclone Docker image and push it to GHCR with version tags.

#### Scenario: Push rclone image on main branch

- **WHEN** code is pushed to the `main` branch
- **THEN** the job SHALL build `Dockerfile.rclone` and push to `ghcr.io/${{ github.repository }}-rclone` with tags `latest` and the computed version string

### Requirement: Job dependencies

The docker-publish job SHALL depend on the version and build jobs.

#### Scenario: Run after build completes

- **WHEN** the CI workflow runs
- **THEN** docker-publish SHALL only start after both `version` and `build` jobs complete successfully

### Requirement: Version tagging

Images SHALL be tagged using the same version string computed by the version job.

#### Scenario: Version tag format

- **WHEN** the version job outputs `v1.2.3`
- **THEN** images SHALL be tagged with both `ghcr.io/mnsrulz/mzworker-go:v1.2.3` and `ghcr.io/mnsrulz/mzworker-go:latest`

### Requirement: Packages write permission

The workflow SHALL request `packages: write` permission to push images to GHCR.

#### Scenario: Permission granted

- **WHEN** the workflow runs
- **THEN** the `packages: write` permission SHALL be included in the top-level permissions block
