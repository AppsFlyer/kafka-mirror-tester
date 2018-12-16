package message

import (
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

const (
	// KeySequence identifies the sequence number header
	KeySequence = "seq"

	// KeyProducerID identifies the producer ID header
	KeyProducerID = "id"
)

// Data represent the data sent in a message.
type Data struct {
	ProducerID        types.ProducerID
	Sequence          types.SequenceNumber
	ProducerTimestamp time.Time
	ConsumerTimestamp time.Time
	Latency           time.Duration
	Payload           []byte
}

// Create a mew message with headers, timestamp and size.
// Does not set TopicPartition.
func Create(
	id types.ProducerID,
	seq types.SequenceNumber,
	size types.MessageSize,
) *kafka.Message {
	payload := make([]byte, size)
	return &kafka.Message{
		Value:         payload,
		Timestamp:     time.Now().UTC(),
		TimestampType: kafka.TimestampCreateTime,
		Headers: []kafka.Header{
			{
				Key:   KeyProducerID,
				Value: []byte(id),
			},
			{
				Key:   KeySequence,
				Value: []byte(strconv.FormatInt(int64(seq), 10)),
			},
		},
	}
}

// Extract the data from the message and set timestamp and latencies
func Extract(msg *kafka.Message) *Data {
	now := time.Now().UTC()
	return &Data{
		ProducerID:        getProducerID(msg),
		Sequence:          getSequence(msg),
		ProducerTimestamp: msg.Timestamp,
		ConsumerTimestamp: now,
		Latency:           now.Sub(msg.Timestamp),
		Payload:           msg.Value,
	}
}

func getProducerID(msg *kafka.Message) types.ProducerID {
	return types.ProducerID(getHeader(msg, KeyProducerID))
}

func getSequence(msg *kafka.Message) types.SequenceNumber {
	str := string(getHeader(msg, KeySequence))
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalf("Malformed Sequence Number %s. %+v", str, err)
	}
	return types.SequenceNumber(i)
}

func getHeader(msg *kafka.Message, key string) []byte {
	for _, h := range msg.Headers {
		if h.Key == key {
			return h.Value
		}
	}
	// header not found
	return nil
}
