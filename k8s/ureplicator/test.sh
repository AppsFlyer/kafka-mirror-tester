#!/bin/bash

set -x
# Test Kafka to see if a topic had been replicated
until POD=$(kubectl --context eu-west-1.k8s.local -n kafka-destination get po kafka-destination-0 | grep Running)
do
  echo "KAFKA on destination isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi

until POD=$(kubectl --context us-east-1.k8s.local -n kafka-source get po kafka-source-0 | grep Running)
do
  echo "KAFKA on source isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi


# Run end to end tests. Produce to the source cluster, consume from the destination cluster
TOPIC="_test_replicator_$(date +%s)"
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; echo '                                     >>>>>>>>>>>>>  GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic $TOPIC"
kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic $TOPIC --max-messages 1"
