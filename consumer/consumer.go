// Package consumer implements the consumption, performance measurement and validation logic of the test
package consumer

import (
	"context"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const (
	sessionTimeoutMs = 6000
	autoOffsetReset  = "earliest"
)

var clientID string

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Can't get hostname %+v", err)
	}
	clientID = fmt.Sprintf("kafka-mirror-tester-%s", hostname)
}

// ConsumeAndAnalyze consumes messages from the kafka toipc and analyzes their correctness and performance.
// The function blocks forever (or until the context is cancled)
func ConsumeAndAnalyze(
	ctx context.Context,
	brokers types.Brokers,
	topics types.Topics,
	group types.ConsumerGroup,
	initialSequence types.SequenceNumber,
) {
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

	consumeForever(ctx, c, initialSequence)
}

func consumeForever(
	ctx context.Context,
	c *kafka.Consumer,
	initialSequence types.SequenceNumber,
) {
	log.SetLevel(log.TraceLevel)
	for {
		select {
		case <-ctx.Done():
			log.Infof("Done. %s", ctx.Err())
			return
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Debugf("AssignedPartitions %v", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Debugf("RevokedPartitions %v", e)
				c.Unassign()
			case *kafka.Message:
				// TODO
				log.Tracef("Message on %s:\n%s\n", e.TopicPartition, string(e.Value))
			case kafka.PartitionEOF:
				log.Debugf("PartitionEOF Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				log.Errorf("Error: %v\n", e)
			}
		}
	}
}
