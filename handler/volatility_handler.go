package handler

import (
	"context"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/mehdihadeli/go-mediatr"
)

type VolatilityHandler struct {
	executor QueryExecutor
}

func NewVolatilityHandler(executor QueryExecutor) *VolatilityHandler {
	return &VolatilityHandler{executor: executor}
}

var VolatilityQuerySchema = z.Struct(z.Shape{
	"Symbol":       z.String().Required(),
	"LookbackDays": z.Int().Required(),
	"Mode":         z.String().Required(),
	"Delta":        z.Int().Optional(),
	"Strike":       z.Int().Optional(),
	"ExpiryMode":   z.String().Optional(),
	"DTE":          z.Int().Optional(),
	"Expiration":   z.String().Optional(),
})

func (r *VolatilityQuery) Validate() error {
	if errs := VolatilityQuerySchema.Validate(r); errs != nil {
		return fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}
	return nil
}

func init() {
	RegisterStruct[VolatilityQuery, *QueryResponse]("volatility-query", func(d Deps) mediatr.RequestHandler[*VolatilityQuery, *QueryResponse] {
		return NewVolatilityHandler(d.Executor)
	})
}

func (h *VolatilityHandler) Handle(ctx context.Context, req *VolatilityQuery) (*QueryResponse, error) {
	sql := h.buildSQL(req)

	columns, rows, err := h.executor(ctx, req.Symbol, sql, 99999)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

func (h *VolatilityHandler) buildSQL(req *VolatilityQuery) string {
	delta := req.Delta
	mode := strings.ToLower(req.Mode)
	expiryMode := strings.ToLower(req.ExpiryMode)

	var whereClause strings.Builder
	fmt.Fprintf(&whereClause, "WHERE quote_date >= (current_date - %d)", req.LookbackDays)

	qualifyClause := ""
	useQualifyClause := false
	partitionOrderColumn := ""

	switch expiryMode {
	case "rolling":
		useQualifyClause = true
		fmt.Fprintf(&whereClause, " AND dte >= %d", req.DTE)
	case "fixed":
		fmt.Fprintf(&whereClause, " AND expiration_date = '%s'", req.Expiration)
	}

	if mode == "strike" {
		fmt.Fprintf(&whereClause, " AND strike_price = %d", req.Strike)
	} else {
		useQualifyClause = true
		if mode == "atm" {
			partitionOrderColumn = ", price_strike_diff"
		} else {
			partitionOrderColumn = ", delta_diff"
		}
	}

	if useQualifyClause {
		qualifyClause = fmt.Sprintf(`
        QUALIFY ROW_NUMBER() OVER (
                PARTITION BY quote_date, option_type
                ORDER BY dte ASC %s
        ) = 1`, partitionOrderColumn)
	}

	return fmt.Sprintf(`
    PIVOT (
    SELECT quote_date as dt, option_type, underlying_close_price as close,
    expiration_date as expiry, strike_price, round((bid_price + ask_price) / 2, 2) as mid_price,
    implied_volatility as iv, underlying_iv30 as iv30,
    underlying_iv30_percentile as iv30_percentile
    FROM (
        SELECT *, abs(delta) AS abs_delta,
                abs(strike_price - underlying_close_price) AS price_strike_diff,
                abs(abs(delta) - %d) AS delta_diff
        FROM
        base
    )
    %s
    %s
    ORDER BY dt
    )
    ON option_type
    USING FIRST(strike_price) AS strike, FIRST(mid_price) AS mid, FIRST(iv) AS iv
    GROUP BY dt, close, iv30, iv30_percentile, expiry
    ORDER BY dt
    `, delta, whereClause.String(), qualifyClause)
}
