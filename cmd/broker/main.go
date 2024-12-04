package main

func main() {
	// log.Println("OtterMq is starting...")
	// b := broker.NewBroker()
	// go b.Start(":5672")

	// b.CreateQueue("testQueue")
	// go b.Publish("testQueue", "Hello, OtterMq!")

	// msg := <-b.Consume("testQueue")
	// log.Printf("Received message: %s (ID: %s)", msg.Content, msg.ID)

	// ackResponse, err := b.TestProcessCommand("ACK " + msg.ID)
	// if err != nil {
	// 	log.Printf("Failed to ACK message: %v", err)
	// }
	// log.Printf(ackResponse)

	// queueNames := b.ListQueues()
	// queues := strings.Join(queueNames, ", ")
	// fmt.Printf("Queues: %s\n", queues)
}
