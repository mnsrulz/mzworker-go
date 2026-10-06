package handler

import (
	"context"
	"fmt"

	z "github.com/Oudwins/zog"
	"github.com/mehdihadeli/go-mediatr"
)

type BullRunSignalHandler struct {
	executor QueryExecutor
}

func NewBullRunSignalHandler(executor QueryExecutor) *BullRunSignalHandler {
	return &BullRunSignalHandler{executor: executor}
}

var BullRunSignalQuerySchema = z.Struct(z.Shape{
	"Symbol":       z.String().Required(),
	"LookbackDays": z.Int().Required(),
})

func (r *BullRunSignalQuery) Validate() error {
	if errs := BullRunSignalQuerySchema.Validate(r); errs != nil {
		return fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}
	if r.LookbackDays <= 0 {
		return fmt.Errorf("validation failed: LookbackDays must be positive")
	}
	return nil
}

func init() {
	RegisterStruct[BullRunSignalQuery, *QueryResponse]("bull-run-signal-query", func(d Deps) mediatr.RequestHandler[*BullRunSignalQuery, *QueryResponse] {
		return NewBullRunSignalHandler(d.Executor)
	})
}

func (h *BullRunSignalHandler) Handle(ctx context.Context, req *BullRunSignalQuery) (*QueryResponse, error) {
	sql := fmt.Sprintf(bullRunSeriesSQL, req.LookbackDays)

	columns, rows, err := h.executor(ctx, req.Symbol, sql, 99999)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{Columns: columns, Rows: rows}, nil
}

// bullRunSeriesSQL is the weekly bull-run entry signal series. It consumes the
// scaffold-provided "dataset" CTE; %d receives the lookback day count.
const bullRunSeriesSQL = `WITH
opt30 AS (
  SELECT quote_date, expiration_date, dte, option_type, strike_price, implied_volatility, underlying_close_price
  FROM dataset WHERE dte BETWEEN 18 AND 42
),
exp30 AS (
  SELECT quote_date, expiration_date,
         ROW_NUMBER() OVER (PARTITION BY quote_date ORDER BY ABS(dte-30), expiration_date) AS rn
  FROM (SELECT DISTINCT quote_date, expiration_date, dte FROM opt30)
),
p30 AS (
  SELECT o.* FROM opt30 o JOIN exp30 e USING (quote_date, expiration_date) WHERE e.rn = 1
),
iv30 AS (
  SELECT quote_date,
         MAX(underlying_close_price) AS spot,
         ARG_MIN(implied_volatility, ABS(strike_price - underlying_close_price)) FILTER (WHERE option_type='C') AS atm_iv,
         ARG_MIN(implied_volatility, ABS(strike_price - underlying_close_price*1.15)) FILTER (WHERE option_type='C') AS c15_iv,
         ARG_MIN(implied_volatility, ABS(strike_price - underlying_close_price*0.85)) FILTER (WHERE option_type='P') AS p15_iv
  FROM p30 GROUP BY quote_date
),
opt10 AS (
  SELECT quote_date, expiration_date, dte, strike_price, implied_volatility, underlying_close_price
  FROM dataset WHERE option_type='C' AND dte BETWEEN 5 AND 16
),
exp10 AS (
  SELECT quote_date, expiration_date,
         ROW_NUMBER() OVER (PARTITION BY quote_date ORDER BY ABS(dte-10), expiration_date) AS rn
  FROM (SELECT DISTINCT quote_date, expiration_date, dte FROM opt10)
),
front AS (
  SELECT q.quote_date,
         ARG_MIN(q.implied_volatility, ABS(q.strike_price - q.underlying_close_price)) AS front_iv
  FROM opt10 q JOIN exp10 e USING (quote_date, expiration_date)
  WHERE e.rn = 1 GROUP BY q.quote_date
),
flow AS (
  SELECT quote_date,
         SUM(option_volume) AS tot_vol,
         SUM(CASE WHEN volume_oi_ratio >= 2 AND dte > 7 THEN 1 ELSE 0 END) AS n_unusual,
         SUM(CASE WHEN option_type='P' THEN option_volume ELSE 0 END) AS put_vol,
         SUM(CASE WHEN option_type='C' THEN option_volume ELSE 0 END) AS call_vol,
         SUM(CASE WHEN option_type='C' AND strike_price > underlying_close_price THEN option_volume ELSE 0 END)
           / NULLIF(SUM(CASE WHEN option_type='C' THEN option_volume ELSE 0 END),0) AS otm_call_share,
         MAX(underlying_iv30_percentile) AS ivpct
  FROM dataset WHERE dte > 0 GROUP BY quote_date
),
daily AS (
  SELECT i.quote_date, i.spot, i.atm_iv, i.c15_iv, i.p15_iv, f.front_iv,
         fl.tot_vol, fl.n_unusual, fl.put_vol, fl.call_vol, fl.otm_call_share, fl.ivpct
  FROM iv30 i JOIN front f USING (quote_date) JOIN flow fl USING (quote_date)
),
weekly AS (
  SELECT date_trunc('week', quote_date) AS wk,
         MAX(spot) AS px,
         AVG(atm_iv) AS atm30, AVG(c15_iv) AS c15, AVG(p15_iv) AS p15, AVG(front_iv) AS front10,
         AVG(ivpct) AS ivpct,
         SUM(tot_vol) AS tot_vol,
         SUM(n_unusual) AS n_unusual,
         SUM(put_vol)::FLOAT / NULLIF(SUM(call_vol),0) AS pc_vol,
         AVG(otm_call_share) AS otm_call_share
  FROM daily GROUP BY 1
),
z AS (
  SELECT *,
    (tot_vol  - AVG(tot_vol)  OVER w8) / NULLIF(STDDEV_SAMP(tot_vol)  OVER w8,0) AS vol_z,
    (n_unusual - AVG(n_unusual) OVER w8) / NULLIF(STDDEV_SAMP(n_unusual) OVER w8,0) AS unus_z
  FROM weekly
  WINDOW w8 AS (ORDER BY wk ROWS BETWEEN 8 PRECEDING AND 1 PRECEDING)
),
scored AS (
  SELECT z.*,
    AVG(px) OVER (ORDER BY wk ROWS BETWEEN 12 PRECEDING AND CURRENT ROW) AS sma13,
    AVG(px) OVER (ORDER BY wk ROWS BETWEEN 16 PRECEDING AND 4 PRECEDING) AS sma13_prev,
    LEAD(px, 4) OVER (ORDER BY wk) AS px_fwd4,
    LEAD(px, 8) OVER (ORDER BY wk) AS px_fwd8,
    LAG(px, 4) OVER (ORDER BY wk) AS px_lag4,
    (CASE WHEN ivpct <= 25 THEN 1 ELSE 0 END) AS b_cheap,
    (CASE WHEN unus_z <= -1 THEN 1 ELSE 0 END) AS b_drought,
    (CASE WHEN vol_z <= -1 THEN 1 ELSE 0 END) AS b_dryup,
    (CASE WHEN c15 - atm30 > 0 THEN 1 ELSE 0 END) AS b_wing
  FROM z
),
final AS (
  SELECT *,
    b_cheap + b_drought + b_dryup + b_wing AS score,
    (sma13 IS NOT NULL AND px > sma13 AND sma13 > sma13_prev) AS trend_up
  FROM scored
)
SELECT wk,
       ROUND(px, 1) AS px,
       ROUND(100.0 * (px / NULLIF(px_lag4, 0) - 1), 1) AS px4w_chg,
       ROUND(atm30, 1) AS atm30,
       ROUND(c15 - atm30, 1) AS c_skew,
       ROUND(p15 - atm30, 1) AS p_skew,
       ROUND(front10 - atm30, 1) AS term,
       ROUND(ivpct, 0) AS ivpct,
       tot_vol,
       ROUND(vol_z, 1) AS vol_z,
       n_unusual,
       ROUND(unus_z, 1) AS unus_z,
       ROUND(pc_vol, 2) AS pc,
       score,
       b_cheap, b_drought, b_dryup, b_wing,
       trend_up,
       (score >= 3 AND b_cheap = 1 AND trend_up) AS rule_pass,
       ROUND(100.0 * (px_fwd4 / px - 1), 1) AS fwd4w_pct,
       ROUND(100.0 * (px_fwd8 / px - 1), 1) AS fwd8w_pct
FROM final
WHERE wk >= (current_date - %d)
ORDER BY wk
`
