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

k8s-all: k8s-create-clusters k8s-kafkas-setup k8s-ureplicator-setup
	sleep 60 # Wait for all clusters to be set up
	make k8s-run-tests

k8s-create-clusters:
	# TODO: set up a cluster. And then another cluster

k8s-kafkas-setup:
	kubectl apply -f k8s/kafka-source
	kubectl apply -f k8s/kafka-destination

k8s-ureplicator-setup:
	kubectl apply -f k8s/ureplicator

k8s-run-tests:
	kubectl apply -f k8s/tester
