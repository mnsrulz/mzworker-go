package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mehdihadeli/go-mediatr"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mnsrulz/mzworker-go/handler"
)

type Consumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	queueName   string
	concurrency int
	sem         chan struct{}
}

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

func NewConsumer(amqpURL, queueName string, concurrency int) (*Consumer, error) {
	if concurrency <= 0 {
		concurrency = 5
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AMQP: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:        conn,
		channel:     ch,
		queueName:   q.Name,
		concurrency: concurrency,
		sem:         make(chan struct{}, concurrency),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	msgs, err := c.channel.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register AMQP consumer: %v", err)
	}

	log.Printf("AMQP consumer started on queue: %s (concurrency: %d)", c.queueName, c.concurrency)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.sem <- struct{}{}
				go func() {
					defer func() { <-c.sem }()
					c.handleMessage(msg)
				}()
			}
		}
	}()
}

func (c *Consumer) Stop() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	var envelope struct {
		RequestType string `json:"requestType"`
	}
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		log.Printf("AMQP: failed to parse envelope: %v", err)
		if msg.ReplyTo != "" {
			c.publishError(msg.ReplyTo, msg.CorrelationId, fmt.Sprintf("invalid JSON: %v", err))
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
			c.publishError(msg.ReplyTo, msg.CorrelationId, err.Error())
		}
		_ = msg.Ack(false)
		return
	}

	resp, err := dispatchMediatr(context.Background(), request)
	if err != nil {
		log.Printf("AMQP: dispatch failed: %v", err)
		if msg.ReplyTo != "" {
			c.publishError(msg.ReplyTo, msg.CorrelationId, err.Error())
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

	err = c.channel.Publish("", msg.ReplyTo, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: msg.CorrelationId,
		Body:          body,
	})
	if err != nil {
		log.Printf("AMQP: failed to publish response: %v", err)
		_ = msg.Ack(false)
		return
	}

	log.Printf("AMQP: response sent (%T)", resp)
	_ = msg.Ack(false)
}

func dispatchMediatr(ctx context.Context, request any) (any, error) {
	switch req := request.(type) {
	case *handler.DynamicSQLQuery:
		return mediatr.Send[*handler.DynamicSQLQuery, *handler.QueryResponse](ctx, req)
	case *handler.OHLCQuery:
		return mediatr.Send[*handler.OHLCQuery, *handler.QueryResponse](ctx, req)
	case *handler.VolatilityQuery:
		return mediatr.Send[*handler.VolatilityQuery, *handler.QueryResponse](ctx, req)
	case *handler.OptionsStatQuery:
		return mediatr.Send[*handler.OptionsStatQuery, *handler.QueryResponse](ctx, req)
	case *handler.ExpectedMoveQuery:
		return mediatr.Send[*handler.ExpectedMoveQuery, *handler.QueryResponse](ctx, req)
	case *handler.PingRequest:
		return mediatr.Send[*handler.PingRequest, *handler.PingResponse](ctx, req)
	default:
		return nil, fmt.Errorf("unknown request type: %T", request)
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

func (c *Consumer) publishError(replyTo, correlationId, errMsg string) {
	resp := amqpResponse[any]{Error: errMsg}
	body, _ := json.Marshal(resp)
	_ = c.channel.Publish("", replyTo, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationId,
		Body:          body,
	})
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
