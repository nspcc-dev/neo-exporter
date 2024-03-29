#!/usr/bin/make -f
SHELL = bash

REPO ?= $(shell go list -m)
VERSION ?= $(shell git describe --tags --dirty --match "v*" --always --abbrev=8 2>/dev/null || cat VERSION 2>/dev/null || echo "develop")
HUB_TAG ?= "$(shell echo ${VERSION} | sed 's/^v//')"

REPO = nspccdev
APP = neo-exporter
BINARY = ./bin/${APP}
SRC = ./cmd/${APP}/

.PHONY: bin image up up-testnet up-devenv down down-testnet down-devenv clean lint test cover

bin:
	@echo "Build neo-exporter binary"
	CGO_ENABLED=0 \
	go build -v -trimpath \
	-ldflags "-X main.Version=$(VERSION)" \
	-o ${BINARY} ${SRC}

image:
	@echo "Build neo-exporter docker image"
	@docker build \
		--rm \
		-f Dockerfile \
		--build-arg VERSION=$(VERSION) \
		-t ${REPO}/${APP}:$(HUB_TAG) .

fmt:
	@gofmt -l -w -s $$(find . -type f -name '*.go'| grep -v "/vendor/")

clean:
	rm -f ${BINARY}

version:
	@echo ${VERSION}

# Run linters
lint:
	@golangci-lint --timeout=5m run

# Run tests
test:
	@go test ./... -cover

# Run tests with race detection and produce coverage output
cover:
	@go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html
