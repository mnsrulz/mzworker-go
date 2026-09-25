package handler

import (
	"context"

	"github.com/mehdihadeli/go-mediatr"
)

type QueryExecutor func(ctx context.Context, symbol, sqlStr string, limit int32) ([]string, [][]interface{}, error)

type QueryHandler struct {
	executor QueryExecutor
}

func NewQueryHandler(executor QueryExecutor) *QueryHandler {
	return &QueryHandler{executor: executor}
}

func init() {
	RegisterStruct[DynamicSQLQuery, *QueryResponse]("dynamic-sql-query", func(d Deps) mediatr.RequestHandler[*DynamicSQLQuery, *QueryResponse] {
		return NewQueryHandler(d.Executor)
	})
}

func (h *QueryHandler) Handle(ctx context.Context, req *DynamicSQLQuery) (*QueryResponse, error) {
	columns, rows, err := h.executor(ctx, req.Symbol, req.Query, int32(req.Limit))
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}
