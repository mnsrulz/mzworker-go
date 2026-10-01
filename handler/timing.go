package handler

import (
	"context"
	"log"
	"time"

	"github.com/mehdihadeli/go-mediatr"
)

type TimingBehavior struct{}

func (t *TimingBehavior) Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error) {
	start := time.Now()
	resp, err := next(ctx)
	if err != nil {
		log.Printf("Pipeline: %T failed after %s: %v", request, time.Since(start), err)
	} else {
		log.Printf("Pipeline: %T completed in %s", request, time.Since(start))
	}
	return resp, err
}
