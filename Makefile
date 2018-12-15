setup:
	@echo For mac:  brew install librdkafka
	@echo For linux install librdkafka-dev

dep-ensure:
	dep ensure

generate:
	go generate ./...

build: dep-ensure generate
	go build ./...

run: dep-ensure generate
	go run main.go

test: dep-ensure generate
	go test ./...
