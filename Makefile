setup:
	@echo For mac:  brew install librdkafka
	@echo For linux install librdkafka-dev

dep-ensure:
	dep ensure

build: dep-ensure
	go build ./...

run-producer: dep-ensure
	go run main.go
	go run main.go produce --bootstrap-server localhost:9093 --id $(hostname) --message-size 100 --throughput 10 --topics topic

test: dep-ensure
	go test ./...
