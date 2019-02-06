LOCAL_IP := `ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -1`

###########################
# Build, test, run
###########################
go-setup:
	@echo For mac:  brew install librdkafka
	@echo For linux install librdkafka-dev

go-dep-ensure:
	dep ensure

go-build: go-dep-ensure go-generate go-test
	go build ./...

go-run-producer:
	# Check out http://localhost:8001/metrics
	go run main.go produce --bootstrap-servers localhost:9093 --id $$(hostname) --message-size 100 --throughput 10 --topics topic1,topic2 --use-message-headers

go-run-consumer:
	# Check out http://localhost:8000/metrics
	go run main.go consume --bootstrap-servers localhost:9093 --consumer-group group-4 --topics topic1,topic2 --use-message-headers

go-test:
	go test ./...

go-generate:
	go generate ./...

#########################
# Docker
#########################
go-docker-build: go-dep-ensure go-test
	docker build . -t rantav/kafka-mirror-tester:latest

go-docker-push: go-docker-build
	# push to dockerhub
	docker push rantav/kafka-mirror-tester

go-docker-run-consumer:
	# Check out http://localhost:8000/metrics
	docker run -p 8000:8000 rantav/kafka-mirror-tester consume --bootstrap-servers $(LOCAL_IP):9093 --consumer-group group-4 --topics topic1,topic2

go-docker-run-producer:
	# Check out http://localhost:8001/metrics
	docker run rantav/kafka-mirror-tester produce --bootstrap-servers $(LOCAL_IP):9093 --id $$(hostname) --message-size 100 --throughput 10 --topics topic1,topic2

go-release: go-docker-push
