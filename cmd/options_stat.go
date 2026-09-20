package cmd

import (
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var optionsStatCmd = &cobra.Command{
	Use:   "options-stat",
	Short: "Query options statistics for a symbol",
	Long:  "Query aggregate options statistics (OI, price, delta, count) pivoted by option type.",
	RunE:  optionsStatRunE,
}

func init() {
	optionsStatCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	optionsStatCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	optionsStatCmd.Flags().IntP("lookback", "d", 30, "Lookback days")
	optionsStatCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")

	_ = optionsStatCmd.MarkFlagRequired("symbol")
	_ = optionsStatCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(optionsStatCmd)
}

func optionsStatRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	lookback, _ := cmd.Flags().GetInt("lookback")
	format, _ := cmd.Flags().GetString("format")

	q := &handler.OptionsStatQuery{
		Symbol:       symbol,
		LookbackDays: lookback,
	}

	resp, err := mediatr.Send[*handler.OptionsStatQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("options-stat query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
