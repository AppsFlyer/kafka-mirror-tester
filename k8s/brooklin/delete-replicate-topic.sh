#!/bin/bash

set -x
set -e

if [ "$#" -ne 1 ]; then
    echo "Illegal number of parameters. Looking for topic name"
    exit 1
fi

topic_name=$1

brooklin_pod=$(kubectl --context eu-west-1.k8s.local get pods -n brooklin -l app=brooklin -o 'jsonpath={.items[0].metadata.name}')
kubectl --context eu-west-1.k8s.local exec -n brooklin $brooklin_pod -- bash -c "unset JMX_OPTS; unset JMX_PORT; unset OPTS; \$BROOKLIN_HOME/bin/brooklin-rest-client.sh -o DELETE -u http://localhost:32311/ -n mirror-$topic_name 2> /dev/null"
