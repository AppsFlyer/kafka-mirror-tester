package producer

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dustin/go-humanize"
	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
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
) {
	initPrometheus(numTopics, numPartitions)
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

func initPrometheus(numTopics, numPartitions uint) {
	messageCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "messages_produced",
		Help: "Number of messages produced to kafka.",
	})
	prometheus.MustRegister(messageCounter)
	bytesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bytes_produced",
		Help: "Number of bytes produced to kafka.",
	})
	prometheus.MustRegister(bytesCounter)

	topicsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "producer_number_of_topics",
		Help: "Number of topics that the producer writes to.",
	})
	prometheus.MustRegister(topicsGauge)
	topicsGauge.Add(float64(numTopics))

	partitionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "producer_number_of_partitions",
		Help: "Number of partitions of each topic that the producer writes to.",
	})
	prometheus.MustRegister(partitionsGauge)
	partitionsGauge.Add(float64(numPartitions))

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8001", nil)
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
	}
}
