package query

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type InternalQueryExecutor struct {
	optionsDir string
	ohlcDir    string
	staticDir  string
	symbol     string
}

func NewInternalQueryExecutor(symbol, dataDir string) *InternalQueryExecutor {
	return &InternalQueryExecutor{
		symbol:     strings.ToUpper(symbol),
		optionsDir: filepath.Join(dataDir, "w2-output"),
		ohlcDir:    filepath.Join(dataDir, "ohlc"),
		staticDir:  filepath.Join(dataDir, "static"),
	}
}

func (e *InternalQueryExecutor) BuildBaseQueryCte() string {
	return fmt.Sprintf(`expirations AS (
    SELECT expiration, isWeekly as is_weekly, isMonthly as is_monthly
    FROM '%s/options-expirations-summary.json'
), T AS (
    SELECT DISTINCT dt, open as underlying_open_price, high as underlying_high_price, low as underlying_low_price,
    close as underlying_close_price, volume as underlying_volume,
    last_value(nullif(iv30, 0) IGNORE NULLS) OVER (
        ORDER BY dt
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) as underlying_iv30,
    symbol as underlying_symbol,
    CAST(100 * PERCENT_RANK() OVER (PARTITION BY symbol ORDER BY iv30) AS DECIMAL(10, 2)) AS underlying_iv30_percentile
    FROM '%s/*.parquet' WHERE replace(symbol, '^', '') = '%s'
    AND underlying_close_price > 0
), T2 AS (
    SELECT *,
    DATE_DIFF('day', dt, expiration) AS dte
    FROM '%s/symbol=%s/*.parquet'
    WHERE open_interest > 0 OR bid > 0 OR ask > 0 OR volume > 0
), weekly_expiries AS (
    SELECT expiration FROM expirations WHERE is_weekly = 1
), base AS (
    SELECT T.dt AS quote_date,
    expiration AS expiration_date,
    dte,
    option_symbol AS option_ticker,
    option_type,
    strike AS strike_price,
    open_interest,
    volume AS option_volume,
    delta, gamma, vega, theta, rho,
    theo AS theoretical_price,
    iv * 100 AS implied_volatility,
    open AS option_open_price,
    high AS option_high_price,
    bid AS bid_price,
    ask AS ask_price,
    underlying_symbol,
    underlying_open_price, underlying_high_price, underlying_low_price, underlying_close_price,
    underlying_iv30, underlying_volume,
    underlying_iv30_percentile
    FROM T2
    JOIN T ON T.dt = T2.dt
), calc AS (
    SELECT *,
    dayofweek(CAST(quote_date as date)) AS quote_dow,
    dayofweek(expiration_date) AS expiration_dow,
    round((bid_price + ask_price) / 2, 2) AS mid_price,
    round((bid_price + ask_price) / 2, 2) AS option_close_price,
    abs(strike_price - underlying_close_price) AS strike_distance,
    abs(strike_price - underlying_close_price) / underlying_close_price * 100 AS strike_distance_pct,
    CASE WHEN expirations.is_weekly = 1 THEN 1 ELSE 0 END AS is_weekly_expiration,
    CASE WHEN expirations.is_monthly = 1 THEN 1 ELSE 0 END AS is_monthly_expiration
    FROM base
    JOIN expirations ON base.expiration_date = expirations.expiration
), ranked AS (
    SELECT
        *,
        row_number() OVER (
            PARTITION BY quote_date, expiration_date, option_type
            ORDER BY strike_distance, strike_price
        ) AS atm_rank
    FROM calc
), dataset AS (
    SELECT quote_date, expiration_date, expiration_dow, quote_dow,
    is_weekly_expiration, is_monthly_expiration, dte,
    option_ticker, option_type, strike_price,
    open_interest, option_volume,
    delta, gamma, vega, theta, rho,
    theoretical_price, implied_volatility,
    option_open_price, option_high_price, option_close_price,
    bid_price, ask_price, mid_price,
    CASE
        WHEN open_interest > 1000 AND option_volume > 100 THEN 'HIGH'
        WHEN open_interest > 100 THEN 'MEDIUM'
        ELSE 'LOW'
    END AS liquidity_tier,
    option_volume / NULLIF(open_interest, 0) AS volume_oi_ratio,
    underlying_symbol,
    underlying_open_price, underlying_high_price, underlying_low_price, underlying_close_price,
    underlying_iv30, underlying_volume,
    underlying_iv30_percentile,
    CASE
        WHEN atm_rank = 1 THEN 'ATM'
        WHEN (
            (option_type = 'call' AND strike_price < underlying_close_price) OR
            (option_type = 'put'  AND strike_price > underlying_close_price)
        ) THEN 'ITM'
        ELSE 'OTM'
    END AS moneyness,
    CASE
        WHEN (option_type = 'C' AND strike_price < underlying_close_price) OR
            (option_type = 'P'  AND strike_price > underlying_close_price)
            THEN -(strike_distance_pct)
        ELSE strike_distance_pct
    END AS moneyness_percent,
    CASE
        WHEN dte <= 7 THEN '0-7D'
        WHEN dte <= 30 THEN '7-30D'
        WHEN dte <= 90 THEN '30-90D'
        ELSE '90D+'
    END AS expiry_bucket
    FROM ranked
)`, e.staticDir, e.ohlcDir, e.symbol, e.optionsDir, e.symbol)
}

func (e *InternalQueryExecutor) BuildFullQuery(handlerSQL string, limit int32) string {
	cte := e.BuildBaseQueryCte()

	limitClause := ""
	l := EnforceLimit(limit)
	if l > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", l)
	}

	return fmt.Sprintf(`WITH %s
SELECT * FROM (
    SELECT * FROM (
        %s
    ) LIMITED_CTE %s
)`, cte, handlerSQL, limitClause)
}

func (e *InternalQueryExecutor) Execute(ctx context.Context, handlerSQL string, limit int32) ([]string, [][]interface{}, error) {
	fullSQL := e.BuildFullQuery(handlerSQL, limit)

	if err := ValidateQuery(fullSQL); err != nil {
		return nil, nil, err
	}

	return ExecuteQuery(ctx, fullSQL, 0)
}
