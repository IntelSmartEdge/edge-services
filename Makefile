# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2019-2020 Intel Corporation

export GO111MODULE = on

.PHONY: \
	clean \
	networkedge \
	lint test help build
COPY_DOCKERFILES := $(shell /usr/bin/cp -rfT ./build/ ./dist/)
VER ?= 1.0

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo ""
	@echo "Build all the required Docker images for OpenNESS' deployment mode:"
	@echo "  networkedge            to build components of Network Edge deployment (Edge DNS Service, Certificate Signer)"
	@echo ""
	@echo "Helper targets:"
	@echo "  clean                  to clean build artifacts"
	@echo "  lint                   to run linter on Go code"
	@echo "  test                   to run tests on Go code"
	@echo "  test-cov               to run coverage tests on Go code"
	@echo "  help                   to show this message"
	@echo "  build                  to build all executables without images"
	@echo ""
	@echo "Single targets:"
	@echo "  certsigner             to build only docker image of the Certificate Signer"

networkedge: certsigner certrequester

clean:
	rm -rf ./dist

test:
	http_proxy= https_proxy= HTTP_PROXY= HTTPS_PROXY= ginkgo -v -r  -gcflags=-l --randomizeSuites --failOnPending --skipPackage=vendor,edgednscli

test-cov:
	rm -rf coverage.out*
	http_proxy= https_proxy= HTTP_PROXY= HTTPS_PROXY= ginkgo -v -r --randomizeSuites --failOnPending --skipPackage=vendor,edgednscli \
	 -gcflags=-l -cover -coverprofile=coverage.out -outputdir=.
	sed '1!{/^mode/d;}' coverage.out > coverage.out.fix
	go tool cover -html=coverage.out.fix

certsigner:
	CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -o ./dist/$@/$@ ./cmd/$@
ifndef SKIP_DOCKER_IMAGES
	VER=${VER} docker-compose build $@
endif

certrequester:
	CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -o ./dist/$@/$@ ./cmd/$@
ifndef SKIP_DOCKER_IMAGES
	VER=${VER} docker-compose build $@
endif

build:
	$(MAKE) SKIP_DOCKER_IMAGES=1 networkedge
