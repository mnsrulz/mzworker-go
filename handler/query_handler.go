package handler

import (
	"context"
	"fmt"

	z "github.com/Oudwins/zog"
)

type QueryExecutor func(ctx context.Context, symbol, sqlStr string, limit int32) ([]string, [][]interface{}, error)

type QueryHandler struct {
	executor QueryExecutor
}

func NewQueryHandler(executor QueryExecutor) *QueryHandler {
	return &QueryHandler{executor: executor}
}

func (h *QueryHandler) Handle(ctx context.Context, req *DynamicSQLQuery) (*QueryResponse, error) {
	errs := DynamicSQLQuerySchema.Validate(req)
	if errs != nil {
		return nil, fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}

	columns, rows, err := h.executor(ctx, req.Symbol, req.Query, int32(req.Limit))
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}
