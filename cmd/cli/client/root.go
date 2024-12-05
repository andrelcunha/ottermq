package client

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ottermq",
	Short: "OtterMq is a lightweight message broker",
	Long:  `OtterMq is a lightweight message broker for learning and educational purposes.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
