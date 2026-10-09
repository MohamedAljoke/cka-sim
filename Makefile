# Shortcuts for the whole cycle. `make` alone lists them.
export GOTOOLCHAIN := local

BIN     := bin/cka-sim
PORT    ?= 8080
MINUTES ?= 120
N       ?= 16
TASKS   ?=

.PHONY: help build install up down reset status exam study shell open resume check test selftest vendor clean
.DEFAULT_GOAL := help

help:
	@echo "Environment"
	@echo "  make reset              drop everything and build it again (down + up, ~5 min)"
	@echo "  make up | down          create / delete the clusters and the base host"
	@echo "  make status             what is running, memory, current session"
	@echo ""
	@echo "Practise"
	@echo "  make exam               timed exam on the current clusters   MINUTES=60 N=8"
	@echo "  make study              no clock, check/reset/solution per task   TASKS=\"tr-kubelet\""
	@echo "  make open               open the panel: questions left, a terminal on base right"
	@echo "  make shell              or work from your own terminal: you are candidate@base"
	@echo "  make resume             bring back the panel of the current exam or study session"
	@echo ""
	@echo "Develop"
	@echo "  make check              what CI runs     make selftest   every task: 0 untouched, full marks solved"
	@echo "  make vendor             refetch the browser terminal's xterm.js into web/vendor"
	@echo ""
	@echo "Typical:  make reset && make exam MINUTES=60    then in another terminal:  make open"

build:
	@go build -o $(BIN) ./cmd/cka-sim

install:
	go install ./cmd/cka-sim

up: build
	$(BIN) up

down: build
	$(BIN) down

reset: build
	$(BIN) reset

status: build
	@$(BIN) status

# The clusters come from `make up` or `make reset`, so the exam does not rebuild them again.
exam: build
	$(BIN) exam -fresh=false -minutes $(MINUTES) -n $(N) -port $(PORT)

study: build
	$(BIN) study -fresh=false -port $(PORT) $(TASKS)

resume: build
	@if grep -q '"study": true' $${XDG_CACHE_HOME:-$$HOME/.cache}/cka-sim/exam.json 2>/dev/null; \
		then $(BIN) study -resume -port $(PORT); else $(BIN) exam -resume -port $(PORT); fi

shell: build
	@$(BIN) shell

# explorer.exe opens the Windows browser from WSL; xdg-open covers plain Linux.
open:
	@url=http://localhost:$(PORT); \
	echo "opening $$url in your browser (or open it yourself)"; \
	if command -v explorer.exe >/dev/null; then explorer.exe $$url || true; \
	elif command -v xdg-open >/dev/null; then xdg-open $$url; \
	else echo "open $$url"; fi

test:
	go test -race ./...

check: test
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	go vet ./...
	@for f in tasks/lib.sh tasks/*/*/*.sh; do bash -n "$$f" || exit 1; done

selftest: build
	$(BIN) selftest $(TASKS)

# The browser terminal's xterm.js, committed under web/vendor so the panel works offline.
XTERM     := @xterm/xterm@5.5.0
XTERM_FIT := @xterm/addon-fit@0.10.0
vendor:
	@mkdir -p web/vendor
	curl -fsSL https://cdn.jsdelivr.net/npm/$(XTERM)/lib/xterm.js -o web/vendor/xterm.js
	curl -fsSL https://cdn.jsdelivr.net/npm/$(XTERM)/css/xterm.css -o web/vendor/xterm.css
	curl -fsSL https://cdn.jsdelivr.net/npm/$(XTERM_FIT)/lib/addon-fit.js -o web/vendor/addon-fit.js

clean:
	rm -rf bin
