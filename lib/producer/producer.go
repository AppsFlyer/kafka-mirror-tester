package producer

// The producer package is responsible for producing messages and repotring success/failure WRT
// delivery as well as capacity (is it able to produce the required throughput)

import (
	"context"
	"math"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/time/rate"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/admin"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/types"
)

const (
	// How much burst we allow for the rate limiter.
	// We provide a 0.1 burst ratio which means that at times the rate might go up to 10% or the desired rate (but not for log)
	// This is done in order to conpersate for slow starts.
	burstRatio = 0.1

	// Number of messages per producer that we allow in-flight before waiting and flushing
	inFlightThreshold = 100000000
)

// ProduceToTopics spawms multiple producer threads and produces to all topics
func ProduceToTopics(
	brokers types.Brokers,
	id types.ProducerID,
	throughput types.Throughput,
	size types.MessageSize,
	initialSequence types.SequenceNumber,
	topicsString string,
	numPartitions, numReplicas uint,
	useMessageHeaders bool,
	retentionMs uint,
) {
	// Count the total number of errors on this topic
	errorCounter := uint64(0)
	topics := strings.Split(topicsString, ",")
	producers := mapset.NewSet()
	ctx := context.Background()
	go monitor(ctx, &errorCounter, monitoringFrequency, throughput, id, uint(len(topics)), numPartitions, producers)

	var wg sync.WaitGroup
	for _, topic := range topics {
		t := types.Topic(topic)
		wg.Add(1)
		go func(topic types.Topic, partitions, replicas uint) {
			admin.MustCreateTopic(ctx, brokers, t, partitions, replicas, retentionMs)
			ProduceForever(
				ctx,
				brokers,
				t,
				id,
				initialSequence,
				numPartitions,
				throughput,
				size,
				useMessageHeaders,
				&errorCounter,
				producers)
			wg.Done()
		}(t, numPartitions, numReplicas)
	}
	wg.Wait()
}

// ProduceForever will produce messages to the topic forver or until canceled by the context.
// It will try to acheive the desired throughput and if not - will log that. It will not exceed the throughput (measured by number of messages per second)
// throughput is limited to 1M messages per second.
func ProduceForever(
	ctx context.Context,
	brokers types.Brokers,
	topic types.Topic,
	id types.ProducerID,
	initialSequence types.SequenceNumber,
	numPartitions uint,
	throughput types.Throughput,
	messageSize types.MessageSize,
	useMessageHeaders bool,
	errorCounter *uint64,
	producers mapset.Set,
) {
	log.Infof("Starting the producer. brokers=%s, topic=%s id=%s throughput=%d size=%d initialSequence=%d",
		brokers, topic, id, throughput, messageSize, initialSequence)
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      string(brokers),
		"queue.buffering.max.ms": "1000",
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	defer p.Close()
	producers.Add(p)
	producerForeverWithProducer(
		ctx,
		p,
		topic,
		id,
		initialSequence,
		numPartitions,
		throughput,
		messageSize,
		useMessageHeaders,
		errorCounter)
}

// producerForeverWithWriter produces kafka messages forever or until the context is canceled.
// adheeers to maintaining the desired throughput.
func producerForeverWithProducer(
	ctx context.Context,
	p *kafka.Producer,
	topic types.Topic,
	producerID types.ProducerID,
	initialSequence types.SequenceNumber,
	numPartitions uint,
	throughput types.Throughput,
	messageSize types.MessageSize,
	useMessageHeaders bool,
	errorCounter *uint64,
) {
	// the rate limiter regulates the producer by limiting its throughput (messages/sec)
	limiter := rate.NewLimiter(rate.Limit(throughput), int(math.Ceil(float64(throughput)*burstRatio)))

	// Sequence number per message
	seq := initialSequence

	go eventsProcessor(p, errorCounter)

	topicString := string(topic)
	tp := kafka.TopicPartition{Topic: &topicString, Partition: kafka.PartitionAny}
	for ; ; seq++ {
		err := limiter.Wait(ctx)
		if err != nil {
			log.Errorf("Error waiting %+v", err)
			continue
		}
		messageKey := types.MessageKey(uint(seq) % numPartitions)
		scopedSeq := seq / types.SequenceNumber(numPartitions)
		produceMessage(ctx, p, tp, producerID, messageKey, scopedSeq, messageSize, useMessageHeaders)
	}
}

// produceMessage produces a single message to kafka.
// message production is asyncrounous on the ProducerChannel
func produceMessage(
	ctx context.Context,
	p *kafka.Producer,
	topicPartition kafka.TopicPartition,
	producerID types.ProducerID,
	messageKey types.MessageKey,
	seq types.SequenceNumber,
	messageSize types.MessageSize,
	useMessageHeaders bool,
) {
	if p.Len() > inFlightThreshold {
		p.Flush(1)
	}
	m := message.Create(producerID, messageKey, seq, messageSize, useMessageHeaders)
	m.TopicPartition = topicPartition
	p.ProduceChannel() <- m
	log.Tracef("Producing %s...", m)
}

// eventsProcessor processes the events emited by the producer p.
// It then logs errors and increased the passed-by-reference errors counter and updates the throughput counter
func eventsProcessor(
	p *kafka.Producer,
	errorCounter *uint64,
) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Errorf("Delivery failed: %v", m.TopicPartition.Error)
				atomic.AddUint64(errorCounter, 1)
				messageSendErrors.Inc()
			} else {
				reportMessageSent(m)
			}
		default:
			log.Infof("Ignored event: %s", ev)
		}
	}
}
