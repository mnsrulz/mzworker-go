package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

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

	response, err := toAmqpResponse("req-1", &handler.PingResponse{
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

	const expected = `{"requestId":"req-1","hasError":false,"value":{"message":"pong","server_time":"2026-09-18T23:11:05Z"}}`
	if string(body) != expected {
		t.Fatalf("unexpected response: %s", body)
	}
}

type fakeAcknowledger struct {
	mu      sync.Mutex
	acks    int
	nacks   int
	rejects int
	requeue bool
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.acks++
	return nil
}

func (f *fakeAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nacks++
	f.requeue = requeue
	return nil
}

func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejects++
	f.requeue = requeue
	return nil
}

type fakeSession struct {
	deliveries chan amqp.Delivery
	stateCh    chan *amqp.StateChanged
	consumed   chan struct{}
	consumeErr error

	mu         sync.Mutex
	publishErr error
	published  []amqp.Publishing
	replyTos   []string
	closed     bool
}

func newFakeSession() *fakeSession {
	return &fakeSession{
		deliveries: make(chan amqp.Delivery),
		stateCh:    make(chan *amqp.StateChanged, 8),
		consumed:   make(chan struct{}),
	}
}

func (f *fakeSession) Consume(queue string) (<-chan amqp.Delivery, error) {
	if f.consumeErr != nil {
		return nil, f.consumeErr
	}
	select {
	case <-f.consumed:
	default:
		close(f.consumed)
	}
	return f.deliveries, nil
}

func (f *fakeSession) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.replyTos = append(f.replyTos, key)
	if f.publishErr != nil {
		return f.publishErr
	}
	f.published = append(f.published, msg)
	return nil
}

func (f *fakeSession) StateChanges() <-chan *amqp.StateChanged {
	return f.stateCh
}

func (f *fakeSession) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
}

func (f *fakeSession) isClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func waitFor(t *testing.T, timeout time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func newTestConsumer(sess session) *Consumer {
	return &Consumer{
		queueName:    "test-queue",
		concurrency:  1,
		sess:         sess,
		sem:          make(chan struct{}, 1),
		drainTimeout: time.Second,
	}
}

func sessionConsumed(s *fakeSession) func() bool {
	return func() bool {
		select {
		case <-s.consumed:
			return true
		default:
			return false
		}
	}
}

func TestNewConsumerFailsFastWhenBrokerUnreachable(t *testing.T) {
	if _, err := NewConsumer("", "q", 1); err == nil {
		t.Fatal("expected validation error for empty URL")
	}
	if _, err := NewConsumer("amqp://127.0.0.1:1", "q", 1); err == nil {
		t.Fatal("expected constructor to fail fast when broker is unreachable")
	}
}

func TestStartStopDrainsAndClosesConnection(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	c.Start(context.Background())
	waitFor(t, 2*time.Second, "session to consume", sessionConsumed(sess))
	c.Stop()

	if !sess.isClosed() {
		t.Fatal("expected connection closed on Stop")
	}
	select {
	case <-c.done:
	default:
		t.Fatal("expected consuming goroutine to have exited")
	}
}

func TestSessionLoopExitsCleanlyOnContextCancel(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.sessionLoop(ctx) }()
	waitFor(t, 2*time.Second, "session to consume", sessionConsumed(sess))

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("expected nil on graceful cancel, got %v", err)
	}
}

func TestSessionLoopReturnsTerminalReasonOnDeliveryChannelClose(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	done := make(chan error, 1)
	go func() { done <- c.sessionLoop(context.Background()) }()
	waitFor(t, 2*time.Second, "session to consume", sessionConsumed(sess))

	close(sess.deliveries)
	if err := <-done; err == nil {
		t.Fatal("expected terminal reason when delivery channel closes")
	}
}

func TestSessionLoopReturnsTerminalReasonOnStateClosed(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	done := make(chan error, 1)
	go func() { done <- c.sessionLoop(context.Background()) }()
	waitFor(t, 2*time.Second, "session to consume", sessionConsumed(sess))

	sess.stateCh <- &amqp.StateChanged{From: amqp.StateReconnecting, To: amqp.StateClosed, Err: errors.New("recovery exhausted")}
	err := <-done
	if err == nil || err.Error() != "connection recovery terminated: recovery exhausted" {
		t.Fatalf("expected wrapped terminal error, got %v", err)
	}
}

func TestNonTerminalStateDoesNotTearDown(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- c.sessionLoop(ctx) }()
	waitFor(t, 2*time.Second, "session to consume", sessionConsumed(sess))

	sess.stateCh <- &amqp.StateChanged{From: amqp.StateOpen, To: amqp.StateReconnecting}
	sess.stateCh <- &amqp.StateChanged{From: amqp.StateReconnecting, To: amqp.StateOpen}

	select {
	case err := <-done:
		t.Fatalf("session loop must not exit on non-terminal states, got %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if sess.isClosed() {
		t.Fatal("expected session untouched while library recovers")
	}
}

func newTestDelivery(ack amqp.Acknowledger) amqp.Delivery {
	return amqp.Delivery{
		Acknowledger:  ack,
		DeliveryTag:   1,
		ReplyTo:       "reply.q",
		CorrelationId: "corr-1",
		Body:          []byte(`{"requestType":"ping"}`),
	}
}

func TestHandleMessageNacksWhenReplyPublishFails(t *testing.T) {
	if err := handler.Init(t.TempDir()); err != nil {
		t.Fatalf("handler.Init: %v", err)
	}
	sess := newFakeSession()
	sess.publishErr = errors.New("publish failed")
	c := newTestConsumer(sess)

	ack := &fakeAcknowledger{}
	c.handleMessage(sess, newTestDelivery(ack))

	ack.mu.Lock()
	defer ack.mu.Unlock()
	if ack.nacks != 1 {
		t.Fatalf("expected 1 nack, got %d", ack.nacks)
	}
	if !ack.requeue {
		t.Fatal("expected nack with requeue=true")
	}
	if ack.acks != 0 {
		t.Fatalf("expected no ack, got %d", ack.acks)
	}
}

func TestHandleMessageAcksWhenReplyPublishSucceeds(t *testing.T) {
	if err := handler.Init(t.TempDir()); err != nil {
		t.Fatalf("handler.Init: %v", err)
	}
	sess := newFakeSession()
	c := newTestConsumer(sess)

	ack := &fakeAcknowledger{}
	c.handleMessage(sess, newTestDelivery(ack))

	ack.mu.Lock()
	acks := ack.acks
	nacks := ack.nacks
	ack.mu.Unlock()
	if acks != 1 || nacks != 0 {
		t.Fatalf("expected 1 ack and 0 nacks, got %d acks and %d nacks", acks, nacks)
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()
	if len(sess.replyTos) != 1 || sess.replyTos[0] != "reply.q" {
		t.Fatalf("expected reply published to reply.q, got %v", sess.replyTos)
	}
	if len(sess.published) != 1 {
		t.Fatalf("expected 1 published reply, got %d", len(sess.published))
	}

	var env replyEnvelope
	if err := json.Unmarshal(sess.published[0].Body, &env); err != nil {
		t.Fatalf("failed to unmarshal reply envelope: %v", err)
	}
	if env.RequestId != "" {
		t.Fatalf("expected empty requestId (body has none), got %q", env.RequestId)
	}
	if env.HasError {
		t.Fatal("expected hasError=false on success")
	}
	if env.Value == nil {
		t.Fatal("expected value to be present")
	}
}

func TestHandleMessageNacksWhenNoSession(t *testing.T) {
	if err := handler.Init(t.TempDir()); err != nil {
		t.Fatalf("handler.Init: %v", err)
	}
	c := newTestConsumer(nil)

	ack := &fakeAcknowledger{}
	c.handleMessage(nil, newTestDelivery(ack))

	ack.mu.Lock()
	defer ack.mu.Unlock()
	if ack.nacks != 1 || !ack.requeue {
		t.Fatalf("expected requeue nack, got nacks=%d requeue=%v", ack.nacks, ack.requeue)
	}
}

func TestToAmqpResponseColumnarExactJSON(t *testing.T) {
	resp := &handler.QueryResponse{
		Columns: []string{"dt", "last_close", "straddle_price", "expiry"},
		Rows: [][]any{
			{"2025-10-20", 336.02, 58.83, "2025-11-21"},
			{"2025-10-27", nil, 59.1, "2025-11-21"},
		},
	}

	envelope, err := toAmqpResponse("req-1", resp)
	if err != nil {
		t.Fatalf("toAmqpResponse: %v", err)
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	const want = `{"requestId":"req-1","hasError":false,"value":{"dt":["2025-10-20","2025-10-27"],"last_close":[336.02,null],"straddle_price":[58.83,59.1],"expiry":["2025-11-21","2025-11-21"]}}`
	if string(body) != want {
		t.Fatalf("unexpected columnar reply:\n got: %s\nwant: %s", body, want)
	}
}

func TestHandleMessageRepliesErrorEnvelopeOnInvalidJSON(t *testing.T) {
	sess := newFakeSession()
	c := newTestConsumer(sess)

	ack := &fakeAcknowledger{}
	delivery := amqp.Delivery{
		Acknowledger:  ack,
		DeliveryTag:   1,
		ReplyTo:       "reply.q",
		CorrelationId: "corr-1",
		Body:          []byte("not json"),
	}
	c.handleMessage(sess, delivery)

	sess.mu.Lock()
	body := string(sess.published[0].Body)
	sess.mu.Unlock()
	const want = `{"requestId":"","hasError":true,"value":{}}`
	if body != want {
		t.Fatalf("unexpected error reply: got %s, want %s", body, want)
	}

	ack.mu.Lock()
	defer ack.mu.Unlock()
	if ack.acks != 1 || ack.nacks != 0 {
		t.Fatalf("expected 1 ack and 0 nacks, got %d acks and %d nacks", ack.acks, ack.nacks)
	}
}

func TestHandleMessageEchoesRequestIdOnSuccess(t *testing.T) {
	if err := handler.Init(t.TempDir()); err != nil {
		t.Fatalf("handler.Init: %v", err)
	}
	sess := newFakeSession()
	c := newTestConsumer(sess)

	ack := &fakeAcknowledger{}
	delivery := amqp.Delivery{
		Acknowledger:  ack,
		DeliveryTag:   1,
		ReplyTo:       "reply.q",
		CorrelationId: "corr-1",
		Body:          []byte(`{"requestType":"ping","requestId":"d14eb4a0-be85-4034-a37f-969c90a9f024"}`),
	}
	c.handleMessage(sess, delivery)

	sess.mu.Lock()
	var env replyEnvelope
	if err := json.Unmarshal(sess.published[0].Body, &env); err != nil {
		sess.mu.Unlock()
		t.Fatalf("failed to unmarshal reply: %v", err)
	}
	sess.mu.Unlock()

	if env.RequestId != "d14eb4a0-be85-4034-a37f-969c90a9f024" {
		t.Fatalf("expected requestId echoed, got %q", env.RequestId)
	}
	if env.HasError {
		t.Fatal("expected hasError=false")
	}
	if env.Value == nil {
		t.Fatal("expected value present")
	}
}
