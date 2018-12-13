package producer

import (
	"context"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

// The producer package is responsible for producing messages and repotring success/failure WRT
// delivery as well as capacity (is it able to produce the required throughput)

// ProduceForever will produce messages to the topic forver or until canceled by the context.
// It will try to acheive the desired throughput and if not - will log that. It will not exceed the throughput (measured by number of messages per second)
// throughput is limited to 1M messages per second.
func ProduceForever(ctx context.Context,
	brokers types.Brokers,
	topic types.Topic,
	id types.ProducerID,
	initialSequence types.SequenceNumber,
	throughput types.Throughput,
	messageSize types.MessageSize,
) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    string(topic),
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()
	producerForeverWithWriter(ctx, w, id, initialSequence, throughput, messageSize)
}

// producerForeverWithWriter produces kafka messages forever or until the context is canceled.
// adheeers tp maintaining the desired throughput.
func producerForeverWithWriter(ctx context.Context,
	writer *kafka.Writer,
	id types.ProducerID,
	initialSequence types.SequenceNumber,
	throughput types.Throughput,
	messageSize types.MessageSize,
) {
	// the rate limiter is responsible for regulating the traffic by limiting its throughput (messages/sec)
	rateMicroseconds := time.Duration(1e6 / throughput)
	limiter := time.Tick(rateMicroseconds * time.Microsecond)

	seq := initialSequence

	go monitor(ctx, writer, 5*time.Second)

	errors := make(chan string)

	for ; ; seq++ {
		select {
		case <-limiter:
			produceMessage(ctx, errors, writer, id, seq, messageSize)
		case <-ctx.Done():
			log.Infof("Main producer look done. %s", ctx.Err())
			return
		}
	}
}

// produceMessage produces a message to kafka.
// It silently fails and report the errors to the channel.
func produceMessage(ctx context.Context,
	errors chan string,
	writer *kafka.Writer,
	id types.ProducerID,
	seq types.SequenceNumber,
	messageSize types.MessageSize,
) {
	go func() {

		key := fmt.Sprintf("%s.%d", id, seq)
		m := kafka.Message{
			Key:   []byte(key),
			Value: []byte(message.Format(id, seq, messageSize)),
		}
		log.Trace("writing ", key, time.Now())
		err := writer.WriteMessages(ctx, m)
		if err != nil {
			log.Errorf("ERROR: %s", err)
			errors <- err.Error()
		}
	}()
}

// periodically monitors the kafka writer.
// Blocks forever or until canceled.
func monitor(ctx context.Context, writer *kafka.Writer, frequency time.Duration) {
	ticker := time.Tick(frequency)
	for {
		select {
		case <-ticker:
			printWriterStats(writer)
		case <-ctx.Done():
			log.Infof("Monitor done. %s", ctx.Err())
			return
		}
	}
}

// Prints some runtime stats such as errors, throughputs etc
func printWriterStats(writer *kafka.Writer) {
	stats := writer.Stats()
	log.Infof(`Recent stats:
	Writes: %d
	Messages: %d
	Bytes: %d
	Max Retries: %d
	Average Batch Size: %d
	Queue Length: %d

	Errors: %d
	`,
		stats.Writes, stats.Messages, stats.Bytes, stats.Retries.Max,
		stats.BatchSize.Avg, stats.QueueLength, stats.Errors)
}
