package producer

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/deckarep/golang-set"
	"github.com/dustin/go-humanize"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const monitoringFrequency = 5 * time.Second

var (
	// messageRateCounter is used in order to observe the actual throughput
	messageRateCounter *ratecounter.RateCounter
	messageCounter     prometheus.Counter
	// bytesRateCounter measures the actual throughput in bytes
	bytesRateCounter *ratecounter.RateCounter
	bytesCounter     prometheus.Counter

	topicsGauge     prometheus.Gauge
	partitionsGauge prometheus.Gauge

	// number of messages that are bandwidth throttled or kafka-server throttled.
	// This is the number of messages that were supposed to be sent but got throttled and are lagging behind.
	badwidthThrottledMessages prometheus.Counter

	// Number of currently client-side in-flight messages (messages buffered but not yet sent)
	inflightMessageCount prometheus.GaugeFunc
)

func init() {
	messageRateCounter = ratecounter.NewRateCounter(monitoringFrequency)
	bytesRateCounter = ratecounter.NewRateCounter(monitoringFrequency)
}

func reportMessageSent(m *kafka.Message) {
	messageRateCounter.Incr(1)
	messageCounter.Inc()
	l := len(m.Value)
	bytesRateCounter.Incr(int64(l))
	bytesCounter.Add(float64(l))
}

// periodically monitors the kafka writer.
// Blocks forever or until canceled.
func monitor(
	ctx context.Context,
	errorCounter *uint64,
	frequency time.Duration,
	desiredThroughput types.Throughput,
	id types.ProducerID,
	numTopics, numPartitions uint,
	producers mapset.Set,
) {
	initPrometheus(numTopics, numPartitions, producers)
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

func initPrometheus(
	numTopics, numPartitions uint,
	producers mapset.Set,
) {
	messageCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_produced",
		Help: "Number of messages produced to kafka.",
	})
	bytesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bytes_produced",
		Help: "Number of bytes produced to kafka.",
	})
	topicsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "producer_number_of_topics",
		Help: "Number of topics that the producer writes to.",
	})
	topicsGauge.Add(float64(numTopics))

	partitionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "producer_number_of_partitions",
		Help: "Number of partitions of each topic that the producer writes to.",
	})
	partitionsGauge.Add(float64(numPartitions))

	badwidthThrottledMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bandwidth_throttled_messages",
		Help: "Number of messages throttled after sending.",
	})
	inflightMessageCount = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "in_flight_message_count",
		Help: "Number of currently in-flight messages (client side)",
	}, inFlightMessageCounter(producers))
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8001", nil)
}

func inFlightMessageCounter(producers mapset.Set) func() float64 {
	return func() float64 {
		sum := 0
		for p := range producers.Iterator().C {
			producer := p.(*kafka.Producer)
			sum += producer.Len()
		}
		return float64(sum)
	}
}

// Prints some runtime stats such as errors, throughputs etc
func printStats(
	errorCounter *uint64,
	frequency time.Duration,
	desiredThroughput types.Throughput,
	id types.ProducerID,
) {
	frequencySeconds := int64(frequency / time.Second)
	messageThroughput := messageRateCounter.Rate() / frequencySeconds
	bytesThroughput := uint64(bytesRateCounter.Rate() / frequencySeconds)
	errors := atomic.LoadUint64(errorCounter)
	log.Infof(`Recent stats for %s:
	Throughput: %d messages / sec
	Throughput: %s / sec
	Total errors: %d
	`, id, messageThroughput, humanize.Bytes(bytesThroughput), errors)

	// How much slack we're willing to take if throughput is lower than desired
	const slack = .9

	if float32(messageThroughput) < float32(desiredThroughput)*slack {
		log.Warnf("Actual throughput is < desired throughput. %d < %d", messageThroughput, desiredThroughput)
		badwidthThrottledMessages.Add(float64(desiredThroughput) - float64(messageThroughput))
	}
}
