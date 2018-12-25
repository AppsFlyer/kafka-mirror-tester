package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/consumer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

var (
	cTopics            *string
	cBootstraServers   *string
	consumerGroup      *string
	cUseMessageHeaders *bool
)

// consumeCmd represents the consume command
var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Consume messages from kafka and aggregate results",
	Long: `Consumes messages from kafka and collects statistics about them.
Namely latency statistics and sequence number bookeeping.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		brokers := types.Brokers(*cBootstraServers)
		ts := types.Topics(strings.Split(*cTopics, ","))
		initialSequence := types.SequenceNumber(0)
		cg := types.ConsumerGroup(*consumerGroup)
		consumer.ConsumeAndAnalyze(ctx, brokers, ts, cg, initialSequence, *cUseMessageHeaders)
	},
}

func init() {
	rootCmd.AddCommand(consumeCmd)

	cTopics = consumeCmd.Flags().String("topics", "", "List of topics to consume from (coma separated)")
	consumeCmd.MarkFlagRequired("topics")
	cBootstraServers = consumeCmd.Flags().String("bootstrap-servers", "", "List of host:port bootstrap servers (coma separated)")
	consumeCmd.MarkFlagRequired("bootstrap-servers")
	consumerGroup = consumeCmd.Flags().String("consumer-group", "", "The kafka consumer group name")
	consumeCmd.MarkFlagRequired("consumer-group")
	cUseMessageHeaders = consumeCmd.Flags().Bool("use-message-headers", false, "Whether to use message headers to pass metadata or use the payload instead")
}
