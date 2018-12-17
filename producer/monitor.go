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

// periodically monitors the kafka writer.
// Blocks forever or until canceled.
func monitor(
	ctx context.Context,
	messageCounter *ratecounter.RateCounter,
	bytesCounter *ratecounter.RateCounter,
	errorCounter *uint,
	frequency time.Duration,
	desiredThroughput types.Throughput,
) {
	ticker := time.Tick(frequency)
	for {
		select {
		case <-ticker:
			printStats(messageCounter, bytesCounter, errorCounter, frequency, desiredThroughput)
		case <-ctx.Done():
			log.Infof("Monitor done. %s", ctx.Err())
			return
		}
	}
}

// Prints some runtime stats such as errors, throughputs etc
func printStats(
	messageCounter *ratecounter.RateCounter,
	bytesCounter *ratecounter.RateCounter,
	errorCounter *uint,
	frequency time.Duration,
	desiredThroughput types.Throughput,
) {
	frequencySeconds := int64(frequency / time.Second)
	messageThroughput := messageCounter.Rate() / frequencySeconds
	bytesThroughput := uint64(bytesCounter.Rate() / frequencySeconds)
	log.Infof(`Recent stats:
	Throughput: %d messages / sec
	Throughput: %s / sec
	Total errors: %d
	`, messageThroughput, humanize.Bytes(bytesThroughput), *errorCounter)

	// How much slack we're willing to take if throughput is lower than desired
	const slack = .9

	if float32(messageThroughput) < float32(desiredThroughput)*slack {
		log.Warnf("Actual throughput is < desired throughput. %d < %d", messageThroughput, desiredThroughput)
	}
}
