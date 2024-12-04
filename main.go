package main

import (
	"log"

	"github.com/andrelcunha/ottermq/cmd/broker"
)

func main() {
	log.Println("OtterMq is starting...")
	broker := broker.NewBroker()
	go broker.Start(":5672")

	broker.CreateQueue("testQueue")
	go broker.Publish("testQueue", "Hello, OtterMq!")

	msg := <-broker.Consume("testQueue")
	log.Println("Received message:", msg)
}
