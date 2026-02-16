.PHONY: acceptance
acceptance:
	./scripts/acceptance.sh

.PHONY: qa
qa:
	./scripts/caiged "$(PWD)" --task qa
