#!/bin/bash
set +x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd $DIR

until LB="http://$(kubectl --context us-east-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.status.loadBalancer.ingress[0].hostname}"):$(kubectl --context us-east-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.spec.ports[0].port}")"
do
  echo "Prometheus on us-east-1 isn't ready yet"
  sleep 5
done

sed "s|__US_EAST_1_PROMETHEUS__|$LB|" grafana-datasources.yaml.tmpl > grafana-datasources.yaml

