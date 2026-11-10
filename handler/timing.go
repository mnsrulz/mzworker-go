package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mehdihadeli/go-mediatr"
)

type TimingBehavior struct{}

func (t *TimingBehavior) Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error) {
	start := time.Now()
	resp, err := next(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Pipeline: %T failed after %s: %v", request, time.Since(start), err))
	} else {
		slog.Info(fmt.Sprintf("Pipeline: %T completed in %s", request, time.Since(start)))
	}
	return resp, err
}
