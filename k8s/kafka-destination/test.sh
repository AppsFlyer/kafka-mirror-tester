#!/bin/bash

set -x
# Test ZK
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
kubectl --context eu-west-1.k8s.local -n kafka-destination wait --for=condition=Ready pod/pzoo-destination-0 --timeout=-1s

# wait some, to make sure ZK is with us
sleep 20

kubectl --context eu-west-1.k8s.local exec -n kafka-destination pzoo-destination-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0"

# Test Kafka from the inside
kubectl --context eu-west-1.k8s.local -n kafka-destination wait --for=condition=Ready pod/kafka-destination-0 --timeout=-1s

# wait some, to make sure kafka is with us
sleep 20

TOPIC="_test_destination_$(date +%s)"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; echo '                                 >>>>>>>>>>>>>  DESTINATION GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic $TOPIC"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic $TOPIC --max-messages 1"
