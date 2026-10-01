package amqp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mnsrulz/mzworker-go/handler"
)

type replyEnvelope struct {
	RequestId string `json:"requestId"`
	HasError  bool   `json:"hasError"`
	Value     any    `json:"value"`
}

// columnarValue serializes as {"column": [v, v, ...], ...} preserving the
// query's column order (encoding/json would sort plain map keys).
type columnarValue struct {
	keys []string
	cols map[string][]any
}

func newColumnarValue(columns []string, rows [][]any) columnarValue {
	cols := make(map[string][]any, len(columns))
	for _, name := range columns {
		cols[name] = make([]any, 0, len(rows))
	}
	for _, row := range rows {
		for j, cell := range row {
			if j < len(columns) {
				name := columns[j]
				cols[name] = append(cols[name], toCellValue(cell))
			}
		}
	}
	return columnarValue{keys: columns, cols: cols}
}

func (c columnarValue) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, name := range c.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(c.cols[name])
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(val)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func toCellValue(v any) any {
	switch val := v.(type) {
	case nil, string, float64, int64, bool:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (c *Consumer) handleMessage(sess session, msg amqp.Delivery) {
	var envelope struct {
		RequestType string `json:"requestType"`
		RequestId   string `json:"requestId"`
	}
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		log.Printf("AMQP: failed to parse envelope: %v", err)
		if msg.ReplyTo != "" {
			c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, "")
		}
		_ = msg.Ack(false)
		return
	}

	factory, ok := handler.GetRequestFactory(envelope.RequestType)
	if !ok {
		log.Printf("AMQP: unknown request type: %s", envelope.RequestType)
		if msg.ReplyTo != "" {
			c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, envelope.RequestId)
		}
		_ = msg.Ack(false)
		return
	}

	request, err := factory(msg.Body)
	if err != nil {
		log.Printf("AMQP: failed to create request: %v", err)
		if msg.ReplyTo != "" {
			c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, envelope.RequestId)
		}
		_ = msg.Ack(false)
		return
	}

	resp, err := handler.Dispatch(context.Background(), request)
	if err != nil {
		log.Printf("AMQP: dispatch failed: %v", err)
		if msg.ReplyTo != "" {
			c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, envelope.RequestId)
		}
		_ = msg.Ack(false)
		return
	}

	if msg.ReplyTo == "" {
		log.Printf("AMQP: no reply_to set, dropping response")
		_ = msg.Ack(false)
		return
	}

	amqpResp, err := toAmqpResponse(envelope.RequestId, resp)
	if err != nil {
		log.Printf("AMQP: failed to convert response: %v", err)
		c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, envelope.RequestId)
		_ = msg.Ack(false)
		return
	}
	body, err := json.Marshal(amqpResp)
	if err != nil {
		log.Printf("AMQP: failed to marshal response: %v", err)
		c.publishErrorReply(sess, msg.ReplyTo, msg.CorrelationId, envelope.RequestId)
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

// publishErrorReply sends the reference consumer's failure shape:
// {"requestId": "...", "hasError": true, "value": {}}.
func (c *Consumer) publishErrorReply(s session, replyTo, correlationID, requestId string) {
	body, err := json.Marshal(replyEnvelope{RequestId: requestId, HasError: true, Value: struct{}{}})
	if err != nil {
		log.Printf("AMQP: failed to marshal error reply: %v", err)
		return
	}
	if err := c.publishReply(s, replyTo, correlationID, body); err != nil {
		log.Printf("AMQP: failed to publish error response: %v", err)
	}
}

func toAmqpResponse(requestID string, resp any) (replyEnvelope, error) {
	switch response := resp.(type) {
	case *handler.QueryResponse:
		return replyEnvelope{
			RequestId: requestID,
			Value:     newColumnarValue(response.Columns, response.Rows),
		}, nil
	case *handler.PingResponse:
		return replyEnvelope{RequestId: requestID, Value: response}, nil
	default:
		return replyEnvelope{}, fmt.Errorf("unknown response type: %T", resp)
	}
}
