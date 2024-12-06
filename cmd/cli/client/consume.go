package client

import (
	"github.com/spf13/cobra"
)

var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Consume messages from a queue",
	Run: func(cmd *cobra.Command, args []string) {
		// b := broker.NewBroker()
		// msg := <-b.Consume(queueName)
		// fmt.Printf("Consumed message: %s (ID: %s)\n", msg.Content, msg.ID)
	},
}

func init() {
	consumeCmd.Flags().StringVarP(&queueName, "queue", "q", "", "Name of the queue")
	consumeCmd.MarkFlagRequired("queue")
	rootCmd.AddCommand(consumeCmd)
}
