.PHONY: build test cover lint install clean

build:
	go build -o bin/kl ./cmd/kl

test:
	go test -race ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

install:
	go install ./cmd/kl

clean:
	rm -rf bin coverage.out
