package client

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/broker"
	"github.com/spf13/cobra"
)

var message string
var exchangeName string
var routingKey string

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to a queue",
	Run: func(cmd *cobra.Command, args []string) {
		b := broker.NewBroker()
		b.Publish(exchangeName, routingKey, message)
		fmt.Printf("Message '%s' published to exchange %s\n", message, exchangeName)
	},
}

func init() {
	publishCmd.Flags().StringVarP(&exchangeName, "exchange", "x", "", "Name of the exchange")
	publishCmd.Flags().StringVarP(&routingKey, "routingKey", "r", "", "Name of the routing key")
	publishCmd.Flags().StringVarP(&message, "message", "m", "", "Message to publish")
	publishCmd.MarkFlagRequired("queue")
	publishCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(publishCmd)
}
