package producer

import (
	"context"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/paulbellamy/ratecounter"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const monitoringFrequency = 5 * time.Second

var (
	// messageCounter is used in order to observe the actual throughput
	messageCounter *ratecounter.RateCounter
	// bytesCounter measures the actual throughput in bytes
	bytesCounter *ratecounter.RateCounter
)

func init() {
	messageCounter = ratecounter.NewRateCounter(monitoringFrequency)
	bytesCounter = ratecounter.NewRateCounter(monitoringFrequency)
}

// periodically monitors the kafka writer.
// Blocks forever or until canceled.
func monitor(
	ctx context.Context,
	errorCounter *uint,
	frequency time.Duration,
	desiredThroughput types.Throughput,
	id types.ProducerID,
) {
	ticker := time.Tick(frequency)
	for {
		select {
		case <-ticker:
			printStats(errorCounter, frequency, desiredThroughput, id)
		case <-ctx.Done():
			log.Infof("Monitor done. %s", ctx.Err())
			return
		}
	}
}

// Prints some runtime stats such as errors, throughputs etc
func printStats(
	errorCounter *uint,
	frequency time.Duration,
	desiredThroughput types.Throughput,
	id types.ProducerID,
) {
	frequencySeconds := int64(frequency / time.Second)
	messageThroughput := messageCounter.Rate() / frequencySeconds
	bytesThroughput := uint64(bytesCounter.Rate() / frequencySeconds)
	log.Infof(`Recent stats for %s:
	Throughput: %d messages / sec
	Throughput: %s / sec
	Total errors: %d
	`, id, messageThroughput, humanize.Bytes(bytesThroughput), *errorCounter)

	// How much slack we're willing to take if throughput is lower than desired
	const slack = .9

	if float32(messageThroughput) < float32(desiredThroughput)*slack {
		log.Warnf("Actual throughput is < desired throughput. %d < %d", messageThroughput, desiredThroughput)
	}
}
