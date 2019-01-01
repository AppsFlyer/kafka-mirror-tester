// Package consumer implements the consumption, performance measurement and validation logic of the test
package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const (

	// kafka consumer session timeout
	sessionTimeoutMs = 6000

	// For the purpose of performance monitoring we always want to start with the latest messages
	autoOffsetReset = "latest"
)

// clientID is a friendly name for the client so that monitoring tool know who we are.
var clientID string

func init() {

	//log.SetLevel(log.TraceLevel)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Can't get hostname %+v", err)
	}
	clientID = fmt.Sprintf("kafka-mirror-tester-%s", hostname)
}

// ConsumeAndAnalyze consumes messages from the kafka topic and analyzes their correctness and performance.
// The function blocks forever (or until the context is cancled, or until a signal is sent)
func ConsumeAndAnalyze(
	ctx context.Context,
	brokers types.Brokers,
	topics types.Topics,
	group types.ConsumerGroup,
	initialSequence types.SequenceNumber,
	useMessageHeaders bool,
) {
	log.Infof("Starting the consumer. brokers=%s, topics=%s group=%s initialSequence=%d",
		brokers, topics, group, initialSequence)
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               string(brokers),
		"group.id":                        string(group),
		"session.timeout.ms":              sessionTimeoutMs,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"client.id":                       clientID,
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": true,
		"auto.offset.reset":    autoOffsetReset,
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}
	defer c.Close()
	log.Debugf("Created Consumer %v\n", c)

	err = c.SubscribeTopics([]string(topics), nil)

	if err != nil {
		log.Fatalf("Failed to subscribe to topics %s: %s\n", topics, err)
	}

	serveConsumerUI()

	consumeForever(ctx, c, initialSequence, useMessageHeaders)
}

// loops through the kafka consumer channel and consumes all events
// The loop runs forever until the context is cancled or a signal is sent (SIGINT or SIGTERM)
func consumeForever(
	ctx context.Context,
	c *kafka.Consumer,
	initialSequence types.SequenceNumber,
	useMessageHeaders bool,
) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigchan:
			log.Infof("Caught signal %v: terminating", sig)
			// TODO: Write a summary message to the console before quitting.
			return
		case <-ctx.Done():
			log.Infof("Done. %s", ctx.Err())
			return
		case ev := <-c.Events():
			// Most events are typically juse messages, still we are also interested in
			// Partition changes, EOF and Errors
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Infof("AssignedPartitions %v", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Infof("RevokedPartitions %v", e)
				c.Unassign()
			case *kafka.Message:
				processMessage(e, useMessageHeaders)
			case kafka.PartitionEOF:
				log.Debugf("PartitionEOF Reached %v", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				log.Errorf("Error: %+v", e)
			}
		}
	}
}

// Process a single message, keeping track of latency data and sequence numbers.
func processMessage(
	msg *kafka.Message,
	useMessageHeaders bool,
) {
	data := message.Extract(msg, useMessageHeaders)
	log.Tracef("Data: %s", data)
	validateSequence(data)
	collectThroughput(data)
	collectLatencyStats(data)
}
