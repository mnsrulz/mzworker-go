package cmd

import (
	"fmt"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/spf13/cobra"

	"github.com/mnsrulz/mzworker-go/handler"
)

var volatilityCmd = &cobra.Command{
	Use:   "volatility",
	Short: "Query volatility surface for a symbol",
	Long:  "Query implied volatility data for a specific stock symbol with pivot by option type.",
	RunE:  volatilityRunE,
}

func init() {
	volatilityCmd.Flags().StringP("symbol", "s", "", "Stock symbol (e.g. AAPL, NVDA)")
	volatilityCmd.Flags().String("data-dir", "", "Data directory (e.g. /Users/mz/Documents/ODATA)")
	volatilityCmd.Flags().IntP("lookback", "d", 30, "Lookback days")
	volatilityCmd.Flags().StringP("mode", "m", "atm", "Query mode: atm, delta, strike")
	volatilityCmd.Flags().Int("delta", 0, "Delta value (for delta mode)")
	volatilityCmd.Flags().Int("strike", 0, "Strike price (for strike mode)")
	volatilityCmd.Flags().String("expiry-mode", "rolling", "Expiry mode: rolling, fixed")
	volatilityCmd.Flags().Int("dte", 0, "Days to expiration (for rolling expiry)")
	volatilityCmd.Flags().String("expiration", "", "Expiration date YYYY-MM-DD (for fixed expiry)")
	volatilityCmd.Flags().StringP("format", "f", "json", "Output format: json, table, raw")

	_ = volatilityCmd.MarkFlagRequired("symbol")
	_ = volatilityCmd.MarkFlagRequired("data-dir")

	RootCmd.AddCommand(volatilityCmd)
}

func volatilityRunE(cmd *cobra.Command, args []string) error {
	dataDir, _ := cmd.Flags().GetString("data-dir")
	if err := registerMediatr(dataDir); err != nil {
		return err
	}

	symbol, _ := cmd.Flags().GetString("symbol")
	lookback, _ := cmd.Flags().GetInt("lookback")
	mode, _ := cmd.Flags().GetString("mode")
	delta, _ := cmd.Flags().GetInt("delta")
	strike, _ := cmd.Flags().GetInt("strike")
	expiryMode, _ := cmd.Flags().GetString("expiry-mode")
	dte, _ := cmd.Flags().GetInt("dte")
	expiration, _ := cmd.Flags().GetString("expiration")
	format, _ := cmd.Flags().GetString("format")

	q := &handler.VolatilityQuery{
		Symbol:       symbol,
		LookbackDays: lookback,
		Mode:         mode,
		Delta:        delta,
		Strike:       strike,
		ExpiryMode:   expiryMode,
		DTE:          dte,
		Expiration:   expiration,
	}

	resp, err := mediatr.Send[*handler.VolatilityQuery, *handler.QueryResponse](cmd.Context(), q)
	if err != nil {
		return fmt.Errorf("volatility query failed: %w", err)
	}

	return Output(resp.Columns, resp.Rows, format)
}
