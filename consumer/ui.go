package consumer

import (
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/zserge/metric"
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
		htmlUI()
	})
}

// Create an endpoint which serves HTML UI.
func htmlUI() {
	expvar.Publish("Message received latency", metricHistogram)
	expvar.Publish("bytes received", bytesReceivedMetric)
	expvar.Publish("same message as current sequence", sameMessagesMetric)
	expvar.Publish("old message (old sequence nunmber)", oldMessagesMetric)
	expvar.Publish("correct message in order", inOrderMessagesMetric)
	expvar.Publish("skipped messages", skippedMessagesMetric)

	http.Handle("/debug/metrics", metric.Handler(metric.Exposed))
	go http.ListenAndServe(":8000", nil)
}

// Periodically emit statistics to the terminal.
func terminalUI() {
	ticker := time.Tick(terminalReportingFrequency)
	const terminalWidth = 50
	lastRead := uint64(0)
	lastBytes := uint64(0)
	go func() {
		for {
			<-ticker
			inOrder := atomic.LoadUint64(&inOrderMessagesCount)
			bytesTotal := atomic.LoadUint64(&bytesReceivedCount)
			readRate := int((inOrder - lastRead) / uint64((terminalReportingFrequency / time.Second)))
			bytesRate := uint64((bytesTotal - lastBytes) / uint64((terminalReportingFrequency / time.Second)))
			metrics := tachymeterHistogram.Calc()
			tachymeterHistogram.Reset()

			fmt.Printf("\n\n\n\tSTATS\n")
			//print a visual histogram of latencies
			fmt.Println(metrics.Histogram.String(terminalWidth))
			// print statistics about latencies
			fmt.Println(metrics.String())
			fmt.Printf("\nRead rate: %d messages/sec \t Byte rate: %s/sec \n", readRate, humanize.Bytes(bytesRate))
			fmt.Printf("\nsameMessagesCount=%d, oldMessagesCount=%d, inOrderMessagesCount=%d, skippedMessagesCount=%d",
				atomic.LoadUint64(&sameMessagesCount),
				atomic.LoadUint64(&oldMessagesCount),
				inOrder,
				atomic.LoadUint64(&skippedMessagesCount))
			lastRead = inOrder
			lastBytes = bytesTotal
		}
	}()
}
