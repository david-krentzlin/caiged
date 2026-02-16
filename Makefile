CAIGED_BIN ?= ./caiged/caiged
INSTALL_DIR ?= $(HOME)/.local/bin
CAIGED_MAKEFILE ?= caiged/Makefile

.PHONY: caiged
caiged:
	$(MAKE) -f $(CAIGED_MAKEFILE) build

.PHONY: install
install: caiged
	$(MAKE) -f $(CAIGED_MAKEFILE) install INSTALL_DIR=$(INSTALL_DIR)

.PHONY: acceptance
acceptance:
	./scripts/acceptance.sh

.PHONY: qa
qa: caiged
	$(CAIGED_BIN) "$(PWD)" --spin qa
