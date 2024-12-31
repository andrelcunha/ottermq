package web

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func GetBrokerClient(c *Config) (*amqp091.Connection, error) {
	username := c.Username
	password := c.Password
	host := c.BrokerHost
	port := c.BrokerPort

	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)
	brokerAddr := fmt.Sprintf("%s:%s", host, port)
	tries := 1
	log.Printf("Connecting to %s\n", brokerAddr)
	for {
		if tries > 3 {
			return nil, fmt.Errorf("Broker did not respond after 3 retries")
		}
		conn, err := amqp091.Dial(connectionString)
		if err == nil && conn != nil {
			// log.Println("Connected.")
			return conn, nil
		}
		wait := 3 * tries
		log.Printf("Broker is not ready. Retrying in %d seconds...\n", wait)
		time.Sleep(time.Duration(wait) * time.Second)
		tries++
	}
}
