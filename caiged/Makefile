CAIGED_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REPO_ROOT := $(shell cd $(CAIGED_DIR)/.. && pwd -P)
BUILD_BIN ?= $(CAIGED_DIR)/caiged
INSTALL_DIR ?= $(HOME)/.local/bin
INSTALL_BIN ?= $(INSTALL_DIR)/caiged
MAN_DIR ?= $(HOME)/.local/share/man/man1
MAN_SOURCE_DIR := $(REPO_ROOT)/man/man1

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build              - Build the caiged binary"
	@echo "  make install            - Install caiged and man pages"
	@echo "  make uninstall          - Uninstall caiged and man pages"
	@echo "  make test               - Run unit tests"
	@echo "  make test-integration   - Run integration tests (requires Docker)"
	@echo "  make test-all           - Run all tests (unit + integration)"
	@echo "  make clean              - Clean build artifacts"

.DEFAULT_GOAL := help

.PHONY: build
build:
	go -C $(CAIGED_DIR) build -ldflags "-X github.com/david-krentzlin/caiged/caiged/cmd.defaultRepoPath=$(REPO_ROOT)" -o $(BUILD_BIN) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_BIN) $(INSTALL_BIN)
	@echo "Installing man pages..."
	mkdir -p $(MAN_DIR)
	cp $(MAN_SOURCE_DIR)/*.1 $(MAN_DIR)/
	@echo "Installed man pages to $(MAN_DIR)"
	@echo "Run 'man caiged' to view documentation"

.PHONY: uninstall
uninstall:
	rm -f $(INSTALL_BIN)
	rm -f $(MAN_DIR)/caiged.1
	rm -f $(MAN_DIR)/caiged-build.1
	rm -f $(MAN_DIR)/caiged-connect.1
	rm -f $(MAN_DIR)/caiged-port.1
	rm -f $(MAN_DIR)/caiged-session.1
	@echo "Uninstalled caiged and man pages"

.PHONY: test
test:
	go test ./internal/... -v

.PHONY: test-integration
test-integration:
	@echo "Running integration tests (requires Docker)..."
	go test -tags=integration ./integration/... -v -timeout 5m

.PHONY: test-all
test-all: test test-integration

.PHONY: clean
clean:
	rm -f $(BUILD_BIN)
	@echo "Cleaned build artifacts"
