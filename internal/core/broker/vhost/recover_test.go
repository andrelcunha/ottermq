package vhost

import (
	"context"
	"net"
	"testing"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

func TestHandleBasicRecover_RequeueTrue(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)
	var conn net.Conn = nil

	// Create queue
	q, err := vh.CreateQueue("q1", nil)
	if err != nil {
		t.Fatalf("CreateQueue failed: %v", err)
	}

	// Setup channel delivery state with unacked messages
	key := ConnectionChannelKey{conn, 1}
	ch := &ChannelDeliveryState{Unacked: make(map[uint64]*DeliveryRecord)}
	vh.mu.Lock()
	vh.ChannelDeliveries[key] = ch
	vh.mu.Unlock()

	// Add unacked messages
	m1 := amqp.Message{ID: "msg1", Body: []byte("test1")}
	m2 := amqp.Message{ID: "msg2", Body: []byte("test2")}

	ch.mu.Lock()
	ch.Unacked[1] = &DeliveryRecord{
		DeliveryTag: 1,
		ConsumerTag: "ctag1",
		QueueName:   "q1",
		Message:     m1,
	}
	ch.Unacked[2] = &DeliveryRecord{
		DeliveryTag: 2,
		ConsumerTag: "ctag1",
		QueueName:   "q1",
		Message:     m2,
	}
	ch.mu.Unlock()

	// Call recover with requeue=true
	if err := vh.HandleBasicRecover(conn, 1, true); err != nil {
		t.Fatalf("HandleBasicRecover failed: %v", err)
	}

	// Verify unacked is cleared
	ch.mu.Lock()
	unackedCount := len(ch.Unacked)
	ch.mu.Unlock()
	if unackedCount != 0 {
		t.Errorf("expected 0 unacked after recover, got %d", unackedCount)
	}

	// Verify messages requeued
	if q.Len() != 2 {
		t.Errorf("expected 2 messages in queue, got %d", q.Len())
	}

	// Verify messages marked as redelivered
	if !vh.shouldRedeliver("msg1") {
		t.Error("msg1 should be marked for redelivery")
	}
	if !vh.shouldRedeliver("msg2") {
		t.Error("msg2 should be marked for redelivery")
	}
}

func TestHandleBasicRecover_RequeueFalse_ConsumerExists(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)
	var conn net.Conn = nil

	// Create queue
	if _, err := vh.CreateQueue("q1", nil); err != nil {
		t.Fatalf("CreateQueue failed: %v", err)
	}

	// Register consumer
	consumer := &Consumer{
		Tag:        "ctag1",
		Channel:    1,
		QueueName:  "q1",
		Connection: conn,
		Active:     true,
		Props:      &ConsumerProperties{NoAck: false},
	}

	consumerKey := ConsumerKey{Channel: 1, Tag: "ctag1"}
	vh.mu.Lock()
	vh.Consumers[consumerKey] = consumer
	channelKey := ConnectionChannelKey{conn, 1}
	vh.ConsumersByChannel[channelKey] = []*Consumer{consumer}
	vh.mu.Unlock()

	// Setup mock framer that always succeeds
	vh.framer = &mockFramer{}

	// Setup channel delivery state with unacked message
	ch := &ChannelDeliveryState{Unacked: make(map[uint64]*DeliveryRecord)}
	vh.mu.Lock()
	vh.ChannelDeliveries[channelKey] = ch
	vh.mu.Unlock()

	m1 := amqp.Message{ID: "msg1", Body: []byte("test1")}
	ch.mu.Lock()
	ch.LastDeliveryTag = 1 // Set counter so next delivery gets tag 2
	ch.Unacked[1] = &DeliveryRecord{
		DeliveryTag: 1,
		ConsumerTag: "ctag1",
		QueueName:   "q1",
		Message:     m1,
	}
	ch.mu.Unlock()

	// Call recover with requeue=false
	if err := vh.HandleBasicRecover(conn, 1, false); err != nil {
		t.Fatalf("HandleBasicRecover failed: %v", err)
	}

	// Verify new delivery was created with new tag (old tag 1 was cleared, new tag assigned)
	ch.mu.Lock()
	_, oldExists := ch.Unacked[1]
	newUnackedCount := len(ch.Unacked)
	var newTag uint64
	for tag := range ch.Unacked {
		newTag = tag
		break
	}
	ch.mu.Unlock()

	if oldExists {
		t.Error("old delivery tag 1 should be cleared")
	}
	if newUnackedCount != 1 {
		t.Errorf("expected 1 new unacked after redelivery, got %d", newUnackedCount)
	}
	if newTag != 2 {
		t.Errorf("expected new delivery tag 2 (ch.LastDeliveryTag incremented from 1), got %d", newTag)
	}
}

func TestHandleBasicRecover_RequeueFalse_ConsumerGone(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)
	var conn net.Conn = nil

	// Create queue
	q, err := vh.CreateQueue("q1", nil)
	if err != nil {
		t.Fatalf("CreateQueue failed: %v", err)
	}

	// Setup channel delivery state with unacked message, but NO consumer
	key := ConnectionChannelKey{conn, 1}
	ch := &ChannelDeliveryState{Unacked: make(map[uint64]*DeliveryRecord)}
	vh.mu.Lock()
	vh.ChannelDeliveries[key] = ch
	vh.mu.Unlock()

	m1 := amqp.Message{ID: "msg1", Body: []byte("test1")}
	ch.mu.Lock()
	ch.Unacked[1] = &DeliveryRecord{
		DeliveryTag: 1,
		ConsumerTag: "ctag-missing",
		QueueName:   "q1",
		Message:     m1,
	}
	ch.mu.Unlock()

	// Call recover with requeue=false
	if err := vh.HandleBasicRecover(conn, 1, false); err != nil {
		t.Fatalf("HandleBasicRecover failed: %v", err)
	}

	// Verify unacked is cleared
	ch.mu.Lock()
	unackedCount := len(ch.Unacked)
	ch.mu.Unlock()
	if unackedCount != 0 {
		t.Errorf("expected 0 unacked after recover, got %d", unackedCount)
	}

	// Verify message requeued (fallback when consumer is gone)
	if q.Len() != 1 {
		t.Errorf("expected 1 message requeued, got %d", q.Len())
	}

	// Verify message marked as redelivered
	if !vh.shouldRedeliver("msg1") {
		t.Error("msg1 should be marked for redelivery")
	}
}

func TestHandleBasicRecover_RequeueFalse_DeliveryFails(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)
	var conn net.Conn = nil

	// Create queue
	q, err := vh.CreateQueue("q1", nil)
	if err != nil {
		t.Fatalf("CreateQueue failed: %v", err)
	}

	// Register consumer
	consumer := &Consumer{
		Tag:        "ctag1",
		Channel:    1,
		QueueName:  "q1",
		Connection: conn,
		Active:     true,
		Props:      &ConsumerProperties{NoAck: false},
	}

	consumerKey := ConsumerKey{Channel: 1, Tag: "ctag1"}
	vh.mu.Lock()
	vh.Consumers[consumerKey] = consumer
	channelKey := ConnectionChannelKey{conn, 1}
	vh.ConsumersByChannel[channelKey] = []*Consumer{consumer}
	vh.mu.Unlock()

	// Setup mock framer that FAILS
	vh.framer = &mockFramerFail{}

	// Setup channel delivery state with unacked message
	ch := &ChannelDeliveryState{Unacked: make(map[uint64]*DeliveryRecord)}
	vh.mu.Lock()
	vh.ChannelDeliveries[channelKey] = ch
	vh.mu.Unlock()

	m1 := amqp.Message{ID: "msg1", Body: []byte("test1")}
	ch.mu.Lock()
	ch.Unacked[1] = &DeliveryRecord{
		DeliveryTag: 1,
		ConsumerTag: "ctag1",
		QueueName:   "q1",
		Message:     m1,
	}
	ch.mu.Unlock()

	// Call recover with requeue=false
	if err := vh.HandleBasicRecover(conn, 1, false); err != nil {
		t.Fatalf("HandleBasicRecover failed: %v", err)
	}

	// Verify old unacked is cleared
	ch.mu.Lock()
	unackedCount := len(ch.Unacked)
	ch.mu.Unlock()
	if unackedCount != 0 {
		t.Errorf("expected 0 unacked after failed redelivery, got %d", unackedCount)
	}

	// Verify message requeued (fallback on delivery failure)
	if q.Len() != 1 {
		t.Errorf("expected 1 message requeued after failed delivery, got %d", q.Len())
	}

	// Verify message marked as redelivered
	if !vh.shouldRedeliver("msg1") {
		t.Error("msg1 should be marked for redelivery after failed delivery")
	}
}

func TestHandleBasicRecover_NoChannelState(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)
	var conn net.Conn = nil

	// Call recover without setting up channel state
	err := vh.HandleBasicRecover(conn, 1, true)
	if err == nil {
		t.Error("expected error when channel state missing, got nil")
	}
}

func TestRedeliveredMarkLifecycle(t *testing.T) {
	vh := NewVhost("test-vhost", 1000, nil)

	// Mark a message as redelivered
	vh.markAsRedelivered("msg1")

	// Verify it's marked
	if !vh.shouldRedeliver("msg1") {
		t.Error("msg1 should be marked for redelivery")
	}

	// Clear the mark
	vh.clearRedeliveredMark("msg1")

	// Verify it's cleared
	if vh.shouldRedeliver("msg1") {
		t.Error("msg1 should not be marked after clearing")
	}
}

// Mock framer that succeeds
type mockFramer struct{}

func (m *mockFramer) ReadFrame(conn net.Conn) ([]byte, error) { return nil, nil }
func (m *mockFramer) ParseFrame(frame []byte) (any, error)    { return nil, nil }
func (m *mockFramer) Handshake(configurations *map[string]any, conn net.Conn, connCtxt context.Context) (*amqp.ConnectionInfo, error) {
	return nil, nil
}
func (m *mockFramer) CreateBasicReturnFrame(channel uint16, replyCode uint16, replyText, exchange, routingKey string) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBasicDeliverFrame(channel uint16, consumerTag, exchange, routingKey string, deliveryTag uint64, redelivered bool) []byte {
	return []byte{}
}
func (m *mockFramer) CreateHeaderFrame(channel, classID uint16, msg amqp.Message) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBodyFrame(channel uint16, body []byte) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBasicGetEmptyFrame(channel uint16) []byte { return []byte{} }
func (m *mockFramer) CreateBasicGetOkFrame(channel uint16, exchange, routingkey string, msgCount uint32) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBasicConsumeOkFrame(channel uint16, consumerTag string) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBasicCancelOkFrame(channel uint16, consumerTag string) []byte {
	return []byte{}
}
func (m *mockFramer) CreateBasicRecoverOkFrame(channel uint16) []byte {
	return []byte{}
}
func (m *mockFramer) CreateQueueDeclareOkFrame(request *amqp.RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte {
	return []byte{}
}
func (m *mockFramer) CreateQueueBindOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramer) CreateQueueDeleteOkFrame(request *amqp.RequestMethodMessage, messageCount uint32) []byte {
	return []byte{}
}
func (m *mockFramer) CreateExchangeDeclareFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramer) CreateExchangeDeleteFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramer) CreateChannelOpenOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramer) CreateChannelCloseOkFrame(channel uint16) []byte { return []byte{} }
func (m *mockFramer) CreateConnectionCloseOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramer) CreateCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod uint16, replyText string) []byte {
	return []byte{}
}
func (m *mockFramer) SendFrame(conn net.Conn, frame []byte) error {
	return nil // Success
}

// Mock framer that fails
type mockFramerFail struct{}

func (m *mockFramerFail) ReadFrame(conn net.Conn) ([]byte, error) { return nil, nil }
func (m *mockFramerFail) ParseFrame(frame []byte) (any, error)    { return nil, nil }
func (m *mockFramerFail) Handshake(configurations *map[string]any, conn net.Conn, connCtxt context.Context) (*amqp.ConnectionInfo, error) {
	return nil, nil
}
func (m *mockFramerFail) CreateBasicReturnFrame(channel uint16, replyCode uint16, replyText, exchange, routingKey string) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBasicDeliverFrame(channel uint16, consumerTag, exchange, routingKey string, deliveryTag uint64, redelivered bool) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateHeaderFrame(channel, classID uint16, msg amqp.Message) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBodyFrame(channel uint16, body []byte) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBasicGetEmptyFrame(channel uint16) []byte { return []byte{} }
func (m *mockFramerFail) CreateBasicGetOkFrame(channel uint16, exchange, routingkey string, msgCount uint32) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBasicConsumeOkFrame(channel uint16, consumerTag string) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBasicCancelOkFrame(channel uint16, consumerTag string) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateBasicRecoverOkFrame(channel uint16) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateQueueDeclareOkFrame(request *amqp.RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateQueueBindOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateQueueDeleteOkFrame(request *amqp.RequestMethodMessage, messageCount uint32) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateExchangeDeclareFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateExchangeDeleteFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateChannelOpenOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateChannelCloseOkFrame(channel uint16) []byte { return []byte{} }
func (m *mockFramerFail) CreateConnectionCloseOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte{}
}
func (m *mockFramerFail) CreateCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod uint16, replyText string) []byte {
	return []byte{}
}
func (m *mockFramerFail) SendFrame(conn net.Conn, frame []byte) error {
	return &net.OpError{Op: "write", Err: net.ErrClosed}
}
