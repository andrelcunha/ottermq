package broker

import "testing"

func TestProcessCommand(t *testing.T) {
	t.Log("OtterMq is starting...")
	b := NewBroker()
	go b.Start(":5672")

	b.CreateQueue("testQueue")
	go b.Publish("testQueue", "Hello, OtterMq!")

	msg := <-b.Consume("testQueue")
	t.Logf("Received message: %s (ID: %s)", msg.Content, msg.ID)

	ackResponse, err := b.processCommand("ACK " + msg.ID)
	if err != nil {
		t.Fatalf("Failed to ACK message: %v", err)
	}
	t.Log(ackResponse)
	return
}
