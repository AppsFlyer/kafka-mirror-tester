package message

import (
	"fmt"
	"time"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

// Data represent the data sent in a message.
type Data struct {
	ProducerID        types.ProducerID
	Sequence          types.SequenceNumber
	ProducerTimestamp time.Time
	ConsumerTimestamp time.Time
	Latency           time.Duration // In nanoseconds
	Topic             types.Topic
	Payload           []byte
}

func (d Data) String() string {
	return fmt.Sprintf("message.Data[ProducerID=%s, Topic=%s, Sequence=%d, Latency=%dms len(Payload)=%db]",
		d.ProducerID, d.Topic, d.Sequence, d.LatencyMS(), len(d.Payload))
}

// LatencyMS returns the latency in ms
func (d Data) LatencyMS() int64 {
	return int64(d.Latency / 1e6)
}
