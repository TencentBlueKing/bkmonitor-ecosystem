GO ?= go
PYTHON ?= python
PIP ?= pip
SHELL := bash

.PHONY: help
help:
	@echo "Make Targets: "
	@echo "init: Download dependencies"
	@echo "render: Render the documents in the templates directory to docs"
	@echo "lint: Lint code"

.PHONY: mod
init:
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install -v github.com/incu6us/goimports-reviser/v3@latest
	uv sync
	uv run pre-commit install




.PHONY: lint
lint:
	diff -u <(echo -n) <(gofumpt -w "./examples/go-examples")

.PHONY: render
render:
	$(PYTHON) ./tools/formatter/main.py

	@if [ -f "tools/formatter/context/inner.py" ]; then \
  		$(PYTHON) ./tools/sync/inner/docs_mappings_validate.py; \
		mv -f docs/inner/inner/cookbook/Quickstarts/events/README.md docs/inner/cookbook/Quickstarts/events/README.md; \
		mv -f docs/inner/inner/cookbook/Quickstarts/metrics/README.md docs/inner/cookbook/Quickstarts/metrics/README.md; \
		mv -f docs/inner/inner/README.md docs/inner/README.md; \
		\
  		cp -f docs/inner/cpp/otlp/README.md examples/cpp-examples/helloworld/README.md; \
		cp -f docs/inner/js/otlp/README.md examples/js-examples/helloworld/README.md; \
		cp -f docs/inner/java/otlp/README.md examples/java-examples/helloworld/README.md; \
		cp -f docs/inner/java/otlp-spring-boot-starter/README.md examples/java-examples/spring-boot-starter/README.md; \
		cp -f docs/inner/java/skywalking-agent/README.md examples/java-examples/skywalking-agent/README.md; \
		cp -f docs/inner/go/otlp/README.md examples/go-examples/helloworld/README.md; \
		cp -f docs/inner/go/jaeger-client-demo/README.md examples/go-examples/jaeger-client-demo/README.md; \
		cp -f docs/inner/go/jaeger-ot-demo/README.md examples/go-examples/jaeger-ot-demo/README.md; \
		cp -f docs/inner/python/otlp/README.md examples/python-examples/helloworld/README.md; \
		cp -f docs/inner/python/otlp-automatic/README.md examples/python-examples/helloworld-automatic/README.md; \
		cp -f docs/inner/python/profiling/README.md examples/python-examples/profiling/README.md; \
		cp -f docs/inner/python/patch-profiling/README.md examples/python-examples/patch-profiling/README.md; \
		cp -f docs/inner/python/jaeger-ot-demo/README.md examples/python-examples/jaeger-ot-demo/README.md; \
		cp -f docs/inner/go/trpc-otlp-oteam-open/README.md examples/go-examples/trpc-otlp-oteam-open/README.md; \
		\
		cp -f docs/inner/cookbook/Quickstarts/events/http/curl.md examples/events/curl/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/events/http/java.md examples/events/java/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/events/http/go.md examples/events/go/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/events/http/python.md examples/events/python/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/events/http/cpp.md examples/events/cpp/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/http/curl.md examples/metrics/http/curl/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/http/java.md examples/metrics/http/java/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/http/python.md examples/metrics/http/python/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/http/go.md examples/metrics/http/go/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/http/cpp.md examples/metrics/http/cpp/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/sdks/java.md examples/metrics/sdks/java/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/sdks/python.md examples/metrics/sdks/python/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/sdks/go.md examples/metrics/sdks/go/README.md; \
		cp -f docs/inner/cookbook/Quickstarts/metrics/sdks/cpp.md examples/metrics/sdks/cpp/README.md; \
	else \
		echo "非 inner 环境，跳过复制命令..."; \
	fi

