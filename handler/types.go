package handler

import (
	"fmt"

	z "github.com/Oudwins/zog"
)

type DynamicSQLQuery struct {
	Symbol string `json:"symbol"`
	SQL    string `json:"sql"`
	Limit  int    `json:"limit"`
}

type OHLCQuery struct {
	Symbol       string `json:"symbol"`
	LookbackDays int    `json:"lookbackDays"`
}

type VolatilityQuery struct {
	Symbol       string `json:"symbol"`
	LookbackDays int    `json:"lookbackDays"`
	Mode         string `json:"mode"`
	Delta        int    `json:"delta"`
	Strike       int    `json:"strike"`
	ExpiryMode   string `json:"expiryMode"`
	DTE          int    `json:"dte"`
	Expiration   string `json:"expiration"`
}

type OptionsStatQuery struct {
	Symbol       string `json:"symbol"`
	LookbackDays int    `json:"lookbackDays"`
}

type ExpectedMoveQuery struct {
	Symbol       string `json:"symbol"`
	LookbackDays int    `json:"lookbackDays"`
	ExpiryMode   string `json:"expiryMode"`
}

type BullRunSignalQuery struct {
	Symbol       string `json:"symbol"`
	LookbackDays int    `json:"lookbackDays"`
}

type QueryResponse struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

var DynamicSQLQuerySchema = z.Struct(z.Shape{
	"Symbol": z.String().Required(),
	"SQL":    z.String().Required(),
	"Limit":  z.Int().Optional(),
})

var OHLCQuerySchema = z.Struct(z.Shape{
	"Symbol":       z.String().Required(),
	"LookbackDays": z.Int().Required(),
})

func (r *DynamicSQLQuery) Validate() error {
	if errs := DynamicSQLQuerySchema.Validate(r); errs != nil {
		return fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}
	return nil
}

func (r *OHLCQuery) Validate() error {
	if errs := OHLCQuerySchema.Validate(r); errs != nil {
		return fmt.Errorf("validation failed: %s", z.Issues.Prettify(errs))
	}
	if r.LookbackDays <= 0 {
		return fmt.Errorf("validation failed: LookbackDays must be positive")
	}
	return nil
}
