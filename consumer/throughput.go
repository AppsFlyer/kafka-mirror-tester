package consumer

import (
	"sync/atomic"

	"github.com/zserge/metric"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
)

var (
	// For each measurement there are two kind of counters, one is a simple counter that
	// simply keeps count of how many such events occured.
	// And the other is a metric.Metric instance which measures temporary values of that count
	// (e.g. last minute, last 15 minutes etc)
	bytesMetric   metric.Metric
	bytesCount    uint64
	messageMetric metric.Metric
	messageCount  uint64
)

func init() {
	bytesMetric = metric.NewCounter(measurementIntervals...)
	messageMetric = metric.NewCounter(measurementIntervals...)
}

// Count the total throughput (message count and byte count)
func collectThroughput(data *message.Data) {
	bytes := data.TotalPayloadLength
	bytesMetric.Add(float64(bytes))
	atomic.AddUint64(&bytesCount, bytes)
	messageMetric.Add(1)
	atomic.AddUint64(&messageCount, 1)
}
