#!/bin/bash

set -x
# Test ZK from the inside
#kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
until POD=$(kubectl --context us-east-1.k8s.local -n kafka-source get po pzoo-source-0 | grep Running)
do
  echo "ZK on source isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi
kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0

# Test ZK from the outside
zookeeper-shell $(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po pzoo-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'):2181 get /brokers/ids/0

# Test Kafka from the inside
until POD=$(kubectl --context us-east-1.k8s.local -n kafka-source get po kafka-source-0 | grep Running)
do
  echo "ZK on source isn't ready yet"
  sleep 20
  FIRST_TIME=1
done
if [ $FIRST_TIME ]; then
    sleep 60
fi
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; echo '\t\t >>>>>>>>>>>>>  GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic test_topic_1"
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic test_topic_1 --max-messages 1"

# Test kafka from the outside
kafka-console-consumer --bootstrap-server $(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po kafka-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'):9093 --topic test_topic_1 --from-beginning --max-messages 1
