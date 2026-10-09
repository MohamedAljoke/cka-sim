# Shortcuts for developing cka-sim from source. `make` alone lists them.

WEB := ../web

.PHONY: help doctor up down shell serve web dev test try selftest

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

doctor: ## check this machine can run cka-sim
	go run ./cmd/cli doctor

up: ## create the cluster (does nothing if it is already up)
	go run ./cmd/cli up

down: ## delete the cluster
	go run ./cmd/cli down

shell: ## open a shell on the base host, where the exam starts
	go run ./cmd/cli shell

serve: ## run the backend for the page on 127.0.0.1:7070
	go run ./cmd/server

web: $(WEB)/node_modules ## serve the page with hot reload on http://localhost:5173
	cd $(WEB) && npm run dev

# The binary runs directly, not via go run, so Ctrl-C reaches it and the trap can wait for it.
dev: up $(WEB)/node_modules ## cluster up, then the backend and the page together
	@go build -o bin/server ./cmd/server
	@bin/server & server=$$!; trap 'kill $$server 2>/dev/null; wait $$server' INT TERM EXIT; cd $(WEB) && npm run dev

test: ## vet and test the Go code
	go vet ./... && go test ./...

selftest: ## prove every task is graded fairly (needs the cluster)
	go test ./catalog -run TestSelftest -selftest -count=1 -v -timeout 30m

try: ## prove a task works on the cluster: make try TASK=<id>
	@scripts/try-task.sh $(TASK)

$(WEB)/node_modules: $(WEB)/package-lock.json
	cd $(WEB) && npm ci
	@touch $@
