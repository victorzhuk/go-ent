.PHONY: help build test lint fmt clean validate-plugin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VCS_REF ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_PKG := github.com/victorzhuk/go-ent/internal/version
LDFLAGS := -ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).vcsRef=$(VCS_REF)"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build:
	@go build $(LDFLAGS) -o bin/ent ./cmd/go-ent
	@echo "Build complete: bin/ent"

test:
	@go test -race -cover ./...

lint:
	@golangci-lint run ./...

fmt:
	@goimports -w .

clean:
	@rm -rf bin/
	@rm -f ent
	@echo "Clean complete"

validate-plugin:
	@echo "Validating plugin JSON files..."
	@for file in plugins/*/.claude-plugin/plugin.json; do \
		echo "Checking $$file..."; \
		jq empty "$$file" || exit 1; \
	done
	@echo "All plugin JSON files are valid"
