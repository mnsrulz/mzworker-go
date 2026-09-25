package handler

import (
	"context"
	"time"

	"github.com/mehdihadeli/go-mediatr"
)

type PingRequest struct{}

type PingResponse struct {
	Message    string `json:"message"`
	ServerTime string `json:"server_time"`
}

type PingHandler struct{}

func (r *PingRequest) Validate() error { return nil }

func init() {
	RegisterStruct[PingRequest, *PingResponse]("ping", func(_ Deps) mediatr.RequestHandler[*PingRequest, *PingResponse] {
		return &PingHandler{}
	})
}

func (h *PingHandler) Handle(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Message:    "pong",
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
