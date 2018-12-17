package main

import (
	"context"
	"time"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/consumer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/producer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	brokers := types.Brokers("localhost:9093")
	topic := types.Topic("topic")
	id := types.ProducerID("1")
	throughput := types.Throughput(400000)
	messageSize := types.MessageSize(1000)
	initialSequence := types.SequenceNumber(0)

	consumer.ConsumeAndAnalyze(ctx, brokers, types.Topics([]string{string(topic)}), types.ConsumerGroup("group-1"), 0)

	time.AfterFunc(100*time.Second, func() {
		cancel()
	})
	producer.ProduceForever(ctx, brokers, topic, id, initialSequence, throughput, messageSize)
}
