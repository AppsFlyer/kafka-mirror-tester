package main

import (
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/cmd"
)

func main() {
	cmd.Execute()
	// TODO: This code is temporary and will be replaced by proper CLI flags
	//ctx := context.Background()
	//brokers := types.Brokers("localhost:9093")
	//topic := types.Topic("topic")
	//id := types.ProducerID("1")
	//throughput := types.Throughput(100000)
	//messageSize := types.MessageSize(10)
	//initialSequence := types.SequenceNumber(0)

	//runConsumer := true
	//if runConsumer {
	//consumer.ConsumeAndAnalyze(ctx, brokers, types.Topics([]string{string(topic)}), types.ConsumerGroup("group-4"), 0)
	//} else {
	//producer.ProduceForever(ctx, brokers, topic, id, initialSequence, throughput, messageSize)
	//}
}
