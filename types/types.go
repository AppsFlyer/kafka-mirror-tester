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

// SequenceNumber represents a sequence number in a message used for testing the order of the message
type SequenceNumber uint64
