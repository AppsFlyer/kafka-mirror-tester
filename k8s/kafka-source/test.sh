#!/bin/bash

set -x
# Test ZK
#kubectl exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
kubectl exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0

# Test Kafka
kubectl exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; echo 'GREAT SUCCESS!' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic test_topic_1"
kubectl exec -n kafka-source kafka-source-1 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic test_topic_1 --max-messages 1"

# Test kafka from the outside
kafka-console-consumer --bootstrap-server $(kubectl get node $(kubectl -n kafka-source get po kafka-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'):9093 --topic test_topic_1 --from-beginning --max-messages 1
