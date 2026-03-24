GO ?= go
SHELL := bash
APP    := TestApp
SERVER := HelloGo
IMAGE_NAME ?= taf-demo
IMAGE_URL ?= mirrors.tencent.com/crayon/taf-demo
IMAGE_VERSION ?= latest

COLLECTOR_HOST ?= "bkm-collector.bkmonitor-operator"

.PHONY: help
help:
	@echo "Make Targets: "
	@echo " mod: Download and tidy dependencies"
	@echo " lint: Lint Go code"
	@echo " dev: Run and test program locally"
	@echo " docker-image: Build image"
	@echo " docker-run: Run image"
	@echo " docker-push: Push image to remote"

.PHONY: mod
mod:
	$(GO) mod download
	$(GO) mod tidy

.PHONY: lint
lint: mod
	diff -u <(echo -n) <(gofumpt -w .)

.PHONY: dev
dev: mod
	tars2go -outdir=. SayHello.tars
	$(GO) build -o ./dist/$(SERVER)
	./dist/$(SERVER) --config=config/config.conf

.PHONY: docker-image
docker-image:
	docker build --build-arg SERVER=$(SERVER) -t $(IMAGE_NAME):$(IMAGE_VERSION) .

.PHONY: docker-run
docker-run:
	docker run --init -e COLLECTOR_HOST=$(COLLECTOR_HOST) -e SERVER=$(SERVER) $(IMAGE_NAME):$(IMAGE_VERSION)

.PHONY: docker-push
docker-push:
	docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(IMAGE_URL):$(IMAGE_VERSION)
	docker push $(IMAGE_URL):$(IMAGE_VERSION)
