export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

KIND_CLUSTER_NAME   ?= kind
DOCKER_REPOSITORY   ?= onosproject/
ONOS_CONFIG_MODELS_VERSION ?= latest
GOLANG_BUILD_VERSION  := v0.6.3

linters: # @HELP examines Go source code and reports coding problems
	golangci-lint run --timeout 30m

license_check: # @HELP examine and ensure license headers exist
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR}

gofmt: # @HELP run the Go format validation
	bash -c "diff -u <(echo -n) <(gofmt -d pkg/)"

PHONY:build
build: # @HELP build all libraries
build: linters license_check gofmt

compile-plugins: # @HELP compile standard plugins
compile-plugins:
	go run github.com/onosproject/config-models/cmd/config-model compile-plugin --name test --version 1.0.0 --module test@2020-11-18=plugins/test/test@2020-11-18.yang --output plugins/test

images: # @HELP build Docker images
images:
	docker build . -f build/plugin-compiler/Dockerfile \
		--build-arg GOLANG_BUILD_VERSION=${GOLANG_BUILD_VERSION} \
		-t onosproject/config-plugin-compiler:${ONOS_CONFIG_MODELS_VERSION}

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image onosproject/config-plugin-compiler:${ONOS_CONFIG_MODELS_VERSION}

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor
	go clean -testcache github.com/onosproject/config-models/...

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
