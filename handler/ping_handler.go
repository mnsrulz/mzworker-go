package handler

import (
	"context"
	"time"
)

type PingRequest struct{}

type PingResponse struct {
	Message    string `json:"message"`
	ServerTime string `json:"server_time"`
}

type PingHandler struct{}

func (h *PingHandler) Handle(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Message:    "pong",
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
