package vhost

import (
	"fmt"
	"testing"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

func TestQueueCapacityExceeds1000(t *testing.T) {
	// Create a test queue with a buffer size of 100000
	queue := NewQueue("testQueue", 100000)
	
	// Try to push more than 1000 messages
	t.Logf("Starting to push 2000 messages to queue...")
	
	for i := 0; i < 2000; i++ {
		msg := amqp.Message{
			ID:   fmt.Sprintf("msg-%d", i),
			Body: []byte(fmt.Sprintf("Test message %d", i)),
		}
		
		// Push the message
		queue.Push(msg)
	}
	
	finalCount := queue.Len()
	t.Logf("Final queue length: %d", finalCount)
	
	if finalCount <= 1000 {
		t.Errorf("FAIL: Queue is still limited to 1000 messages (actual: %d)", finalCount)
	} else {
		t.Logf("SUCCESS: Queue can now hold more than 1000 messages! (actual: %d)", finalCount)
	}
	
	// Verify we can hold at least 2000 messages
	if finalCount < 2000 {
		t.Errorf("Expected queue to hold 2000 messages, got %d", finalCount)
	}
}
