package client

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/broker"
	"github.com/spf13/cobra"
)

var message string

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to a queue",
	Run: func(cmd *cobra.Command, args []string) {
		b := broker.NewBroker()
		b.Publish(queueName, message)
		fmt.Printf("Message '%s' published to queue %s\n", message, queueName)
	},
}

func init() {
	publishCmd.Flags().StringVarP(&queueName, "queue", "q", "", "Name of the queue")
	publishCmd.Flags().StringVarP(&message, "message", "m", "", "Message to publish")
	publishCmd.MarkFlagRequired("queue")
	publishCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(publishCmd)
}
