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

// Topic describes a name of a kafka topic
type Topic string

// Topics is just an array of topics
type Topics []string

// SequenceNumber represents a sequence number in a message used for testing the order of the message
type SequenceNumber int64

// ConsumerGroup for kafka
type ConsumerGroup string
