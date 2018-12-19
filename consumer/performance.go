package consumer

import (
	"github.com/jamiealquiza/tachymeter"
	"github.com/zserge/metric"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
)

const (
	// Define a sample size of 500. This affects memory consumption v/s precision.
	// It's probably OK to increase this number by a lot but didn't test it yet.
	tachymeterSampleSize = 500
)

var (
	// We use two tools to measure the time performance. One is useful due to
	// it's web interface and the other is useful as a CLI interface.

	// This one has a nice web UI
	metricHistogram metric.Metric

	// And this one has a mice text UI.
	tachymeterHistogram *tachymeter.Tachymeter

	// Define the time windows in which metrics are aggregater for.
	// How to read this? "20s1s" means a chart will be displayed for 20 seconds and each
	// item in this chart is a 1 second average.
	tachymeterMeasurementWindows = []string{"20s1s", "1m1s", "2m1s", "15m30s", "1h1m"}
)

func init() {
	metricHistogram = metric.NewHistogram(tachymeterMeasurementWindows...)
	tachymeterHistogram = tachymeter.New(&tachymeter.Config{Size: tachymeterSampleSize})
}

// Collect the latency stats from the data into the various counters.
func collectLatencyStats(data *message.Data) {
	metricHistogram.Add(float64(data.LatencyMS()))
	tachymeterHistogram.AddTime(data.Latency)
}
