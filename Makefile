# Shortcuts for developing cka-sim from source. `make` alone lists them.

PLATFORM := apps/platform
LAB := apps/lab
WEB := apps/web
AUTH := apps/auth
MCP := apps/mcp
# make up creates it; the lab and the server both get it.
TOKEN := CKA_SIM_AGENT_TOKEN="$$(cat $(HOME)/.config/cka-sim/agent-token)"
# Signs login sessions; the platform reads the same file to check them.
AUTH_SECRET := $(HOME)/.config/cka-sim/auth-secret
AUTH_ENV := CKA_SIM_AUTH_SECRET="$$(cat $(AUTH_SECRET))"

.PHONY: help doctor up down shell serve web dev db auth mcp test boundaries selftest

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

serve: $(AUTH_SECRET) ## run the backend for the page on 127.0.0.1:7070
	cd $(PLATFORM) && $(TOKEN) $(AUTH_ENV) go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binaries run directly, not via go run, so Ctrl-C reaches them and the trap can wait for them.
dev: up db $(WEB)/node_modules $(AUTH_SECRET) ## lab and Postgres up, then auth, the backend, MCP and the page together
	@cd $(PLATFORM) && go build -o bin/server ./cmd/server
	@cd $(AUTH) && go build -o bin/server ./cmd/server
	@cd $(MCP) && go build -o bin/server ./cmd/server
	@$(TOKEN) $(AUTH_ENV) $(PLATFORM)/bin/server & server=$$!; $(AUTH_ENV) $(AUTH)/bin/server & auth=$$!; \
	  $(MCP)/bin/server & mcp=$$!; \
	  trap 'kill $$server $$auth $$mcp 2>/dev/null; wait $$server $$auth $$mcp' INT TERM EXIT; cd $(WEB) && npm run dev

db: ## start Postgres, then apply migrations and seeds
	docker compose up -d --wait postgres
	@$(MAKE) --no-print-directory -C $(AUTH) migrate seed

auth: $(AUTH_SECRET) ## run the auth server on 127.0.0.1:7272 (make db first)
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

$(AUTH_SECRET):
	@mkdir -p $(dir $@)
	@umask 077 && openssl rand -hex 32 > $@

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
