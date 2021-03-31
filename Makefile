#!/usr/bin/make -f
SHELL = bash

REPO ?= $(shell go list -m)
VERSION ?= $(shell git describe --tags --always)
HUB_TAG ?= "$(shell echo ${VERSION} | sed 's/^v//')"

REPO = nspccdev
APP = neofs-net-monitor
BINARY = ./bin/${APP}
SRC = ./cmd/${APP}/

.PHONY: bin image up up-devenv down down-devenv clean

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
	$(shell docker-compose -f docker/docker-compose.yml up -d)

up-devenv:
	$(shell docker-compose -f docker/docker-compose.devenv.yml up -d)

down:
	$(shell docker-compose -f docker/docker-compose.yml down)

down-devenv:
	$(shell docker-compose -f docker/docker-compose.devenv.yml down)

clean:
	rm ${BINARY}