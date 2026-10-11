# The lab on this machine: the image GKE runs, as a privileged container. `make` alone lists the targets.

NAME       := cka-sim-lab
IMAGE      := cka-sim-lab:dev
# The lab's own Docker: keeps the node image and task images between runs.
VOLUME     := cka-sim-lab-docker
PORT       := 7171
TOKEN_FILE := $(HOME)/.config/cka-sim/agent-token
# The platform's catalog decides which images a lab preloads.
IMAGES     := $(shell paste -sd, ../platform/catalog/images.txt)

.PHONY: help doctor image up down clean shell

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-7s %s\n", $$1, $$2}'

doctor: ## check this machine can run a lab
	go run ./cmd/lab doctor

image: ## build the lab image
	docker build -q -f deploy/lab.Dockerfile -t $(IMAGE) . >/dev/null

$(TOKEN_FILE):
	@mkdir -p $(dir $@)
	@umask 077 && openssl rand -hex 32 > $@

up: doctor image $(TOKEN_FILE) ## start the lab container, or reuse it; waits until its cluster is up
	@if [ "$$(docker inspect -f '{{.State.Running}}' $(NAME) 2>/dev/null)" = true ]; then \
	  [ "$$(docker inspect -f '{{.Image}}' $(NAME))" = "$$(docker image inspect -f '{{.Id}}' $(IMAGE))" ] || \
	    echo "the lab runs an older image; make down up to use the new one"; \
	elif docker inspect $(NAME) >/dev/null 2>&1; then \
	  docker start $(NAME) >/dev/null; \
	else \
	  docker run -d --privileged --name $(NAME) \
	    -v /lib/modules:/lib/modules:ro -v $(VOLUME):/var/lib/docker -p 127.0.0.1:$(PORT):7171 \
	    -e CKA_SIM_AGENT_TOKEN="$$(cat $(TOKEN_FILE))" -e CKA_SIM_IMAGES=$(IMAGES) $(IMAGE) >/dev/null && \
	  echo "building the lab's cluster (about 3 minutes the first time)"; \
	fi
	@start=$$(date +%s); until curl -sf -o /dev/null http://127.0.0.1:$(PORT)/healthz; do \
	  [ "$$(docker inspect -f '{{.State.Running}}' $(NAME))" = true ] || { docker logs --tail 30 $(NAME); exit 1; }; \
	  sleep 2; \
	done; echo "lab ready at http://127.0.0.1:$(PORT) after $$(( $$(date +%s) - start ))s"

# lab down first: the volume outlives the container, and a cluster left in it would look up to the next lab.
down: ## delete the lab's cluster and its container; keeps the downloaded images
	@docker exec $(NAME) lab down >/dev/null 2>&1 || true
	@docker rm -f $(NAME) >/dev/null 2>&1 || true
	@echo "lab $(NAME) is gone"

clean: down ## also delete the downloaded images
	@docker volume rm $(VOLUME) >/dev/null 2>&1 || true

shell: ## open a shell on the base host, where the exam starts
	docker exec -it $(NAME) lab shell
