package vhost

import (
	"net"
	"testing"
	"time"
)

// MockConnection implements net.Conn for testing
type MockConnection struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	closed     bool
	id         string
}

func NewMockConnection(id string) *MockConnection {
	return &MockConnection{
		localAddr:  &MockAddr{"tcp", "127.0.0.1:5672"},
		remoteAddr: &MockAddr{"tcp", "127.0.0.1:12345"},
		id:         id,
	}
}

func (m *MockConnection) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConnection) Write(b []byte) (n int, err error) { return len(b), nil }
func (m *MockConnection) Close() error                      { m.closed = true; return nil }
func (m *MockConnection) LocalAddr() net.Addr               { return m.localAddr }
func (m *MockConnection) RemoteAddr() net.Addr              { return m.remoteAddr }
func (m *MockConnection) SetDeadline(t time.Time) error     { return nil }
func (m *MockConnection) SetReadDeadline(t time.Time) error { return nil }
func (m *MockConnection) SetWriteDeadline(t time.Time) error { return nil }

// MockAddr implements net.Addr for testing
type MockAddr struct {
	network string
	address string
}

func (m *MockAddr) Network() string { return m.network }
func (m *MockAddr) String() string  { return m.address }

func createTestVHost() *VHost {
	vh := NewVhost("test-vhost", 1000, nil) // Using nil persistence for tests
	
	// Add test queue
	vh.Queues["test-queue"] = &Queue{
		Name: "test-queue",
		Props: &QueueProperties{
			Passive:    false,
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  nil,
		},
	}
	
	// Add exclusive test queue
	vh.Queues["exclusive-queue"] = &Queue{
		Name: "exclusive-queue",
		Props: &QueueProperties{
			Passive:    false,
			Durable:    false,
			AutoDelete: false,
			Exclusive:  true,
			NoWait:     false,
			Arguments:  nil,
		},
	}
	
	return vh
}

func TestRegisterConsumer_ValidConsumer(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	consumer := NewConsumer(conn, 1, "test-queue", "test-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer)
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Check primary registry
	key := ConsumerKey{Channel: 1, Tag: "test-consumer"}
	registeredConsumer, exists := vh.Consumers[key]
	if !exists {
		t.Error("Consumer not found in primary registry")
	} else {
		if !registeredConsumer.Active {
			t.Error("Consumer should be active")
		}
		if registeredConsumer.QueueName != "test-queue" {
			t.Errorf("Expected queue name 'test-queue', got '%s'", registeredConsumer.QueueName)
		}
	}
	
	// Check queue index
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 1 {
		t.Errorf("Expected 1 consumer in queue index, got %d", len(queueConsumers))
	} else if queueConsumers[0].Tag != "test-consumer" {
		t.Errorf("Expected consumer tag 'test-consumer', got '%s'", queueConsumers[0].Tag)
	}
	
	// Check channel index
	channelConsumers := vh.ConsumersByChannel[1]
	if len(channelConsumers) != 1 {
		t.Errorf("Expected 1 consumer in channel index, got %d", len(channelConsumers))
	} else if channelConsumers[0].Tag != "test-consumer" {
		t.Errorf("Expected consumer tag 'test-consumer', got '%s'", channelConsumers[0].Tag)
	}
}

func TestRegisterConsumer_NonExistentQueue(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	consumer := NewConsumer(conn, 1, "non-existent-queue", "test-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer)
	
	if err == nil {
		t.Error("Expected error for non-existent queue")
	}
	
	expectedError := "queue non-existent-queue does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
	
	// Check that consumer was not registered
	if len(vh.Consumers) != 0 {
		t.Errorf("Expected 0 consumers, got %d", len(vh.Consumers))
	}
}

func TestRegisterConsumer_DuplicateConsumer(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register first consumer
	consumer1 := NewConsumer(conn, 1, "test-queue", "duplicate-tag", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer1)
	if err != nil {
		t.Fatalf("Failed to register first consumer: %v", err)
	}
	
	// Try to register second consumer with same tag and channel
	consumer2 := NewConsumer(conn, 1, "test-queue", "duplicate-tag", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err = vh.RegisterConsumer(consumer2)
	
	if err == nil {
		t.Error("Expected error for duplicate consumer")
	}
	
	expectedError := "consumer with tag duplicate-tag already exists on channel 1"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
	
	// Check that only one consumer exists
	if len(vh.Consumers) != 1 {
		t.Errorf("Expected 1 consumer, got %d", len(vh.Consumers))
	}
}

func TestRegisterConsumer_ExclusiveConsumerExists(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register exclusive consumer
	exclusiveConsumer := NewConsumer(conn, 1, "test-queue", "exclusive-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: true,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(exclusiveConsumer)
	if err != nil {
		t.Fatalf("Failed to register exclusive consumer: %v", err)
	}
	
	// Try to register another consumer on the same queue
	normalConsumer := NewConsumer(conn, 2, "test-queue", "normal-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err = vh.RegisterConsumer(normalConsumer)
	
	if err == nil {
		t.Error("Expected error when trying to add consumer to queue with exclusive consumer")
	}
	
	expectedError := "exclusive consumer already exists for queue test-queue"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRegisterConsumer_ExclusiveConsumerWithExistingConsumers(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register normal consumer first
	normalConsumer := NewConsumer(conn, 1, "test-queue", "normal-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(normalConsumer)
	if err != nil {
		t.Fatalf("Failed to register normal consumer: %v", err)
	}
	
	// Try to register exclusive consumer when others exist
	exclusiveConsumer := NewConsumer(conn, 2, "test-queue", "exclusive-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: true,
		Arguments: nil,
	})
	
	err = vh.RegisterConsumer(exclusiveConsumer)
	
	if err == nil {
		t.Error("Expected error when trying to add exclusive consumer when other consumers exist")
	}
	
	expectedError := "cannot add exclusive consumer when other consumers exist for queue test-queue"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRegisterConsumer_MultipleConsumersOnDifferentChannels(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register consumer on channel 1
	consumer1 := NewConsumer(conn, 1, "test-queue", "consumer-1", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer1)
	if err != nil {
		t.Fatalf("Failed to register consumer 1: %v", err)
	}
	
	// Register consumer on channel 2
	consumer2 := NewConsumer(conn, 2, "test-queue", "consumer-2", &ConsumerProperties{
		NoAck:     true,
		Exclusive: false,
		Arguments: nil,
	})
	
	err = vh.RegisterConsumer(consumer2)
	if err != nil {
		t.Fatalf("Failed to register consumer 2: %v", err)
	}
	
	// Check that both consumers are registered
	if len(vh.Consumers) != 2 {
		t.Errorf("Expected 2 consumers, got %d", len(vh.Consumers))
	}
	
	// Check queue index has both consumers
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 2 {
		t.Errorf("Expected 2 consumers in queue index, got %d", len(queueConsumers))
	}
	
	// Check channel indexes
	channel1Consumers := vh.ConsumersByChannel[1]
	if len(channel1Consumers) != 1 {
		t.Errorf("Expected 1 consumer on channel 1, got %d", len(channel1Consumers))
	}
	
	channel2Consumers := vh.ConsumersByChannel[2]
	if len(channel2Consumers) != 1 {
		t.Errorf("Expected 1 consumer on channel 2, got %d", len(channel2Consumers))
	}
}

func TestCancelConsumer_ValidConsumer(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register consumer first
	consumer := NewConsumer(conn, 1, "test-queue", "test-consumer", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer)
	if err != nil {
		t.Fatalf("Failed to register consumer: %v", err)
	}
	
	// Cancel the consumer
	err = vh.CancelConsumer(1, "test-consumer")
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Check that consumer was removed from all indexes
	key := ConsumerKey{Channel: 1, Tag: "test-consumer"}
	if _, exists := vh.Consumers[key]; exists {
		t.Error("Consumer should have been removed from primary registry")
	}
	
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 0 {
		t.Errorf("Expected 0 consumers in queue index, got %d", len(queueConsumers))
	}
	
	channelConsumers := vh.ConsumersByChannel[1]
	if len(channelConsumers) != 0 {
		t.Errorf("Expected 0 consumers in channel index, got %d", len(channelConsumers))
	}
}

func TestCancelConsumer_NonExistentConsumer(t *testing.T) {
	vh := createTestVHost()
	
	err := vh.CancelConsumer(1, "non-existent-consumer")
	
	if err == nil {
		t.Error("Expected error for non-existent consumer")
	}
	
	expectedError := "consumer with tag non-existent-consumer on channel 1 does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCleanupChannel(t *testing.T) {
	vh := createTestVHost()
	conn := NewMockConnection("test-conn")
	
	// Register multiple consumers on the same channel
	consumer1 := NewConsumer(conn, 1, "test-queue", "consumer-1", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	consumer2 := NewConsumer(conn, 1, "test-queue", "consumer-2", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer1)
	if err != nil {
		t.Fatalf("Failed to register consumer 1: %v", err)
	}
	
	err = vh.RegisterConsumer(consumer2)
	if err != nil {
		t.Fatalf("Failed to register consumer 2: %v", err)
	}
	
	// Verify consumers are registered
	if len(vh.Consumers) != 2 {
		t.Fatalf("Expected 2 consumers before cleanup, got %d", len(vh.Consumers))
	}
	
	// Cleanup the channel
	vh.CleanupChannel(1)
	
	// Check that all consumers on the channel were removed
	if len(vh.Consumers) != 0 {
		t.Errorf("Expected 0 consumers after cleanup, got %d", len(vh.Consumers))
	}
	
	// Check that queue index is cleaned up
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 0 {
		t.Errorf("Expected 0 consumers in queue index after cleanup, got %d", len(queueConsumers))
	}
	
	// Check that channel index is cleaned up
	if _, exists := vh.ConsumersByChannel[1]; exists {
		t.Error("Channel index should have been cleaned up")
	}
}

func TestCleanupConnection(t *testing.T) {
	vh := createTestVHost()
	conn1 := NewMockConnection("conn-1")
	conn2 := NewMockConnection("conn-2")
	
	// Register consumers on different connections and channels
	consumer1 := NewConsumer(conn1, 1, "test-queue", "consumer-1", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	consumer2 := NewConsumer(conn1, 2, "test-queue", "consumer-2", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	consumer3 := NewConsumer(conn2, 1, "test-queue", "consumer-3", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	err := vh.RegisterConsumer(consumer1)
	if err != nil {
		t.Fatalf("Failed to register consumer 1: %v", err)
	}
	
	err = vh.RegisterConsumer(consumer2)
	if err != nil {
		t.Fatalf("Failed to register consumer 2: %v", err)
	}
	
	err = vh.RegisterConsumer(consumer3)
	if err != nil {
		t.Fatalf("Failed to register consumer 3: %v", err)
	}
	
	// Verify all consumers are registered
	if len(vh.Consumers) != 3 {
		t.Fatalf("Expected 3 consumers before cleanup, got %d", len(vh.Consumers))
	}
	
	// Cleanup connection 1
	vh.CleanupConnection(conn1)
	
	// Check that only consumers from conn1 were removed
	if len(vh.Consumers) != 1 {
		t.Errorf("Expected 1 consumer after cleanup (from conn2), got %d", len(vh.Consumers))
	}
	
	// Check that remaining consumer is from conn2
	for _, consumer := range vh.Consumers {
		if consumer.Connection != conn2 {
			t.Error("Remaining consumer should be from conn2")
		}
	}
	
	// Check that queue index still has one consumer
	queueConsumers := vh.ConsumersByQueue["test-queue"]
	if len(queueConsumers) != 1 {
		t.Errorf("Expected 1 consumer in queue index after cleanup, got %d", len(queueConsumers))
	}
}

func TestNewConsumer_EmptyTag(t *testing.T) {
	conn := NewMockConnection("test-conn")
	
	consumer := NewConsumer(conn, 1, "test-queue", "", &ConsumerProperties{
		NoAck:     false,
		Exclusive: false,
		Arguments: nil,
	})
	
	if consumer.Tag == "" {
		t.Error("Consumer tag should have been generated when empty")
	}
	
	if len(consumer.Tag) < 10 { // UUID should be much longer
		t.Errorf("Generated consumer tag seems too short: '%s'", consumer.Tag)
	}
	
	// Check other properties
	if consumer.Channel != 1 {
		t.Errorf("Expected channel 1, got %d", consumer.Channel)
	}
	
	if consumer.QueueName != "test-queue" {
		t.Errorf("Expected queue name 'test-queue', got '%s'", consumer.QueueName)
	}
	
	if consumer.Connection != conn {
		t.Error("Consumer connection should match provided connection")
	}
}

func TestNewConsumer_WithTag(t *testing.T) {
	conn := NewMockConnection("test-conn")
	
	consumer := NewConsumer(conn, 2, "my-queue", "my-consumer-tag", &ConsumerProperties{
		NoAck:     true,
		Exclusive: true,
		Arguments: map[string]any{"key": "value"},
	})
	
	if consumer.Tag != "my-consumer-tag" {
		t.Errorf("Expected tag 'my-consumer-tag', got '%s'", consumer.Tag)
	}
	
	if consumer.Channel != 2 {
		t.Errorf("Expected channel 2, got %d", consumer.Channel)
	}
	
	if consumer.QueueName != "my-queue" {
		t.Errorf("Expected queue name 'my-queue', got '%s'", consumer.QueueName)
	}
	
	if consumer.Connection != conn {
		t.Error("Consumer connection should match provided connection")
	}
	
	if consumer.Props == nil {
		t.Error("Consumer properties should not be nil")
	} else {
		if !consumer.Props.NoAck {
			t.Error("Consumer NoAck should be true")
		}
		
		if !consumer.Props.Exclusive {
			t.Error("Consumer Exclusive should be true")
		}
		
		if consumer.Props.Arguments == nil {
			t.Error("Consumer arguments should not be nil")
		} else if consumer.Props.Arguments["key"] != "value" {
			t.Error("Consumer arguments not preserved correctly")
		}
	}
}