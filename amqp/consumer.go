package amqp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultDrainTimeout = 30 * time.Second

type Consumer struct {
	queueName    string
	concurrency  int
	sess         session
	sem          chan struct{}
	drainTimeout time.Duration

	cancel   context.CancelFunc
	done     chan struct{}
	handlers sync.WaitGroup
}

// NewConsumer validates input and dials the broker exactly once. There is no
// dial retry: a failure here is returned so the application fails to boot.
// All runtime reconnection is handled by the library's built-in recovery.
func NewConsumer(amqpURL, queueName string, concurrency int) (*Consumer, error) {
	if amqpURL == "" || queueName == "" {
		return nil, errors.New("amqp URL and queue name are required")
	}
	if concurrency <= 0 {
		concurrency = 5
	}
	amqp.SetLogger(log.Default())
	sess, err := dialSession(amqpURL, queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AMQP: %w", err)
	}
	return &Consumer{
		queueName:    queueName,
		concurrency:  concurrency,
		sess:         sess,
		sem:          make(chan struct{}, concurrency),
		drainTimeout: defaultDrainTimeout,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})
	go func() {
		defer close(c.done)
		log.Printf("AMQP consumer started on queue: %s (concurrency: %d)", c.queueName, c.concurrency)
		reason := c.sessionLoop(runCtx)
		c.drain()
		c.sess.Close()
		if reason != nil && runCtx.Err() == nil {
			log.Fatalf("AMQP: connection closed terminally: %v", reason)
		}
	}()
}

func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.done != nil {
		<-c.done
	}
}

// sessionLoop consumes until ctx is cancelled (returns nil) or the connection
// dies terminally (returns the reason; the caller fails the process).
func (c *Consumer) sessionLoop(ctx context.Context) error {
	sess := c.sess
	msgs, err := sess.Consume(c.queueName)
	if err != nil {
		return fmt.Errorf("failed to register AMQP consumer: %w", err)
	}
	states := sess.StateChanges()

	for {
		select {
		case <-ctx.Done():
			return nil
		case sc, ok := <-states:
			if !ok {
				return errors.New("connection state listener closed")
			}
			if sc.To == amqp.StateClosed {
				if sc.Err != nil {
					return fmt.Errorf("connection recovery terminated: %w", sc.Err)
				}
				return errors.New("connection closed")
			}
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("delivery channel closed")
			}
			select {
			case <-ctx.Done():
				return nil
			case c.sem <- struct{}{}:
			}
			c.handlers.Add(1)
			go func(msg amqp.Delivery) {
				defer c.handlers.Done()
				defer func() { <-c.sem }()
				c.handleMessage(sess, msg)
			}(msg)
		}
	}
}

func (c *Consumer) drain() {
	done := make(chan struct{})
	go func() {
		c.handlers.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(c.drainTimeout):
		log.Printf("AMQP: timed out after %s waiting for in-flight handlers", c.drainTimeout)
	}
}
