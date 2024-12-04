package broker

import (
	"testing"
	"time"
)

func TestProcessCommand(t *testing.T) {
	t.Log("OtterMq is starting...")
	b := NewBroker()
	go b.Start(":5672")

	// Test CreateQueue command
	resp, err := b.TestProcessCommand("CREATE_QUEUE testQueue")
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}
	t.Log(resp)

	if _, ok := b.queues["testQueue"]; !ok {
		t.Fatalf("Queue testQueue was not created")
	}

	// Test Publish command
	resp, err = b.TestProcessCommand("PUBLISH testQueue Hello, OtterMq!")
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}
	t.Log(resp)

	// Verify the message was published
	queue, ok := b.queues["testQueue"]
	if !ok {
		t.Fatalf("Queue testQueue does not exist")
	}

	select {
	case msg := <-queue.messages:
		if msg.Content != "Hello, OtterMq!" {
			t.Fatalf("Expected message 'Hello, OtterMq!', but got '%s'", msg.Content)
		}
		t.Logf("Received message: %s (ID: %s)", msg.Content, msg.ID)

		// Test ACK command
		resp, err = b.TestProcessCommand("ACK " + msg.ID)
		if err != nil {
			t.Fatalf("Failed to ACK message: %v", err)
		}
		t.Log(resp)

		// Verify the message was acknowledged
		if _, ok := b.unackedMessages[msg.ID]; ok {
			t.Fatalf("Message %s was not acknowledged", msg.ID)
		}
	case <-time.After(time.Second):
		t.Fatalf("Did not receive message in time")
	}

	// Test ListQueues command
	resp, err = b.TestProcessCommand("LIST_QUEUES")
	if err != nil {
		t.Fatalf("Failed to list queues: %v", err)
	}
	if resp != "Queues: testQueue" {
		t.Fatalf("Expected 'Queues: testQueue', but got '%s'", resp)
	}
	t.Log(resp)

	// Test DeleteQueue command
	resp, err = b.TestProcessCommand("DELETE_QUEUE testQueue")
	if err != nil {
		t.Fatalf("Failed to delete queue: %v", err)
	}
	t.Log(resp)

	// Verify the queue was deleted
	if _, ok := b.queues["testQueue"]; ok {
		t.Fatalf("Queue testQueue was not deleted")
	}
}

func (b *Broker) TestProcessCommand(command string) (string, error) {
	return b.processCommand(command)
}
