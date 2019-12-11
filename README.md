# Kafka Mirror Tester

Kafka mirror tester is a tool meant to test the performance and correctness of apache kafka mirroring.
Mirroring is not one of kafka's built in properties, but there are 3rd party tools that implement mirroring, namely:

1. Kafka's [Mirror Maker](https://kafka.apache.org/documentation.html#basic_ops_mirror_maker), a relatively simple tool within the kafka project with some known limitations
2. [Confluent's Replicator](https://docs.confluent.io/current/multi-dc-replicator/index.html), a paid for tool from confluent.
3. Uber's open source [uReplicator](https://github.com/uber/uReplicator)
4. LinkedIn [Brooklin](https://github.com/linkedin/brooklin)

This test tool is indifferent to the underlying mirroring tool so it is able to test all the above mentioned replicators.

*The current implementation supports Uber's uReplicator and Linkedin's Brooklin.*

Presentation on this project: https://speakerdeck.com/rantav/infrastructure-testing-using-kubernetes


## High level design

Mirroring typically takes place between two datacenters as described below:

```
----------------------                        --------------------------------------
|                    |                        |                                    |
|  Source DC         |                        | Destination DC                     |
|                    |                        |                                    |
|  ----------------  |                        | --------------    ---------------- |
|  | Source Kafka |  | - - - - - - - - - - -> | | Replicator | -> | Target Kafka | |
|  ________________  |                        | --------------    ---------------- |
|                    |                        |                                    |
----------------------                        --------------------------------------
```

The test tool has the following goals in mind:

1. Correctness, mainly completeness - that all messages sent to Source arrived in order at Destination (per partition). At least once semantic.
2. Performance - how long does it take for messages to get replicated and sent to Destination. This, of course, takes into consideration the laws of manure, e.g. inherent line latency.

The test harness is therefore comprised  of two components: The `producer` and the `consumer`

### The producer
The producer writes messages with sequence numbers and timestamps.

### The consumer
The consumer reads messages and looks into the sequence numbers and timestamps to determine correctness and performance.
This assumes the producer and consumer's clocks are in sync (we don't require the punctuation of atomic clocks, but we do realize that out of sync clocks will influence accuracy)

## Lower level design

The producer writes its messages to the source kafka, adding it's `producer-id`, `sequence`, `timestamp` and a `payload`.
The producer is capable of throttling it's throughput so that we'd achieve predictable throughput.

```
----------------------                        --------------------------------------
|                    |                        |                                    |
|  Source DC         |                        | Destination DC                     |
|                    |                        |                                    |
|  ----------------  |                        | --------------    ---------------- |
|  | Source Kafka |  | - - - - - - - - - - -> | | Replicator | -> | Target Kafka | |
|  ----------------  |                        | --------------    ---------------- |
|    ↑               |                        |                             |      |
|    |               |                        |                             |      |
|    |               |                        |                             ↓      |
|  ------------      |                        |                       ------------ |
|  | producer |      |                        |                       | consumer | |
|  ------------      |                        |                       ------------ |
----------------------                        --------------------------------------
```

### Message format

We aim for a simple, low overhead message format utilizing Kafka's built in header fields. And where headers are not supported (shamefully, the current reality with both uReplicator and Brooklin), we utilize an in-body message format.

Message format:
There are two variants of message formats, one that uses kafka headers and the other that does not. 
We implement two formats because while headers are nicer and easier to use, neither uReplicator nor Brooklin currently support them.

Message format with headers: (for simplicity, we use a json format but of course in Kafka it's all binary)
```
{
    value: payload,               // Payload size is determined by the user.
    timestamp: produceTime,       // The producer embeds a timestamp in UTC
    headers: {
        id: producer-id,
        seq: sequence-number
    }
}
```

Message format without headers (encoded in the message body itself):
```
+-------------------------------------------------+
| producer-id;sequence-number;timestamp;payload...|
+-------------------------------------------------+
```

We add the `producer-id` so that we can run the producers on multiple hosts and still be able to make sure that all messages arrived.

### Producer

Command line arguments:

`--id`: Producer ID. May be the hostname etc.

`--topics`: List of topic names to write to, separated by comas

`--throughput`: Number of messages per second per topic

`--message-size`: Message size, including the header section (producer-id;sequence-number;timestamp;). The minimal message size is around 30 bytes then due to a typical header length

`--bootstrap-server`: A kafka server from which to bootstrap

`--use-message-headers`: Whether to use message headers to encode metadata (or encode it within the payload)

The producer would generate messages containing the header and adding the payload for as long as needed in order to reach the `message-size` and send them to kafka.
It will try to achieve the desired throughput (send batches and in parallel) but will not exceed it. If it is unable to achieve the desired throughput we'll emit a log warning and continue. We also keep an eye on that using Prometheus and Grafana in the big picture.
The throughput is measured as the number of messages / second / topic.

### Consumer

Command line arguments:

`--topics`: List of topic names to read from, separated by comas

`--bootstrap-server`: A kafka server from which to bootstrap

`--use-message-headers`: Whether to use message headers to encode metadata (or encode it within the payload)

The consumer would read the messages from each of the topics and calculate correctness and performance.

Correctness is determined by the combination of `topic`, `producer-id` and `sequence-number` (e.g. if a specific producer has gaps that means we're  missing messages).

There is a fine point to mention in that respect. When operating with multiple partitions we utilize Kafka's message `key` in order to ensure message routing correctness. When multiple consumers read, (naturally from multiple partitions) we want each consumer to be able to read *all* sequential messages *in the order* they were sent. To achieve that we use Kafka's message routing abilities such that messages with the same key always routed to the same partition. What matters is the number of partitions in the destination cluster. To achieve linearity we sequence the messages modulo the number of partitions in the destination cluster. This way, all ascending sequence numbers are sent to the same partition in the same order and clients are then able to easily verify that all messages arrived in the order they were sent.


Latency is determined by the time gap between the `timestamp` and the current local consumer time. The consumer then emits a histogram of latency buckets.

## Open for discussion

1. If the last message from a producer got lost we don't know about it. If all messages from a specific producer got lost, we won't know about it either (although it's possible to manually audit that). If a certain partition is not replicated we can only see it by means of traffic volume monitoring, not precise counts.

# Using it.
The tools in this project expect some familiarity with 3rd party tools, namely Kubernetes and AWS. We don't expect expert level but some familiarity with the tools is very helpful.

For details how to run it see [Running it](running.md)

# Results

We have [benchmark results for uReplicator](results-ureplicator.md) and [benchmark results for brooklin](results-brooklin.md)
