# ==============================================================================
# protoc-gen-jsonschema — Makefile
#
# Common targets:
#   make help          List all targets
#   make fmt           Format Go + non-Go (oxfmt) files
#   make fmt-check     Verify formatting without writing (CI gate)
#   make lint          Run golangci-lint
#   make test          Run unit tests
#   make build         Build the protoc-gen-jsonschema plugin binary
#   make install       go install the plugin into $GOBIN
#   make buf-lint      Lint proto files with buf
#   make buf-format    Format proto files with buf
#   make buf-generate  Run buf generate in the repo root
#   make buf-push      Publish the module to the Buf Schema Registry (BSR)
#   make pre-commit    Run the configured pre-commit hooks
# ==============================================================================

PROJECT_ROOT := $(abspath .)
MODULE       := github.com/sunerpy/protoc-gen-jsonschema
BINARY       := protoc-gen-jsonschema
DIST_DIR     := dist
CMD_PKG      := ./cmd/protoc-gen-jsonschema
VERSION_PKG  := main

# Version metadata injected at build time (falls back to git describe).
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

# Go files to format/lint (exclude vendored, generated, and the example module).
GO_FILES := $(shell find . -name '*.go' \
	-not -path './vendor/*' \
	-not -path './example/*' \
	-not -name '*.pb.go')

OXFMT_IGNORE := --ignore-path "$(PROJECT_ROOT)/.oxfmtignore"

# Tool availability guards.
HAS_OXFMT       := $(shell command -v oxfmt 2>/dev/null)
HAS_GOFUMPT     := $(shell command -v gofumpt 2>/dev/null)
HAS_GOIMPORTS   := $(shell command -v goimports 2>/dev/null)
HAS_GOLANGCI    := $(shell command -v golangci-lint 2>/dev/null)
HAS_BUF         := $(shell command -v buf 2>/dev/null)
HAS_PRECOMMIT   := $(shell command -v pre-commit 2>/dev/null)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------------------
# Formatting
# ------------------------------------------------------------------------------

.PHONY: fmt
fmt: fmt-go fmt-oxfmt ## Format Go and non-Go files
	@echo "Formatting complete."

.PHONY: fmt-go
fmt-go: ## Format Go code (goimports + gofumpt)
	@echo "Formatting Go code..."
ifndef HAS_GOIMPORTS
	@echo "goimports not found. Install: go install golang.org/x/tools/cmd/goimports@latest"
else
	@echo "$(GO_FILES)" | tr ' ' '\n' | xargs -r -P 4 goimports -w -local $(MODULE)
endif
ifndef HAS_GOFUMPT
	@echo "gofumpt not found. Install: go install mvdan.cc/gofumpt@latest"
else
	@echo "$(GO_FILES)" | tr ' ' '\n' | xargs -r -P 4 gofumpt -extra -w
endif

.PHONY: fmt-oxfmt
fmt-oxfmt: ## Format non-Go files (YAML/Markdown/JSON) with oxfmt
ifndef HAS_OXFMT
	@echo "oxfmt not found, skipping non-Go formatting. Install from https://github.com/oxc-project/oxfmt"
else
	@echo "Formatting non-Go files with oxfmt..."
	@oxfmt --write --no-error-on-unmatched-pattern $(OXFMT_IGNORE) "$(PROJECT_ROOT)"
endif

.PHONY: fmt-check
fmt-check: ## Verify formatting without writing (CI gate)
	@echo "Checking Go formatting..."
ifdef HAS_GOFUMPT
	@unformatted=$$(echo "$(GO_FILES)" | tr ' ' '\n' | xargs -r gofumpt -extra -l); \
	if [ -n "$$unformatted" ]; then \
		echo "The following Go files are not formatted:"; echo "$$unformatted"; \
		echo "Run 'make fmt' to fix."; exit 1; \
	fi
else
	@echo "gofumpt not found, skipping Go format check."
endif
ifdef HAS_OXFMT
	@echo "Checking non-Go formatting..."
	@oxfmt --check --no-error-on-unmatched-pattern $(OXFMT_IGNORE) "$(PROJECT_ROOT)"
else
	@echo "oxfmt not found, skipping non-Go format check."
endif

# ------------------------------------------------------------------------------
# Linting
# ------------------------------------------------------------------------------

.PHONY: lint
lint: ## Run golangci-lint
ifndef HAS_GOLANGCI
	@echo "golangci-lint not found. Install: https://golangci-lint.run/welcome/install/"
	@echo "Running 'go vet' instead..."
	@go vet ./...
else
	@echo "Running golangci-lint..."
	@golangci-lint run ./...
endif

# ------------------------------------------------------------------------------
# Testing & building
# ------------------------------------------------------------------------------

.PHONY: test
test: ## Run unit tests
	@echo "Running tests..."
	@go test ./... -count=1

.PHONY: test-race
test-race: ## Run unit tests with the race detector
	@CGO_ENABLED=1 go test ./... -count=1 -race

.PHONY: build
build: ## Build the plugin binary into $(DIST_DIR)
	@echo "Building $(BINARY) $(VERSION) ($(COMMIT))..."
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=0 go build -ldflags="-s -w -X $(VERSION_PKG).version=$(VERSION)" -o $(DIST_DIR)/$(BINARY) $(CMD_PKG)
	@echo "Built $(DIST_DIR)/$(BINARY)"

.PHONY: install
install: ## go install the plugin into $GOBIN
	@echo "Installing $(BINARY)..."
	@go install $(CMD_PKG)

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf $(DIST_DIR) $(BINARY)

# ------------------------------------------------------------------------------
# buf / BSR
# ------------------------------------------------------------------------------

.PHONY: buf-lint
buf-lint: ## Lint proto files with buf
ifndef HAS_BUF
	@echo "buf not found. Install: https://buf.build/docs/installation"; exit 1
else
	@echo "Linting proto files with buf..."
	@buf lint
endif

.PHONY: buf-format
buf-format: ## Format proto files with buf
ifndef HAS_BUF
	@echo "buf not found. Install: https://buf.build/docs/installation"; exit 1
else
	@echo "Formatting proto files with buf..."
	@buf format -w
endif

.PHONY: buf-generate
buf-generate: ## Run buf generate in the repo root
ifndef HAS_BUF
	@echo "buf not found. Install: https://buf.build/docs/installation"; exit 1
else
	@echo "Running buf generate..."
	@buf generate
endif

.PHONY: buf-push
buf-push: buf-lint ## Publish the module to the Buf Schema Registry (BSR)
ifndef HAS_BUF
	@echo "buf not found. Install: https://buf.build/docs/installation"; exit 1
else
	@echo "Pushing module to BSR with label $(VERSION)..."
	@buf push --label "$(VERSION)"
endif

# ------------------------------------------------------------------------------
# Pre-commit
# ------------------------------------------------------------------------------

.PHONY: pre-commit
pre-commit: ## Run configured pre-commit hooks against all files
ifndef HAS_PRECOMMIT
	@echo "pre-commit not found. Install: https://pre-commit.com/#install"; exit 1
else
	@pre-commit run --all-files
endif

.PHONY: hooks
hooks: ## Install the git pre-commit hooks
ifndef HAS_PRECOMMIT
	@echo "pre-commit not found. Install: https://pre-commit.com/#install"; exit 1
else
	@pre-commit install
	@echo "pre-commit hooks installed."
endif

# ------------------------------------------------------------------------------
# Aggregate gates
# ------------------------------------------------------------------------------

.PHONY: check
check: fmt-check lint buf-lint test ## Run the full local verification suite
	@echo "All checks passed."
