package query

import (
	"testing"
)

func TestValidateQuerySelect(t *testing.T) {
	err := ValidateQuery("SELECT * FROM 'data/options_data/symbol=AMZN/*.parquet'")
	if err != nil {
		t.Errorf("Expected no error for SELECT query, got: %v", err)
	}
}

func TestValidateQueryInsertRejected(t *testing.T) {
	err := ValidateQuery("INSERT INTO table VALUES (1)")
	if err == nil {
		t.Error("Expected error for INSERT query")
	}
	if err.(*QueryError).Code != "NOT_SELECT" {
		t.Errorf("Expected NOT_SELECT error code, got: %s", err.(*QueryError).Code)
	}
}

func TestValidateQueryDropRejected(t *testing.T) {
	err := ValidateQuery("DROP TABLE users")
	if err == nil {
		t.Error("Expected error for DROP query")
	}
	if err.(*QueryError).Code != "NOT_SELECT" {
		t.Errorf("Expected NOT_SELECT error code, got: %s", err.(*QueryError).Code)
	}
}

func TestValidateQueryForbiddenPath(t *testing.T) {
	err := ValidateQuery("SELECT * FROM '/etc/passwd'")
	if err == nil {
		t.Error("Expected error for forbidden path")
	}
	if err.(*QueryError).Code != "PATH_NOT_ALLOWED" {
		t.Errorf("Expected PATH_NOT_ALLOWED error code, got: %s", err.(*QueryError).Code)
	}
}

func TestEnforceLimitDefault(t *testing.T) {
	if EnforceLimit(0) != DefaultLimit {
		t.Errorf("Expected default limit for 0, got: %d", EnforceLimit(0))
	}
	if EnforceLimit(-1) != DefaultLimit {
		t.Errorf("Expected default limit for -1, got: %d", EnforceLimit(-1))
	}
}

func TestEnforceLimitCustom(t *testing.T) {
	if EnforceLimit(100) != 100 {
		t.Errorf("Expected 100, got: %d", EnforceLimit(100))
	}
}

func TestEnforceLimitMaxCap(t *testing.T) {
	if EnforceLimit(50000) != MaxLimit {
		t.Errorf("Expected max limit, got: %d", EnforceLimit(50000))
	}
}
