package types

// Common type definitions for this project

// Brokers is a coma seperated string of host:port
type Brokers string

// Throughput describes a message send throughput measured by number of messages per second
type Throughput uint

// MessageSize describes a message size in bytes
type MessageSize uint

// ProducerID describes an ID for a producer
type ProducerID string

// MessageKey describes a message key in kafka. We define them as uint b/c we want 
// to enfoce that as part of the business logic. 
// One thing worth mentioning is that in Kafka message keys are not generally required to be unique.
// In fact we use them simply for message routing b/w partitions
// and we expect each key to repeat many times
type MessageKey uint

// Topic describes a name of a kafka topic
type Topic string

// Topics is just an array of topics
type Topics []string

// SequenceNumber represents a sequence number in a message used for testing the order of the message
type SequenceNumber int64

// ConsumerGroup for kafka
type ConsumerGroup string
