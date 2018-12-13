dep-ensure:
	dep ensure

build: dep-ensure
	go build ./...

run: dep-ensure
	go run main.go

test: dep-ensure
	go test ./...
