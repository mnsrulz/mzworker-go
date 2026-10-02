package handler

import (
	"context"
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
)

type OHLCHandler struct {
	executor QueryExecutor
}

func NewOHLCHandler(executor QueryExecutor) *OHLCHandler {
	return &OHLCHandler{executor: executor}
}

func init() {
	RegisterStruct[OHLCQuery, *QueryResponse]("ohlc-query", func(d Deps) mediatr.RequestHandler[*OHLCQuery, *QueryResponse] {
		return NewOHLCHandler(d.Executor)
	})
}

func (h *OHLCHandler) Handle(ctx context.Context, req *OHLCQuery) (*QueryResponse, error) {
	sql := buildOHLCQuery(req.LookbackDays)

	columns, rows, err := h.executor(ctx, req.Symbol, sql, 99999)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

func buildOHLCQuery(lookbackDays int) string {
	return fmt.Sprintf(`
	SELECT DISTINCT
		strftime(CASE
			WHEN dayofweek(CAST(dt AS DATE)) = 1 THEN CAST(dt AS DATE) - INTERVAL 3 DAY
			ELSE CAST(dt AS DATE) - INTERVAL 1 DAY
		END, '%%Y-%%m-%%d') AS dt,
		underlying_open_price AS open,
		underlying_high_price AS high,
		underlying_low_price AS low,
		underlying_close_price AS close,
		underlying_iv30 AS iv30
	FROM T
	WHERE T.dt >= current_date - %d
		AND dayofweek(dt) <> 6
	ORDER BY dt
	`, lookbackDays)
}
