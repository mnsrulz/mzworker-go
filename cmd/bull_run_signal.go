package cmd

import (
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var bullRunSignalCmd = &cobra.Command{
	Use:   "bull-run-signal",
	Short: "Query bull-run signal series for a symbol",
	Long:  "Query the weekly bull-run entry signal series (cheap vol, positioning, wing, trend gate) for a specific symbol.",
	RunE:  bullRunSignalRunE,
}

func init() {
	bullRunSignalCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	bullRunSignalCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	bullRunSignalCmd.Flags().IntP("lookback", "d", 730, "Lookback days")
	bullRunSignalCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")

	_ = bullRunSignalCmd.MarkFlagRequired("symbol")
	_ = bullRunSignalCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(bullRunSignalCmd)
}

func bullRunSignalRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	lookback, _ := cmd.Flags().GetInt("lookback")
	format, _ := cmd.Flags().GetString("format")

	q := &handler.BullRunSignalQuery{
		Symbol:       symbol,
		LookbackDays: lookback,
	}

	resp, err := mediatr.Send[*handler.BullRunSignalQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("bull-run-signal query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
