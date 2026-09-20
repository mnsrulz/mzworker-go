package amqp

import (
	"encoding/json"
	"testing"

	"github.com/mnsrulz/mzworker-go/handler"
)

func TestPingAMQPRequestAndResponse(t *testing.T) {
	factory, ok := handler.GetRequestFactory("ping")
	if !ok {
		t.Fatal("expected ping request factory")
	}

	request, err := factory([]byte(`{"requestType":"ping"}`))
	if err != nil {
		t.Fatalf("failed to create ping request: %v", err)
	}
	if _, ok := request.(*handler.PingRequest); !ok {
		t.Fatalf("expected *handler.PingRequest, got %T", request)
	}

	response, err := toAmqpResponse(&handler.PingResponse{
		Message:    "pong",
		ServerTime: "2026-09-18T23:11:05Z",
	})
	if err != nil {
		t.Fatalf("failed to create ping response: %v", err)
	}

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal ping response: %v", err)
	}

	const expected = `{"data":{"message":"pong","server_time":"2026-09-18T23:11:05Z"}}`
	if string(body) != expected {
		t.Fatalf("unexpected response: %s", body)
	}
}
