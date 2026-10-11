# cka-sim on GKE: the platform on the system pool, labs from a warm pool on their own nodes. A spike, kept in
# its own GCP project so `make down` removes every trace. `make` alone lists the targets.

PROJECT ?= cka-sim-gke-maljoke
BILLING ?= 019195-7CBF40-70FE42
REGION  := southamerica-east1
ZONE    := $(REGION)-a
CLUSTER := cka-sim
REPO    := cka-sim
IMAGE   := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/lab:latest
APPS     := $(abspath ../../apps)
LAB_DIR  := $(APPS)/lab
PLATFORM := $(APPS)/platform
# The platform's catalog decides which images a lab preloads.
IMAGES   := $(shell paste -sd, $(PLATFORM)/catalog/images.txt)
# Lab pods get their own range, so the firewall can tell them from GKE's system pods.
LAB_PODS := 10.80.0.0/16
APP_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/app:latest
SITE_PASSWORD := $(HOME)/.config/cka-sim/gke-password
POOL    ?= 1
SANDBOX := v1.0.4
GC      := gcloud --project $(PROJECT)

.PHONY: help project cluster image sandbox pool wall site url site-down down

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

project: ## create the project, link billing, enable the APIs (once)
	gcloud projects create $(PROJECT) --name "cka-sim GKE spike"
	gcloud billing projects link $(PROJECT) --billing-account $(BILLING)
	$(GC) services enable container.googleapis.com artifactregistry.googleapis.com
	$(GC) artifacts repositories create $(REPO) --repository-format docker --location $(REGION)

cluster: ## create the cluster: a small system pool and an empty, tainted "labs" pool
	$(GC) container clusters create $(CLUSTER) --zone $(ZONE) \
	  --num-nodes 1 --machine-type e2-medium --disk-size 30 --enable-autoscaling --min-nodes 1 --max-nodes 2 \
	  --workload-pool $(PROJECT).svc.id.goog --enable-dataplane-v2
	$(GC) compute networks subnets update default --region $(REGION) --add-secondary-ranges labs-pods=$(LAB_PODS)
	@# Labs scale from 0: a node per lab, removed when its lab is gone.
	$(GC) container node-pools create labs --pod-ipv4-range labs-pods --cluster $(CLUSTER) --zone $(ZONE) \
	  --machine-type e2-standard-2 --image-type UBUNTU_CONTAINERD --disk-size 50 \
	  --num-nodes 0 --enable-autoscaling --min-nodes 0 --max-nodes 2 \
	  --node-labels pool=labs --node-taints lab=true:NoSchedule --tags cka-sim-lab \
	  --workload-metadata GKE_METADATA
	$(GC) container clusters get-credentials $(CLUSTER) --zone $(ZONE)

image: ## build and push the lab image
	$(GC) auth configure-docker $(REGION)-docker.pkg.dev --quiet
	docker build -f $(LAB_DIR)/deploy/lab.Dockerfile -t $(IMAGE) $(LAB_DIR)
	docker push $(IMAGE)

sandbox: ## install agent-sandbox (the warm pool controller)
	kubectl apply -f https://github.com/kubernetes-sigs/agent-sandbox/releases/download/$(SANDBOX)/sandbox.yaml
	kubectl apply -f https://github.com/kubernetes-sigs/agent-sandbox/releases/download/$(SANDBOX)/extensions.yaml

pool: ## keep POOL warm labs (POOL=1); POOL=0 empties it
	POOL_SIZE=$(POOL) IMAGE=$(IMAGE) IMAGES=$(IMAGES) envsubst < $(LAB_DIR)/deploy/gke/lab-pool.yaml | kubectl apply -f -

wall: ## firewall: no lab node or lab pod reaches another lab node (agent-sandbox's NetworkPolicy covers pod to pod)
	@# Above GKE's own allow rules (priority 1000). Source tags match the lab nodes' own addresses, the range their
	@# pods. System pods stay allowed: the control plane reaches each kubelet through one (konnectivity).
	$(GC) compute firewall-rules create cka-sim-labs-apart --network default --direction INGRESS --priority 900 \
	  --action DENY --rules all --target-tags cka-sim-lab --source-tags cka-sim-lab --source-ranges $(LAB_PODS)

site: image ## the app on the platform pool, a warm lab in the pool, HTTPS + password; `make url` prints the link
	$(GC) compute addresses describe cka-sim-site --region $(REGION) >/dev/null 2>&1 || \
	  $(GC) compute addresses create cka-sim-site --region $(REGION)
	docker build -f $(PLATFORM)/deploy/gke/app.Dockerfile -t $(APP_IMAGE) $(APPS)
	docker push $(APP_IMAGE)
	@test -s $(SITE_PASSWORD) || (mkdir -p $(dir $(SITE_PASSWORD)) && umask 077 && openssl rand -base64 15 > $(SITE_PASSWORD))
	@kubectl create secret generic cka-sim-site --dry-run=client -o yaml \
	  --from-literal=hash="$$(docker run --rm --entrypoint caddy $(APP_IMAGE) hash-password --plaintext "$$(cat $(SITE_PASSWORD))")" | kubectl apply -f -
	@# The token only ever lives in the cluster: the app and the labs read the same Secret.
	@kubectl get secret cka-sim-agent >/dev/null 2>&1 || \
	  kubectl create secret generic cka-sim-agent --from-literal=token="$$(openssl rand -hex 32)"
	$(MAKE) --no-print-directory pool POOL=$(POOL)
	@ip=$$($(GC) compute addresses describe cka-sim-site --region $(REGION) --format='value(address)'); \
	APP_IMAGE=$(APP_IMAGE) SITE_IP=$$ip SITE_HOST=$$(echo $$ip | tr . -).sslip.io envsubst < $(PLATFORM)/deploy/gke/app.yaml | kubectl apply -f -
	kubectl rollout restart deployment/app

url: ## print the site's link and password
	@echo "https://$$($(GC) compute addresses describe cka-sim-site --region $(REGION) --format='value(address)' | tr . -).sslip.io"
	@echo "user cka, password $$(cat $(SITE_PASSWORD))"

site-down: ## remove the app, its load balancer, its IP, and every lab
	kubectl delete deployment app --ignore-not-found
	kubectl delete sandboxclaims -l app.kubernetes.io/managed-by=cka-sim --ignore-not-found
	$(MAKE) --no-print-directory pool POOL=0
	kubectl delete service site secret cka-sim-site --ignore-not-found
	$(GC) compute addresses delete cka-sim-site --region $(REGION) --quiet

down: ## delete the whole project, cluster and images included
	gcloud projects delete $(PROJECT) --quiet
