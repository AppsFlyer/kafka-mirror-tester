#!/bin/bash

set -x
set -e

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters. Looking for topic name"
    exit 1
fi

topic_name=$1

kafka_source_ip=$(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po kafka-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}')
brooklin_pod=$(kubectl --context eu-west-1.k8s.local get pods -n brooklin -l app=brooklin -o 'jsonpath={.items[0].metadata.name}')
kubectl --context eu-west-1.k8s.local exec -n brooklin $brooklin_pod -- bash -c "unset JMX_OPTS; unset JMX_PORT; unset OPTS; \$BROOKLIN_HOME/bin/brooklin-rest-client.sh -o CREATE -u http://localhost:32311/ -n mirror-$topic_name -s \"kafka://$kafka_source_ip:9093/^$topic_name$\" -c kafkaMirroringC -t kafkaTP -m '{\"owner\":\"test-user\",\"system.reuseExistingDestination\":\"false\",\"system.destination.identityPartitioningEnabled\":true}' 2>/dev/null"
