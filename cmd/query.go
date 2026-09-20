package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var queryCmd = &cobra.Command{
	Use:   "query [SQL]",
	Short: "Execute SQL queries against DuckDB with CTE scaffold",
	Long:  "Execute arbitrary SQL against DuckDB. The query is wrapped in a CTE scaffold providing tables: expirations, T, T2, base, calc, ranked, dataset.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  queryRunE,
}

func init() {
	queryCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	queryCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	queryCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")
	queryCmd.Flags().Int32P("limit", "l", 1000, "Maximum rows to return")

	_ = queryCmd.MarkFlagRequired("symbol")
	_ = queryCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(queryCmd)
}

func queryRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	sqlStr := ""
	if len(args) > 0 {
		sqlStr = args[0]
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return cmd.Help()
		}

		scanner := bufio.NewScanner(os.Stdin)
		var sb strings.Builder
		for scanner.Scan() {
			sb.WriteString(scanner.Text())
			sb.WriteString(" ")
		}
		sqlStr = strings.TrimSpace(sb.String())
		if sqlStr == "" {
			return fmt.Errorf("no SQL provided")
		}
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	format, _ := cmd.Flags().GetString("format")
	limit, _ := cmd.Flags().GetInt32("limit")

	q := &handler.DynamicSQLQuery{
		Symbol: symbol,
		Query:  sqlStr,
		Limit:  int(limit),
	}

	resp, err := mediatr.Send[*handler.DynamicSQLQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
