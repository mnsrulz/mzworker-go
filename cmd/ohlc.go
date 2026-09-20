package cmd

import (
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var ohlcCmd = &cobra.Command{
	Use:   "ohlc",
	Short: "Query OHLC data for a symbol",
	Long:  "Query Open-High-Low-Close data for a specific stock symbol.",
	RunE:  ohlcRunE,
}

func init() {
	ohlcCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	ohlcCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	ohlcCmd.Flags().StringP("from", "F", "", "Start date (YYYY-MM-DD)")
	ohlcCmd.Flags().StringP("to", "T", "", "End date (YYYY-MM-DD)")
	ohlcCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")
	ohlcCmd.Flags().Int32P("limit", "l", 1000, "Maximum rows to return")

	_ = ohlcCmd.MarkFlagRequired("symbol")
	_ = ohlcCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(ohlcCmd)
}

func ohlcRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	format, _ := cmd.Flags().GetString("format")
	limit, _ := cmd.Flags().GetInt32("limit")

	q := &handler.OHLCQuery{
		Symbol: symbol,
		From:   from,
		To:     to,
		Limit:  int(limit),
	}

	resp, err := mediatr.Send[*handler.OHLCQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("ohlc query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
