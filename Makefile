#!/usr/bin/make -f
SHELL = bash

REPO ?= $(shell go list -m)
VERSION ?= $(shell git describe --tags --match "v*" --always --abbrev=8 2>/dev/null || cat VERSION 2>/dev/null || echo "develop")
HUB_TAG ?= "$(shell echo ${VERSION} | sed 's/^v//')"

REPO = nspccdev
APP = neofs-net-monitor
BINARY = ./bin/${APP}
SRC = ./cmd/${APP}/

LOCODE_DIR = ./locode
LOCODE_FILE = locode_db.gz
LOCODE_DB_URL = https://github.com/nspcc-dev/neofs-locode-db/releases/download/v0.2.1/locode_db.gz

.PHONY: bin image up up-testnet up-devenv down down-testnet down-devenv clean locode

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

locode:
	@mkdir -p ${LOCODE_DIR}
	@echo "⇒ Download NeoFS LOCODE database from ${LOCODE_DB_URL}"
	@curl \
    		-sSL "${LOCODE_DB_URL}" \
    		-o ${LOCODE_DIR}/${LOCODE_FILE}
	gzip -dfk ${LOCODE_DIR}/${LOCODE_FILE}

up: locode
	@docker-compose -f docker/docker-compose.yml up -d

up-testnet: locode
	@docker-compose -f docker/docker-compose.testnet.yml up -d

up-devenv: locode
	@docker-compose -f docker/docker-compose.devenv.yml up -d

down:
	@docker-compose -f docker/docker-compose.yml down

down-testnet:
	@docker-compose -f docker/docker-compose.testnet.yml down

down-devenv:
	@docker-compose -f docker/docker-compose.devenv.yml down

clean:
	rm -f ${BINARY}
	rm -rf ${LOCODE_DIR}

version:
	@echo ${VERSION}
