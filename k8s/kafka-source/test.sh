kubectl exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 create /foo bar
kubectl exec -n kafka-source pzoo-source-0 -- /opt/kafka/bin/zookeeper-shell.sh localhost:2181 get /foo
