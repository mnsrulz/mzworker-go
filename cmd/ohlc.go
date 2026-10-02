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
	ohlcCmd.Flags().IntP("lookback", "d", 30, "Lookback days")
	ohlcCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")

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
	lookback, _ := cmd.Flags().GetInt("lookback")
	format, _ := cmd.Flags().GetString("format")

	q := &handler.OHLCQuery{
		Symbol:       symbol,
		LookbackDays: lookback,
	}

	resp, err := mediatr.Send[*handler.OHLCQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("ohlc query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
