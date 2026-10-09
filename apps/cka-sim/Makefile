# Shortcuts for developing cka-sim from source. `make` alone lists them.

WEB := ../web

.PHONY: help doctor up down shell web dev test

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-7s %s\n", $$1, $$2}'

doctor: ## check this machine can run cka-sim
	go run ./cmd doctor

up: ## create the cluster (does nothing if it is already up)
	go run ./cmd up

down: ## delete the cluster
	go run ./cmd down

shell: ## open a shell on the control plane
	go run ./cmd shell

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

dev: up web ## cluster up, then the page

test: ## vet and test the Go code
	go vet ./... && go test ./...

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
