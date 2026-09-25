package handler

import (
	"context"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/mehdihadeli/go-mediatr"
)

type OptionsStatHandler struct {
	executor QueryExecutor
}

func NewOptionsStatHandler(executor QueryExecutor) *OptionsStatHandler {
	return &OptionsStatHandler{executor: executor}
}

var OptionsStatQuerySchema = z.Struct(z.Shape{
	"Symbol":       z.String().Required(),
	"LookbackDays": z.Int().Required(),
})

func (r *OptionsStatQuery) Validate() error {
	if errs := OptionsStatQuerySchema.Validate(r); errs != nil {
		return fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}
	return nil
}

func init() {
	RegisterStruct[OptionsStatQuery, *QueryResponse]("options-stat-query", func(d Deps) mediatr.RequestHandler[*OptionsStatQuery, *QueryResponse] {
		return NewOptionsStatHandler(d.Executor)
	})
}

func (h *OptionsStatHandler) Handle(ctx context.Context, req *OptionsStatQuery) (*QueryResponse, error) {
	sql := buildOptionsStatSQL(req)

	columns, rows, err := h.executor(ctx, req.Symbol, sql, 99999)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

func buildOptionsStatSQL(req *OptionsStatQuery) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, `
    PIVOT (
        SELECT T2.dt, T2.option_type, T.underlying_close_price AS close,
            SUM(T2.open_interest) AS total_oi,
            ROUND(SUM(T2.open_interest * T2.theo) * 100) AS total_price,
            ROUND(SUM(T2.open_interest * abs(T2.delta))) AS total_delta,
            CAST(COUNT(DISTINCT T2.option_symbol) AS INTEGER) AS options_count
        FROM T2
        JOIN T ON T.dt = T2.dt
        WHERE T2.dt >= current_date - %d
        GROUP BY T2.dt, T2.option_type, T.underlying_close_price
    )
    ON option_type
    USING FIRST(total_oi) AS oi, FIRST(total_price) AS price, FIRST(total_delta) AS delta, FIRST(options_count) AS options
    GROUP BY dt, close
    ORDER BY dt
    `, req.LookbackDays)

	return sb.String()
}
