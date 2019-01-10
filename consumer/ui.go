package consumer

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	terminalReportingFrequency = 10 * time.Second
)

var (
	// once is used for one-time initialization that we don't want to embed in the init function.
	once sync.Once
)

// Serve the different UIs for viewing metrics.
func serveConsumerUI() {
	once.Do(func() {
		terminalUI()
		initPrometheus()
	})
}

func initPrometheus() {
	latencySummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "message_arrival_latency_ms",
		Help:       "Latency in ms for message arrival e2e.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
	})
	prometheus.MustRegister(latencySummary)
	sameMessagesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "same_message_count",
		Help: "Number of times the same message was consumed.",
	})
	prometheus.MustRegister(sameMessagesCounter)
	oldMessagesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "old_message_count",
		Help: "Number of times an old message was consumed.",
	})
	prometheus.MustRegister(oldMessagesCounter)
	inOrderMessagesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "in_order_message_count",
		Help: "Number of times a message was received in order (this is the happy path).",
	})
	prometheus.MustRegister(inOrderMessagesCounter)
	skippedMessagesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "skipped_message_count",
		Help: "Number of times a message was skipped.",
	})
	prometheus.MustRegister(skippedMessagesCounter)
	messageCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "messages_consumed",
		Help: "Number of messages consumed from kafka.",
	})
	prometheus.MustRegister(messageCounter)

	bytesCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bytes_consumed",
		Help: "Number of bytes consumed from kafka.",
	})
	prometheus.MustRegister(bytesCounter)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8000", nil)
}

// Periodically emit statistics to the terminal.
func terminalUI() {
	ticker := time.Tick(terminalReportingFrequency)
	const terminalWidth = 50
	var (
		lastMessages uint64
		lastBytes    uint64
	)
	go func() {
		for {
			<-ticker
			messages := atomic.LoadUint64(&messageCount)
			bytes := atomic.LoadUint64(&bytesCount)
			reportingFrequencySec := uint64((terminalReportingFrequency / time.Second))
			messageRate := int((messages - lastMessages) / reportingFrequencySec)
			bytesRate := uint64((bytes - lastBytes) / reportingFrequencySec)
			metrics := tachymeterHistogram.Calc()
			tachymeterHistogram.Reset()

			fmt.Printf("\n\n\n\tSTATS\n")
			//print a visual histogram of latencies
			fmt.Println(metrics.Histogram.String(terminalWidth))
			// print statistics about latencies
			fmt.Println(metrics.String())
			fmt.Printf("\nRead rate: %d messages/sec \t Byte rate: %s/sec \n", messageRate, humanize.Bytes(bytesRate))
			fmt.Printf("\nsameMessagesCount=%d, oldMessagesCount=%d, inOrderMessagesCount=%d, skippedMessagesCount=%d",
				atomic.LoadUint64(&sameMessagesCount),
				atomic.LoadUint64(&oldMessagesCount),
				atomic.LoadUint64(&inOrderMessagesCount),
				atomic.LoadUint64(&skippedMessagesCount))
			lastMessages = messages
			lastBytes = bytes
		}
	}()
}
