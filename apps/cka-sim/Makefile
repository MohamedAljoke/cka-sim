# Shortcuts for developing cka-sim from source. `make` alone lists them.

WEB := ../web

.PHONY: help doctor up down shell serve web dev test

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-7s %s\n", $$1, $$2}'

doctor: ## check this machine can run cka-sim
	go run ./cmd/cli doctor

up: ## create the cluster (does nothing if it is already up)
	go run ./cmd/cli up

down: ## delete the cluster
	go run ./cmd/cli down

shell: ## open a shell on the control plane
	go run ./cmd/cli shell

serve: ## run the backend for the page on 127.0.0.1:7070
	go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# kill 0 stops the other server when one of them exits.
dev: up $(WEB)/node_modules ## cluster up, then the backend and the page together
	@trap 'kill 0' EXIT; go run ./cmd/server & cd $(WEB) && npm run dev

test: ## vet and test the Go code
	go vet ./... && go test ./...

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
