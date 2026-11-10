package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mehdihadeli/go-mediatr"
)

type Validatable interface {
	Validate() error
}

type ValidationBehavior struct{}

func (v *ValidationBehavior) Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error) {
	slog.Info(fmt.Sprintf("Pipeline: validating request %T", request))
	if validatable, ok := request.(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			return nil, err
		}
	}
	return next(ctx)
}
