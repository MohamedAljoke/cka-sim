# CKA question bank

These are general practice questions, to be turned into cka-sim tasks later (`task.md`, `setup.sh`,
`check.sh`, `solution.sh`, `explain.md`). Each one names the skill it tests, not the final names,
namespaces or paths. Those are chosen when the question becomes a task.

## Sources

Checked on 2026-10-09.

- **Official curriculum:** [CNCF CKA Curriculum v1.35](https://github.com/cncf/curriculum)
  (`CKA_Curriculum_v1.35.pdf`). It gives the domains, weights and competencies.
- **Exam page:** [Linux Foundation CKA](https://training.linuxfoundation.org/certification/certified-kubernetes-administrator-cka/).
  It gives the format, duration, Kubernetes version and simulator details.
- **Third-party study guides:** used only to cross-check the task count and the new topics (Helm,
  Kustomize, CRDs, Gateway API), not for official facts:
  - [certland.net: CKA Certified Kubernetes Administrator guide 2026](https://certland.net/blog/cka-certified-kubernetes-administrator-guide-2026/)
  - [examcert.app: CKA study guide 2026](https://www.examcert.app/blog/cka-study-guide-2026/)
- **Our own tasks:** the 8 tasks on branch `archive/v1` (`tasks/`) and `wl-scale` on `main`
  (`apps/cka-sim/catalog/`).
- **Common practice topics:** what shows up again and again in killer.sh-style mock exams.

## The exam, in short

| | |
|---|---|
| Format | Online, proctored, performance-based: solve tasks from a command line |
| Duration | 2 hours |
| Tasks | Not published. The killer.sh simulator has 17 per attempt; reports range 15–20 |
| Kubernetes | v1.35, tracking the latest minor version within about 4–8 weeks of its release |
| Allowed docs | kubernetes.io/docs and a few project docs; check the LF candidate handbook for the current list |

Things the exam environment does that our tasks should copy:
- Each task names the cluster or host to ssh into first.
- Answers are often written to a file (`/opt/course/<task>/...`).
- Many tasks say "do not modify X".
- Some fixes must survive a reboot.

## Domains and weights

| Weight | Domain | Official competencies |
|---|---|---|
| 30% | Troubleshooting | Troubleshoot clusters and nodes · Troubleshoot cluster components · Monitor cluster and application resource usage · Manage and evaluate container output streams · Troubleshoot services and networking |
| 25% | Cluster Architecture, Installation & Configuration | Manage RBAC · Prepare underlying infrastructure for installing a cluster · Create and manage clusters using kubeadm · Manage the lifecycle of clusters · Implement and configure a highly-available control plane · Use Helm and Kustomize to install cluster components · Understand extension interfaces (CNI, CSI, CRI, etc.) · Understand CRDs, install and configure operators |
| 20% | Services & Networking | Understand connectivity between Pods · Define and enforce NetworkPolicies · Use ClusterIP, NodePort, LoadBalancer Service types and endpoints · Use the Gateway API to manage Ingress traffic · Know how to use Ingress controllers and Ingress resources · Understand and use CoreDNS |
| 15% | Workloads & Scheduling | Understand Deployments, rolling updates and rollbacks · Use ConfigMaps and Secrets to configure applications · Configure workload autoscaling · Understand the primitives for robust, self-healing deployments · Configure Pod admission and scheduling (limits, node affinity, etc.) |
| 10% | Storage | Implement StorageClasses and dynamic provisioning · Configure volume types, access modes and reclaim policies · Manage PersistentVolumes and PersistentVolumeClaims |

## How to read the lists

- **Covered** shows where a task already exists: `main` means `apps/cka-sim/catalog/<id>`, and `v1`
  means `archive/v1:tasks/<domain>/<id>`.
- **Needs** lists what the cluster must have beyond a plain 1-control-plane, 1-worker kind cluster.
  This decides when we can build the task.
- **Level:** ★ quick (under 5 min), ★★ normal, ★★★ long or multi-step.

---

## 1. Troubleshooting (30%)

### Clusters and nodes

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T1 | A worker node is `NotReady`. Find out why and fix it so that it survives a reboot. Example causes: kubelet stopped or disabled, wrong flag in the kubelet config, bad CA path. | ★★ | systemd kubelet on node | v1 `tr-kubelet` |
| T2 | A node is `NotReady` because the container runtime is down or its socket path is wrong in the kubelet config. Fix it. | ★★ | containerd on node | |
| T3 | A node shows `DiskPressure` or `MemoryPressure` and Pods are being evicted. Find the cause and free the node. | ★★★ | fill a disk on node | |
| T4 | A node was cordoned and left `SchedulingDisabled`. Find which one and make it schedulable again without restarting anything. | ★ | | |
| T5 | Drain a node for maintenance (DaemonSets, emptyDir and unmanaged Pods present), then return it to service. | ★★ | 2+ workers | |
| T6 | The kubelet client certificate on a node has expired or is invalid. Get the node back. | ★★★ | cert tampering | |

### Cluster components

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T7 | New Pods stay `Pending` forever. The scheduler is broken through its static Pod manifest. Fix it without touching the Pods. | ★★ | static Pods | v1 `tr-scheduler` |
| T8 | `kubectl` cannot reach the API server. The kube-apiserver manifest has a bad flag, port, cert path or etcd endpoint. Fix it. | ★★★ | static Pods | |
| T9 | Deployments don't create Pods, or scaling has no effect. kube-controller-manager is broken. Fix it. | ★★ | static Pods | |
| T10 | Write how kube-scheduler, controller-manager, etcd and DNS are run in this cluster (static Pod, Deployment, systemd process) to a file. | ★ | | |
| T11 | Write the order of commands you would use to check a cluster's health to a file (`crictl ps`, `journalctl -u kubelet`, `/var/log/pods`). | ★ | | |

### Resource usage

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T12 | Find the Pod using the most CPU (or memory) in a namespace or across the cluster, and write its name to a file. | ★ | metrics-server | |
| T13 | Write a script or command that shows node resource usage sorted by memory. | ★ | metrics-server | |
| T14 | Write all Pods sorted by creation time or by restart count to a file (`--sort-by`, `custom-columns`). | ★ | | |

### Container output streams (logs)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T15 | A Pod is in `CrashLoopBackOff`. Read the logs of the **previous** container and fix the cause (wrong command, missing env, missing ConfigMap key). | ★★ | | |
| T16 | Save the logs of one container in a multi-container Pod to a file, and only the lines that match an error. | ★ | | |
| T17 | Add a sidecar container that streams a log file the main container writes to an emptyDir, so `kubectl logs` shows it. | ★★ | | |
| T18 | A control-plane component's container is gone from `kubectl`. Find its logs with `crictl logs` or `/var/log/pods` on the node. | ★★ | | |
| T19 | Write the events for a failing Pod, sorted by time, to a file. | ★ | | |

### Services and networking

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T20 | A Service times out. Its selector or `targetPort` doesn't match the Pods. Fix the Service, not the Deployment. | ★★ | | v1 `tr-service` |
| T21 | DNS lookups fail inside Pods. CoreDNS is scaled to 0, its ConfigMap is broken, or kube-dns has no endpoints. Fix it. | ★★ | | |
| T22 | Pods on one node cannot reach Pods on another. The CNI config is missing or broken on that node. Fix it. | ★★★ | CNI on node | |
| T23 | kube-proxy is not running, or uses the wrong mode, so ClusterIPs don't route. Fix it. | ★★ | | |
| T24 | An Ingress or HTTPRoute returns 404/503. Find the wrong backend name, port or path and fix it. | ★★ | ingress/gateway controller | |

---

## 2. Cluster Architecture, Installation & Configuration (25%)

### RBAC

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A1 | Create a ServiceAccount, a Role and a RoleBinding with least privilege, then check with `kubectl auth can-i --as=system:serviceaccount:...`. | ★★ | | v1 `ar-rbac` |
| A2 | Create a ClusterRole that can only `list` and `get` nodes and PersistentVolumes, and bind it to a user or group. | ★ | | |
| A3 | Bind a **ClusterRole** with a **RoleBinding** so the subject gets its rights in one namespace only. Explain the difference in a file. | ★★ | | |
| A4 | Create a new user from a CSR (key → CSR object → approve → kubeconfig entry), then give that user read-only access to one namespace. | ★★★ | | |
| A5 | A Pod's ServiceAccount gets `forbidden` when it lists Pods. Find the missing rule and add it. | ★★ | | |

### Infrastructure and kubeadm

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A6 | Prepare a node for kubeadm: load kernel modules, set sysctls (`ip_forward`, `bridge-nf-call-iptables`), disable swap and install the container runtime. | ★★ | bare node | |
| A7 | Join a new worker node to the cluster: create the token and CA hash on the control plane, then run `kubeadm join`. | ★★★ | extra node not joined | |
| A8 | Upgrade the control plane by one minor version with kubeadm (`plan`, `apply`, then kubelet/kubectl), then upgrade a worker. | ★★★ | versioned packages | |
| A9 | Check which kubeadm certificates expire soonest and renew them. Write their expiry dates to a file. | ★★ | | |
| A10 | Find the kubelet config file, the static Pod path and the cgroup driver on a node, and write them to a file. | ★ | | |

### Cluster lifecycle and etcd

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A11 | Back up etcd with `etcdctl snapshot save` (endpoints and certs), then restore from an older snapshot and get the cluster healthy. | ★★★ | etcd static Pod | v1 `ar-etcd` |
| A12 | Find etcd's version, data dir and member list, and write them to a file. | ★ | | |

### Highly-available control plane

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A13 | List the control-plane nodes and the etcd members in an HA cluster. Find the one etcd member that is unhealthy. | ★★ | 3 control planes | |
| A14 | Explain in a file what `--control-plane-endpoint` is for, and which component load-balances the API servers here. | ★ | HA cluster | |

### Helm and Kustomize

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A15 | Add a Helm repo and install a chart in a namespace with overridden values (`--set` or a values file). Then upgrade it, then roll it back. | ★★ | helm + local chart/repo | |
| A16 | Find a broken Helm release (`helm ls -a`, `pending-install`/`failed`) and delete it. Write the chart's default values to a file. | ★★ | helm | |
| A17 | Render a chart with `helm template` and install only the manifests it creates, without Helm managing them. | ★★ | helm | |
| A18 | Use a Kustomize overlay to change the replicas, image tag and a label on a base, and apply it with `kubectl apply -k`. | ★★ | | |
| A19 | Create a ConfigMap and a Secret from files with a Kustomize generator, and roll them into a Deployment. | ★★ | | |

### Extension interfaces (CNI, CSI, CRI)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A20 | Find which CNI plugin is installed, where its config lives (`/etc/cni/net.d`) and the Pod CIDR, and write them to a file. | ★ | | |
| A21 | Find the container runtime and its socket, and list the running containers on a node with `crictl`. | ★ | | |
| A22 | Install a CNI plugin from a given manifest on a cluster that has none, so the nodes become `Ready`. | ★★★ | cluster without CNI | |
| A23 | List the CSI drivers in the cluster and write which StorageClass uses which provisioner to a file. | ★ | CSI driver | |

### CRDs and operators

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A24 | Install a CRD from a manifest and create a custom resource of that kind. Write its `kubectl explain` fields to a file. | ★★ | | |
| A25 | Install an operator (manifest or Helm), then create its custom resource so the operator builds the workload. | ★★★ | operator image | |
| A26 | List the CRDs of a group (e.g. `*.cert-manager.io`) and the custom resources that exist for one of them. | ★ | CRDs installed | |

---

## 3. Services & Networking (20%)

### Pod connectivity

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N1 | From a temporary Pod, check whether Pod A can reach Pod B by IP and by Service DNS name. Write the results to a file. | ★ | | |
| N2 | Find a Pod's IP, its node and the node's Pod CIDR, and write them to a file. | ★ | | |

### NetworkPolicies

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N3 | Allow ingress to `app=api` only from one label in the same namespace and from any Pod in another namespace, on one port. | ★★ | policy-enforcing CNI | v1 `nw-netpol` |
| N4 | Default-deny all ingress and egress in a namespace, then allow egress to DNS (UDP/TCP 53) only. | ★★ | policy-enforcing CNI | |
| N5 | Allow egress from a backend to a database Pod on one port and to nothing else. | ★★ | policy-enforcing CNI | |
| N6 | Several NetworkPolicies are given. Pick the one that meets the requirement with least privilege and apply it. | ★★ | policy-enforcing CNI | |
| N7 | Fix a policy whose `namespaceSelector` and `podSelector` are in separate list items (OR) when they should be one item (AND). | ★★ | policy-enforcing CNI | |

### Service types and endpoints

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N8 | Expose a Deployment with a ClusterIP Service on a different port than the container port. | ★ | | |
| N9 | Expose a Deployment with a NodePort Service on a fixed nodePort, and reach it from the node. | ★ | | |
| N10 | Create a LoadBalancer Service and explain why it stays `<pending>` with no cloud provider. | ★★ | LB provider optional | |
| N11 | Create a Service with no selector and a manual EndpointSlice that points to an outside IP. | ★★ | | |
| N12 | A Service has no endpoints. Find out why (the Pods aren't Ready because of a failing readiness probe) and fix it. | ★★ | | |

### Gateway API

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N13 | Create a Gateway (with a given GatewayClass) and an HTTPRoute that sends `/api` to one Service and `/` to another. | ★★ | Gateway API CRDs + controller | |
| N14 | Split traffic between two backends with weights (canary 80/20) in an HTTPRoute. | ★★ | Gateway API | |
| N15 | Migrate an existing Ingress (host, paths, TLS) to a Gateway plus HTTPRoute, keeping the same behaviour. | ★★★ | Gateway API | |
| N16 | Add a header-based route match, or a request header modifier filter, to an HTTPRoute. | ★★ | Gateway API | |

### Ingress

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N17 | Create an Ingress with two host or path rules to two Services and a given `ingressClassName`, and test it with `curl`. | ★★ | ingress controller | |
| N18 | Add TLS to an Ingress using a Secret made from a given cert and key. | ★★ | ingress controller | |

### CoreDNS

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N19 | Write the FQDN of a Service, and of a Pod by its IP, to a file, and check both with `nslookup` from a Pod. | ★ | | |
| N20 | Change the CoreDNS config to add a custom domain or stub zone, or forward an upstream zone, then check it resolves. | ★★★ | | |
| N21 | Change the cluster domain the kubelet hands to Pods, or find it, and write it to a file. | ★ | | |

---

## 4. Workloads & Scheduling (15%)

### Deployments, rollouts and rollbacks

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W1 | Scale a Deployment and make sure all replicas are Ready. | ★ | | main `wl-scale` |
| W2 | Roll back a broken rollout to the last good revision, write the revision number to a file, and set `maxSurge`/`maxUnavailable`. | ★★ | | v1 `wl-rollout` |
| W3 | Update a Deployment's image and record the change cause. Pause the rollout, make two changes, then resume it. | ★★ | | |
| W4 | Change a Deployment to the `Recreate` strategy and explain when you would choose it, in a file. | ★ | | |

### ConfigMaps and Secrets

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W5 | Create a ConfigMap from literals and a file. Mount one key as a file at a given path, and expose another as an env var. | ★★ | | |
| W6 | Create a Secret and use it as env vars (`envFrom`) and as a read-only volume. | ★★ | | |
| W7 | Decode a value from an existing Secret and write it to a file. | ★ | | |
| W8 | Make a ConfigMap immutable, then change the value the app sees by creating a new ConfigMap and rolling the Deployment. | ★★ | | |

### Autoscaling

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W9 | Create an HPA (`autoscaling/v2`) for a Deployment: min/max replicas and a CPU utilization target. | ★★ | metrics-server | |
| W10 | Add a scale-down stabilization window or behaviour policy to an HPA. | ★★ | metrics-server | |
| W11 | An HPA shows `<unknown>` targets. Find out why (missing resource requests or metrics-server) and fix it. | ★★ | metrics-server | |
| W12 | Explain or set a VPA in recommendation-only mode, if the CRD exists. | ★★ | VPA CRDs | |

### Self-healing primitives

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W13 | Add liveness, readiness and startup probes (HTTP, TCP and exec) to a Deployment. | ★★ | | |
| W14 | Create a DaemonSet that runs on every node, including the control plane (tolerations). | ★★ | | |
| W15 | Create a StatefulSet with a headless Service and `volumeClaimTemplates`, and check the stable Pod names and PVCs. | ★★★ | | |
| W16 | Create a Job with `completions`, `parallelism` and `backoffLimit`, and a CronJob with a schedule and history limits. | ★★ | | |
| W17 | Add an init container that waits for a Service before the main container starts. | ★★ | | |
| W18 | Create a PodDisruptionBudget so a drain never takes a Deployment below N Pods. | ★★ | 2+ workers | |

### Admission and scheduling

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W19 | Set resource requests and limits on a Pod. Add a LimitRange and a ResourceQuota to a namespace, and show a Pod being rejected. | ★★ | | |
| W20 | Schedule a Pod only on nodes with a label, first with `nodeSelector`, then with required and preferred node affinity. | ★★ | | |
| W21 | Taint a node, then run a Pod that tolerates the taint and one that doesn't. | ★★ | | |
| W22 | Spread a Deployment's replicas across nodes with pod anti-affinity or `topologySpreadConstraints`. | ★★ | 2+ workers | |
| W23 | A Pod stays `Pending`. Read the events (insufficient CPU, an unmatched affinity, an untolerated taint) and fix the Pod spec. | ★★ | | |
| W24 | Run a Pod on a specific node by name without the scheduler (`nodeName`), or as a static Pod on a worker. | ★ | | |
| W25 | Create a PriorityClass and use it so a critical Pod evicts lower-priority Pods. | ★★ | | |
| W26 | Set a Pod `securityContext`: `runAsUser`, `readOnlyRootFilesystem` and dropped capabilities. | ★★ | | |

---

## 5. Storage (10%)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| S1 | Create a hostPath PV (capacity, access mode, `Retain`, storageClassName), a PVC that binds to it, and a Pod that writes to it. | ★★ | | v1 `st-pvc` |
| S2 | Create a StorageClass with a provisioner, `reclaimPolicy` and `volumeBindingMode: WaitForFirstConsumer`, and make it the default. | ★★ | dynamic provisioner | |
| S3 | Create a PVC with no PV and get it bound through dynamic provisioning. Check that the PV it creates gets deleted with the PVC. | ★★ | dynamic provisioner | |
| S4 | A PVC stays `Pending`. Find out why (wrong class, access mode, or too big a size) and fix the PVC without changing the PV. | ★★ | | |
| S5 | Expand a PVC (`allowVolumeExpansion`) and check the new size inside the Pod. | ★★ | expandable provisioner | |
| S6 | Find the PV a PVC is bound to and its reclaim policy. Change the policy from `Delete` to `Retain` so data survives deleting the PVC. | ★ | | |
| S7 | Reuse a `Released` PV: clear its `claimRef` so a new PVC can bind to it. | ★★ | | |
| S8 | Share data between two containers in a Pod with an `emptyDir` volume, then with a memory-backed one. | ★ | | |
| S9 | Explain in a file the difference between `ReadWriteOnce`, `ReadOnlyMany`, `ReadWriteMany` and `ReadWriteOncePod`. | ★ | | |

---

## Coverage so far

| Domain | Weight | Questions | Covered |
|---|---|---|---|
| Troubleshooting | 30% | 24 | 3 (v1) |
| Cluster Architecture | 25% | 26 | 2 (v1) |
| Services & Networking | 20% | 21 | 1 (v1) |
| Workloads & Scheduling | 15% | 26 | 1 (main) + 1 (v1) |
| Storage | 10% | 9 | 1 (v1) |

## What the cluster still needs

These are grouped by how many questions each one unblocks:

1. **2+ worker nodes:** T5, W18, W22, and many scheduling questions read better with them.
2. **metrics-server:** T12, T13, W9–W11.
3. **NetworkPolicy-enforcing CNI** (Calico or Cilium instead of kindnet): N3–N7.
4. **Gateway API CRDs + a controller** (e.g. Envoy Gateway or NGINX Gateway Fabric): N13–N16, T24.
5. **Ingress controller:** N17, N18.
6. **helm binary on the control-plane node, plus a local chart/repo:** A15–A17. The exam has no
   internet, so we can't rely on it either.
7. **Dynamic provisioner** (kind's `local-path` already does this; check its expansion support for S5).
8. **Spare unjoined node, and versioned kubeadm packages:** A7, A8. These are the hardest in kind;
   leave them last.
9. **3 control-plane nodes:** A13, A14.

## A suggested order for turning these into tasks

1. **Bring back the v1 tasks:** `tr-kubelet`, `tr-scheduler`, `tr-service`, `ar-etcd`, `ar-rbac`,
   `st-pvc`, `wl-rollout`, `nw-netpol`.
2. **Quick, no-extra-setup wins:** T4, T15, T19, T21, A2, A5, A10, A18, N8, N9, N12, W5, W6, W13,
   W19–W21, W23, S4, S6.
3. **High-weight troubleshooting:** T2, T8, T9, T22, T23.
4. **New curriculum topics:** Helm (A15), Kustomize (A18), CRDs/operators (A24–A25), Gateway API (N13–N15).
5. **The long ones:** kubeadm join and upgrade (A7, A8), HA (A13).
