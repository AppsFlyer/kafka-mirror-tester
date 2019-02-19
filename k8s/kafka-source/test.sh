#!/bin/bash

set -x
# Test ZK from the inside
#kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
#kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
kubectl --context us-east-1.k8s.local -n kafka-source wait --for=condition=Ready pod/pzoo-source-0 --timeout=-1s

# wait some, to make sure ZK is with us
sleep 20

kubectl --context us-east-1.k8s.local exec -n kafka-source pzoo-source-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /brokers/ids/0"

# Test ZK from the outside. We assume there's a zookeeper-shell installed locally on the developer's laptop
zookeeper-shell $(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po pzoo-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'):2181 get /brokers/ids/0


kubectl --context us-east-1.k8s.local -n kafka-source wait --for=condition=Ready pod/kafka-source-0 --timeout=-1s

# wait some, to make sure kafka is with us
sleep 20

TOPIC="_test_source_$(date +%s)"
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; echo '                                  >>>>>>>>>>>>>  SOURCE GREAT SUCCESS! <<<<<<<<<<<<<<<<' | /opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic $TOPIC"
kubectl --context us-east-1.k8s.local exec -n kafka-source kafka-source-0 -- bash -c "unset JMX_PORT; /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --from-beginning --topic $TOPIC --max-messages 1"

# Test kafka from the outside. This assumes there's a locally installed kafka-console-consumer script
kafka-console-consumer --bootstrap-server $(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po kafka-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}'):9093 --topic $TOPIC --from-beginning --max-messages 1
