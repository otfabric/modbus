# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: help all build build-cmd build-examples lint vet test check clean
help: ## This help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

BIN_DIR := bin # Output directory for generated binaries
# Use all packages except /examples (multiple main programs in one folder)
PKGS := $(shell go list ./... | grep -v '/examples$$' | sed 's,^github.com/otfabric/modbus,.,')

all: build ## Default target: build cmd + examples apps

build: build-cmd build-examples ## Build all app entrypoints

build-cmd: ## Build binaries from cmd/*.go
	@mkdir -p $(BIN_DIR)
	@for src in cmd/*.go; do \
		name="$$(basename "$$src" .go)"; \
		go build -o "$(BIN_DIR)/$$name" "$$src"; \
	done

build-examples: ## Build binaries from examples/*.go
	@mkdir -p $(BIN_DIR)
	@for src in examples/**/*.go; do \
		name="$$(basename "$$src" .go)"; \
		go build -o "$(BIN_DIR)/example-$$name" "$$src"; \
	done

lint: ## Run linter (golangci-lint preferred, fallback golint)
	@echo "Running linter on packages: $(PKGS)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(PKGS); \
	elif command -v golint >/dev/null 2>&1; then \
		golint $(PKGS); \
	else \
		echo "No Go linter found. Install golangci-lint or golint."; \
		exit 1; \
	fi

vet: ## Run go vet on project packages
	@echo "Running go vet on packages: $(PKGS)"
	@go vet $(PKGS)

test: ## Run all tests on project packages
	@echo "Running tests on packages: $(PKGS)"
	@go test $(PKGS)

check: lint vet test ## Run lint + vet + test

clean: ## Remove generated binaries
	@rm -rf $(BIN_DIR)
