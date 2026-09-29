package amqp

import (
	"fmt"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// defaultRetryInterval is handed to the library's own recovery config.
const defaultRetryInterval = 5 * time.Second

type session interface {
	Consume(queue string) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	StateChanges() <-chan *amqp.StateChanged
	Close()
}

func dialSession(amqpURL, queueName string) (session, error) {
	cfg := amqp.Config{
		Recovery: &amqp.Recovery{
			ReconnectionConfig: &amqp.ReconnectionConfig{
				MaxRetryCount: math.MaxInt32,
				RetryInterval: defaultRetryInterval,
			},
		},
	}
	conn, err := amqp.DialConfig(amqpURL, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AMQP: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	stateCh := make(chan *amqp.StateChanged, 8)
	conn.NotifyStateChange(stateCh)
	return &amqpConnSession{
		conn:    conn,
		channel: ch,
		stateCh: stateCh,
	}, nil
}

type amqpConnSession struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	stateCh <-chan *amqp.StateChanged
}

func (s *amqpConnSession) Consume(queue string) (<-chan amqp.Delivery, error) {
	return s.channel.Consume(queue, "", false, false, false, false, nil)
}

func (s *amqpConnSession) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return s.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (s *amqpConnSession) StateChanges() <-chan *amqp.StateChanged {
	return s.stateCh
}

func (s *amqpConnSession) Close() {
	_ = s.channel.Close()
	_ = s.conn.Close()
}
