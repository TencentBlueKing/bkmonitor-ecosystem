SHELL := bash
SERVER=profiling-js
IMAGE_URL ?= mirrors.tencent.com/crayon/profiling-js
IMAGE_VERSION ?= latest

.PHONY: help
help:
	@echo "Make Targets: "
	@echo " docker-image: Build image"
	@echo " docker-push: Push image to remote"

.PHONY: docker-image
docker-image:
	docker build -t $(SERVER):$(IMAGE_VERSION) .

.PHONY: docker-push
docker-push:
	docker tag $(SERVER):$(IMAGE_VERSION) $(IMAGE_URL):$(IMAGE_VERSION)
	docker push $(IMAGE_URL):$(IMAGE_VERSION)
