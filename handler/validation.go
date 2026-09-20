package handler

import (
	"context"
	"log"

	"github.com/mehdihadeli/go-mediatr"
)

type ValidationBehavior struct{}

func (v *ValidationBehavior) Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error) {
	log.Printf("Pipeline: validating request %T", request)
	return next(ctx)
}
