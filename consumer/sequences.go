package consumer

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zserge/metric"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

var (
	// Map of producer:topic -> latest sequence number
	receivedSequenceNumbers map[string]types.SequenceNumber

	// For each measurement there are two kind of counters, one is a simple counter that
	// simply keeps count of how many such events occured.
	// And the other is a metric.Metric instance which measures temporary values of that count
	// (e.g. last minute, last 15 minutes etc)
	sameMessagesMetric    metric.Metric
	sameMessagesCount     uint64
	oldMessagesMetric     metric.Metric
	oldMessagesCount      uint64
	inOrderMessagesMetric metric.Metric
	inOrderMessagesCount  uint64
	skippedMessagesMetric metric.Metric
	skippedMessagesCount  uint64

	measurementIntervals = []string{"1m1s", "15m10s", "1h1m"}
)

func init() {
	receivedSequenceNumbers = make(map[string]types.SequenceNumber)
	sameMessagesMetric = metric.NewCounter(measurementIntervals...)
	oldMessagesMetric = metric.NewCounter(measurementIntervals...)
	inOrderMessagesMetric = metric.NewCounter(measurementIntervals...)
	skippedMessagesMetric = metric.NewCounter(measurementIntervals...)
}

// For each message validates that the sequence numnber that corresponds to the producer and the topic
// are in order.
// If they are not in order, will log it and accumulate in counters.
func validateSequence(data *message.Data) {
	seq := data.Sequence
	key := createSeqnenceNumberKey(data.ProducerID, data.Topic)
	latestSeq, exists := receivedSequenceNumbers[key]
	if !exists {
		// key not found, let's insert it first
		if seq != 0 {
			log.Warnf("Received initial sequence number > 0. topic=%s producer=%s number=%d",
				data.Topic, data.ProducerID, data.Sequence)
		}
		receivedSequenceNumbers[key] = seq
		log.Tracef("Message received first of it's producer-topic: %s", data)
		inOrderMessagesMetric.Add(1)
		inOrderMessagesCount++
		return
	}

	switch {
	case seq == latestSeq:
		// Same message twice? That's OK, let's just log it
		log.Infof("Received the same message again: %s", data)
		sameMessagesMetric.Add(1)
		sameMessagesCount++
	case seq < latestSeq:
		// Received an old message
		log.Infof("Received old data. Current seq=%d, but received %s", latestSeq, data)
		oldMessagesMetric.Add(1)
		oldMessagesCount++
	case seq == latestSeq+1:
		// That's just perfect!
		log.Tracef("Message received in order %s", data)
		inOrderMessagesMetric.Add(1)
		inOrderMessagesCount++
	case seq > latestSeq+1:
		// skipped a few sequences :-(
		howMany := seq - latestSeq
		log.Warnf("Skipped a few messages (%d messages). Current seq=%d, received %s",
			howMany, latestSeq, data)
		skippedMessagesMetric.Add(float64(howMany))
		skippedMessagesCount += uint64(howMany)
	}
	receivedSequenceNumbers[key] = seq
}

// create a key for the sequence number map
func createSeqnenceNumberKey(pid types.ProducerID, topic types.Topic) string {
	return fmt.Sprintf("%s:%s", pid, topic)
}
