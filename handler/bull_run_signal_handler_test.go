package handler

import (
	"context"
	"strings"
	"testing"
)

func TestBullRunSignalValidateInvalid(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
		// Init may have been called already; ignore duplicate behavior registration
	}
	q := &BullRunSignalQuery{LookbackDays: 30}
	if err := q.Validate(); err == nil {
		t.Fatal("expected validation error for missing Symbol")
	}
}

func TestBullRunSignalValidateNonPositiveLookback(t *testing.T) {
	q := &BullRunSignalQuery{Symbol: "IBIT", LookbackDays: 0}
	if err := q.Validate(); err == nil {
		t.Fatal("expected validation error for non-positive LookbackDays")
	}
}

func TestBullRunSignalValidateValid(t *testing.T) {
	q := &BullRunSignalQuery{Symbol: "IBIT", LookbackDays: 730}
	if err := q.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestBullRunSignalRequestTypeRegistered(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
		// ignore if already initialized
	}
	factory, ok := GetRequestFactory("bull-run-signal-query")
	if !ok {
		t.Fatal("request type bull-run-signal-query not registered")
	}
	if factory == nil {
		t.Fatal("expected non-nil request factory")
	}
}

func TestDispatchInvalidBullRunSignalViaPipeline(t *testing.T) {
	if err := Init("/tmp/test-data"); err != nil {
		// ignore if already initialized
	}
	ctx := context.Background()
	invalid := &BullRunSignalQuery{Symbol: "", LookbackDays: 0}
	_, err := Dispatch(ctx, invalid)
	if err == nil {
		t.Fatal("expected dispatch validation error")
	}
	if err != nil && len(err.Error()) == 0 {
		t.Fatal("expected non-empty error")
	}
}

func TestBullRunSeriesSQLSanity(t *testing.T) {
	if !strings.Contains(bullRunSeriesSQL, "FROM dataset") {
		t.Error("series SQL must consume the scaffold-provided dataset CTE")
	}
	if !strings.Contains(bullRunSeriesSQL, "rule_pass") {
		t.Error("series SQL must expose rule_pass")
	}
	if !strings.Contains(bullRunSeriesSQL, "WHERE wk >= (current_date - %d)") {
		t.Error("series SQL must window by lookback days via current_date")
	}
	if strings.Contains(bullRunSeriesSQL, "2025-01-01") {
		t.Error("series SQL must not contain a hardcoded start date")
	}
	if !strings.Contains(bullRunSeriesSQL, "ORDER BY wk") {
		t.Error("series SQL must order by week")
	}
}
