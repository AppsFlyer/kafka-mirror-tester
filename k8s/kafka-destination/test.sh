#!/bin/bash

set -x
# Test ZK
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo

until POD=$(kubectl --context eu-west-1.k8s.local -n kafka-destination get po pzoo-destination-0 | grep Running)
do
  echo "ZK on destination isn't ready yet"
  sleep 5
done
kubectl --context eu-west-1.k8s.local exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0

# Test Kafka from the inside
until POD=$(kubectl --context eu-west-1.k8s.local -n kafka-destination get po kafka-destination-1 | grep Running)
do
  echo "KAFKA on destination isn't ready yet"
  sleep 5
done
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; echo '>>>>>>>>>>>>>  GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic test_topic_2"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-1 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic test_topic_2 --max-messages 1"
