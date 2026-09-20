package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
	"github.com/mnsrulz/mzworker-go/query"
)

var RootCmd = &cobra.Command{
	Use:   "mzworker-go",
	Short: "CLI tool for querying DuckDB with options data",
	Long:  "CLI tool for querying DuckDB with SQL. Callable from opencode MCP/skills.",
	RunE:  rootRunE,
}

func init() {
	RootCmd.Flags().Bool("schema", false, "Show available tables and columns")
}

func rootRunE(cmd *cobra.Command, args []string) error {
	if schema, _ := cmd.Flags().GetBool("schema"); schema {
		printSchema()
		return nil
	}
	return cmd.Help()
}

func Execute() error {
	return RootCmd.Execute()
}

func Output(columns []string, rows [][]interface{}, format string) error {
	switch format {
	case "table":
		return formatTable(os.Stdout, columns, rows)
	case "raw":
		return formatRaw(os.Stdout, columns, rows)
	default:
		return formatJSON(os.Stdout, columns, rows)
	}
}

func formatJSON(w *os.File, columns []string, rows [][]interface{}) error {
	objects := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		obj := make(map[string]interface{}, len(columns))
		for j, col := range columns {
			obj[col] = row[j]
		}
		objects[i] = obj
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

func formatTable(w *os.File, columns []string, rows [][]interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for i, col := range columns {
		if i > 0 {
			_, _ = fmt.Fprint(tw, "\t")
		}
		_, _ = fmt.Fprint(tw, strings.ToUpper(col))
	}
	_, _ = fmt.Fprintln(tw)

	for _, row := range rows {
		for i, val := range row {
			if i > 0 {
				_, _ = fmt.Fprint(tw, "\t")
			}
			_, _ = fmt.Fprint(tw, formatValue(val))
		}
		_, _ = fmt.Fprintln(tw)
	}

	return tw.Flush()
}

func formatRaw(w *os.File, columns []string, rows [][]interface{}) error {
	type rawResponse struct {
		Columns []string        `json:"columns"`
		Rows    [][]interface{} `json:"rows"`
	}
	resp := rawResponse{Columns: columns, Rows: rows}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", v)
}

func registerMediatr(dataDir string) error {
	if err := mediatr.RegisterRequestPipelineBehaviors(&handler.ValidationBehavior{}); err != nil {
		return fmt.Errorf("failed to register pipeline behavior: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(handler.NewQueryHandler(newInternalQueryExecutor(dataDir))); err != nil {
		return fmt.Errorf("failed to register query handler: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(handler.NewOHLCHandler(newInternalQueryExecutor(dataDir))); err != nil {
		return fmt.Errorf("failed to register OHLC handler: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(handler.NewVolatilityHandler(newInternalQueryExecutor(dataDir))); err != nil {
		return fmt.Errorf("failed to register volatility handler: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(handler.NewOptionsStatHandler(newInternalQueryExecutor(dataDir))); err != nil {
		return fmt.Errorf("failed to register options-stat handler: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(handler.NewExpectedMoveHandler(newInternalQueryExecutor(dataDir))); err != nil {
		return fmt.Errorf("failed to register expected-move handler: %w", err)
	}
	if err := mediatr.RegisterRequestHandler(&handler.PingHandler{}); err != nil {
		return fmt.Errorf("failed to register ping handler: %w", err)
	}
	return nil
}

func newInternalQueryExecutor(dataDir string) handler.QueryExecutor {
	return func(ctx context.Context, symbol, sql string, limit int32) ([]string, [][]interface{}, error) {
		executor := query.NewInternalQueryExecutor(symbol, dataDir)
		return executor.Execute(ctx, sql, limit)
	}
}
