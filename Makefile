.PHONY: build test lint install clean

build:
	go build -o bin/kie-cli ./cmd/kie-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/kie-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/kie-mcp ./cmd/kie-mcp

install-mcp:
	go install ./cmd/kie-mcp

build-all: build build-mcp
