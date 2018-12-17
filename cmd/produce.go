package cmd

import (
	"context"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/producer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

var (
	// CLI args
	producerID       *string
	pTopics          *string
	throughput       *uint
	messageSize      *uint
	pBootstraServers *string
)

// produceCmd represents the produce command
var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Producer messages to kafka",
	Long: `The producer is a high-throughput kafka message producer.
	It sends sequence numbered and timestamped messages to kafka where by the consumer reads and validates. `,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		brokers := types.Brokers(*pBootstraServers)
		id := types.ProducerID(*producerID)
		through := types.Throughput(*throughput)
		size := types.MessageSize(*messageSize)
		initialSequence := types.SequenceNumber(0)
		var wg sync.WaitGroup
		for _, topic := range strings.Split(*pTopics, ",") {
			t := types.Topic(topic)
			wg.Add(1)
			go func(topic types.Topic) {
				producer.ProduceForever(ctx, brokers, t, id, initialSequence, through, size)
				wg.Done()
			}(t)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(produceCmd)
	producerID = produceCmd.Flags().String("id", "", "ID of the producer. You can use the hostname command")
	produceCmd.MarkFlagRequired("id")
	pTopics = produceCmd.Flags().String("topics", "", "List of topics to produce to (coma separated)")
	produceCmd.MarkFlagRequired("topics")
	throughput = produceCmd.Flags().Uint("throughput", 0, "Number of messages to send to each topic per second")
	produceCmd.MarkFlagRequired("throughput")
	messageSize = produceCmd.Flags().Uint("message-size", 0, "Message size to send (in bytes)")
	produceCmd.MarkFlagRequired("message-size")
	pBootstraServers = produceCmd.Flags().String("bootstrap-servers", "", "List of host:port bootstrap servers (coma separated)")
	produceCmd.MarkFlagRequired("bootstrap-servers")
}
