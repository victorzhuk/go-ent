.PHONY: help build test lint fmt clean validate-plugin skill-validate skill-sync skill-quality

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VCS_REF ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_PKG := github.com/victorzhuk/go-ent/internal/version
LDFLAGS := -ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).vcsRef=$(VCS_REF)"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build binary to bin/ent
	@go build $(LDFLAGS) -o bin/ent ./cmd/go-ent
	@echo "Build complete: bin/ent"

test: ## Run all tests with race detection and coverage
	@go test -race -cover ./...

lint: ## Run golangci-lint
	@golangci-lint run ./...

fmt: ## Format code with goimports
	@goimports -w .

clean: ## Remove build artifacts
	@rm -rf bin/
	@rm -f ent
	@echo "Clean complete"

validate-plugin: ## Validate plugin JSON files
	@echo "Validating plugin JSON files..."
	@for file in plugins/*/.claude-plugin/plugin.json; do \
		echo "Checking $$file..."; \
		jq empty "$$file" || exit 1; \
	done
	@echo "All plugin JSON files are valid"

skill-validate: ## Validate all skills with strict mode
	@echo "Validating skills..."
	@go run ./cmd/cli validate skills --strict

skill-sync: ## Sync skills from plugins to .claude directory
	@echo "Syncing skills..."
	@go run ./cmd/cli sync skills

skill-quality: ## Generate quality report for all skills
	@echo "Getting skill quality report..."
	@go run ./cmd/cli quality skills
