.PHONY: build test lint fmt clean validate-plugin prepare-templates help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VCS_REF ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.vcsRef=$(VCS_REF)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

prepare-templates: ## Copy templates to cmd/goent for embedding
	@rm -rf cmd/goent/templates
	@cp -r templates cmd/goent/templates
	@echo "Templates prepared for embedding"

build: prepare-templates ## Build CLI binary to dist/goent
	@mkdir -p dist
	@go build $(LDFLAGS) -o dist/goent ./cmd/goent
	@echo "Build complete: dist/goent"

test: ## Run tests with race detector and coverage
	@go test -race -cover ./...

lint: ## Run golangci-lint
	@golangci-lint run ./...

fmt: ## Format code with goimports
	@goimports -w .

clean: ## Remove dist/ and build artifacts
	@rm -rf dist/
	@rm -rf cmd/goent/templates
	@rm -f goent
	@echo "Clean complete"

validate-plugin: ## Validate plugin.json structure
	@echo "Validating plugin JSON files..."
	@for file in plugins/*/.claude-plugin/plugin.json; do \
		echo "Checking $$file..."; \
		jq empty "$$file" || exit 1; \
	done
	@echo "All plugin JSON files are valid"
