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

// Create a mew message with headers, timestamp and size.
// Does not set TopicPartition.
func Create(
	id types.ProducerID,
	seq types.SequenceNumber,
	size types.MessageSize,
	useMessageHeaders bool,
) *kafka.Message {
	msg := &kafka.Message{
		Timestamp:     time.Now().UTC(),
		TimestampType: kafka.TimestampCreateTime,
	}
	if useMessageHeaders {
		msg.Value = make([]byte, size)
		msg.Headers = []kafka.Header{
			{
				Key:   KeyProducerID,
				Value: []byte(id),
			},
			{
				Key:   KeySequence,
				Value: []byte(strconv.FormatInt(int64(seq), 10)),
			},
		}
	} else {
		msg.Value = []byte(format(id, seq, size))
	}
	return msg
}

// Extract the data from the message and set timestamp and latencies
func Extract(
	msg *kafka.Message,
	useMessageHeaders bool,
) *Data {
	now := time.Now().UTC()
	var topic types.Topic
	if msg.TopicPartition.Topic != nil {
		topic = types.Topic(*msg.TopicPartition.Topic)
	} else {
		topic = types.Topic("")
	}
	data := &Data{
		ProducerTimestamp: msg.Timestamp,
		ConsumerTimestamp: now,
		Latency:           now.Sub(msg.Timestamp),
		Topic:             topic,
		Payload:           msg.Value,
	}
	if useMessageHeaders {
		data.ProducerID = getProducerID(msg)
		data.Sequence = getSequence(msg)
		data.Payload = msg.Value
	} else {
		parsed, err := parse(string(msg.Value))
		if err != nil {
			log.Errorf("Error parsing message %s", string(msg.Value))
			return data
		}
		data.ProducerID = parsed.producerID
		data.Sequence = parsed.sequence
		data.Payload = parsed.payload
	}
	return data
}

func getProducerID(msg *kafka.Message) types.ProducerID {
	v := getHeader(msg, KeyProducerID)
	if v == nil {
		return types.ProducerID("")
	}
	return types.ProducerID(string(v))
}

func getSequence(msg *kafka.Message) types.SequenceNumber {
	str := string(getHeader(msg, KeySequence))
	if str == "" {
		return -1
	}
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
