#!/bin/bash

set -x
# Test Kafka to see if a topic had been replicated
kubectl --context eu-west-1.k8s.local -n kafka-destination wait --for=condition=Ready pod/kafka-destination-0 --timeout=-1s
kubectl --context us-east-1.k8s.local -n kafka-source wait --for=condition=Ready pod/kafka-source-0 --timeout=-1s

kubectl --context eu-west-1.k8s.local -n brooklin wait --for=condition=Available deployment/brooklin --timeout=-1s
while [[ $(kubectl --context eu-west-1.k8s.local get pods -n brooklin -l app=brooklin -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' | cut -d' ' -f1) != "True" ]]; do echo "waiting for brooklin pod..." && sleep 10; done

# Run end to end tests. Produce to the source cluster, consume from the destination cluster
TOPIC="_test_replicator_$(date +%s)"
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; echo '                                     >>>>>>>>>>>>>  REPLICATOR GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic $TOPIC"

$(dirname "$0")/replicate-topic.sh $TOPIC

kubectl --context eu-west-1.k8s.local exec -n kafka-destination kafka-destination-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic $TOPIC --max-messages 1"

$(dirname "$0")/delete-replicate-topic.sh $TOPIC