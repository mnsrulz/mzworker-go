package handler

import (
	"context"
	"fmt"
	"strings"

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
	sql := buildOHLCQuery(req.Symbol, req.From, req.To)

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	columns, rows, err := h.executor(ctx, req.Symbol, sql, int32(limit))
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

func buildOHLCQuery(symbol, from, to string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "SELECT dt, underlying_symbol as symbol, underlying_open_price as open, underlying_high_price as high, underlying_low_price as low, underlying_close_price as close, underlying_volume as volume, underlying_iv30 as iv30 FROM T")

	fmt.Fprintf(&sb, " WHERE underlying_symbol = '%s'", strings.ToUpper(symbol))

	if from != "" {
		fmt.Fprintf(&sb, " AND dt >= '%s'", from)
	}
	if to != "" {
		fmt.Fprintf(&sb, " AND dt <= '%s'", to)
	}

	sb.WriteString(" ORDER BY dt")
	return sb.String()
}
