CAIGED_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REPO_ROOT := $(shell cd $(CAIGED_DIR)/.. && pwd -P)
BUILD_BIN ?= $(CAIGED_DIR)/caiged
INSTALL_DIR ?= $(HOME)/.local/bin
INSTALL_BIN ?= $(INSTALL_DIR)/caiged

.PHONY: build
build:
	go -C $(CAIGED_DIR) build -ldflags "-X github.com/david-krentzlin/caiged/caiged/cmd.defaultRepoPath=$(REPO_ROOT)" -o $(BUILD_BIN) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_BIN) $(INSTALL_BIN)
