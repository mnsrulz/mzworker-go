package handler

import "encoding/json"

type RequestFactory func(payload []byte) (any, error)

func jsonRequestFactory[T any](payload []byte) (any, error) {
	var request T
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, err
	}
	return &request, nil
}

var registry = map[string]RequestFactory{
	"dynamic-sql-query":   jsonRequestFactory[DynamicSQLQuery],
	"ohlc-query":          jsonRequestFactory[OHLCQuery],
	"volatility-query":    jsonRequestFactory[VolatilityQuery],
	"options-stat-query":  jsonRequestFactory[OptionsStatQuery],
	"expected-move-query": jsonRequestFactory[ExpectedMoveQuery],
	"ping":                jsonRequestFactory[PingRequest],
}

func GetRequestFactory(requestType string) (RequestFactory, bool) {
	f, ok := registry[requestType]
	return f, ok
}
