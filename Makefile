# Shortcuts for developing cka-sim from source. `make` alone lists them.

-include .env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' .env 2>/dev/null)

PLATFORM := apps/platform
LAB := apps/lab
ASSISTANT := apps/assistant
SHARED := apps/go-shared
WEB := apps/web
COMPOSE := docker compose -f infra/local/compose.yml
BUF := go run github.com/bufbuild/buf/cmd/buf@v1.74.0
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0
# The Kubernetes docs the tutor searches: v1.37's live on kubernetes/website's main until release-1.37 is cut.
DOCS_REF ?= main
DOCS_DIR := $(HOME)/.cache/cka-sim/website
DOCS_INDEX := $(CURDIR)/$(ASSISTANT)/data/docs-index.gob.gz

.PHONY: help up down serve web dev lab lab-logs assistant docs-index test e2e lab-e2e selftest eval lint proto boundaries

help:
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-10s %s\n", $$1, $$2}'

up: ## build and run cka-sim the way a user does, Docker only, on http://localhost:7080
	$(COMPOSE) up -d --build --wait
	@echo "cka-sim is on http://localhost:$${CKA_SIM_PORT:-7080}"

down: ## stop cka-sim, from make up or make lab; keeps the pulled node image for the next cluster
	$(COMPOSE) down

serve: ## run the backend for the page on 127.0.0.1:7070
	cd $(PLATFORM) && go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binary runs directly, not via go run, so Ctrl-C reaches it and the trap can wait for it.
dev: $(WEB)/node_modules ## the backend and the page together
	@cd $(PLATFORM) && go build -o bin/server ./cmd/server
	@$(PLATFORM)/bin/server & server=$$!; \
	  trap 'kill $$server 2>/dev/null; wait $$server' INT TERM EXIT; cd $(WEB) && npm run dev

lab: ## build and start only the lab, for make dev; its agent listens on 127.0.0.1:7171
	$(COMPOSE) up -d --build --wait lab

lab-logs: ## follow the lab's log, where a failed cluster build explains itself
	$(COMPOSE) logs -f lab

assistant: ## serve the tutor's MCP tools on 127.0.0.1:7373/mcp for your own AI CLI; needs make serve or make dev
	cd $(ASSISTANT) && CKA_SIM_DOCS_INDEX=$${CKA_SIM_DOCS_INDEX:-$(DOCS_INDEX)} go run ./cmd/assistant

docs-index: ## embed the Kubernetes docs for the tutor's search_docs with Ollama; re-runs embed only changed pages
	@test -d $(DOCS_DIR)/.git || git clone --depth 1 --filter=blob:none --sparse -b $(DOCS_REF) \
	  https://github.com/kubernetes/website $(DOCS_DIR)
	git -C $(DOCS_DIR) sparse-checkout set content/en/docs content/en/examples
	git -C $(DOCS_DIR) fetch --depth 1 origin $(DOCS_REF) && git -C $(DOCS_DIR) checkout -q FETCH_HEAD
	cd $(ASSISTANT) && go run ./cmd/docs -dir $(DOCS_DIR) -out $(DOCS_INDEX) -ref $$(git -C $(DOCS_DIR) rev-parse HEAD)

test: boundaries $(WEB)/node_modules ## vet and test the Go code, type-check the page
	cd $(SHARED) && go vet ./... && go test ./...
	cd $(PLATFORM) && go vet ./... && go test $$(go list ./... | grep -v /e2e$$)
	cd $(LAB) && go vet ./... && go test $$(go list ./... | grep -v /e2e$$)
	cd $(ASSISTANT) && go vet ./... && go test $$(go list ./... | grep -v /e2e$$)
	@# Never cached: the e2e tests build their binary in a subprocess, so Go can't see when its code changes.
	cd $(PLATFORM) && go test -count=1 ./e2e
	cd $(LAB) && go test -count=1 ./e2e
	cd $(ASSISTANT) && go test -count=1 ./e2e
	cd $(WEB) && npx tsc

e2e: ## run the platform's, the lab's and the assistant's end-to-end tests against their real binaries
	cd $(PLATFORM) && go test ./e2e -count=1 -v
	cd $(LAB) && go test ./e2e -count=1 -v
	cd $(ASSISTANT) && go test ./e2e -count=1 -v

lab-e2e: ## build and delete a real kind cluster through the agent on this machine's Docker (minutes)
	cd $(LAB) && go test ./e2e -count=1 -v -timeout 30m -run TestARealCluster -create

selftest: ## set up, solve and grade every task on the running lab (long); one task: make selftest TASK=<id>
	cd $(PLATFORM) && go test ./catalog -run 'TestSelftest$(if $(TASK),/^$(TASK)$$)' -selftest -count=1 -v -timeout 90m

eval: ## ask the tutor about broken, then solved tasks on the running lab, judged by Gemini; make eval TASK=a,b
	cd $(ASSISTANT) && go run ./cmd/eval $(if $(TASK),-tasks $(TASK))

lint: ## lint every Go module with the rules in .golangci.yml
	@for mod in apps/*/go.mod; do (cd $$(dirname $$mod) && $(GOLANGCI_LINT) run ./...) || exit 1; done

proto: ## lint the contracts in apps/go-shared/contract and regenerate their Go code
	cd $(SHARED) && $(BUF) lint && $(BUF) generate

boundaries: ## fail if an app imports another app (only apps/go-shared is shared)
	@bad=0; for mod in apps/*/go.mod; do \
	  app=$$(basename $$(dirname $$mod)); \
	  hits=$$(grep -rnoE 'cka-sim/apps/[a-z-]+' --include='*.go' --include=go.mod apps/$$app \
	    | grep -vE "cka-sim/apps/($$app|go-shared)$$"); \
	  if [ -n "$$hits" ]; then echo "$$app imports another app:"; echo "$$hits"; bad=1; fi; \
	done; exit $$bad

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
