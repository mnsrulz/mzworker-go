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
	SELECT
		quote_date AS dt,
		underlying_close_price AS close,
		underlying_iv30 AS iv30,
		underlying_iv30_percentile AS iv_percentile,
		expiration_date AS expiry,
		FIRST(strike_price) FILTER (WHERE option_type = 'call') AS cs,
		FIRST(mid_price) FILTER (WHERE option_type = 'call') AS cp,
		FIRST(implied_volatility) FILTER (WHERE option_type = 'call') AS cv,
		FIRST(strike_price) FILTER (WHERE option_type = 'put') AS ps,
		FIRST(mid_price) FILTER (WHERE option_type = 'put') AS pp,
		FIRST(implied_volatility) FILTER (WHERE option_type = 'put') AS pv
	FROM (
		SELECT *
		FROM (
			SELECT *,
				round((bid_price + ask_price) / 2, 2) AS mid_price,
				abs(strike_price - underlying_close_price) AS price_strike_diff,
				abs(abs(delta) - %d) AS delta_diff
			FROM base
		) enriched
		%s
		%s
	) filtered
	GROUP BY quote_date, underlying_close_price, underlying_iv30,
		underlying_iv30_percentile, expiration_date
	ORDER BY dt
	`, delta, whereClause.String(), qualifyClause)
}
