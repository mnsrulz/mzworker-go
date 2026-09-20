package handler

import z "github.com/Oudwins/zog"

type DynamicSQLQuery struct {
	Symbol string `json:"symbol"`
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
}

type OHLCQuery struct {
	Symbol string `json:"symbol"`
	From   string `json:"from"`
	To     string `json:"to"`
	Limit  int    `json:"limit"`
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

type QueryResponse struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

var DynamicSQLQuerySchema = z.Struct(z.Shape{
	"Symbol": z.String().Required(),
	"Query":  z.String().Required(),
	"Limit":  z.Int().Optional(),
})

var OHLCQuerySchema = z.Struct(z.Shape{
	"Symbol": z.String().Required(),
	"From":   z.String().Optional(),
	"To":     z.String().Optional(),
	"Limit":  z.Int().Optional(),
})
