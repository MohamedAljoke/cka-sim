# cka-sim on GKE: Terraform and the lab improvements

Each section says what is wrong today (with the file), what to change, **how to test it**, and what the test
teaches you. The tests are the point: each one proves a boundary or a behaviour instead of trusting the config.

The assistant's improvements live in [IMPROVEMENTS-ASSISTANT.md](IMPROVEMENTS-ASSISTANT.md).

## Where it stands

```
GKE cluster (zonal, Dataplane V2, Workload Identity)
├─ default-pool (e2-medium, 1–3)   app + mcp + cloud-sql-proxy, auth, agent-sandbox controller
└─ labs pool (e2-standard-2, 0–2)  ONE privileged lab pod per node: dockerd → kind (3 nodes) → agent :7171
   taint lab=true, label pool=labs, tag cka-sim-lab, pod range 10.80.0.0/16
```

The design is already the right one: the lab pod is privileged, so **the GKE node (a VM) is the boundary
between students**. Every point below either keeps that true (node SA, node reuse, private nodes) or makes it
cheaper and sturdier (Terraform, restarts, boot time, CI).

The same pattern runs Transfeera's CircleCI runner (`terraform-live/shared-services/aws/us-east-1/eks`): a
controller turns a request into a pod on a tainted, autoscaled pool. One difference to notice: EKS needed
`k8s.io/cluster-autoscaler/node-template/*` tags on the ASG to scale from 0; GKE's autoscaler reads the node
pool's labels and taints itself.

## 1. Terraform instead of `infra/gke/Makefile`

The Makefile works, but it can't show a diff, can't detect drift, and `make down` only works because the
whole project is deleted. Keep the Makefile for builds; move infrastructure to Terraform.

### Layout: two stacks, like `terraform-live`'s `eks/` and `eks/kubernetes/`

```
infra/terraform/
├─ gcp/        project APIs, network, NAT, GKE cluster + node pools, firewall, Artifact Registry,
│              Cloud SQL, service accounts + IAM, static IP, budget
└─ cluster/    in-cluster platform: agent-sandbox (kustomize), secrets, RBAC
```

`cluster/` needs the cluster to exist before its providers can connect, so it gets its own state. App, auth
and the lab pool template stay as YAML applied by CI with an image digest (the `terraform-live` vs `k8s` repo
split): they change on every deploy, infrastructure doesn't.

State goes in a GCS bucket with versioning. Commit `.terraform.lock.hcl` so provider versions are pinned.
Generated secrets (`random_password`) end up in state, so restrict who can read the bucket.

### Makefile target → Terraform

| Make target | Terraform resources |
|---|---|
| `project` | `google_project_service` (container, artifactregistry, sqladmin, aiplatform), `google_artifact_registry_repository` |
| `cluster` | `google_compute_network`, `google_compute_subnetwork` (secondary ranges), `google_container_cluster`, two `google_container_node_pool` |
| `wall` | `google_compute_firewall` |
| `db` | `google_sql_database_instance`, `google_sql_database`, `google_sql_user`, `random_password` |
| `ai`, `db` identities | `google_service_account`, `google_project_iam_member`, `google_service_account_iam_member` (Workload Identity) |
| `site` (IP, secrets) | `google_compute_address`; in `cluster/`: `kubernetes_secret_v1` from `random_password` |
| `sandbox` | `cluster/`: vendored or kustomized agent-sandbox YAML with the controller-args patch |
| *(new)* | `google_compute_router` + `google_compute_router_nat`, `google_billing_budget` |

### The core of `gcp/`

```hcl
resource "google_compute_network" "main" {
  name                    = "cka-sim"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "gke" {
  name                     = "gke"
  region                   = var.region
  network                  = google_compute_network.main.id
  ip_cidr_range            = "10.10.0.0/20"
  private_ip_google_access = true # private nodes reach Artifact Registry and Google APIs without NAT

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.20.0.0/16"
  }
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.30.0.0/20"
  }
  # Lab pods get their own range so the firewall can tell them from GKE's system pods (konnectivity).
  secondary_ip_range {
    range_name    = "labs-pods"
    ip_cidr_range = "10.80.0.0/16"
  }
}

resource "google_compute_router" "nat" {
  name    = "cka-sim-nat"
  region  = var.region
  network = google_compute_network.main.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "cka-sim-nat"
  router                             = google_compute_router.nat.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_container_cluster" "main" {
  name       = "cka-sim"
  location   = var.zone
  network    = google_compute_network.main.id
  subnetwork = google_compute_subnetwork.gke.id

  remove_default_node_pool = true
  initial_node_count       = 1
  deletion_protection      = false # a spike: `terraform destroy` must work

  datapath_provider = "ADVANCED_DATAPATH"
  workload_identity_config {
    workload_pool = "${var.project}.svc.id.goog"
  }
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = var.admin_cidr
      display_name = "admin"
    }
  }
  cluster_autoscaling {
    autoscaling_profile = "OPTIMIZE_UTILIZATION" # removes empty (used) lab nodes sooner, see §3
  }
}

resource "google_container_node_pool" "labs" {
  name               = "labs"
  cluster            = google_container_cluster.main.id
  location           = var.zone
  initial_node_count = 0

  autoscaling {
    min_node_count = 0
    max_node_count = var.max_labs # also your cost cap
  }
  network_config {
    pod_range = "labs-pods"
  }
  node_config {
    machine_type    = "e2-standard-2"
    image_type      = "UBUNTU_CONTAINERD"
    disk_size_gb    = 50
    tags            = ["cka-sim-lab"]
    labels          = { pool = "labs" }
    service_account = google_service_account.lab_nodes.email # §2
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    taint {
      key    = "lab"
      value  = "true"
      effect = "NO_SCHEDULE"
    }
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }
}

resource "google_compute_firewall" "labs_apart" {
  name          = "cka-sim-labs-apart"
  network       = google_compute_network.main.name
  direction     = "INGRESS"
  priority      = 900 # above GKE's own allow rules (1000)
  target_tags   = ["cka-sim-lab"]
  source_tags   = ["cka-sim-lab"]
  source_ranges = ["10.80.0.0/16"]
  deny {
    protocol = "all"
  }
}
```

The `system` node pool follows the same shape (e2-medium, 1–2 nodes, no taint, its own service account).

`master_authorized_networks_config` blocks GitHub-hosted runners, whose IPs change. When CI applies
`cluster/`, use GKE's DNS-based control plane endpoint or a runner inside the VPC.

### Test it

1. **Parity.** Apply into a new project (`project = "cka-sim-tf-<you>"`) and run the HOSTING.md checks:
   claim a warm lab, run `wl-scale` setup/check, `ssh` to a node from the terminal, End lab.
2. **Idempotence.** Run `terraform plan -detailed-exitcode` right after `apply`. Exit code 0 means no diff.
   Exit code 2 means something the provider doesn't round-trip, and that would show up as fake drift
   forever.
3. **Drift.** Change something by hand (`gcloud container node-pools update labs ... --max-nodes 3`), then
   run `plan` again. Terraform should report the drift and put it back.
4. **Teardown.** `terraform destroy`, then `gcloud compute instances list`, `gcloud compute addresses list`
   and `gcloud sql instances list` should all come back empty. Leaked resources are the ones that cost money.

You learn the difference between declarative state and a script: a plan is a diff you can review in a PR.

## 2. The lab nodes run as the Compute Engine default service account

`infra/gke/Makefile` `cluster` creates the labs pool without `--service-account`, so its nodes use the
project's default compute SA. HOSTING.md already says to give them an account with no roles. That still needs
doing. A student who escapes their privileged pod is root on the node and can fetch that SA's token from the
metadata server on the host network. VPC firewall rules don't apply to metadata server traffic, so IAM is the
only real wall here.

### Change

- Create a `cka-sim-lab-nodes` SA with `roles/container.defaultNodeServiceAccount` (logs and metrics), plus
  `roles/artifactregistry.reader` **on the labs repository only**. The node pulls the lab image with this SA,
  so "no roles at all" would break image pulls.
- Split Artifact Registry into a `labs` repo and a `platform` repo, so a stolen lab-node token can't pull the
  app image.

```hcl
resource "google_service_account" "lab_nodes" {
  account_id = "cka-sim-lab-nodes"
}

resource "google_project_iam_member" "lab_nodes_default" {
  project = var.project
  role    = "roles/container.defaultNodeServiceAccount"
  member  = "serviceAccount:${google_service_account.lab_nodes.email}"
}

resource "google_artifact_registry_repository_iam_member" "lab_nodes_pull" {
  project    = var.project
  location   = var.region
  repository = google_artifact_registry_repository.labs.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.lab_nodes.email}"
}
```

### Test it

This also answers HOSTING.md's open question, "What a lab can get from the metadata server".

```sh
# Today: which identity do lab nodes run as?
gcloud container node-pools describe labs --cluster cka-sim --zone southamerica-east1-a \
  --format='value(config.serviceAccount)'                    # "default" before the fix

# What an escaped student gets: run on the lab node itself (IAP ssh once nodes are private)
gcloud compute ssh <lab-node> --zone southamerica-east1-a --tunnel-through-iap -- \
  'curl -s -H "Metadata-Flavor: Google" \
   http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/email'
T=$(gcloud compute ssh <lab-node> ... -- 'curl -s -H "Metadata-Flavor: Google" \
   http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token' | jq -r .access_token)
curl -s -H "Authorization: Bearer $T" \
  "https://cloudresourcemanager.googleapis.com/v1/projects/cka-sim-gke-maljoke:getIamPolicy" -X POST
# Expect 403 after the fix. Also try listing Cloud SQL instances and the platform repo: both 403.

# What the pod (not the node) gets: Workload Identity's identity, which should have no IAM grants.
kubectl exec <lab-pod> -- sh -c 'wget -qO- --header "Metadata-Flavor: Google" \
  http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/email'
```

You learn the difference between the pod's identity (Workload Identity) and the node's identity, and that
`privileged` hands the student the node's.

## 3. A used node can host the next student

HOSTING.md, "Node reuse": when a claim ends, agent-sandbox refills the warm pool, and the new lab can land on
the node the last student was root on. That student could have left anything on the host: a process, a
modified kubelet, files on disk.

### Change

In `apps/platform/internal/sandbox/gke`, `Destroy` taints the node **before** it deletes the claim. The
autoscaler then removes the empty node, and the refill lands on a fresh one.

```go
func (p *Provider) Destroy(ctx context.Context, b sandbox.Box) error {
    c, err := p.kube.getClaim(ctx, b.ID)
    if err != nil {
        return err
    }
    node, err := p.kube.podNode(ctx, c.Status.Sandbox.Name) // pod.spec.nodeName
    if err != nil {
        return err
    }
    // Taint first: in the gap after the delete, the pool's refill could land on this node.
    if err := p.kube.taintUsed(ctx, node); err != nil {
        return err
    }
    return p.kube.deleteClaim(ctx, b.ID)
}
```

- `taintUsed`: `PATCH /api/v1/nodes/<name>` with `Content-Type: application/json-patch+json` and
  `[{"op":"add","path":"/spec/taints/-","value":{"key":"cka-sim/used","value":"true","effect":"NoSchedule"}}]`.
  Appending works because lab nodes always carry `lab=true` already. `kube.do` currently hard-codes
  `application/json`, so it needs a content-type parameter.
- RBAC: add a `ClusterRole` with `nodes: [get, patch]` for `cka-sim-app`. Nodes are cluster-scoped, so the
  current `Role` can't grant it.
- `ReapClaims` at startup must taint the nodes of the claims it deletes too.
- Capacity: a used node lingers until scale-down (about 10 minutes by default, sooner with
  `OPTIMIZE_UTILIZATION`). Size `max_labs` for that.

### Test it

```sh
kubectl get pods -l app=cka-sim-lab -o wide              # note NODE of the claimed lab
# Prove the student owns the node: write a marker on the HOST disk from the privileged pod
kubectl exec <lab-pod> -- sh -c 'mkdir -p /mnt/h && mount /dev/sda1 /mnt/h && touch /mnt/h/var/cka-sim-marker'
# End lab in the UI, then:
kubectl get node <node> -o jsonpath='{.spec.taints}'     # contains cka-sim/used
kubectl get pods -l app=cka-sim-lab -o wide -w           # the refill lands on a DIFFERENT node
kubectl get nodes -l pool=labs -w                        # the used node disappears after scale-down
```

Before the fix, run the same steps with `POOL=1` and `max-nodes 1`. The refill lands on the same node and
the marker is still there. That's the bug, made visible. Add a `gke_test.go` case asserting that the taint
call happens before the delete; the fake `kube` already exists there.

You learn that the autoscaler deletes nodes that hold only DaemonSets, and why ordering matters in
cleanup code.

## 4. Public nodes and a public control plane

HOSTING.md, "Not checked yet": nodes have public IPs, the default network allows SSH from anywhere, and the
GKE API's public endpoint is reachable from a lab.

### Change

This is already in the §1 snippet: `enable_private_nodes`, a dedicated VPC (no `default-allow-ssh`), Cloud
NAT for egress, `master_authorized_networks_config` restricted to your IP, and an IAP SSH rule if you want
node access:

```hcl
resource "google_compute_firewall" "iap_ssh" {
  name          = "cka-sim-iap-ssh"
  network       = google_compute_network.main.name
  direction     = "INGRESS"
  source_ranges = ["35.235.240.0/20"] # Google's IAP range
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}
```

The student's node still has to reach the control plane, because its kubelet does. What protects the API
from an escaped student is the Node authorizer: a kubelet's credentials only read objects bound to its own
node. That includes the lab pod's own Secret, which is why §5 matters.

### Test it

```sh
kubectl get nodes -o wide                                  # EXTERNAL-IP is <none>
kubectl exec <lab-pod> -- wget -qO- -T5 https://<public-endpoint>/version   # times out from the lab
curl -k -m5 https://<public-endpoint>/version              # works from your IP only
kubectl exec <lab-pod> -- wget -qO- -T5 https://example.com # still works, through Cloud NAT
```

You learn private clusters, why NAT becomes mandatory with them, and what Node authorization limits.

## 5. One agent token is shared by every lab

`cka-sim-agent` is a single Secret. Every lab pod gets it as an env var, and the app uses it to talk to
every lab. An escaped student can read it from `/proc/1/environ` of the lab container or through the
kubelet's credentials. The VPC firewall is the only thing that stops them using it against another lab, so
one firewall gap means control of every lab.

### Change

Make the secret that lives in the lab worthless:

- The app holds an Ed25519 private key. Labs get only the public key, from a ConfigMap.
- For each call, the app signs a short token: `aud` set to the sandbox name, `exp` about 60 seconds away.
- The agent (`apps/lab/internal/agent.checkToken`) checks the signature, checks `exp`, and checks that `aud`
  equals its own pod name (from the downward API, `POD_NAME`).

### Test it

- Unit tests in `apps/lab/internal/agent`: a wrong `aud`, an expired token, a bad signature and a missing
  header each give 401. This is intricate logic, so it's worth isolating.
- End to end: from the app pod, mint a token for lab A and call lab B's `/run` with it. Expect 401. Read
  the env of a lab pod and confirm it holds no private key.

You learn audience-bound, short-lived credentials as opposed to shared bearer secrets.

## 6. Every app restart ends every lab and exam

- `apps/platform/cmd/server/main.go`: `ReapClaims` deletes **all** `cka-sim` claims at startup.
- `desk.Desks` keeps account → lab in memory.
- `apps/platform/deploy/gke/app.yaml` uses `strategy: Recreate`, and `make site` always runs
  `rollout restart`.
- `exam.DefaultDir()` writes to `~/.local/state/cka-sim/exams` inside the container. That isn't on the PVC
  (the PVC is mounted only at Caddy's data dir).

So every deploy ends every student's lab, and running exams too.

### Change

- **Reattach instead of reap.** Claims already carry `cka-sim/account=<id>`. At startup, list them, rebuild
  each account's `Lab` in `Ready` state with that claim's box, and only delete claims past their max
  lifetime. Kubernetes is the source of truth for "which lab is whose".
- **Exams in Postgres.** The app already has a database. Move `FileStore` behind the same interface.
- **A hard max lifetime.** The idle reaper skips labs with a running exam (`desk.idle`), and there's no cap
  otherwise. Add `CKA_SIM_LAB_MAX` (for example 3h) and check it against the claim's `creationTimestamp`, so
  it still holds after a restart.

### Test it

```sh
# Start a lab and an exam, answer one task, then:
kubectl get sandboxclaims -l app.kubernetes.io/managed-by=cka-sim   # note the name
kubectl rollout restart deployment/app && kubectl rollout status deployment/app
kubectl get sandboxclaims -l app.kubernetes.io/managed-by=cka-sim   # same name, not deleted
# Reload the page: same lab, exam timer still counting, the answered task still answered.

# Max lifetime: deploy with CKA_SIM_LAB_MAX=5m, start a lab, and watch the claim disappear.
kubectl get sandboxclaims -w
```

Add a `gke_test.go` case for the reattach logic (claims with labels → boxes), using the existing fake
`kube`.

You learn the difference between stateless and stateful servers, and how to reconcile against the real
world on startup, which is what controllers do.

## 7. Warm labs take 2–4 minutes because they build at boot

`lab-start.sh` → `lab up` builds the node image and creates kind on every new pod (HOSTING.md: 2–4 minutes,
plus about 3.5 minutes for a new node). With private nodes, every pull from Docker Hub also leaves through
the same Cloud NAT IPs, and Docker Hub's anonymous pull limit is per IP.

### Change

- Build the kind node image (`apps/lab/internal/cluster/node.Dockerfile`) in CI and push it to Artifact
  Registry. `lab up` then pulls it instead of building it.
- Add Artifact Registry **remote repositories** for `docker.io` and `registry.k8s.io`, and point the lab's
  dockerd (`registry-mirrors`) and the kind nodes' containerd at them. That gives you no rate limits and a
  cache close to the nodes.
- Use **immutable image tags**. Today the pool runs `lab:latest`, so a node can pull a different image than
  the one the template was written for. Have CI substitute `lab@sha256:...` into `lab-pool.yaml`.

### Test it

```sh
kubectl get sandboxtemplate lab -o jsonpath='{..image}'     # a digest, not :latest
# Time it: node Ready → pod Ready, before and after
kubectl get events --field-selector involvedObject.kind=Pod -w
kubectl exec <lab-pod> -- docker info | grep -A2 'Registry Mirrors'
```

You learn image supply chains, pull-through caches, and where a cold start's time actually goes.

## 8. CI never runs a real lab

`.github/workflows/ci.yml` runs `go vet` and `go test`. `make selftest`, which proves that every task scores
0 before its solution and full marks after, only runs by hand.

### Change

- **Nightly selftest:** an `ubuntu-latest` job running `make up && make selftest`. GitHub's Linux runners
  allow privileged Docker, so this runs the lab exactly as a laptop does.
- **Terraform in PRs:** `terraform fmt -check`, `validate`, and `plan` posted to the PR. Authenticate with
  Workload Identity Federation (`google-github-actions/auth` with a `workload_identity_provider`), so there
  is no SA key in GitHub secrets.
- **Deploy from CI:** move `make site`'s builds and `kubectl apply` into a workflow on `main`, and push to
  Artifact Registry with the same WIF identity.

### Test it

- Break a task's `check.sh` on a branch. The nightly job, or a manual `workflow_dispatch`, must fail.
- Open a PR that changes `max_labs`. The plan comment should show exactly one in-place update.

## 9. Lab egress is open to the internet

The template allows `0.0.0.0/0` minus private ranges. That's fine for image pulls, but it also allows
mining, scanning and spam from your project.

Do this after §7: once images come from Artifact Registry over Private Google Access, a lab only needs a few
destinations. Those are the public site (`CKA_SIM_MCP_URL` comes back through it) and the Claude API for
the assistant. Restricting by domain name (FQDN) needs more than VPC firewall rules (they match IPs). Look at
Secure Web Proxy or a small egress proxy before you choose one.

### Test it

```sh
kubectl exec <lab-pod> -- wget -qO- -T5 https://example.com          # blocked
kubectl exec <lab-pod> -- wget -qO- -T5 https://api.anthropic.com    # allowed
# And from inside a kind node, which is the student's real position:
kubectl exec <lab-pod> -- docker exec cka-sim-control-plane curl -m5 https://example.com
```

## Order

1. §1, Terraform at parity with the Makefile, no behaviour change. Prove it with the parity and idempotence
   tests.
2. §2 node SA and §3 node reuse. These are the two real isolation gaps.
3. §4 private nodes + NAT, then §7 mirrors and prebuilt images. NAT makes the rate limit urgent.
4. §6 restarts and max lifetime. Deploys stop hurting students.
5. §5 per-lab credentials, §8 CI, §9 egress.

Also add a `google_billing_budget` with alerts at 50% and 90% of the credit (the currency must match the
billing account's). `max_labs` limits node count, and the budget alert catches everything else.
