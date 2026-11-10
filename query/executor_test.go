package query

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func captureSlog(t *testing.T, level slog.Level, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})))
	defer slog.SetDefault(prev)
	fn()
	return buf.String()
}

func TestExecuteQueryLogsSQLAtDebug(t *testing.T) {
	out := captureSlog(t, slog.LevelDebug, func() {
		if _, _, err := ExecuteQuery(context.Background(), "SELECT 1 AS x", 0); err != nil {
			t.Fatalf("ExecuteQuery failed: %v", err)
		}
	})

	if !strings.Contains(out, `msg="executing query"`) {
		t.Fatalf("expected executed-SQL record, got %q", out)
	}
	if !strings.Contains(out, `sql="SELECT 1 AS x LIMIT 1000"`) {
		t.Fatalf("expected full SQL including limit, got %q", out)
	}
}

func TestExecuteQueryNoSQLAtInfo(t *testing.T) {
	out := captureSlog(t, slog.LevelInfo, func() {
		if _, _, err := ExecuteQuery(context.Background(), "SELECT 1 AS x", 0); err != nil {
			t.Fatalf("ExecuteQuery failed: %v", err)
		}
	})

	if strings.Contains(out, "executing query") {
		t.Fatalf("expected no executed-SQL record at info, got %q", out)
	}
}

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
