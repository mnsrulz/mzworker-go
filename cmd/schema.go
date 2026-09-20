package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type schemaColumn struct {
	Name string
	Type string
	Desc string
}

type schemaTable struct {
	Name    string
	Desc    string
	Columns []schemaColumn
}

var schema = []schemaTable{
	{
		Name: "expirations",
		Desc: "Static expiration calendar",
		Columns: []schemaColumn{
			{"expiration", "DATE", "Expiration date"},
			{"is_weekly", "BOOLEAN", "Weekly expiration flag"},
			{"is_monthly", "BOOLEAN", "Monthly expiration flag"},
		},
	},
	{
		Name: "T",
		Desc: "Daily OHLC + IV30 for the underlying",
		Columns: []schemaColumn{
			{"dt", "DATE", "Quote date"},
			{"underlying_open_price", "DOUBLE", "Open price"},
			{"underlying_high_price", "DOUBLE", "High price"},
			{"underlying_low_price", "DOUBLE", "Low price"},
			{"underlying_close_price", "DOUBLE", "Close price"},
			{"underlying_volume", "BIGINT", "Volume"},
			{"underlying_iv30", "DOUBLE", "30-day IV (forward-filled)"},
			{"underlying_symbol", "VARCHAR", "Stock symbol"},
			{"underlying_iv30_percentile", "DECIMAL", "IV30 percentile rank"},
		},
	},
	{
		Name: "dataset",
		Desc: "Full options chain view (main table to query)",
		Columns: []schemaColumn{
			{"quote_date", "DATE", "Quote date"},
			{"expiration_date", "DATE", "Expiration date"},
			{"expiration_dow", "INT", "Day of week for expiration"},
			{"quote_dow", "INT", "Day of week for quote date"},
			{"is_weekly_expiration", "INT", "1 if weekly expiration"},
			{"is_monthly_expiration", "INT", "1 if monthly expiration"},
			{"dte", "INT", "Days to expiration"},
			{"option_ticker", "VARCHAR", "Option symbol"},
			{"option_type", "VARCHAR", "'call' or 'put'"},
			{"strike_price", "DOUBLE", "Strike price"},
			{"open_interest", "BIGINT", "Open interest"},
			{"option_volume", "BIGINT", "Option volume"},
			{"delta", "DOUBLE", "Delta"},
			{"gamma", "DOUBLE", "Gamma"},
			{"vega", "DOUBLE", "Vega"},
			{"theta", "DOUBLE", "Theta"},
			{"rho", "DOUBLE", "Rho"},
			{"theoretical_price", "DOUBLE", "Theoretical price"},
			{"implied_volatility", "DOUBLE", "IV (%)"},
			{"option_open_price", "DOUBLE", "Option open"},
			{"option_high_price", "DOUBLE", "Option high"},
			{"option_close_price", "DOUBLE", "Option close (mid)"},
			{"bid_price", "DOUBLE", "Bid price"},
			{"ask_price", "DOUBLE", "Ask price"},
			{"mid_price", "DOUBLE", "Mid price (bid+ask)/2"},
			{"liquidity_tier", "VARCHAR", "HIGH/MEDIUM/LOW"},
			{"volume_oi_ratio", "DOUBLE", "Volume / open interest"},
			{"underlying_symbol", "VARCHAR", "Stock symbol"},
			{"underlying_open_price", "DOUBLE", "Underlying open"},
			{"underlying_high_price", "DOUBLE", "Underlying high"},
			{"underlying_low_price", "DOUBLE", "Underlying low"},
			{"underlying_close_price", "DOUBLE", "Underlying close"},
			{"underlying_iv30", "DOUBLE", "30-day IV"},
			{"underlying_volume", "BIGINT", "Underlying volume"},
			{"underlying_iv30_percentile", "DECIMAL", "IV30 percentile"},
			{"moneyness", "VARCHAR", "ATM/ITM/OTM"},
			{"moneyness_percent", "DOUBLE", "Signed distance %"},
			{"expiry_bucket", "VARCHAR", "0-7D/7-30D/30-90D/90D+"},
		},
	},
}

func printSchema() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, t := range schema {
		_, _ = fmt.Fprintf(w, "TABLE: %s\t%s\n", t.Name, t.Desc)
		_, _ = fmt.Fprintf(w, "  COLUMN\tTYPE\tDESCRIPTION\n")
		for _, c := range t.Columns {
			_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\n", c.Name, c.Type, c.Desc)
		}
		_, _ = fmt.Fprintln(w)
	}

	_ = w.Flush()
}
