package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/producer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

var (
	// CLI args
	producerID         *string
	pTopics            *string
	throughput         *uint
	messageSize        *uint
	pBootstraServers   *string
	pUseMessageHeaders *bool
	pNumPartitions     *uint
	pNumReplicas       *uint
	pRetention         *uint
)

// produceCmd represents the produce command
var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Produce messages to kafka",
	Long: `The producer is a high-throughput kafka message producer.
	It sends sequence numbered and timestamped messages to kafka where by the consumer reads and validates. `,
	Run: func(cmd *cobra.Command, args []string) {
		brokers := types.Brokers(*pBootstraServers)
		id := types.ProducerID(*producerID)
		through := types.Throughput(*throughput)
		size := types.MessageSize(*messageSize)
		initialSequence := types.SequenceNumber(0)
		producer.ProduceToTopics(brokers, id, through, size, initialSequence, *pTopics, *pNumPartitions, *pNumReplicas, *pUseMessageHeaders, *pRetention)
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
	pUseMessageHeaders = produceCmd.Flags().Bool("use-message-headers", false, "Whether to use message headers to pass metadata or use the payload instead")
	pNumPartitions = produceCmd.Flags().Uint("num-partitions", 1, "Number of partitions to create per each topic (if the topics are new)")
	pNumReplicas = produceCmd.Flags().Uint("num-replicas", 1, "Number of replicas to create per each topic (if the topics are new)")
	pRetention = produceCmd.Flags().Uint("retention", 30000, "Data retention for the created topics. In ms.")
}
