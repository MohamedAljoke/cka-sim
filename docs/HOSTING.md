# Hosting: labs on demand on Kubernetes

Locally, the lab is one container that is always there. Hosted, each student gets a lab of their own: the
platform creates it as a pod when they click **Start lab**, on a node of its own that the cluster adds for it,
and deletes it when they stop. The lab image, the agent and the platform's use cases are the same in both
places. Only the code that brings the machine up differs.

> [!NOTE]
> Students are a fixed list of names, each with a token their invite link carries; there are no accounts or
> sign-ups. An IP allowlist in front keeps everyone else out.

**Contents:** [1. A lab is a lab anywhere](#1-a-lab-is-a-lab-anywhere) · [2. Start lab, step by step](#2-start-lab-step-by-step) ·
[3. The lab pod](#3-the-lab-pod) · [4. Students and their links](#4-students-and-their-links) ·
[5. Who can reach the lab](#5-who-can-reach-the-lab) ·
[6. The cluster: Terraform](#6-the-cluster-terraform) · [7. The app: kustomize and Caddy](#7-the-app-kustomize-and-caddy) ·
[8. Run it](#8-run-it)

## 1. A lab is a lab anywhere

The platform's lab use cases need two things: a `LabProvider` (Get, Start, Stop) and a `ScriptRunner`.
`labagent.Lab` (`apps/platform/internal/infra/adapters/labagent/lab.go`) provides both, and it is the only code
that talks to an agent. What it doesn't know is where the agent runs. It asks a `Machine` for that:

```go
type Machine interface {
	Up(ctx context.Context) error                       // make sure the machine exists
	Down(ctx context.Context) (removed bool, err error) // remove it, cluster and all
	Where(ctx context.Context) (Place, error)           // the agent's URL, or why there is none yet
}
```

| Machine | Up | Down | Where |
| --- | --- | --- | --- |
| `labagent.Fixed` (local) | nothing: the container is always there | nothing, so Stop asks the agent to delete its cluster | the one URL from `CKA_SIM_LAB_URL` |
| `kube.Pod` (hosted) | creates the pod | deletes the pod, taking the cluster with it | the pod's IP once it's Ready; until then a Starting lab with a step, or Failed |

**A lab per student.** Every lab use case takes the owner, the student's name. `labagent.Fleet`
(`labagent/fleet.go`) makes each owner's `Lab` the first time it's asked for one, on the machine `machineOf(owner)`
gives:

| `CKA_SIM_LABS` | `machineOf(owner)` |
| --- | --- |
| `local` | the same `Fixed` container for everyone; locally the only owner is `local` anyway |
| `kubernetes` | a `kube.Pod` named `lab-<owner>`, labelled `cka-sim/owner=<owner>` |

The owner also keys the rest of a student's work: the touched tasks and the script lock, and the exam, saved as
`<state>/exams/<owner>.json`. `cmd/server/main.go` wires it. Nothing in `kube` is tied to GKE, so EKS or any
other cluster would be a new Terraform folder, not new Go code.

## 2. Start lab, step by step

```mermaid
sequenceDiagram
    participant B as Browser
    participant P as Platform
    participant K as Kubernetes API
    participant N as Autoscaler
    participant A as Agent (lab pod)

    B->>P: POST /api/lab, with alice's cookie
    P->>K: create pod "lab-alice" in cka-sim-labs
    P-->>B: Starting · "Waiting for a node"
    N->>K: the labs pool grows from 0 to 1 node
    loop the page polls GET /api/lab
        P->>K: get pod "lab-alice"
        P-->>B: "Pulling the lab image", then "Starting Docker and the agent"
    end
    Note over K,A: the readiness probe on /healthz passes
    P->>A: GetCluster: none
    P->>A: CreateCluster, since the pod exists only to hold this lab
    A-->>P: Creating · kind's live build steps
    B->>P: WebSocket /ws/terminal, once Ready
    P->>A: OpenShell stream
```

The page's state comes from two places:
- **While the pod is not ready**, from the pod itself: `kube.stateOf` reads its conditions and phase.
- **Once it is Ready**, from the agent, as it does locally.

| The pod | The lab |
| --- | --- |
| not found | None |
| not scheduled | Starting: "Waiting for a node" |
| scheduled, Pending | Starting: "Pulling the lab image" |
| running, not Ready | Starting: "Starting Docker and the agent" |
| Ready | whatever the agent says; an agent with no cluster is told to build one |
| `ErrImagePull`, `ImagePullBackOff`, `CreateContainer*Error` | Failed, with the reason: these don't fix themselves |
| Failed or Succeeded | Failed: "the lab pod stopped" |
| being deleted | Stopping, and Start refuses until it has gone |

**Stop lab** deletes that student's pod. The kind cluster lived in the pod's emptyDir, so it goes too, and the
cluster autoscaler removes the empty node about 10 minutes later. Other students' labs are untouched.

## 3. The lab pod

`kube.PodSettings.spec()` (`apps/platform/internal/infra/adapters/kube/pod.go`) is the Compose file's lab
container, written as a pod:

| Field | Why |
| --- | --- |
| `privileged: true` | dockerd and kind's nodes make cgroups, mounts and networks |
| `emptyDir` at `/var/lib/docker` | The inner Docker's storage; it can't be the node's overlayfs |
| `hostPath /lib/modules`, read-only | kube-proxy and the CNI inside kind load kernel modules |
| `nodeSelector pool=labs` and a toleration for `lab=true:NoSchedule` | Only the lab lands on the labs pool, and it lands nowhere else |
| requests of 2 CPU and 8 Gi | Takes a whole `e2-standard-4`, so each lab is alone on its node |
| `restartPolicy: Never` | A crashed lab shows as Failed, and Start makes a fresh one |
| token from Secret `cka-sim-agent` | The agent accepts only the app's bearer token, and it never sits in a manifest |
| no ServiceAccount token, no service links | Nothing in the lab can talk to the hosting cluster's API |
| readiness probe on `/healthz` | "Ready" means the agent answers |

The labs pool runs `UBUNTU_CONTAINERD` nodes, because kind needs a regular Linux's kernel modules and cgroups,
which Container-Optimized OS doesn't give it.

## 4. Students and their links

Two tokens guard two different hops, and the browser only ever holds the first:

| Hop | Token | Where it lives |
| --- | --- | --- |
| Browser → platform | the student's own, from their link | the `cka_sim_student` cookie |
| Platform → agent | the agent token, one for every lab | Secret `cka-sim-agent`; the platform sends it as `Authorization: Bearer` |

**Students** come from `CKA_SIM_STUDENTS`, `alice=<token>,bob=<token>`: a name of up to 30 lowercase letters,
digits and dashes, since it becomes the pod's name, and a token of at least 32 characters. The middleware in
`apps/platform/internal/api/students.go`:

- `GET /api/join?token=…` checks the token, keeps it in an `HttpOnly`, `SameSite=Lax` cookie (`Secure` over
  HTTPS) for 400 days, and redirects to `/`, so the token leaves the address bar;
- every other request is let through only with a cookie it knows, as that student; without one it answers 401
  `student_unknown`, and the page asks for the student's link;
- with `CKA_SIM_STUDENTS` empty, as on `cka-sim up`, every request is the student `local`.

Revoking a student means editing the list and restarting the app.

## 5. Who can reach the lab

Inside the lab, root on a node can break out to the privileged pod, and from there to its node. So the node is
the boundary, and the rules are about who can reach it and what it can reach:

- **The app's RBAC** (`infra/gke/k8s/lab.yaml`): ServiceAccount `cka-sim-app` gets a Role with get, create and
  delete on pods, in `cka-sim-labs` only. It can't touch anything else in the cluster.
- **NetworkPolicy `lab`**, enforced by Dataplane v2:
  - Only pods labelled `app: cka-sim-app` in `cka-sim` reach port 7171, so one student's lab can't reach
    another's.
  - The lab can reach DNS and the internet, for the images its cluster pulls, but no private range and not
    `169.254.0.0/16`, so not the metadata server either.
- **The node itself:** the labs pool is tainted and the pod takes a whole node, so no other workload shares it.

## 6. The cluster: Terraform

`infra/gke/terraform/` makes what lives in GCP:

| Resource | What |
| --- | --- |
| `google_container_cluster.cka_sim` | Zonal, `southamerica-east1-a`, Dataplane v2 (`ADVANCED_DATAPATH`), the default pool removed |
| `google_container_node_pool.system` | `e2-medium`, 1–2 nodes: the app and GKE's own pods |
| `google_container_node_pool.labs` | `e2-standard-4` on Ubuntu, **0 to `var.max_labs` nodes** (default 2), label `pool=labs`, taint `lab=true:NoSchedule` |
| `google_compute_address.site` | The static IP the site answers on |

A labs pool that scales from zero means an idle site pays for no lab node. The price is a cold start: a new
node, the image pull, then the kind build, a few minutes in all. Each lab takes a whole node, so `max_labs` is
also how many students can have a lab at once; past it, a new lab waits for a node.

## 7. The app: kustomize and Caddy

`infra/gke/k8s/` is applied with `kubectl apply -k`:

- **`app.yaml`:**
  - the `app` Deployment, the same `cka-sim-app` image as `cka-sim up`, with `CKA_SIM_LABS=kubernetes`, the
    agent token from Secret `cka-sim-agent` and `CKA_SIM_STUDENTS` from Secret `cka-sim-students`. It pulls
    `:main` on every start (`imagePullPolicy: Always`), since that tag moves with every push;
  - a PVC for the saved exam and Caddy's certificates;
  - a LoadBalancer Service on the static IP.
- **`lab.yaml`:** the RBAC and the NetworkPolicy above.
- **`kustomization.yaml`:** builds two ConfigMaps:
  - `site`, from `site.env`, which `make site` writes from Terraform's output and your address;
  - `caddy`, from the `Caddyfile`.

  Their names carry a hash of their content, so a change restarts the app. A replacement copies `SITE_IP` into
  the Service's `loadBalancerIP`.

**The allowlist lives in Caddy, not on the load balancer.** Caddy serves `https://<ip>.sslip.io` with a Let's
Encrypt certificate. Its `route` drops any client outside `CKA_SIM_ALLOW` before anything is served. Let's
Encrypt still gets through, because Caddy answers its challenges before any route runs. For Caddy to see your
real address, the Service has `externalTrafficPolicy: Local`, so the load balancer doesn't hide the source
address behind a node's.

## 8. Run it

Every command runs against GCP project `cka-sim-gke-maljoke` and the context Terraform prints, never the current
kubectl context.

```sh
make -C infra/gke up                    # the cluster, then the site; about 10 minutes, then prints the URL
make -C infra/gke links                 # each student's link; whoever opens one is that student
make -C infra/gke labs                  # each student's lab pod and its node, then the labs pool's nodes
make -C infra/gke allow                 # your IP changed: let the new one in
make -C infra/gke lab-down NAME=alice   # delete alice's lab pod, as their Stop lab does
make -C infra/gke site-down             # remove the app, the labs and the load balancer; the cluster stays
make -C infra/gke down                  # delete the cluster and the IP too
```

The first `make site` makes the students in `STUDENTS` (default `alice bob`), each with a random token, into
Secret `cka-sim-students`, and the agent token into `cka-sim-agent`. Both are kept after that: to change the
students, delete the Secret and run `make site STUDENTS="…"` again.

**Cost**, roughly: with no lab running, about US$1.5–2 a day, for the system node, the load balancer and the
disks. Each running lab node adds about US$0.20 an hour, until about 10 minutes after its Stop lab.
