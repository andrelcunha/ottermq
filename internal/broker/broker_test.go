package broker

// import (
// 	"testing"
// 	"time"
// )

// func (b *Broker) TestProcessCommand(command string) string {
// 	return b.processCommand(command)
// }

// func (b *Broker) TestConsume(queueName string) <-chan Message {
// 	return b.consume(queueName)
// }

// func TestProcessCommand(t *testing.T) {
// 	t.Log("OtterMq is starting...")
// 	b := NewBroker()
// 	go b.Start(":5672")

// 	// Test CreateQueue command
// 	resp, err := b.TestProcessCommand("CREATE_QUEUE testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to create queue: %v", err)
// 	}
// 	t.Log(resp)

// 	if _, ok := b.Queues["testQueue"]; !ok {
// 		t.Fatalf("Queue testQueue was not created")
// 	}

// 	// Test Publish command
// 	resp, err = b.TestProcessCommand("PUBLISH testQueue Hello, OtterMq!")
// 	if err != nil {
// 		t.Fatalf("Failed to publish message: %v", err)
// 	}
// 	t.Log(resp)

// 	// Verify the message was published
// 	queue, ok := b.Queues["testQueue"]
// 	if !ok {
// 		t.Fatalf("Queue testQueue does not exist")
// 	}

// 	select {
// 	case msg := <-queue.messages:
// 		if msg.Content != "Hello, OtterMq!" {
// 			t.Fatalf("Expected message 'Hello, OtterMq!', but got '%s'", msg.Content)
// 		}
// 		t.Logf("Received message: %s (ID: %s)", msg.Content, msg.ID)

// 		// Test ACK command
// 		resp, err = b.TestProcessCommand("ACK " + msg.ID)
// 		if err != nil {
// 			t.Fatalf("Failed to ACK message: %v", err)
// 		}
// 		t.Log(resp)

// 		// Verify the message was acknowledged
// 		if _, ok := b.UnackedMessages[msg.ID]; ok {
// 			t.Fatalf("Message %s was not acknowledged", msg.ID)
// 		}
// 	case <-time.After(time.Second):
// 		t.Logf("Did not receive message in time")
// 		//	t.Fatalf("Did not receive message in time")
// 	}

// 	// Test ListQueues command
// 	resp, err = b.TestProcessCommand("LIST_QUEUES")
// 	if err != nil {
// 		t.Fatalf("Failed to list queues: %v", err)
// 	}
// 	if resp != "Queues: testQueue" {
// 		t.Fatalf("Expected 'Queues: testQueue', but got '%s'", resp)
// 	}
// 	t.Log(resp)

// 	// Test DeleteQueue command
// 	resp, err = b.TestProcessCommand("DELETE_QUEUE testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to delete queue: %v", err)
// 	}
// 	t.Log(resp)

// 	// Verify the queue was deleted
// 	if _, ok := b.Queues["testQueue"]; ok {
// 		t.Fatalf("Queue testQueue was not deleted")
// 	}
// }

// func TestCreateExchange(t *testing.T) {
// 	b := NewBroker()
// 	resp, err := b.TestProcessCommand("CREATE_EXCHANGE testExchange direct")
// 	if err != nil {
// 		t.Fatalf("Failed to create exchange: %v", err)
// 	}
// 	t.Log(resp)

// 	if _, ok := b.Exchanges["testExchange"]; !ok {
// 		t.Fatalf("Exchange testExchange was not created")
// 	}
// }

// func TestCreateQueue(t *testing.T) {
// 	b := NewBroker()
// 	resp, err := b.TestProcessCommand("CREATE_QUEUE testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to create queue: %v", err)
// 	}
// 	t.Log(resp)

// 	if _, ok := b.Queues["testQueue"]; !ok {
// 		t.Fatalf("Queue testQueue was not created")
// 	}
// }

// func TestBindQueue(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_EXCHANGE testExchange direct")
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")

// 	resp, err := b.TestProcessCommand("BIND_QUEUE testExchange testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to bind queue: %v", err)
// 	}
// 	t.Log(resp)

// 	// exchange, _ := b.exchanges["testExchange"]
// 	// if _, ok := exchange.bindings["testQueue"]; !ok {
// 	// 	t.Fatalf("Queue testQueue was not bound to exchange testExchange")
// 	// }
// }

// func TestPublish(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_EXCHANGE testExchange direct")
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")
// 	b.TestProcessCommand("BIND_QUEUE testExchange testQueue")

// 	resp, err := b.TestProcessCommand("PUBLISH testExchange testQueue Hello, World!")
// 	if err != nil {
// 		t.Fatalf("Failed to publish message: %v", err)
// 	}
// 	t.Log(resp)
// }

// func TestConsume(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")
// 	b.TestProcessCommand("PUBLISH default testQueue Hello, World!")

// 	resp, err := b.TestProcessCommand("CONSUME testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to consume message: %v", err)
// 	}
// 	if resp != "Hello, World!" {
// 		t.Fatalf("Expected 'Hello, World!', but got '%s'", resp)
// 	}
// 	t.Log(resp)
// }

// func TestAcknowledge(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")
// 	b.TestProcessCommand("PUBLISH default testQueue Hello, World!")

// 	msg := <-b.TestConsume("testQueue")
// 	resp, err := b.TestProcessCommand("ACK " + msg.ID)
// 	if err != nil {
// 		t.Fatalf("Failed to acknowledge message: %v", err)
// 	}
// 	t.Log(resp)
// }

// func TestDeleteQueue(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")

// 	resp, err := b.TestProcessCommand("DELETE_QUEUE testQueue")
// 	if err != nil {
// 		t.Fatalf("Failed to delete queue: %v", err)
// 	}
// 	t.Log(resp)

// 	if _, ok := b.Queues["testQueue"]; ok {
// 		t.Fatalf("Queue testQueue was not deleted")
// 	}
// }

// func TestListQueues(t *testing.T) {
// 	b := NewBroker()
// 	b.TestProcessCommand("CREATE_QUEUE testQueue")

// 	resp, err := b.TestProcessCommand("LIST_QUEUES")
// 	if err != nil {
// 		t.Fatalf("Failed to list queues: %v", err)
// 	}
// 	if resp != "Queues: testQueue" {
// 		t.Fatalf("Expected 'Queues: testQueue', but got '%s'", resp)
// 	}
// 	t.Log(resp)
// }
