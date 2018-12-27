LOCAL_IP := `ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -1`

setup:
	@echo For mac:  brew install librdkafka
	@echo For linux install librdkafka-dev

dep-ensure:
	dep ensure

build: dep-ensure generate test
	go build ./...

run-producer:
	go run main.go produce --bootstrap-servers localhost:9093 --id $$(hostname) --message-size 100 --throughput 10 --topics topic1,topic2 --use-message-headers

run-consumer:
	# Check out http://localhost:8000/debug/metrics
	go run main.go consume --bootstrap-servers localhost:9093 --consumer-group group-4 --topics topic1,topic2 --use-message-headers

test:
	go test ./...

generate:
	go generate ./...

docker-build: dep-ensure test
	docker build . -t rantav/kafka-mirror-tester:latest

docker-push: docker-build
	# push to dockerhub
	docker push rantav/kafka-mirror-tester

docker-run-consumer:
	# Check out http://localhost:8000/debug/metrics
	docker run -p 8000:8000 rantav/kafka-mirror-tester consume --bootstrap-servers $(LOCAL_IP):9093 --consumer-group group-4 --topics topic1,topic2

docker-run-producer:
	docker run rantav/kafka-mirror-tester produce --bootstrap-servers $(LOCAL_IP):9093 --id $$(hostname) --message-size 100 --throughput 10 --topics topic1,topic2

release: docker-push

k8s-all: k8s-create-clusters k8s-kafkas-setup k8s-replicator-setup
	sleep 60 # Wait for all clusters to be set up
	make k8s-run-tests

k8s-create-clusters: k8s-create-cluster-us-east-1 k8s-create-cluster-eu-west-1 k8s-wait-for-cluster-us-east-1 k8s-wait-for-cluster-eu-west-1 k8s-allow-kubectl-node-access

k8s-allow-kubectl-node-access:
	kubectl create clusterrolebinding --clusterrole=system:controller:node-controller --serviceaccount=kafka-source:default kubectl-node-access  --context us-east-1.k8s.local
	kubectl create clusterrolebinding --clusterrole=system:controller:node-controller --serviceaccount=kafka-source:default kubectl-node-access  --context eu-west-1.k8s.local

k8s-kafkas-setup:
	kubectl apply -f k8s/kafka-source/ --context us-east-1.k8s.local
	# Punch a hole in the security group so that we can access Kafka from the outside
	aws ec2 authorize-security-group-ingress --group-id $$(aws ec2 describe-security-groups --filters Name=group-name,Values=nodes.us-east-1.k8s.local --region us-east-1 --output text --query 'SecurityGroups[0].GroupId') --protocol tcp --port 9093 --cidr 0.0.0.0/0 --region us-east-1
	# validate
	k8s/kafka-source/test.sh

	kubectl apply -f k8s/kafka-destination --context eu-west-1.k8s.local

k8s-replicator-setup:
	kubectl apply -f k8s/ureplicator --context eu-west-1.k8s.local

k8s-run-tests:
	kubectl apply -f k8s/tester/producer.yaml --context us-east-1.k8s.local
	kubectl apply -f k8s/tester/consumer.yaml --context eu-west-1.k8s.local
	# For logs run:
	# kubectl -n kafka-source logs kafka-mirror-tester-producer --follow --context us-east-1.k8s.local
	# kubectl -n kafka-destination logs kafka-mirror-tester-consumer --follow --context eu-west-1.k8s.local

k8s-delete-all: k8s-delete-tests k8s-delete-replicator k8s-delete-kafkas

k8s-delete-replicator:
	kubectl delete -f k8s/ureplicator --context eu-west-1.k8s.local

k8s-delete-kafkas:
	kubectl delete -f k8s/kafka-source --context us-east-1.k8s.local
	kubectl delete -f k8s/kafka-destination --context eu-west-1.k8s.local

k8s-delete-tests:
	kubectl delete -f k8s/tester/producer.yaml --context us-east-1.k8s.local
	kubectl delete -f k8s/tester/consumer.yaml --context eu-west-1.k8s.local


k8s-create-cluster-us-east-1:
	#aws s3api create-bucket  --bucket us-east-1.k8s.local  --region us-east-1
	kops create cluster --zones us-east-1a,us-east-1b,us-east-1c,us-east-1d,us-east-1e,us-east-1f --node-count 3 --node-size m4.large --master-size t2.small --master-zones us-east-1a --networking calico --cloud aws --cloud-labels "Owner=rantav" --state s3://us-east-1.k8s.local  us-east-1.k8s.local --yes
k8s-delete-cluster-us-east-1:
	kops delete cluster --state s3://us-east-1.k8s.local  us-east-1.k8s.local --yes
k8s-wait-for-cluster-us-east-1:
	kops validate cluster --name us-east-1.k8s.local --state s3://us-east-1.k8s.local; if [ $$? -ne 0 ]; then echo "\n\n	>>>>>	NOT READY YET	\n\n"; 	sleep 10; make k8s-wait-for-cluster-us-east-1; fi

k8s-create-cluster-eu-west-1:
	#aws s3api create-bucket  --bucket eu-west-1.k8s.local --region eu-west-1 --create-bucket-configuration LocationConstraint=eu-west-1
	kops create cluster --zones eu-west-1a,eu-west-1b,eu-west-1c --node-count 3 --node-size m4.large --master-size t2.small --master-zones eu-west-1c --networking calico --cloud aws --cloud-labels "Owner=rantav" --state s3://eu-west-1.k8s.local  eu-west-1.k8s.local --yes
k8s-delete-cluster-eu-west-1:
	kops delete cluster --state s3://eu-west-1.k8s.local  eu-west-1.k8s.local --yes
k8s-wait-for-cluster-eu-west-1:
	kops validate cluster --name eu-west-1.k8s.local --state s3://eu-west-1.k8s.local; if [ $$? -ne 0 ]; then echo "\n\n	>>>>>	NOT READY YET	\n\n"; 	sleep 10; make k8s-wait-for-cluster-eu-west-1; fi


