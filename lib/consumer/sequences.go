package consumer

import (
	"fmt"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/message"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/types"
)

var (
	// Map of producer,topic,key -> latest sequence number recieved from this key
	receivedSequenceNumbers map[string]types.SequenceNumber

	// For each measurement there are two kind of counters, one is a simple counter that
	// simply keeps count of how many such events occured.
	// And the other is a prometheus.Counter instance which measures temporary values of that count
	// (e.g. last minute, last 15 minutes etc)
	sameMessagesCounter    prometheus.Counter
	sameMessagesCount      uint64
	oldMessagesCounter     prometheus.Counter
	oldMessagesCount       uint64
	inOrderMessagesCounter prometheus.Counter
	inOrderMessagesCount   uint64
	skippedMessagesCounter prometheus.Counter
	skippedMessagesCount   uint64
)

func init() {
	receivedSequenceNumbers = make(map[string]types.SequenceNumber)
}

// For each message validates that the sequence numnber that corresponds to the producer and the topic
// are in order.
// If they are not in order, will log it and accumulate in counters.
// The function accesses some global varialbe that aren't thread safe (receivedSequenceNumbers) which makes the function not thread safe by itself.
func validateSequence(data *message.Data) {
	seq := data.Sequence
	key := createSeqnenceNumberKey(data.ProducerID, data.Topic, data.MessageKey)
	latestSeq, exists := receivedSequenceNumbers[key]
	if !exists {
		// key not found, let's insert it first
		if seq != 0 {
			log.Infof("Received initial sequence number > 0. topic=%s producer=%s key=%d number=%d",
				data.Topic, data.ProducerID, data.MessageKey, data.Sequence)
		}
		receivedSequenceNumbers[key] = seq
		log.Tracef("Message received first of it's producer-topic: %s", data)
		inOrderMessagesCounter.Add(1)
		atomic.AddUint64(&inOrderMessagesCount, 1)
		return
	}

	switch {
	case seq == latestSeq:
		// Same message twice? That's OK, let's just log it
		log.Debugf("Received the same message again: %s", data)
		sameMessagesCounter.Add(1)
		atomic.AddUint64(&sameMessagesCount, 1)
	case seq < latestSeq:
		// Received an old message
		log.Debugf("Received old data. Current seq=%d, but received %s", latestSeq, data)
		oldMessagesCounter.Add(1)
		atomic.AddUint64(&oldMessagesCount, 1)
	case seq == latestSeq+1:
		// That's just perfect!
		log.Tracef("Message received in order %s", data)
		inOrderMessagesCounter.Add(1)
		atomic.AddUint64(&inOrderMessagesCount, 1)
	case seq > latestSeq+1:
		// skipped a few sequences :-(
		howMany := seq - latestSeq
		log.Debugf("Skipped a few messages (%d messages). Current seq=%d, received %s",
			howMany, latestSeq, data)
		skippedMessagesCounter.Add(float64(howMany))
		atomic.AddUint64(&skippedMessagesCount, uint64(howMany))
	}
	receivedSequenceNumbers[key] = seq
}

// create a key for the sequence number map
func createSeqnenceNumberKey(pid types.ProducerID, topic types.Topic, messageKey types.MessageKey) string {
	return fmt.Sprintf("%s:%s:%d", pid, topic, messageKey)
}
