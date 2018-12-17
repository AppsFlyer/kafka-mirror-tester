package producer

// The producer package is responsible for producing messages and repotring success/failure WRT
// delivery as well as capacity (is it able to produce the required throughput)

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/paulbellamy/ratecounter"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const (
	// How much burst we allow for the rate limiter.
	// We provide a 0.1 burst ratio which means that at times the rate might go up to 10% or the desired rate (but not for log)
	// This is done in order to conpersate for slow starts.
	burstRatio = 0.1
)

// ProduceForever will produce messages to the topic forver or until canceled by the context.
// It will try to acheive the desired throughput and if not - will log that. It will not exceed the throughput (measured by number of messages per second)
// throughput is limited to 1M messages per second.
func ProduceForever(
	ctx context.Context,
	brokers types.Brokers,
	topic types.Topic,
	id types.ProducerID,
	initialSequence types.SequenceNumber,
	throughput types.Throughput,
	messageSize types.MessageSize,
) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": string(brokers)})
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	defer p.Close()
	producerForeverWithProducer(ctx, p, topic, id, initialSequence, throughput, messageSize)
}

// producerForeverWithWriter produces kafka messages forever or until the context is canceled.
// adheeers to maintaining the desired throughput.
func producerForeverWithProducer(
	ctx context.Context,
	p *kafka.Producer,
	topic types.Topic,
	id types.ProducerID,
	initialSequence types.SequenceNumber,
	throughput types.Throughput,
	messageSize types.MessageSize,
) {
	// the rate limiter regulates the producer by limiting its throughput (messages/sec)
	limiter := rate.NewLimiter(rate.Limit(throughput), int(float32(throughput)*burstRatio))

	// messageCounter is used in order to observe the actual throughput
	messageCounter := ratecounter.NewRateCounter(monitoringFrequency)
	// bytesCounter measures the actual throughput in bytes
	bytesCounter := ratecounter.NewRateCounter(monitoringFrequency)

	// Sequence number per message
	seq := initialSequence

	// Count the total number of errors on this topic
	errorCounter := uint(0)

	go monitor(ctx, messageCounter, bytesCounter, &errorCounter, monitoringFrequency, throughput)

	go eventsProcessor(p, messageCounter, bytesCounter, &errorCounter)

	topicString := string(topic)
	tp := kafka.TopicPartition{Topic: &topicString, Partition: kafka.PartitionAny}
	for ; ; seq++ {
		err := limiter.Wait(ctx)
		if err != nil {
			log.Errorf("Error waiting %+v", err)
			continue
		}
		produceMessage(ctx, p, tp, id, seq, messageSize)
	}
}

// produceMessage produces a single message to kafka.
// message production is asyncrounous on the ProducerChannel
func produceMessage(
	ctx context.Context,
	p *kafka.Producer,
	topicPartition kafka.TopicPartition,
	id types.ProducerID,
	seq types.SequenceNumber,
	messageSize types.MessageSize,
) {
	m := message.Create(id, seq, messageSize)
	m.TopicPartition = topicPartition
	p.ProduceChannel() <- m
	log.Tracef("Producing %s...", m)
}

// eventsProcessor processes the events emited by the producer p.
// It then logs errors and increased the passed-by-reference errors counter and updates the throughput counter
func eventsProcessor(
	p *kafka.Producer,
	messageCounter *ratecounter.RateCounter,
	bytesCounter *ratecounter.RateCounter,
	errorCounter *uint,
) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Errorf("Delivery failed: %v", m.TopicPartition.Error)
				*errorCounter++
			} else {
				messageCounter.Incr(1)
				bytesCounter.Incr(int64(len(m.Value)))
			}
		default:
			log.Infof("Ignored event: %s", ev)
		}
	}
}
