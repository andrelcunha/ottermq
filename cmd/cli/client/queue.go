package client

import (
	"github.com/spf13/cobra"
)

var queueName string

var createQueueCmd = &cobra.Command{
	Use:   "create-queue",
	Short: "Create a new queue",
	Run: func(cmd *cobra.Command, args []string) {
		// b := broker.NewBroker()
		// b.CreateQueue(queueName)
		// fmt.Printf("Queue %s created\n", queueName)
	},
}

func init() {
	createQueueCmd.Flags().StringVarP(&queueName, "name", "n", "", "Name of the queue")
	createQueueCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(createQueueCmd)
}
