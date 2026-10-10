# cka-sim on Fly: the site app and the pool app of warm exam VMs. `make` alone lists the targets.

SITE   := cka-sim-maljoke
LABS   := cka-sim-labs-maljoke
REGION := gru
APPS   := $(abspath ../../..)
MODULE := $(abspath ../..)

.PHONY: help apps image pool list deploy password down

help:
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

apps: ## create both Fly apps and the site's secrets (once)
	fly apps create $(LABS) --org personal
	fly apps create $(SITE) --org personal
	@# The site may only manage the pool app's VMs; the token goes straight into the secret, never on screen.
	fly tokens create deploy -a $(LABS) --name cka-sim-site | grep '^FlyV1 ' | sed 's/^/FLY_API_TOKEN=/' | fly secrets import -a $(SITE) --stage
	@$(MAKE) --no-print-directory password

password: ## set a new site password, saved to ~/.config/cka-sim/fly-password
	@mkdir -p ~/.config/cka-sim && umask 077 && openssl rand -base64 18 > ~/.config/cka-sim/fly-password
	@echo "CKA_SIM_PASSWORD=$$(cat ~/.config/cka-sim/fly-password)" | fly secrets import -a $(SITE) --stage
	@echo "password saved in ~/.config/cka-sim/fly-password"

image: ## build and push the exam VM image
	fly auth docker
	docker build -f lab.Dockerfile -t registry.fly.io/$(LABS):lab $(MODULE)
	docker push registry.fly.io/$(LABS):lab

pool: ## add N warm exam VMs: make pool N=2
	FLY_LABS_APP=$(LABS) FLY_REGION=$(REGION) ./pool.sh fill $(or $(N),1)

list: ## list the exam VMs and their pool state
	FLY_LABS_APP=$(LABS) ./pool.sh list

deploy: ## build and deploy the site
	fly auth docker
	docker build -f server.Dockerfile -t registry.fly.io/$(SITE):latest $(APPS)
	docker push registry.fly.io/$(SITE):latest
	fly deploy -c fly.toml -a $(SITE) --image registry.fly.io/$(SITE):latest --ha=false

down: ## destroy both apps and every VM in them
	fly apps destroy $(SITE) --yes
	fly apps destroy $(LABS) --yes
