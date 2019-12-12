USE_UREPLICATOR := false
USE_BROOKLIN := true
US_EAST_NODE_COUNT := 20
EU_WEST_NODE_COUNT := 38

k8s-all: kops-check-version k8s-create-clusters k8s-monitoring k8s-kafkas-setup k8s-replicator-setup k8s-setup-weave-scope k8s-wait-for-kafkas k8s-run-tests k8s-help-monitoring

k8s-single-cluster: kops-check-version k8s-create-cluster-eu-west-1 k8s-wait-for-cluster-eu-west-1 k8s-allow-kubectl-node-access k8s-monitoring-eu-west-1 k8s-setup-weave-scope-eu-west-1

helm-unsafe:
	kubectl create clusterrolebinding permissive-binding --clusterrole=cluster-admin --user=admin --user=kubelet --group=system:serviceaccounts

kops-check-version:
	kops version | grep "Version 1.10.0" || (echo "Sorry, this was tested with kops Version 1.10.0"; exit 1)

k8s-monitoring: k8s-monitoring-us-east-1 k8s-monitoring-eu-west-1 k8s-monitoring-graphite-exporter k8s-grafana-configure

k8s-monitoring-us-east-1:
	kubectl config set current-context us-east-1.k8s.local
	make k8s-monitoring-create k8s-monitoring-create-dashboard

k8s-monitoring-eu-west-1:
	kubectl config set current-context eu-west-1.k8s.local
	make k8s-monitoring-create k8s-monitoring-create-dashboard

k8s-monitoring-create: k8s-monitoring-create-node-problem-detector k8s-monitoring-install-metrics-server k8s-monitoring-install-kube-state-metrics k8s-monitoring-expose-k8s-services
	# Reference:
	# 	https://www.sumologic.com/blog/cloud/how-to-monitor-kubernetes/
	#	https://coreos.com/operators/prometheus/docs/latest/user-guides/cluster-monitoring.html
	kubectl apply --validate=false -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/prometheus-operator/v0.19.0.yaml || echo aleady exists?

	# Create load balancers for easier access to promethous and graphana
	kubectl --namespace monitoring patch svc/grafana --patch '{"spec": {"type": "LoadBalancer"}}'
	kubectl --namespace monitoring patch svc/prometheus-k8s --patch '{"spec": {"type": "LoadBalancer"}}'

k8s-monitoring-create-dashboard:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/influxdb/heapster.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/influxdb/influxdb.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/heapster/master/deploy/kube-config/rbac/heapster-rbac.yaml
	kubectl apply -f k8s/monitoring/admin-service-account.yaml
	kubectl apply -f k8s/monitoring/admin-cluster-role-binding.yaml

k8s-monitoring-create-node-problem-detector:
	# See more here https://kubernetes.io/docs/tasks/debug-application-cluster/monitor-node-health/
	kubectl apply -f https://k8s.io/examples/debug/node-problem-detector.yaml

k8s-monitoring-install-metrics-server:
	# allows for kubectl top as well as the metrics API at: http://127.0.0.1:8001/apis/metrics.k8s.io/
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/metrics-server/v1.8.x.yaml


k8s-monitoring-install-kube-state-metrics:
	kubectl apply -f k8s/monitoring/kube-state-metrics-service.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-service-account.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-role.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-role-binding.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-deployment.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-cluster-role.yaml
	kubectl apply -f k8s/monitoring/kube-state-metrics-cluster-role-binding.yaml


k8s-monitoring-expose-k8s-services:
	kubectl apply -f k8s/monitoring/monitoring-expose-kube-scheduler.yaml
	kubectl apply -f k8s/monitoring/monitoring-expose-kube-controller-manager.yaml

k8s-create-clusters: k8s-create-cluster-us-east-1 k8s-create-cluster-eu-west-1 k8s-wait-for-cluster-us-east-1 k8s-wait-for-cluster-eu-west-1 k8s-allow-kubectl-node-access

k8s-allow-kubectl-node-access:
	# Various init containers use kubectl to get information about their nodes, so we add this permission for them
	kubectl create clusterrolebinding --clusterrole=system:controller:node-controller --serviceaccount=kafka-source:default kubectl-node-access  --context us-east-1.k8s.local || echo cluster binding alre exists?
	kubectl create clusterrolebinding --clusterrole=system:controller:node-controller --serviceaccount=kafka-destination:default kubectl-node-access  --context eu-west-1.k8s.local || echo cluster binding alre exists?

k8s-kafkas-setup: k8s-kafkas-setup-source k8s-kafkas-setup-destination k8s-kafkas-setup-source-validate k8s-kafkas-setup-destination-validate

k8s-kafkas-setup-source:
	kubectl apply -f k8s/kafka-source/ --context us-east-1.k8s.local
	# Punch a hole in the security group so that we can access Kafka from the outside
	aws ec2 authorize-security-group-ingress --group-id $$(aws ec2 describe-security-groups --filters Name=group-name,Values=nodes.us-east-1.k8s.local --region us-east-1 --output text --query 'SecurityGroups[0].GroupId') --protocol tcp --port 9093 --cidr 0.0.0.0/0 --region us-east-1 || echo already exists?
	# Punch another hole to zookeeper
	aws ec2 authorize-security-group-ingress --group-id $$(aws ec2 describe-security-groups --filters Name=group-name,Values=nodes.us-east-1.k8s.local --region us-east-1 --output text --query 'SecurityGroups[0].GroupId') --protocol tcp --port 2181 --cidr 0.0.0.0/0 --region us-east-1 || echo already exists?
k8s-kafkas-setup-source-validate:
	# validate
	k8s/kafka-source/test.sh
	kubectl --context us-east-1.k8s.local -n kafka-source get po -o wide
	# Logs:
	# stern --context us-east-1.k8s.local -n kafka-source -l app=kafka-source

k8s-kafkas-setup-destination:
	kubectl apply -f k8s/kafka-destination --context eu-west-1.k8s.local
k8s-kafkas-setup-destination-validate:
	k8s/kafka-destination/test.sh
	kubectl --context eu-west-1.k8s.local -n kafka-destination get po -o wide
	# Logs:
	# stern --context eu-west-1.k8s.local -n kafka-destination -l app=kafka-destination
k8s-replicator-setup:
ifeq ($(USE_UREPLICATOR),true)
	k8s/ureplicator/template.sh
	kubectl apply -f k8s/ureplicator --context eu-west-1.k8s.local
	k8s/ureplicator/test.sh
	kubectl --context eu-west-1.k8s.local -n ureplicator get po -o wide
	# View logs:
	# stern --context eu-west-1.k8s.local -n ureplicator -l app=ureplicator
endif
ifeq ($(USE_BROOKLIN),true)
	kubectl apply -f k8s/brooklin --context eu-west-1.k8s.local
	k8s/brooklin/test.sh
	k8s/brooklin/replicate-topic.sh 'topic.*' # Start replicting all topics that start with "topic" (regex)
	kubectl --context eu-west-1.k8s.local -n brooklin get po -o wide
	# View logs:
	# stern --context eu-west-1.k8s.local -n brooklin -l app=brooklin
endif

k8s-ureplicator-visit-controller:
	kubectl --context eu-west-1.k8s.local -n ureplicator port-forward $$(kubectl --context eu-west-1.k8s.local get -n ureplicator pod -l app=ureplicator -l component=controller -o jsonpath='{.items[0].metadata.name}') 9000 &
	open http://localhost:9000/instances/

k8s-setup-weave-scope: k8s-setup-weave-scope-us-east-1 k8s-setup-weave-scope-eu-west-1

k8s-setup-weave-scope-us-east-1:
	kubectl --context us-east-1.k8s.local apply -f "https://cloud.weave.works/k8s/scope.yaml?k8s-version=$$(kubectl --context us-east-1.k8s.local version | base64 | tr -d '\n')"
	kubectl --context us-east-1.k8s.local -n weave patch svc/weave-scope-app --patch '{"spec": {"type": "LoadBalancer"}}'
	# Set up traffic control so that we can manage package loss or others
	kubectl --context us-east-1.k8s.local apply -f https://raw.githubusercontent.com/weaveworks-plugins/scope-traffic-control/master/deployments/k8s-traffic-control.yaml
	@echo " 🍭 Weave Scope us-east-1: http://$$(kubectl --context us-east-1.k8s.local get svc -n weave weave-scope-app -o jsonpath="{.status.loadBalancer.ingress[0].hostname}")"

k8s-setup-weave-scope-eu-west-1:
	kubectl --context eu-west-1.k8s.local apply -f "https://cloud.weave.works/k8s/scope.yaml?k8s-version=$$(kubectl --context eu-west-1.k8s.local version | base64 | tr -d '\n')"
	kubectl --context eu-west-1.k8s.local -n weave patch svc/weave-scope-app --patch '{"spec": {"type": "LoadBalancer"}}'
	# Set up traffic control so that we can manage package loss or others
	kubectl --context eu-west-1.k8s.local apply -f https://raw.githubusercontent.com/weaveworks-plugins/scope-traffic-control/master/deployments/k8s-traffic-control.yaml
	@echo " 🍭 Weave Scope eu-west-1: http://$$(kubectl --context eu-west-1.k8s.local get svc -n weave weave-scope-app -o jsonpath="{.status.loadBalancer.ingress[0].hostname}")"

k8s-wait-for-kafkas:
	# Not sure if it works, let's skip it now... 
	# [ $$(kubectl --context us-east-1.k8s.local -n kafka-source get statefulset kafka-source -o jsonpath='{.spec.replicas}') -eq $$(kubectl --context us-east-1.k8s.local -n kafka-source get pod | grep kafka-source | grep Running | wc -l) ]; \
  # if [ $$? -ne 0 ]; then echo "	>	kafka-source NOT READY YET"; 	sleep 30; make k8s-wait-for-kafkas; fi
	# [ $$(kubectl --context eu-west-1.k8s.local -n kafka-destination get statefulset kafka-destination -o jsonpath='{.spec.replicas}') -eq $$(kubectl --context eu-west-1.k8s.local -n kafka-destination get pod | grep kafka-destination | grep Running | wc -l) ]; \
  # if [ $$? -ne 0 ]; then echo "	>	kafka-destination NOT READY YET"; 	sleep 30; make k8s-wait-for-kafkas; fi

k8s-monitoring-graphite-exporter:
	kubectl apply -f k8s/monitoring/graphite-exporter --context us-east-1.k8s.local
	kubectl apply -f k8s/monitoring/graphite-exporter --context eu-west-1.k8s.local

k8s-grafana-configure:
	k8s/monitoring/patch/template.sh
	kubectl --context eu-west-1.k8s.local -n monitoring patch configmap grafana-datasources --record=true -p "$$(cat ./k8s/monitoring/patch/grafana-datasources.yaml)"
	kubectl --context eu-west-1.k8s.local -n monitoring patch configmap grafana-dashboard-definitions --record=true -p "$$(cat ./k8s/monitoring/patch/grafana-dashboard-definitions.yaml)"
	# Reload configuration by scaling down then up:
	kubectl --context eu-west-1.k8s.local -n monitoring scale replicasets $$(kubectl --context eu-west-1.k8s.local -n monitoring get replicasets | ag grafana | cut -d' ' -f1) --replicas 0
	sleep 10
	kubectl --context eu-west-1.k8s.local -n monitoring scale replicasets $$(kubectl --context eu-west-1.k8s.local -n monitoring get replicasets | ag grafana | cut -d' ' -f1) --replicas 1

k8s-help-monitoring:
	@echo
	@echo ">>> Admin for us-east-1:"
	@kubectl --context us-east-1.k8s.local -n kube-system describe secret $$(kubectl --context us-east-1.k8s.local -n kube-system get secret | grep eks-admin | awk '{print $$1}')
	@echo "	Now run: kubectl --context us-east-1.k8s.local proxy"
	@echo "	And then open http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/"
	@echo "	🍭 Weave Scope us-east-1: http://$$(kubectl --context us-east-1.k8s.local get svc -n weave weave-scope-app -o jsonpath="{.status.loadBalancer.ingress[0].hostname}")"
	@echo "	✅ Prometheus us-east-1: http://$$(kubectl --context us-east-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.status.loadBalancer.ingress[0].hostname}"):$$(kubectl --context us-east-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.spec.ports[0].port}")"
	@echo
	@echo "Monitor low level cluster events:"
	@echo "	kubectl --context us-east-1.k8s.local get events --watch --all-namespaces -o wide"

	@echo
	@echo
	@echo ">>> Admin for eu-west-1:"
	@kubectl --context eu-west-1.k8s.local -n kube-system describe secret $$(kubectl --context eu-west-1.k8s.local -n kube-system get secret | grep eks-admin | awk '{print $$1}')
	@echo "	Now run: kubectl --context eu-west-1.k8s.local proxy --port 8002"
	@echo "	And then open http://127.0.0.1:8002/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/"
	@echo
	@echo "Monitor low level cluster events:"
	@echo "	kubectl --context eu-west-1.k8s.local get events --watch --all-namespaces -o wide"
	@echo
	@echo "	🍭 Weave Scope eu-west-1: http://$$(kubectl --context eu-west-1.k8s.local get svc -n weave weave-scope-app -o jsonpath="{.status.loadBalancer.ingress[0].hostname}")"
	@echo "	✅ Prometheus eu-west-1: http://$$(kubectl --context eu-west-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.status.loadBalancer.ingress[0].hostname}"):$$(kubectl --context eu-west-1.k8s.local get svc --namespace monitoring prometheus-k8s -o jsonpath="{.spec.ports[0].port}")"
	@echo "	✅ Grafana (user/pass: admin/admin): http://$$(kubectl --context eu-west-1.k8s.local get svc --namespace monitoring grafana -o jsonpath="{.status.loadBalancer.ingress[0].hostname}"):$$(kubectl --context eu-west-1.k8s.local get svc --namespace monitoring grafana -o jsonpath="{.spec.ports[0].port}")"

k8s-resize-cluster-us-east-1:
	kops edit ig --name us-east-1.k8s.local --state s3://us-east-1.k8s.local nodes && kops update cluster --name us-east-1.k8s.local --state s3://us-east-1.k8s.local --yes

k8s-resize-cluster-eu-west-1:
	kops edit ig --name eu-west-1.k8s.local --state s3://eu-west-1.k8s.local nodes && kops update cluster --name eu-west-1.k8s.local --state s3://eu-west-1.k8s.local --yes

k8s-run-tests:
	kubectl apply -f k8s/tester/producer.yaml --context us-east-1.k8s.local
	kubectl apply -f k8s/tester/consumer.yaml --context eu-west-1.k8s.local
	# For logs run:
	# stern -l app=kafka-mirror-tester-producer --context us-east-1.k8s.local
	# stern -l app=kafka-mirror-tester-consumer --context eu-west-1.k8s.local

k8s-kafka-shell-destination:
	# And now you can:
	# unset JMX_PORT
	# ./bin/kafka-topics.sh --zookeeper  zookeeper:2181 --list
	# ./bin/kafka-topics.sh --zookeeper  zookeeper:2181 --describe
	# ./bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic test_topic_1
	# bin/kafka-topics.sh --zookeeper  zookeeper:2181 --alter --topic topic5 --partitions 300
	kubectl --context eu-west-1.k8s.local -n kafka-destination exec  kafka-destination-0 -it /bin/bash

k8s-kafka-shell-source:
	# And now you can:
	# unset JMX_PORT
	# bin/kafka-topics.sh --zookeeper  zookeeper:2181 --list
	# bin/kafka-topics.sh --zookeeper  zookeeper:2181 --describe
	# bin/kafka-console-producer.sh --broker-list localhost:9092 --topic test_topic_1
	#
	# Purge messages in topic:
	# bin/kafka-configs.sh --zookeeper zookeeper:2181 --entity-type topics --alter --add-config retention.ms=1000 --entity-name topic1
	# bin/kafka-topics.sh --zookeeper  zookeeper:2181 --alter --topic topic5 --partitions 300
	kubectl --context us-east-1.k8s.local -n kafka-source exec  kafka-source-0 -it /bin/bash

k8s-redeploy-tests: k8s-delete-tests k8s-run-tests

k8s-delete-all-apps: k8s-delete-tests k8s-delete-replicator k8s-delete-kafkas

k8s-delete-replicator:
ifeq ($(USE_UREPLICATOR),true)
	kubectl delete -f k8s/ureplicator --context eu-west-1.k8s.local || echo already deleted?
endif
ifeq ($(USE_BROOKLIN),true)
	kubectl delete -f k8s/brooklin --context eu-west-1.k8s.local || echo already deleted?
endif

k8s-redeploy-replicator: k8s-delete-replicator k8s-replicator-setup

k8s-delete-kafkas:
	kubectl delete -f k8s/kafka-source --context us-east-1.k8s.local || echo already deleted?
	kubectl delete -f k8s/kafka-destination --context eu-west-1.k8s.local || echo already deleted?

k8s-delete-tests:
	kubectl delete -f k8s/tester/producer.yaml --context us-east-1.k8s.local || echo already deleted?
	kubectl delete -f k8s/tester/consumer.yaml --context eu-west-1.k8s.local || echo already deleted?


k8s-create-cluster-us-east-1:
	aws s3api create-bucket  --bucket us-east-1.k8s.local  --region us-east-1 || echo Bucket already exists?
	kops create cluster --zones us-east-1a,us-east-1b,us-east-1c --node-count $(US_EAST_NODE_COUNT) --node-size i3.large --master-size m4.large --master-zones us-east-1a --networking calico --cloud aws --cloud-labels "Owner=rantav" --state s3://us-east-1.k8s.local  us-east-1.k8s.local --yes || echo Aready exists?
k8s-delete-cluster-us-east-1:
	kops delete cluster --state s3://us-east-1.k8s.local  us-east-1.k8s.local --yes
k8s-wait-for-cluster-us-east-1:
	kops validate cluster --name us-east-1.k8s.local --state s3://us-east-1.k8s.local; if [ $$? -ne 0 ]; then echo "\n\n	>>>>>	NOT READY YET	\n\n"; 	sleep 10; make k8s-wait-for-cluster-us-east-1; fi

k8s-create-cluster-eu-west-1:
	aws s3api create-bucket  --bucket eu-west-1.k8s.local --region eu-west-1 --create-bucket-configuration LocationConstraint=eu-west-1 || echo Bucket already exists?
	kops create cluster --zones eu-west-1a,eu-west-1b,eu-west-1c --node-count $(EU_WEST_NODE_COUNT) --node-size i3.large --master-size m4.large --master-zones eu-west-1c --networking calico --cloud aws --cloud-labels "Owner=rantav" --state s3://eu-west-1.k8s.local  eu-west-1.k8s.local --yes || echo Aready exists?
k8s-delete-cluster-eu-west-1:
	kops delete cluster --state s3://eu-west-1.k8s.local  eu-west-1.k8s.local --yes
k8s-wait-for-cluster-eu-west-1:
	kops validate cluster --name eu-west-1.k8s.local --state s3://eu-west-1.k8s.local; if [ $$? -ne 0 ]; then echo "\n\n	>>>>>	NOT READY YET	\n\n"; 	sleep 10; make k8s-wait-for-cluster-eu-west-1; fi

k8s-delete-all:
	make k8s-delete-cluster-eu-west-1& make k8s-delete-cluster-us-east-1
