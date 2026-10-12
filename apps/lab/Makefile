# The lab runs from infra/local/compose.yml; the root Makefile's up, down and shell drive it.

.PHONY: help doctor

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-7s %s\n", $$1, $$2}'

doctor: ## check this machine can run a lab
	go run ./cmd/lab doctor
