GO ?= go
SHELL := bash
SERVER=trpc-otlp-oteam-demo
IMAGE_URL= ?= mirrors.tencent.com/crayon/trpc-otlp-oteam-open-demo
IMAGE_VERSION ?= latest
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

token ?= "fixme"
otlp_endpoint ?= "host.docker.internal:4317"
otlp_http_endpoint ?= "http://host.docker.internal:4318"
local_ip ?= "127.255.255.255"
container_name ?= "trpc-otlp-oteam-open-demo"

.PHONY: help
help:
	@echo "Make Targets: "
	@echo " setup: tRPC setup"
	@echo " pb: Generate protobuf code"
	@echo " pb-clean: Clean protobuf code"
	@echo " mod: Download and tidy dependencies"
	@echo " lint: Lint Go code"
	@echo " dev: Run and test program locally"
	@echo " docker-image: Build image"
	@echo " docker-run: Run image"
	@echo " docker-push: Push image to remote"


.PHONY: setup
setup:
	go install trpc.group/trpc-go/trpc-cmdline/trpc@latest && trpc setup

.PHONY: pb
pb:
	trpc create -p greeter/greeter.proto -o greeter --rpconly --nogomod --mock=false
	cd $(ROOT_DIR)/greeter && \
	go mod init "bk-apm/bkmonitor-ecosystem/examples/go-examples/trpc-otlp-oteam-open/greeter" && \
	go mod tidy
	cd $(ROOT_DIR)

.PHONY: pb-clean
pb-clean:
	rm -f $(ROOT_DIR)/greeter/*.{go,mod,sum};

.PHONY: mod
mod:
	$(GO) mod download
	$(GO) mod tidy

.PHONY: lint
lint: mod
	diff -u <(echo -n) <(gofumpt -w .)

.PHONY: dev
dev: mod
	$(GO) build -o ./dist/$(SERVER)
	token=$(token) otlp_endpoint=$(otlp_endpoint) otlp_http_endpoint=$(otlp_http_endpoint) local_ip=$(local_ip) container_name=$(container_name) ./dist/$(SERVER) -conf=./conf/trpc_go.yaml

.PHONY: docker-image
docker-image:
	docker build -t $(SERVER):$(IMAGE_VERSION) .

.PHONY: docker-run
docker-run:
	docker run -e token=$(token) -e otlp_endpoint=$(otlp_endpoint) otlp_http_endpoint=$(otlp_http_endpoint) $(SERVER):$(IMAGE_VERSION)

.PHONY: docker-push
docker-push:
	docker tag $(SERVER):$(IMAGE_VERSION) $(IMAGE_URL):$(IMAGE_VERSION)
	docker push $(IMAGE_URL):$(IMAGE_VERSION)
