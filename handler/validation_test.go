package handler

import (
	"context"
	"testing"
)

func TestVolatilityValidateInvalid(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
		// Init may have been called already; ignore duplicate behavior registration
	}
	// invalid: missing Symbol, Mode
	q := &VolatilityQuery{LookbackDays: 30}
	if err := q.Validate(); err == nil {
		t.Fatal("expected validation error for missing Symbol/Mode")
	}
}

func TestVolatilityValidateValid(t *testing.T) {
	q := &VolatilityQuery{Symbol: "AAPL", LookbackDays: 30, Mode: "atm"}
	if err := q.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestDispatchInvalidVolatilityViaPipeline(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
		// ignore if already initialized
	}
	ctx := context.Background()
	invalid := &VolatilityQuery{Symbol: "", LookbackDays: 0, Mode: ""}
	_, err := Dispatch(ctx, invalid)
	if err == nil {
		t.Fatal("expected dispatch validation error")
	}
	if err != nil && len(err.Error()) == 0 {
		t.Fatal("expected non-empty error")
	}
}

func TestDispatchPingValid(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
	}
	ctx := context.Background()
	_, err := Dispatch(ctx, &PingRequest{})
	if err != nil {
		t.Fatalf("unexpected error dispatching ping: %v", err)
	}
}

func TestGetRequestFactoryAndDispatchByType(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
	}
	factory, ok := GetRequestFactory("ping")
	if !ok {
		t.Fatal("expected factory for ping")
	}
	payload := []byte(`{"requestType":"ping"}`)
	req, err := factory(payload)
	if err != nil {
		t.Fatalf("factory error: %v", err)
	}
	if _, ok := req.(*PingRequest); !ok {
		t.Fatalf("expected *PingRequest got %T", req)
	}
	resp, err := Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("dispatch error: %v", err)
	}
	if _, ok := resp.(*PingResponse); !ok {
		t.Fatalf("expected *PingResponse got %T", resp)
	}
}
