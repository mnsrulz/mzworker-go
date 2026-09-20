package cmd

import (
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var expectedMoveCmd = &cobra.Command{
	Use:   "expected-move",
	Short: "Query expected move for a symbol",
	Long:  "Calculate expected move (straddle price as % of underlying) for weekly or monthly expirations.",
	RunE:  expectedMoveRunE,
}

func init() {
	expectedMoveCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	expectedMoveCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	expectedMoveCmd.Flags().IntP("lookback", "d", 30, "Lookback days")
	expectedMoveCmd.Flags().String("expiry-mode", "weekly", "Expiry mode: weekly, monthly")
	expectedMoveCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")

	_ = expectedMoveCmd.MarkFlagRequired("symbol")
	_ = expectedMoveCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(expectedMoveCmd)
}

func expectedMoveRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	lookback, _ := cmd.Flags().GetInt("lookback")
	expiryMode, _ := cmd.Flags().GetString("expiry-mode")
	format, _ := cmd.Flags().GetString("format")

	q := &handler.ExpectedMoveQuery{
		Symbol:       symbol,
		LookbackDays: lookback,
		ExpiryMode:   expiryMode,
	}

	resp, err := mediatr.Send[*handler.ExpectedMoveQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("expected-move query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
