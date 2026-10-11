# Shortcuts for developing cka-sim from source. `make` alone lists them.

PLATFORM := apps/platform
LAB := apps/lab
WEB := apps/web
# make up creates it; the lab and the server both get it.
TOKEN := CKA_SIM_AGENT_TOKEN="$$(cat $(HOME)/.config/cka-sim/agent-token)"

.PHONY: help doctor up down shell serve web dev test selftest

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

doctor: ## check this machine can run a lab
	@$(MAKE) --no-print-directory -C $(LAB) doctor

up: ## start the lab container (does nothing if it is already up)
	@$(MAKE) --no-print-directory -C $(LAB) up

down: ## delete the lab
	@$(MAKE) --no-print-directory -C $(LAB) down

shell: ## open a shell on the base host, where the exam starts
	@$(MAKE) --no-print-directory -C $(LAB) shell

serve: ## run the backend for the page on 127.0.0.1:7070
	cd $(PLATFORM) && $(TOKEN) go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binary runs directly, not via go run, so Ctrl-C reaches it and the trap can wait for it.
dev: up $(WEB)/node_modules ## lab up, then the backend and the page together
	@cd $(PLATFORM) && go build -o bin/server ./cmd/server
	@$(TOKEN) $(PLATFORM)/bin/server & server=$$!; trap 'kill $$server 2>/dev/null; wait $$server' INT TERM EXIT; cd $(WEB) && npm run dev

test: ## vet and test the Go code
	cd $(LAB) && go vet ./... && go test ./...
	cd $(PLATFORM) && go vet ./... && go test ./...

selftest: ## prove tasks are graded fairly (needs the lab); one task: make selftest TASK=<id>
	cd $(PLATFORM) && $(TOKEN) go test ./catalog -run 'TestSelftest$(if $(TASK),/^$(TASK)$$)' -selftest -count=1 -v -timeout 30m

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
