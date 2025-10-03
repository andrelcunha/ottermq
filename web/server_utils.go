package web

import (
	"fmt"
	"github.com/rs/zerolog/log"
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
	log.Info().Str("addr", brokerAddr).Msg("Connecting to broker")
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
		log.Warn().Int("wait_seconds", wait).Msg("Broker is not ready. Retrying...")
		time.Sleep(time.Duration(wait) * time.Second)
		tries++
	}
}
