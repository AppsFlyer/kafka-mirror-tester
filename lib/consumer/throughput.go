package consumer

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/message"
)

var (
	bytesCounter   prometheus.Counter
	bytesCount     uint64
	messageCounter prometheus.Counter
	messageCount   uint64
)

// Count the total throughput (message count and byte count)
func collectThroughput(data *message.Data) {
	bytes := data.TotalPayloadLength
	bytesCounter.Add(float64(bytes))
	atomic.AddUint64(&bytesCount, bytes)
	messageCounter.Inc()
	atomic.AddUint64(&messageCount, 1)
}
