#!/bin/bash
set +x

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd $DIR

until IP=$(kubectl --context us-east-1.k8s.local get node $(kubectl --context us-east-1.k8s.local -n kafka-source get po pzoo-source-0 -o jsonpath='{.spec.nodeName}') -o jsonpath='{.status.addresses[?(@.type=="ExternalIP")].address}')
do
  echo "ZK on source isn't ready yet"
  sleep 5
done

sed "s/__SRC_ZK_CONNECT__/$IP:2181/" 25env-config.yml.tmpl > 25env-config.yml

