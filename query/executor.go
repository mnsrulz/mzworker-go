package query

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

type QueryBuilder struct {
	ctes   []string
	symbol string
}

func NewQueryBuilder(symbol string) *QueryBuilder {
	return &QueryBuilder{
		symbol: strings.ToUpper(symbol),
	}
}

func (q *QueryBuilder) AddCTE(name, definition string) *QueryBuilder {
	q.ctes = append(q.ctes, fmt.Sprintf("%s AS (\n%s\n)", name, definition))
	return q
}

func (q *QueryBuilder) WithOHLCData(dataDir string) *QueryBuilder {
	if dataDir == "" {
		dataDir = os.Getenv("OHLC_DATA_DIR")
		if dataDir == "" {
			dataDir = "data/ohlc_data"
		}
	}

	q.AddCTE("ohlc", fmt.Sprintf(`SELECT dt, symbol, open, high, low, close, volume, iv30
FROM '%s/*.parquet'
WHERE symbol = '%s'`, dataDir, q.symbol))
	return q
}

func (q *QueryBuilder) WithOptionsData(dataDir string) *QueryBuilder {
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
		if dataDir == "" {
			dataDir = "data/options_data"
		}
	}

	q.AddCTE("options", fmt.Sprintf(`SELECT *,
DATE_DIFF('day', dt, expiration) AS dte
FROM '%s/symbol=%s/*.parquet'
WHERE open_interest > 0 OR bid > 0 OR ask > 0 OR volume > 0`, dataDir, q.symbol))
	return q
}

func (q *QueryBuilder) Build(sqlStr string, limit int32) string {
	if len(q.ctes) == 0 {
		return AppendLimit(sqlStr, EnforceLimit(limit))
	}

	var sb strings.Builder
	sb.WriteString("WITH ")
	for i, cte := range q.ctes {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(cte)
	}
	sb.WriteString("\n")
	sb.WriteString(sqlStr)

	return AppendLimit(sb.String(), EnforceLimit(limit))
}

func (q *QueryBuilder) Execute(ctx context.Context, sqlStr string, limit int32) ([]string, [][]interface{}, error) {
	fullSQL := q.Build(sqlStr, limit)

	if err := ValidateQuery(fullSQL); err != nil {
		return nil, nil, err
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, nil, &QueryError{Code: "EXECUTION_ERROR", Message: fmt.Sprintf("Failed to open DuckDB: %v", err)}
	}
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, fullSQL)
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
