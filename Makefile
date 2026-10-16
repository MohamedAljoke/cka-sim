# Shortcuts for developing cka-sim from source. `make` alone lists them.

-include .env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' .env 2>/dev/null)

PLATFORM := apps/platform
SHARED := apps/go-shared
WEB := apps/web
BUF := go run github.com/bufbuild/buf/cmd/buf@v1.74.0
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0

.PHONY: help serve web dev test e2e lint proto boundaries

help:
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-10s %s\n", $$1, $$2}'

serve: ## run the backend for the page on 127.0.0.1:7070
	cd $(PLATFORM) && go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binary runs directly, not via go run, so Ctrl-C reaches it and the trap can wait for it.
dev: $(WEB)/node_modules ## the backend and the page together
	@cd $(PLATFORM) && go build -o bin/server ./cmd/server
	@$(PLATFORM)/bin/server & server=$$!; \
	  trap 'kill $$server 2>/dev/null; wait $$server' INT TERM EXIT; cd $(WEB) && npm run dev

test: boundaries $(WEB)/node_modules ## vet and test the Go code, type-check the page
	cd $(SHARED) && go vet ./... && go test ./...
	cd $(PLATFORM) && go vet ./... && go test $$(go list ./... | grep -v /e2e$$)
	@# Never cached: the e2e tests build their binary in a subprocess, so Go can't see when its code changes.
	cd $(PLATFORM) && go test -count=1 ./e2e
	cd $(WEB) && npx tsc

e2e: ## run the platform's end-to-end tests against the real server binary
	cd $(PLATFORM) && go test ./e2e -count=1 -v

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
