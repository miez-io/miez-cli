.PHONY: build test test-race vet lint install clean release-check

BINARY ?= bin/miez
VERSION ?= 0.1.0
LDFLAGS ?= -s -w -X github.com/manuel/miez-cli/internal/cli.Version=$(VERSION)

build:
	mkdir -p "$(dir $(BINARY))"
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BINARY)" ./cmd/miez

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

install:
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/miez

clean:
	rm -f "$(BINARY)"

release-check: test test-race vet build
