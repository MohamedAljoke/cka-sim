# Shortcuts for developing cka-sim from source. `make` alone lists them.

PLATFORM := apps/platform
WEB := apps/web

.PHONY: help serve web dev test boundaries

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
	cd apps/go-shared && go vet ./... && go test ./...
	cd $(PLATFORM) && go vet ./... && go test ./...
	cd $(WEB) && npx tsc

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
