## 1. Behavior and registration

- [x] 1.1 Add `handler/timing.go` with `TimingBehavior`: measure around `next`, log `Pipeline: %T completed in %s` (success) or `Pipeline: %T failed after %s: %v` (error), return result unchanged
- [x] 1.2 Register `&TimingBehavior{}` before `&ValidationBehavior{}` in `handler.Init` (single call, timing outermost)

## 2. Tests and verification

- [x] 2.1 Unit test: success and failure passthrough (response/error unchanged) with the expected log line captured via `log.SetOutput`
- [x] 2.2 Run `go build ./...`, `go vet ./...`, `go test ./... -count=1 -race`
