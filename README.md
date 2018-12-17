# Kafka Mirror Tester

Kafka mirror tester is a tool meant to test the performance and correctness of apache kafka mirroring.
Mirroring is not one of kafka's built in properties, but there are 3rd party tools that implement mirroring, namely:

1. Kafka's [Mirror Maker](https://kafka.apache.org/documentation.html#basic_ops_mirror_maker), a relatively simple tool within the kafka project with some known limitations
2. [Confluent's Replicator](https://docs.confluent.io/current/multi-dc-replicator/index.html), a paid for tool from confluent.
3. Uber's open source [uReplicator](https://github.com/uber/uReplicator)

This test tool is indiferent to the underlying mirroring tool so it is able to test all the above mentioned replicators.

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

1. Test correctness, mainly completeness - that all messages sent to Source arrived at Destination. At least once semantic.
2. Test performance - how long does it take for messages to be replicated and sent to Destination. This of course takes into consideration the laws of manure, e.g. inherent line latency.

The test harness is therefore comprised  of two components: The `producer` and the `consumer`

### The producer
The producer writes messages with sequence numbers and timestamps.

### The consumer
The consumer reads messages and looks into the sequence numbers and timestamps to determine correctness and performance.
This assumes the producer and consumer's clocks are in sync (we don't necessarily require atomic clocks, but we do realize that out of sync clocks will influence accuracy)

## Lower level design

The producer writes its messages to the source kafka, adding it's `producer-id`, `sequence`, `timestamp` and a `payload`.
The producer is capable of throttling it's throughput.

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
We aim for a simple, low overhead message format utilizing Kafka's built in header fields

Message format: (for simplicity, we use a json format but of course in Kafka it's all binary)
```json
{
    value: payload, // Payload size is determined by the user.
    timestamp: produceTime, // The producer embeds a timestamp in UTC
    headers: {
        id: producer-id,
        seq: sequence-number
    }
}
```
We add the `producer-id` so that we can run the producers on multiple hosts and still be able to make sure that all messages arrived.

### Producer

Command line arguments:

`--id`: Producer ID. May be the hostname etc.

`--topics`: List of topic names to write to, separated by comas

`--throughput`: Number of messages per second per topic

`--message-size`: Message size, including the header section (producer-id;sequence-number;timestamp;). The minimal message size is around 30 bytes then due to a typical header length

`--bootstrap-server`: A kafka server from which to bootstrap


The producer would generate messages containing the header and adding the payload for as long as needed in order to reach the `message-size` and send them to kafka.
It will try to achieve the desired throughput (send batches and in prallel) but will not exceed it. If it is unable to achieve the desired throughput we'll emit a log warning and continue.
The throughput is measured as the numnber of messages / second / topic.

### Consumer

Command line arguments:

`--topics`: List of topic names to read from, separated by comas

`--bootstrap-server`: A kafka server from which to bootstrap


The consumer would read the messages from each of the topics and calculate correctness and performance.

Correctness is determined by the combination of `topic`, `producer-id` and `sequence-number` (e.g. if a specific producer has gaps that means we'rea missing messages)

Performance is determined by the time gap between the `timestamp` and the current local consumer time. The consumer then emits a histogram of latency buckets.


## Open for discussion
1. Open question: can we implement multiple consumers in order to increase the consumption throughput? (and at the same time maintain correct bookeeping for correctness and performance)
1. If the last message from a producer got lost we don't know about it. If all messages from a specific producer got lost, we won't know about it either (although it's possible to manually audit that)