#!/bin/bash

set -x
# Test ZK
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo

until POD=$(kubectl --context eu-west-1.k8s.local -n kafka-destination get po pzoo-destination-0 | grep Running)
do
  echo "ZK on destination isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi

kubectl --context eu-west-1.k8s.local exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0

# Test Kafka from the inside
until POD=$(kubectl --context eu-west-1.k8s.local -n kafka-destination get po kafka-destination-0 | grep Running)
do
  echo "KAFKA on destination isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi
TOPIC="_test_destination_$(date +%s)"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; echo '                                 >>>>>>>>>>>>>  DESTINATION GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic $TOPIC"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic $TOPIC --max-messages 1"
