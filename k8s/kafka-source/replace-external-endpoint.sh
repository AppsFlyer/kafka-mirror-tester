#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

until EP=$(kubectl -n kafka-source --context us-east-1.k8s.local get svc kafka-broker-external -o=jsonpath='{.status.loadBalancer.ingress[0].hostname}')
do
  echo "Load balancer still not ready"
  sleep 5
done


sed "s/__EXTERNAL_PLAINTEXT__/$EP/" $DIR/10broker-config.yml.tmpl > $DIR/10broker-config.yml
