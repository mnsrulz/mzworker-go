package main

import (
	"testing"
)

func TestValidateQuerySelect(t *testing.T) {
	err := validateQuery("SELECT * FROM 'data/options_data/symbol=AMZN/*.parquet'")
	if err != nil {
		t.Errorf("Expected no error for SELECT query, got: %v", err)
	}
}

func TestValidateQueryInsertRejected(t *testing.T) {
	err := validateQuery("INSERT INTO table VALUES (1)")
	if err == nil {
		t.Error("Expected error for INSERT query")
	}
	if err.(*QueryError).Code != "NOT_SELECT" {
		t.Errorf("Expected NOT_SELECT error code, got: %s", err.(*QueryError).Code)
	}
}

func TestValidateQueryDropRejected(t *testing.T) {
	err := validateQuery("DROP TABLE users")
	if err == nil {
		t.Error("Expected error for DROP query")
	}
	if err.(*QueryError).Code != "NOT_SELECT" {
		t.Errorf("Expected NOT_SELECT error code, got: %s", err.(*QueryError).Code)
	}
}

func TestValidateQueryForbiddenPath(t *testing.T) {
	err := validateQuery("SELECT * FROM '/etc/passwd'")
	if err == nil {
		t.Error("Expected error for forbidden path")
	}
	if err.(*QueryError).Code != "PATH_NOT_ALLOWED" {
		t.Errorf("Expected PATH_NOT_ALLOWED error code, got: %s", err.(*QueryError).Code)
	}
}

func TestEnforceLimitDefault(t *testing.T) {
	if enforceLimit(0) != defaultLimit {
		t.Errorf("Expected default limit for 0, got: %d", enforceLimit(0))
	}
	if enforceLimit(-1) != defaultLimit {
		t.Errorf("Expected default limit for -1, got: %d", enforceLimit(-1))
	}
}

func TestEnforceLimitCustom(t *testing.T) {
	if enforceLimit(100) != 100 {
		t.Errorf("Expected 100, got: %d", enforceLimit(100))
	}
}

func TestEnforceLimitMaxCap(t *testing.T) {
	if enforceLimit(50000) != maxLimit {
		t.Errorf("Expected max limit, got: %d", enforceLimit(50000))
	}
}
