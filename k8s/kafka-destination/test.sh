kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
kubectl exec -n kafka-destination pzoo-destination-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
