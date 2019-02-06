// Package cmd is a cli layer that takes care of cli args etc.
// powereve by cobra https://github.com/spf13/cobra
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// cobra root cmd
var rootCmd = &cobra.Command{
	Use:   "kafka-mirror-tester",
	Short: "Kafka mirror tester is a test tool for kafka mirroring",
	Long:  `A high throughput producer and consumer that stress kafka and validate message consumption order and latency.`,
}

// Execute is the main entry point for the CLI, using cobra lib.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
