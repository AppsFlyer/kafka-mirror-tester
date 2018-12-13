package main

import (
	"context"
	"time"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/producer"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	brokers := types.Brokers{"localhost:9093"}
	topic := types.Topic("topic")
	id := types.ProducerID("1")
	throughput := types.Throughput(10)
	messageSize := types.MessageSize(100)
	initialSequence := types.SequenceNumber(0)

	time.AfterFunc(10*time.Second, func() {
		cancel()
	})
	producer.ProduceForever(ctx, brokers, topic, id, initialSequence, throughput, messageSize)
}
