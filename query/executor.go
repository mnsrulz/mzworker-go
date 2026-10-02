package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/duckdb/duckdb-go/v2"
)

const (
	DefaultLimit = 1000
	MaxLimit     = 10000
)

var forbiddenPatterns = []string{"/etc/", "/proc/", "/sys/", "/dev/"}

type QueryError struct {
	Code    string
	Message string
}

func (e *QueryError) Error() string {
	return e.Message
}

func ValidateQuery(sqlStr string) error {
	upper := strings.ToUpper(strings.TrimSpace(sqlStr))

	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") {
		return &QueryError{Code: "NOT_SELECT", Message: "Only SELECT queries are allowed"}
	}

	lower := strings.ToLower(sqlStr)
	for _, pattern := range forbiddenPatterns {
		if strings.Contains(lower, pattern) {
			return &QueryError{Code: "PATH_NOT_ALLOWED", Message: fmt.Sprintf("Path not allowed: %s", pattern)}
		}
	}

	return nil
}

func EnforceLimit(requestLimit int32) int32 {
	if requestLimit <= 0 {
		return DefaultLimit
	}
	if requestLimit > MaxLimit {
		return MaxLimit
	}
	return requestLimit
}

func AppendLimit(sqlStr string, limit int32) string {
	if strings.Contains(strings.ToLower(sqlStr), "limit") {
		return sqlStr
	}
	return fmt.Sprintf("%s LIMIT %d", sqlStr, limit)
}

func ExecuteQuery(ctx context.Context, sqlStr string, limit int32) ([]string, [][]interface{}, error) {
	if err := ValidateQuery(sqlStr); err != nil {
		return nil, nil, err
	}

	sqlWithLimit := AppendLimit(sqlStr, EnforceLimit(limit))

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Failed to open DuckDB: %v", err)}
	}
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, sqlWithLimit)
	if err != nil {
		return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Query execution failed: %v", err)}
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Failed to get columns: %v", err)}
	}

	var resultRows [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Failed to scan row: %v", err)}
		}

		row := make([]interface{}, len(columns))
		for i, v := range values {
			switch val := v.(type) {
			case []byte:
				row[i] = string(val)
			case time.Time:
				row[i] = val.Format("2006-01-02")
			case duckdb.Decimal:
				row[i] = val.Float64()
			default:
				row[i] = val
			}
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Row iteration error: %v", err)}
	}

	return columns, resultRows, nil
}
