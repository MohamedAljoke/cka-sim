# Shortcuts for developing cka-sim from source. `make` alone lists them.

PLATFORM := apps/platform
LAB := apps/lab
WEB := apps/web
AUTH := apps/auth
MCP := apps/mcp
# The same Compose project `cka-sim up` runs, so dev and the installed cka-sim share one lab, database and secrets.
COMPOSE := docker compose -f $(CURDIR)/infra/local/compose.yml -f $(CURDIR)/infra/local/compose.dev.yml
secret = "$$($(COMPOSE) run --rm --no-deps -T secrets cat /secrets/$(1))"
TOKEN := CKA_SIM_AGENT_TOKEN=$(call secret,agent-token)
AUTH_ENV := CKA_SIM_AUTH_SECRET=$(call secret,auth-secret)

.PHONY: help doctor secrets up down shell serve web dev db auth mcp test boundaries selftest eval try

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

doctor: ## check this machine can run a lab
	@$(MAKE) --no-print-directory -C $(LAB) doctor

secrets:
	@$(COMPOSE) run --rm secrets

up: doctor ## build and start the lab, stopping cka-sim's own app, auth and mcp; waits until its cluster is up
	@$(COMPOSE) stop app auth mcp
	$(COMPOSE) up -d --build --wait --wait-timeout 1800 lab

down: ## stop everything; keeps the cluster, accounts and images, like cka-sim down
	$(COMPOSE) down

shell: ## open a shell on the base host, where the exam starts
	$(COMPOSE) exec lab lab shell

serve: secrets ## run the backend for the page on 127.0.0.1:7070
	cd $(PLATFORM) && $(TOKEN) $(AUTH_ENV) go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binaries run directly, not via go run, so Ctrl-C reaches them and the trap can wait for them.
dev: up db secrets $(WEB)/node_modules ## lab and Postgres up, then auth, the backend, MCP and the page together
	@cd $(PLATFORM) && go build -o bin/server ./cmd/server
	@cd $(AUTH) && go build -o bin/server ./cmd/server
	@cd $(MCP) && go build -o bin/server ./cmd/server
	@$(TOKEN) $(AUTH_ENV) $(PLATFORM)/bin/server & server=$$!; $(AUTH_ENV) $(AUTH)/bin/server & auth=$$!; \
	  $(MCP)/bin/server & mcp=$$!; \
	  trap 'kill $$server $$auth $$mcp 2>/dev/null; wait $$server $$auth $$mcp' INT TERM EXIT; cd $(WEB) && npm run dev

db: ## start Postgres, then apply migrations and seeds
	$(COMPOSE) up -d --wait postgres
	@$(MAKE) --no-print-directory -C $(AUTH) migrate seed

auth: secrets ## run the auth server on 127.0.0.1:7272 (make db first)
	cd $(AUTH) && $(AUTH_ENV) go run ./cmd/server

mcp: ## run the MCP server for the Assistant on 127.0.0.1:7373 (the lab reaches it as host.docker.internal)
	cd $(MCP) && go run ./cmd/server

test: boundaries ## vet and test the Go code
	cd apps/go-shared && go vet ./... && go test ./...
	cd $(LAB) && go vet ./... && go test ./...
	cd $(PLATFORM) && go vet ./... && go test ./...
	cd $(AUTH) && go vet ./... && go test ./...
	cd $(MCP) && go vet ./... && go test ./...

boundaries: ## fail if an app imports another app (only apps/go-shared is shared)
	@bad=0; for mod in apps/*/go.mod; do \
	  app=$$(basename $$(dirname $$mod)); \
	  hits=$$(grep -rnoE 'cka-sim/apps/[a-z-]+' --include='*.go' --include=go.mod apps/$$app \
	    | grep -vE "cka-sim/apps/($$app|go-shared)$$"); \
	  if [ -n "$$hits" ]; then echo "$$app imports another app:"; echo "$$hits"; bad=1; fi; \
	done; exit $$bad

selftest: ## prove tasks are graded fairly (needs the lab); one task: make selftest TASK=<id>
	cd $(PLATFORM) && $(TOKEN) go test ./catalog -run 'TestSelftest$(if $(TASK),/^$(TASK)$$)' -selftest -count=1 -v -timeout 30m

eval: ## score the Assistant's reviews (needs make dev and Claude logged in once); one task: make eval TASK=<id>
	cd $(PLATFORM) && $(TOKEN) go run ./cmd/eval $(if $(TASK),-task $(TASK))

try: ## build from source and run cka-sim the way users do, Docker only, on http://localhost:7080
	$(COMPOSE) up -d --build --wait --wait-timeout 1800
	@echo "ready on http://localhost:$${CKA_SIM_PORT:-7080}; log in as dev@cka-sim.local, password cka-sim-dev"

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
