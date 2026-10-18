# cka-sim on GKE: the app behind HTTPS and an IP allowlist, and a lab pod per student that the platform makes on
# Start lab, each on a node of its own. Terraform makes the cluster and the IP; k8s/ is what runs on it.
# `make` alone lists the targets.

TF    := terraform -chdir=terraform
# Your address now; `make allow ALLOW="1.2.3.4/32 5.6.7.8/32"` lets in more than one.
ALLOW ?= $(shell curl -fsS https://ifconfig.me)/32
# The students made on the first `make site`, each with a token of its own; `make links` prints their links.
STUDENTS ?= alice bob
# Never the current kubectl context: it may be another project's cluster.
KC     = kubectl --context $(shell $(TF) output -raw kube_context)

.PHONY: help up cluster site allow url links labs lab-down site-down down

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-9s %s\n", $$1, $$2}'

up: cluster site ## everything, from nothing; about 10 minutes

cluster: ## create or update the cluster, its two node pools and the site's IP
	$(TF) init -input=false
	$(TF) apply
	$$($(TF) output -raw get_credentials)

site: ## put the app on the cluster and let only ALLOW in; `make url` prints the link
	$(KC) apply -f k8s/namespaces.yaml
	@# The token only ever lives in the cluster: the app and the lab pod read the same one.
	@$(KC) -n cka-sim get secret cka-sim-agent >/dev/null 2>&1 || { \
	  token=$$(openssl rand -hex 32); \
	  for ns in cka-sim cka-sim-labs; do $(KC) -n $$ns create secret generic cka-sim-agent --from-literal=token=$$token; done; }
	@# Each student and the token their link carries, made once: name=token,name=token.
	@$(KC) -n cka-sim get secret cka-sim-students >/dev/null 2>&1 || \
	  $(KC) -n cka-sim create secret generic cka-sim-students --from-literal=list=$$(for s in $(STUDENTS); do \
	    printf '%s=%s,' $$s $$(openssl rand -hex 32); done)
	@ip=$$($(TF) output -raw site_ip); \
	  printf 'SITE_IP=%s\nCKA_SIM_SITE_HOST=%s.sslip.io\nCKA_SIM_ALLOW=%s\n' "$$ip" "$$ip" "$(ALLOW)" > k8s/site.env
	$(KC) apply -k k8s
	@$(MAKE) --no-print-directory url

allow: site ## the same as site: your address may have changed

url: ## print the site's link
	@echo "https://$$($(TF) output -raw site_ip).sslip.io"

links: ## print each student's link; whoever opens one is that student
	@site=https://$$($(TF) output -raw site_ip).sslip.io; \
	  $(KC) -n cka-sim get secret cka-sim-students -o jsonpath='{.data.list}' | base64 -d | tr ',' '\n' | \
	  while IFS== read -r name token; do [ -n "$$name" ] && echo "$$name  $$site/api/join?token=$$token"; done

labs: ## each student's lab pod and the node it runs on, then the labs pool's nodes
	@$(KC) -n cka-sim-labs get pods -L cka-sim/owner -o wide
	@echo
	@$(KC) get nodes -l pool=labs

lab-down: ## delete one student's lab pod, as their Stop lab does: make lab-down NAME=alice
	@test -n "$(NAME)" || { echo "usage: make lab-down NAME=alice" >&2; exit 2; }
	$(KC) -n cka-sim-labs delete pod lab-$(NAME) --ignore-not-found

site-down: ## remove the app and the lab, and with them the load balancer; the cluster stays
	$(KC) delete namespace cka-sim cka-sim-labs --ignore-not-found

down: site-down ## delete the cluster and the IP too; the project and its other resources stay
	$(TF) destroy
