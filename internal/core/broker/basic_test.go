package broker

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/andrelcunha/ottermq/pkg/persistence/implementations/dummy"
)

// MockConnection implements net.Conn for testing
type MockConnection struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	closed     bool
}

func (m *MockConnection) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConnection) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *MockConnection) Close() error                       { m.closed = true; return nil }
func (m *MockConnection) LocalAddr() net.Addr                { return m.localAddr }
func (m *MockConnection) RemoteAddr() net.Addr               { return m.remoteAddr }
func (m *MockConnection) SetDeadline(t time.Time) error      { return nil }
func (m *MockConnection) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConnection) SetWriteDeadline(t time.Time) error { return nil }

// MockAddr implements net.Addr for testing
type MockAddr struct {
	network string
	address string
}

func (m *MockAddr) Network() string { return m.network }
func (m *MockAddr) String() string  { return m.address }

// MockFramer implements amqp.Framer for testing
type MockFramer struct {
	sentFrames [][]byte
}

func (m *MockFramer) ReadFrame(conn net.Conn) ([]byte, error) { return nil, nil }
func (m *MockFramer) SendFrame(conn net.Conn, frame []byte) error {
	m.sentFrames = append(m.sentFrames, frame)
	return nil
}
func (m *MockFramer) Handshake(configurations *map[string]any, conn net.Conn, connCtxt context.Context) (*amqp.ConnectionInfo, error) {
	return nil, nil
}
func (m *MockFramer) ParseFrame(frame []byte) (any, error) { return nil, nil }
func (m *MockFramer) CreateHeaderFrame(channel, classID uint16, msg amqp.Message) []byte {
	return []byte("header-frame")
}
func (m *MockFramer) CreateBodyFrame(channel uint16, content []byte) []byte {
	return []byte("body-frame")
}

func (m *MockFramer) CreateBasicDeliverFrame(channel uint16, consumerTag, exchange, routingKey string, deliveryTag uint64, redelivered bool) []byte {
	return []byte("basic-deliver")
}
func (m *MockFramer) CreateBasicGetEmptyFrame(channel uint16) []byte {
	return []byte("basic-get-empty")
}
func (m *MockFramer) CreateBasicGetOkFrame(channel uint16, exchange, routingkey string, msgCount uint32) []byte {
	return []byte("basic-get-ok")
}
func (m *MockFramer) CreateBasicConsumeOkFrame(channel uint16, consumerTag string) []byte {
	return []byte("basic-consume-ok:" + consumerTag)
}
func (m *MockFramer) CreateBasicCancelOkFrame(channel uint16, consumerTag string) []byte {
	return []byte("basic-cancel-ok:" + consumerTag)
}
func (m *MockFramer) CreateBasicRecoverOkFrame(channel uint16) []byte {
	return []byte("basic-recover-ok")
}
func (m *MockFramer) CreateQueueDeclareOkFrame(request *amqp.RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte {
	return []byte("queue-declare")
}
func (m *MockFramer) CreateQueueBindOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte("queue-bind-ok")
}
func (m *MockFramer) CreateQueueDeleteOkFrame(request *amqp.RequestMethodMessage, messageCount uint32) []byte {
	return []byte("queue-delete-ok")
}
func (m *MockFramer) CreateExchangeDeclareFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte("exchange-declare")
}
func (m *MockFramer) CreateExchangeDeleteFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte("exchange-delete")
}
func (m *MockFramer) CreateChannelOpenOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte("channel-open-ok")
}
func (m *MockFramer) CreateChannelCloseOkFrame(channel uint16) []byte {
	return []byte("channel-close-ok")
}
func (m *MockFramer) CreateConnectionCloseOkFrame(request *amqp.RequestMethodMessage) []byte {
	return []byte("connection-close-ok")
}
func (m *MockFramer) CreateCloseFrame(channel, replyCode, classID, methodID, closeClassID, closeClassMethod uint16, replyText string) []byte {
	return []byte("close")
}

func createTestBroker() (*Broker, *MockFramer, net.Conn) {
	mockFramer := &MockFramer{}
	broker := &Broker{
		framer:      mockFramer,
		Connections: make(map[net.Conn]*amqp.ConnectionInfo),
		VHosts:      make(map[string]*vhost.VHost),
	}

	// Create test vhost with a test queue
	vh := vhost.NewVhost("test-vhost", 1000, &dummy.DummyPersistence{})
	vh.Queues["test-queue"] = &vhost.Queue{
		Name: "test-queue",
		Props: &vhost.QueueProperties{
			Passive:    false,
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			Arguments:  nil,
		},
	}
	broker.VHosts["test-vhost"] = vh

	conn := &MockConnection{
		localAddr:  &MockAddr{"tcp", "127.0.0.1:5672"},
		remoteAddr: &MockAddr{"tcp", "127.0.0.1:12345"},
	}

	// Initialize connection state
	broker.Connections[conn] = &amqp.ConnectionInfo{
		VHostName: "test-vhost",
		Channels:  make(map[uint16]*amqp.ChannelState),
	}
	broker.Connections[conn].Channels[1] = &amqp.ChannelState{}

	return broker, mockFramer, conn
}

func TestBasicConsumeHandler_ValidConsumer(t *testing.T) {
	broker, mockFramer, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60, // BASIC class
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "test-consumer",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   nil,
		},
	}

	// Call the handler
	result, err := broker.basicConsumeHandler(request, conn, vh)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	// Check that consumer was registered
	consumerKey := vhost.ConsumerKey{Channel: 1, Tag: "test-consumer"}
	consumer, exists := vh.Consumers[consumerKey]
	if !exists {
		t.Error("Consumer was not registered")
	} else {
		if consumer.QueueName != "test-queue" {
			t.Errorf("Expected queue name 'test-queue', got '%s'", consumer.QueueName)
		}
		if consumer.Channel != 1 {
			t.Errorf("Expected channel 1, got %d", consumer.Channel)
		}
		if !consumer.Active {
			t.Error("Consumer should be active")
		}
	}

	// Check that CONSUME_OK frame was sent
	if len(mockFramer.sentFrames) != 1 {
		t.Errorf("Expected 1 frame to be sent, got %d", len(mockFramer.sentFrames))
	} else {
		expectedFrame := "basic-consume-ok:test-consumer"
		if string(mockFramer.sentFrames[0]) != expectedFrame {
			t.Errorf("Expected frame '%s', got '%s'", expectedFrame, string(mockFramer.sentFrames[0]))
		}
	}

	// Check that consumer is indexed correctly
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 1 {
		t.Errorf("Expected 1 consumer for queue, got %d", len(queueConsumers))
	}

	channelKey := vhost.ConnectionChannelKey{
		Connection: conn,
		Channel:    1,
	}
	channelConsumers := vh.ConsumersByChannel[channelKey]
	if len(channelConsumers) != 1 {
		t.Errorf("Expected 1 consumer for channel, got %d", len(channelConsumers))
	}
}

func TestBasicConsumeHandler_EmptyConsumerTag(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "", // Empty tag - should be auto-generated
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   nil,
		},
	}

	_, err := broker.basicConsumeHandler(request, conn, vh)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that a consumer was registered with a generated tag
	if len(vh.Consumers) != 1 {
		t.Errorf("Expected 1 consumer, got %d", len(vh.Consumers))
	}

	// Find the registered consumer
	var registeredConsumer *vhost.Consumer
	for _, consumer := range vh.Consumers {
		registeredConsumer = consumer
		break
	}

	if registeredConsumer == nil {
		t.Fatal("No consumer found")
	}

	if registeredConsumer.Tag == "" {
		t.Error("Consumer tag should have been generated")
	}

	if len(registeredConsumer.Tag) < 10 { // UUID should be much longer
		t.Errorf("Generated consumer tag seems too short: '%s'", registeredConsumer.Tag)
	}
}

func TestBasicConsumeHandler_NonExistentQueue(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "non-existent-queue",
			ConsumerTag: "test-consumer",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   nil,
		},
	}

	_, err := broker.basicConsumeHandler(request, conn, vh)

	if err == nil {
		t.Error("Expected error for non-existent queue")
	}

	expectedError := "queue non-existent-queue does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Check that no consumer was registered
	if len(vh.Consumers) != 0 {
		t.Errorf("Expected 0 consumers, got %d", len(vh.Consumers))
	}
}

func TestBasicConsumeHandler_DuplicateConsumer(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	// Register first consumer
	consumer1 := vhost.NewConsumer(conn, 1, "test-queue", "duplicate-tag", &vhost.ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	err := vh.RegisterConsumer(consumer1)
	if err != nil {
		t.Fatalf("Failed to register first consumer: %v", err)
	}

	// Try to register second consumer with same tag on same channel
	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "duplicate-tag",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   nil,
		},
	}

	_, err = broker.basicConsumeHandler(request, conn, vh)

	if err == nil {
		t.Error("Expected error for duplicate consumer tag")
	}

	expectedError := "consumer with tag duplicate-tag already exists on channel 1"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestBasicConsumeHandler_ExclusiveConsumer(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	// Register exclusive consumer
	request1 := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "exclusive-consumer",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   true, // Exclusive consumer
			NoWait:      false,
			Arguments:   nil,
		},
	}

	_, err := broker.basicConsumeHandler(request1, conn, vh)
	if err != nil {
		t.Fatalf("Failed to register exclusive consumer: %v", err)
	}

	// Try to register another consumer on the same queue
	request2 := &amqp.RequestMethodMessage{
		Channel:  2,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "second-consumer",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   nil,
		},
	}

	// Initialize channel 2
	broker.Connections[conn].Channels[2] = &amqp.ChannelState{}

	_, err = broker.basicConsumeHandler(request2, conn, vh)

	if err == nil {
		t.Error("Expected error when trying to add consumer to queue with exclusive consumer")
	}

	expectedError := "exclusive consumer already exists for queue test-queue"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestBasicConsumeHandler_NoWaitFlag(t *testing.T) {
	broker, mockFramer, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "no-wait-consumer",
			NoLocal:     false,
			NoAck:       false,
			Exclusive:   false,
			NoWait:      true, // No response expected
			Arguments:   nil,
		},
	}

	_, err := broker.basicConsumeHandler(request, conn, vh)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that no frame was sent (due to NoWait flag)
	if len(mockFramer.sentFrames) != 0 {
		t.Errorf("Expected 0 frames to be sent due to NoWait flag, got %d", len(mockFramer.sentFrames))
	}

	// But consumer should still be registered
	consumerKey := vhost.ConsumerKey{Channel: 1, Tag: "no-wait-consumer"}
	if _, exists := vh.Consumers[consumerKey]; !exists {
		t.Error("Consumer should still be registered despite NoWait flag")
	}
}

func TestBasicConsumeHandler_WithArguments(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	arguments := map[string]any{
		"x-priority":     int32(10),
		"x-consumer-tag": "custom-tag",
		"x-exclusive":    true,
	}

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content: &amqp.BasicConsumeContent{
			Queue:       "test-queue",
			ConsumerTag: "args-consumer",
			NoLocal:     false,
			NoAck:       true,
			Exclusive:   false,
			NoWait:      false,
			Arguments:   arguments,
		},
	}

	_, err := broker.basicConsumeHandler(request, conn, vh)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that consumer was registered with correct properties
	consumerKey := vhost.ConsumerKey{Channel: 1, Tag: "args-consumer"}
	consumer, exists := vh.Consumers[consumerKey]
	if !exists {
		t.Fatal("Consumer was not registered")
	}

	if !consumer.Props.NoAck {
		t.Error("Consumer NoAck property should be true")
	}

	if consumer.Props.Exclusive {
		t.Error("Consumer Exclusive property should be false")
	}

	if consumer.Props.Arguments == nil {
		t.Error("Consumer arguments should not be nil")
	} else {
		if len(consumer.Props.Arguments) != len(arguments) {
			t.Errorf("Expected %d arguments, got %d", len(arguments), len(consumer.Props.Arguments))
		}

		for key, expectedValue := range arguments {
			actualValue, exists := consumer.Props.Arguments[key]
			if !exists {
				t.Errorf("Expected argument '%s' not found", key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected argument '%s' value %v, got %v", key, expectedValue, actualValue)
			}
		}
	}
}

func TestBasicConsumeHandler_InvalidContent(t *testing.T) {
	broker, _, conn := createTestBroker()
	vh := broker.VHosts["test-vhost"]

	request := &amqp.RequestMethodMessage{
		Channel:  1,
		ClassID:  60,
		MethodID: uint16(amqp.BASIC_CONSUME),
		Content:  nil, // Invalid content
	}

	_, err := broker.basicConsumeHandler(request, conn, vh)

	if err == nil {
		t.Error("Expected error for invalid content")
	}

	expectedError := "invalid basic consume content"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Check that no consumer was registered
	if len(vh.Consumers) != 0 {
		t.Errorf("Expected 0 consumers, got %d", len(vh.Consumers))
	}
}
