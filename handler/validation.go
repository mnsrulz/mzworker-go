package handler

import (
	"context"
	"log"

	"github.com/mehdihadeli/go-mediatr"
)

type Validatable interface {
	Validate() error
}

type ValidationBehavior struct{}

func (v *ValidationBehavior) Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error) {
	log.Printf("Pipeline: validating request %T", request)
	if validatable, ok := request.(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			return nil, err
		}
	}
	return next(ctx)
}
