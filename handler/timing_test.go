package handler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/mehdihadeli/go-mediatr"
)

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)
	fn()
	return buf.String()
}

func TestTimingBehaviorLogsSuccessAndPassesResponse(t *testing.T) {
	var next mediatr.RequestHandlerFunc = func(ctx context.Context) (any, error) {
		return "pong", nil
	}

	var resp any
	var err error
	out := captureLogs(t, func() {
		resp, err = (&TimingBehavior{}).Handle(context.Background(), &PingRequest{}, next)
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp != "pong" {
		t.Fatalf("expected response passed through, got %v", resp)
	}
	if !strings.Contains(out, "Pipeline: *handler.PingRequest completed in ") {
		t.Fatalf("expected completion log line, got %q", out)
	}
}

func TestTimingBehaviorLogsFailureAndPassesError(t *testing.T) {
	wantErr := errors.New("boom")
	var next mediatr.RequestHandlerFunc = func(ctx context.Context) (any, error) {
		return nil, wantErr
	}

	var resp any
	var err error
	out := captureLogs(t, func() {
		resp, err = (&TimingBehavior{}).Handle(context.Background(), &PingRequest{}, next)
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error passed through, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	if !strings.Contains(out, "Pipeline: *handler.PingRequest failed after ") || !strings.Contains(out, "boom") {
		t.Fatalf("expected failure log line with error, got %q", out)
	}
}
