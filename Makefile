#!/usr/bin/make -f
SHELL = bash

REPO ?= $(shell go list -m)
VERSION ?= $(shell git describe --tags --always)
HUB_TAG ?= "$(shell echo ${VERSION} | sed 's/^v//')"

REPO = nspccdev
APP = neofs-net-monitor
BINARY = ./bin/${APP}
SRC = ./cmd/${APP}/

.PHONY: bin image up up-testnet up-devenv down down-testnet down-devenv clean

bin:
	@echo "Build neofs-net-monitor binary"
	CGO_ENABLED=0 \
	go build -v -trimpath \
	-ldflags "-X main.Version=$(VERSION)" \
	-o ${BINARY} ${SRC}

image:
	@echo "Build neofs-net-monitor docker image"
	@docker build \
		--rm \
		-f Dockerfile \
		--build-arg VERSION=$(VERSION) \
		-t ${REPO}/${APP}:$(HUB_TAG) .

up:
	@docker-compose -f docker/docker-compose.yml up -d

up-testnet:
	@docker-compose -f docker/docker-compose.testnet.yml up -d

up-devenv:
	@docker-compose -f docker/docker-compose.devenv.yml up -d

down:
	@docker-compose -f docker/docker-compose.yml down

down-testnet:
	@docker-compose -f docker/docker-compose.testnet.yml down

down-devenv:
	@docker-compose -f docker/docker-compose.devenv.yml down

clean:
	rm -f ${BINARY}
