#!/usr/bin/make -f
SHELL = bash

REPO ?= $(shell go list -m)
VERSION ?= $(shell set -o pipefail; git describe --tags --dirty --match "v*" --always --abbrev=8 2>/dev/null | sed 's/^v//' || cat VERSION 2>/dev/null || echo "develop")
HUB_TAG ?= ${VERSION}

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

.golangci.yml:
	wget -O $@ https://github.com/nspcc-dev/.github/raw/master/.golangci.yml

# Run linters
lint: .golangci.yml
	@golangci-lint --timeout=5m run

# Run tests
test:
	@go test ./... -cover

# Run tests with race detection and produce coverage output
cover:
	@go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html
