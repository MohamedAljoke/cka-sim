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
AUTH     := $(APPS)/auth
MCP      := $(APPS)/mcp
# The platform's catalog decides which images a lab preloads.
IMAGES   := $(shell paste -sd, $(PLATFORM)/catalog/images.txt)
# Lab pods get their own range, so the firewall can tell them from GKE's system pods.
LAB_PODS := 10.80.0.0/16
APP_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/app:latest
AUTH_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/auth:latest
MCP_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/mcp:latest
DOCS_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT)/$(REPO)/docs:latest
# Docs for the lab's Kubernetes version; v1.37 is current, so it lives on main until release-1.37 is cut.
DOCS_REF ?= main
SQL      := cka-sim-db
INSTANCE := $(PROJECT):$(REGION):$(SQL)
AUTH_GSA := cka-sim-auth@$(PROJECT).iam.gserviceaccount.com
APP_GSA  := cka-sim-app@$(PROJECT).iam.gserviceaccount.com
DB_PASSWORD := $(HOME)/.config/cka-sim/gke-db-password
POOL    ?= 1
SANDBOX := v1.0.4
GC      := gcloud --project $(PROJECT)

.PHONY: help project cluster image sandbox pool wall db ai site docs url site-down ai-down db-down down

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

project: ## create the project, link billing, enable the APIs (once)
	gcloud projects create $(PROJECT) --name "cka-sim GKE spike"
	gcloud billing projects link $(PROJECT) --billing-account $(BILLING)
	$(GC) services enable container.googleapis.com artifactregistry.googleapis.com sqladmin.googleapis.com
	$(GC) artifacts repositories create $(REPO) --repository-format docker --location $(REGION)

cluster: ## create the cluster: a small system pool and an empty, tainted "labs" pool
	$(GC) container clusters create $(CLUSTER) --zone $(ZONE) \
	  --num-nodes 1 --machine-type e2-medium --disk-size 30 --enable-autoscaling --min-nodes 1 --max-nodes 3 \
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
	docker build -f $(LAB_DIR)/deploy/lab.Dockerfile -t $(IMAGE) $(APPS)
	docker push $(IMAGE)

sandbox: ## install agent-sandbox (the warm pool controller)
	kubectl apply -f https://github.com/kubernetes-sigs/agent-sandbox/releases/download/$(SANDBOX)/sandbox.yaml
	kubectl apply -f https://github.com/kubernetes-sigs/agent-sandbox/releases/download/$(SANDBOX)/extensions.yaml
	@# A lab boots kind plus its add-ons in more than the default 5m, so a warm lab was replaced before it was ever Ready.
	kubectl -n agent-sandbox-system patch deployment agent-sandbox-controller --type json -p '[{"op": "replace", \
	  "path": "/spec/template/spec/containers/0/args", \
	  "value": ["--leader-elect=true", "--extensions", "--sandbox-warm-pool-readiness-grace-period=20m"]}]'

pool: ## keep POOL warm labs (POOL=1); POOL=0 empties it
	POOL_SIZE=$(POOL) IMAGE=$(IMAGE) IMAGES=$(IMAGES) envsubst < $(LAB_DIR)/deploy/gke/lab-pool.yaml | kubectl apply -f -

wall: ## firewall: no lab node or lab pod reaches another lab node (agent-sandbox's NetworkPolicy covers pod to pod)
	@# Above GKE's own allow rules (priority 1000). Source tags match the lab nodes' own addresses, the range their
	@# pods. System pods stay allowed: the control plane reaches each kubelet through one (konnectivity).
	$(GC) compute firewall-rules create cka-sim-labs-apart --network default --direction INGRESS --priority 900 \
	  --action DENY --rules all --target-tags cka-sim-lab --source-tags cka-sim-lab --source-ranges $(LAB_PODS)

db: ## Cloud SQL Postgres for auth (smallest tier) and the identity its pod uses to reach it (once)
	$(GC) services enable sqladmin.googleapis.com
	$(GC) sql instances describe $(SQL) >/dev/null 2>&1 || \
	  $(GC) sql instances create $(SQL) --database-version POSTGRES_16 --edition enterprise --tier db-f1-micro \
	    --region $(REGION) --storage-type HDD --storage-size 10 --no-backup --availability-type zonal
	$(GC) sql databases describe cka_sim --instance $(SQL) >/dev/null 2>&1 || \
	  $(GC) sql databases create cka_sim --instance $(SQL)
	@test -s $(DB_PASSWORD) || (mkdir -p $(dir $(DB_PASSWORD)) && umask 077 && openssl rand -hex 24 > $(DB_PASSWORD))
	@$(GC) sql users list --instance $(SQL) --format='value(name)' | grep -qx cka_sim || \
	  $(GC) sql users create cka_sim --instance $(SQL) --password "$$(cat $(DB_PASSWORD))"
	$(GC) iam service-accounts describe $(AUTH_GSA) >/dev/null 2>&1 || \
	  $(GC) iam service-accounts create cka-sim-auth --display-name "cka-sim auth"
	$(GC) projects add-iam-policy-binding $(PROJECT) --member serviceAccount:$(AUTH_GSA) \
	  --role roles/cloudsql.client --condition None >/dev/null
	$(GC) iam service-accounts add-iam-policy-binding $(AUTH_GSA) --role roles/iam.workloadIdentityUser \
	  --member "serviceAccount:$(PROJECT).svc.id.goog[default/cka-sim-auth]" >/dev/null

ai: ## Vertex AI for the Gemini assistant and the identity the app's pod calls it and its docs database with (once)
	$(GC) services enable aiplatform.googleapis.com
	$(GC) iam service-accounts describe $(APP_GSA) >/dev/null 2>&1 || \
	  $(GC) iam service-accounts create cka-sim-app --display-name "cka-sim app"
	$(GC) projects add-iam-policy-binding $(PROJECT) --member serviceAccount:$(APP_GSA) \
	  --role roles/aiplatform.user --condition None >/dev/null
	$(GC) projects add-iam-policy-binding $(PROJECT) --member serviceAccount:$(APP_GSA) \
	  --role roles/cloudsql.client --condition None >/dev/null
	$(GC) iam service-accounts add-iam-policy-binding $(APP_GSA) --role roles/iam.workloadIdentityUser \
	  --member "serviceAccount:$(PROJECT).svc.id.goog[default/cka-sim-app]" >/dev/null

site: image ai ## the app (with MCP) and auth on the platform pool, a warm lab in the pool, HTTPS + login; `make url` prints the link
	$(GC) compute addresses describe cka-sim-site --region $(REGION) >/dev/null 2>&1 || \
	  $(GC) compute addresses create cka-sim-site --region $(REGION)
	docker build -f $(PLATFORM)/deploy/app.Dockerfile -t $(APP_IMAGE) $(APPS)
	docker push $(APP_IMAGE)
	docker build -f $(AUTH)/deploy/auth.Dockerfile -t $(AUTH_IMAGE) $(APPS)
	docker push $(AUTH_IMAGE)
	docker build -f $(MCP)/deploy/mcp.Dockerfile -t $(MCP_IMAGE) $(APPS)
	docker push $(MCP_IMAGE)
	@kubectl get secret cka-sim-auth >/dev/null 2>&1 || \
	  kubectl create secret generic cka-sim-auth --from-literal=secret="$$(openssl rand -hex 32)"
	@kubectl create secret generic cka-sim-db --dry-run=client -o yaml \
	  --from-literal=url="postgres://cka_sim:$$(cat $(DB_PASSWORD))@127.0.0.1:5432/cka_sim?sslmode=disable" | kubectl apply -f -
	@# The token only ever lives in the cluster: the app and the labs read the same Secret.
	@kubectl get secret cka-sim-agent >/dev/null 2>&1 || \
	  kubectl create secret generic cka-sim-agent --from-literal=token="$$(openssl rand -hex 32)"
	$(MAKE) --no-print-directory pool POOL=$(POOL)
	AUTH_IMAGE=$(AUTH_IMAGE) INSTANCE=$(INSTANCE) AUTH_GSA=$(AUTH_GSA) envsubst < $(AUTH)/deploy/gke/auth.yaml | kubectl apply -f -
	@ip=$$($(GC) compute addresses describe cka-sim-site --region $(REGION) --format='value(address)'); \
	APP_IMAGE=$(APP_IMAGE) MCP_IMAGE=$(MCP_IMAGE) SITE_IP=$$ip SITE_HOST=$$ip.sslip.io APP_GSA=$(APP_GSA) PROJECT=$(PROJECT) \
	  INSTANCE=$(INSTANCE) envsubst < $(PLATFORM)/deploy/gke/app.yaml | kubectl apply -f -
	kubectl rollout restart deployment/auth deployment/app

docs: ## embed the Kubernetes docs for the assistant's search_docs into Cloud SQL (a Job; about $0.30 of Vertex the first time)
	docker build -f $(PLATFORM)/deploy/docs.Dockerfile -t $(DOCS_IMAGE) $(APPS) \
	  --build-arg DOCS_SHA=$$(git ls-remote https://github.com/kubernetes/website refs/heads/$(DOCS_REF) | cut -f1)
	docker push $(DOCS_IMAGE)
	kubectl delete job docs-index --ignore-not-found
	DOCS_IMAGE=$(DOCS_IMAGE) INSTANCE=$(INSTANCE) PROJECT=$(PROJECT) \
	  envsubst < $(PLATFORM)/deploy/gke/docs-index.yaml | kubectl apply -f -
	@until kubectl logs -f job/docs-index -c index 2>/dev/null; do sleep 3; done
	@test "$$(kubectl get job docs-index -o jsonpath='{.status.succeeded}')" = 1

url: ## print the site's link and login
	@echo "https://$$($(GC) compute addresses describe cka-sim-site --region $(REGION) --format='value(address)').sslip.io"
	@echo "log in as dev@cka-sim.local or dev2@cka-sim.local, password cka-sim-dev"

site-down: ## remove the app, auth, the load balancer, its IP, and every lab
	kubectl delete deployment app auth --ignore-not-found
	kubectl delete job docs-index --ignore-not-found
	kubectl delete sandboxclaims -l app.kubernetes.io/managed-by=cka-sim --ignore-not-found
	$(MAKE) --no-print-directory pool POOL=0
	kubectl delete service/site service/auth serviceaccount/cka-sim-auth pvc/caddy-data \
	  secret/cka-sim-auth secret/cka-sim-db secret/cka-sim-site --ignore-not-found
	$(GC) compute addresses delete cka-sim-site --region $(REGION) --quiet

ai-down: ## delete the app's Vertex AI identity
	$(GC) iam service-accounts delete $(APP_GSA) --quiet

db-down: ## delete the Cloud SQL instance (its data too) and auth's identity
	$(GC) sql instances delete $(SQL) --quiet
	$(GC) iam service-accounts delete $(AUTH_GSA) --quiet

down: ## delete the whole project, cluster and images included
	gcloud projects delete $(PROJECT) --quiet
