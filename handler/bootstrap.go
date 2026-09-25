package handler

import (
	"context"
	"fmt"
	"sync"

	"github.com/mehdihadeli/go-mediatr"

	"github.com/mnsrulz/mzworker-go/query"
)

var (
	bootstrapMu sync.Mutex
	initialized bool
)

// Init registers the ValidationBehavior once and instantiates all pending handlers
// with the provided dataDir. It is safe to call multiple times; subsequent calls are no-ops.
func Init(dataDir string) error {
	bootstrapMu.Lock()
	defer bootstrapMu.Unlock()
	if initialized {
		return nil
	}
	if err := mediatr.RegisterRequestPipelineBehaviors(&ValidationBehavior{}); err != nil {
		// go-mediatr returns error if already registered; treat as non-fatal if already initialized
		// but first call should succeed; ignore duplicate registration
		if initialized {
			// no-op
		} else {
			return fmt.Errorf("failed to register pipeline behavior: %w", err)
		}
	}
	d := Deps{Executor: newInternalQueryExecutor(dataDir)}
	if err := initRegistry(d); err != nil {
		return err
	}
	initialized = true
	return nil
}

func newInternalQueryExecutor(dataDir string) QueryExecutor {
	return func(ctx context.Context, symbol, sql string, limit int32) ([]string, [][]interface{}, error) {
		executor := query.NewInternalQueryExecutor(symbol, dataDir)
		return executor.Execute(ctx, sql, limit)
	}
}
