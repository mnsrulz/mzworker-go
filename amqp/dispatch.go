package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mnsrulz/mzworker-go/handler"
)

type amqpResponse[T any] struct {
	Data  T      `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type amqpQueryResponse struct {
	Columns []string  `json:"columns"`
	Rows    []amqpRow `json:"rows"`
}

type amqpRow struct {
	Values []amqpValue `json:"values"`
}

type amqpValue struct {
	StringVal *string  `json:"string_val,omitempty"`
	DoubleVal *float64 `json:"double_val,omitempty"`
	IntVal    *int64   `json:"int_val,omitempty"`
	BoolVal   *bool    `json:"bool_val,omitempty"`
}

func (c *Consumer) handleMessage(sess session, msg amqp.Delivery) {
	var envelope struct {
		RequestType string `json:"requestType"`
	}
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		log.Printf("AMQP: failed to parse envelope: %v", err)
		if msg.ReplyTo != "" {
			c.publishError(sess, msg.ReplyTo, msg.CorrelationId, fmt.Sprintf("invalid JSON: %v", err))
		}
		_ = msg.Ack(false)
		return
	}

	factory, ok := handler.GetRequestFactory(envelope.RequestType)
	if !ok {
		log.Printf("AMQP: unknown request type: %s", envelope.RequestType)
		_ = msg.Ack(false)
		return
	}

	request, err := factory(msg.Body)
	if err != nil {
		log.Printf("AMQP: failed to create request: %v", err)
		if msg.ReplyTo != "" {
			c.publishError(sess, msg.ReplyTo, msg.CorrelationId, err.Error())
		}
		_ = msg.Ack(false)
		return
	}

	resp, err := handler.Dispatch(context.Background(), request)
	if err != nil {
		log.Printf("AMQP: dispatch failed: %v", err)
		if msg.ReplyTo != "" {
			c.publishError(sess, msg.ReplyTo, msg.CorrelationId, err.Error())
		}
		_ = msg.Ack(false)
		return
	}

	if msg.ReplyTo == "" {
		log.Printf("AMQP: no reply_to set, dropping response")
		_ = msg.Ack(false)
		return
	}

	amqpResp, err := toAmqpResponse(resp)
	if err != nil {
		log.Printf("AMQP: failed to convert response: %v", err)
		_ = msg.Ack(false)
		return
	}
	body, err := json.Marshal(amqpResp)
	if err != nil {
		log.Printf("AMQP: failed to marshal response: %v", err)
		_ = msg.Ack(false)
		return
	}

	if err := c.publishReply(sess, msg.ReplyTo, msg.CorrelationId, body); err != nil {
		log.Printf("AMQP: failed to publish response: %v; requeueing request", err)
		_ = msg.Nack(false, true)
		return
	}

	log.Printf("AMQP: response sent (%T)", resp)
	_ = msg.Ack(false)
}

func (c *Consumer) publishReply(s session, replyTo, correlationID string, body []byte) error {
	if s == nil {
		return errors.New("no active AMQP session")
	}
	return s.Publish("", replyTo, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		Body:          body,
	})
}

func (c *Consumer) publishError(s session, replyTo, correlationId, errMsg string) {
	resp := amqpResponse[any]{Error: errMsg}
	body, _ := json.Marshal(resp)
	if err := c.publishReply(s, replyTo, correlationId, body); err != nil {
		log.Printf("AMQP: failed to publish error response: %v", err)
	}
}

func toAmqpResponse(resp any) (amqpResponse[any], error) {
	switch response := resp.(type) {
	case *handler.QueryResponse:
		queryResponse := amqpQueryResponse{
			Columns: response.Columns,
			Rows:    make([]amqpRow, len(response.Rows)),
		}
		for i, row := range response.Rows {
			vals := make([]amqpValue, len(row))
			for j, v := range row {
				vals[j] = toAmqpValue(v)
			}
			queryResponse.Rows[i] = amqpRow{Values: vals}
		}
		return amqpResponse[any]{Data: queryResponse}, nil
	case *handler.PingResponse:
		return amqpResponse[any]{Data: response}, nil
	default:
		return amqpResponse[any]{}, fmt.Errorf("unknown response type: %T", resp)
	}
}

func toAmqpValue(v interface{}) amqpValue {
	if v == nil {
		return amqpValue{}
	}
	switch val := v.(type) {
	case string:
		return amqpValue{StringVal: &val}
	case float64:
		return amqpValue{DoubleVal: &val}
	case int64:
		return amqpValue{IntVal: &val}
	case bool:
		return amqpValue{BoolVal: &val}
	default:
		s := fmt.Sprintf("%v", val)
		return amqpValue{StringVal: &s}
	}
}
