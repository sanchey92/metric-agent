GOLANGCI_LINT_VERSION := v2.1.5
GOLANGCI_LINT_BIN := ./bin/golangci-lint
GOFUMPT_BIN := $(shell go env GOPATH)/bin/gofumpt
GO_BIN := $(shell go env GOPATH)/bin
PATH := $(GO_BIN):$(PATH)

install-lint:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@mkdir -p ./bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ./bin $(GOLANGCI_LINT_VERSION)
	@$(GOLANGCI_LINT_BIN) --version

install-gofumpt:
	@echo "Installing gofumpt..."
	@go install mvdan.cc/gofumpt@latest

lint:
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT_BIN) run ./... --config=.golangci.yml

fmt:
	@echo "Formatting code with gofumpt..."
	@$(GOFUMPT_BIN) -l -w .

check: install-gofumpt fmt lint

clean:
	@echo "Cleaning up..."
	@rm -rf ./bin

.PHONY: install-lint install-gofumpt lint fmt check clean