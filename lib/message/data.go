package message

import (
	"fmt"
	"time"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/types"
)

// Data represent the data sent in a message.
type Data struct {
	ProducerID        types.ProducerID
	MessageKey        types.MessageKey
	Sequence          types.SequenceNumber
	ProducerTimestamp time.Time
	ConsumerTimestamp time.Time
	Latency           time.Duration // In nanoseconds
	Topic             types.Topic
	// The actual payload (without metadata)
	Payload []byte
	// The total payload lenght, including metadata sent inside the payload
	TotalPayloadLength uint64
}

func (d Data) String() string {
	return fmt.Sprintf("message.Data[ProducerID=%s, MessageKey=%d, Topic=%s, Sequence=%d, Latency=%dms len(Payload)=%db]",
		d.ProducerID, d.MessageKey, d.Topic, d.Sequence, d.LatencyMS(), len(d.Payload))
}

// LatencyMS returns the latency in ms
func (d Data) LatencyMS() int64 {
	return int64(d.Latency / 1e6)
}

// Data parsed from the payload (when headers are not used)
type parsedData struct {
	producerID types.ProducerID
	sequence   types.SequenceNumber
	timestamp  time.Time
	payload    []byte
}
