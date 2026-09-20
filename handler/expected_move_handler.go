package handler

import (
	"context"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
)

type ExpectedMoveHandler struct {
	executor QueryExecutor
}

func NewExpectedMoveHandler(executor QueryExecutor) *ExpectedMoveHandler {
	return &ExpectedMoveHandler{executor: executor}
}

var ExpectedMoveQuerySchema = z.Struct(z.Shape{
	"Symbol":       z.String().Required(),
	"LookbackDays": z.Int().Required(),
	"ExpiryMode":   z.String().Required(),
})

func (h *ExpectedMoveHandler) Handle(ctx context.Context, req *ExpectedMoveQuery) (*QueryResponse, error) {
	errs := ExpectedMoveQuerySchema.Validate(req)
	if errs != nil {
		return nil, fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}

	sql := buildExpectedMoveSQL(req)

	columns, rows, err := h.executor(ctx, req.Symbol, sql, 99999)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

func buildExpectedMoveSQL(req *ExpectedMoveQuery) string {
	mode := strings.ToLower(req.ExpiryMode)

	isWeeklyFilter := "is_weekly"
	if mode != "weekly" {
		isWeeklyFilter = "is_monthly"
	}

	isWeeklyExpirationCol := "is_weekly_expiration"
	if mode != "weekly" {
		isWeeklyExpirationCol = "is_monthly_expiration"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, `
    WITH opex AS (
        SELECT * FROM (
            WITH opex_cte AS (
                SELECT expiration, LEAD(expiration) OVER (ORDER BY expiration) AS next_opex
                FROM expirations
                WHERE %s = 1
            )
            SELECT next_opex AS expiration, LEAD(quote_date) OVER (ORDER BY quote_date) AS opex_start
            FROM (
                SELECT DISTINCT quote_date, expiration, next_opex
                FROM dataset LEFT JOIN opex_cte ON quote_date = expiration
                WHERE quote_dow NOT IN (6,7)
                ORDER BY 1
            )
        ) WHERE expiration IS NOT NULL
    )
    SELECT quote_date AS dt, underlying_close_price AS last_close, straddle_price, expiration_date AS expiry
    FROM (
        SELECT quote_date, expiration_date, dte, strike_price, underlying_close_price,
            round(SUM(mid_price), 2) AS straddle_price
        FROM dataset JOIN opex ON dataset.quote_date = opex.opex_start AND dataset.expiration_date = opex.expiration
        WHERE moneyness = 'ATM'
        AND quote_date >= current_date - %d
        AND %s = 1
        GROUP BY quote_date, dte, strike_price, expiration_date, underlying_close_price
    )
    ORDER BY quote_date
    `, isWeeklyFilter, req.LookbackDays, isWeeklyExpirationCol)

	return sb.String()
}
