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
  (`apps/platform/catalog/`).
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

- **Covered** shows where a task already exists: `main` means `apps/platform/catalog/<id>`, and `v1`
  means `archive/v1:tasks/<domain>/<id>`.
- **Needs** lists what the cluster must have beyond a plain 1-control-plane, 1-worker kind cluster.
  This decides when we can build the task.
- **Level:** ★ quick (under 5 min), ★★ normal, ★★★ long or multi-step.

---

## 1. Troubleshooting (30%)

### Clusters and nodes

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T1 | A worker node is `NotReady`. Find out why and fix it so that it survives a reboot. Example causes: kubelet stopped or disabled, wrong flag in the kubelet config, bad CA path. | ★★ | systemd kubelet on node | main `tr-kubelet` |
| T2 | A node is `NotReady` because the container runtime is down or its socket path is wrong in the kubelet config. Fix it. | ★★ | containerd on node || main `tr-runtime` |
| T3 | A node shows `DiskPressure` or `MemoryPressure` and Pods are being evicted. Find the cause and free the node. | ★★★ | fill a disk on node | |
| T4 | A node was cordoned and left `SchedulingDisabled`. Find which one and make it schedulable again without restarting anything. | ★ | | main `tr-cordon` |
| T5 | Drain a node for maintenance (DaemonSets, emptyDir and unmanaged Pods present), then return it to service. | ★★ | 2+ workers || main `tr-drain` |
| T6 | The kubelet client certificate on a node has expired or is invalid. Get the node back. | ★★★ | cert tampering || main `tr-kubelet-cert` |

### Cluster components

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T7 | New Pods stay `Pending` forever. The scheduler is broken through its static Pod manifest. Fix it without touching the Pods. | ★★ | static Pods | main `tr-scheduler` |
| T8 | `kubectl` cannot reach the API server. The kube-apiserver manifest has a bad flag, port, cert path or etcd endpoint. Fix it. | ★★★ | static Pods || main `tr-apiserver` |
| T9 | Deployments don't create Pods, or scaling has no effect. kube-controller-manager is broken. Fix it. | ★★ | static Pods || main `tr-controller-manager` |
| T10 | Write how kube-scheduler, controller-manager, etcd and DNS are run in this cluster (static Pod, Deployment, systemd process) to a file. | ★ | || main `tr-components` |
| T11 | Write the order of commands you would use to check a cluster's health to a file (`crictl ps`, `journalctl -u kubelet`, `/var/log/pods`). | ★ | || main `tr-etcd` (folded into T18) |
| T37 | Check control-plane health without the removed `componentstatuses`: `kubectl get --raw='/readyz?verbose'`, `/livez`, and the kubelet's `/healthz`. Write the failing check to a file. | ★ | || main `tr-health` |

### Resource usage

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T12 | Find the Pod using the most CPU (or memory) in a namespace or across the cluster, and write its name to a file. | ★ | metrics-server || main `tr-top-pod` |
| T13 | Write a script or command that shows node resource usage sorted by memory. | ★ | metrics-server || main `tr-top-nodes` |
| T14 | Write all Pods sorted by creation time or by restart count to a file (`--sort-by`, `custom-columns`). | ★ | | main `tr-sort-by` |
| T33 | Install metrics-server from a manifest and get `kubectl top` working on a lab cluster (`--kubelet-insecure-tls`). | ★★ | metrics-server image preloaded || main `tr-metrics-server` |
| T34 | Without metrics-server, find the node with the most *requested* CPU from `kubectl describe node` ("Allocated resources"), and explain how requests differ from actual usage. | ★ | || main `tr-requested-cpu` |
| T35 | Write every Pod in the cluster that has no resource requests (BestEffort QoS) to a file, using custom-columns or jsonpath. | ★ | || main `tr-besteffort` |
| T36 | Change a Pod so its QoS class becomes `Guaranteed`, and write the QoS class of three given Pods to a file. | ★ | || main `tr-qos` |
| T38 | Find which container in a multi-container Pod uses the most memory (`kubectl top pod --containers`). | ★ | metrics-server || main `tr-top-containers` |

### Container output streams (logs)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T15 | A Pod is in `CrashLoopBackOff`. Read the logs of the **previous** container and fix the cause (wrong command, missing env, missing ConfigMap key). | ★★ | | main `tr-crashloop` |
| T16 | Save the logs of one container in a multi-container Pod to a file, and only the lines that match an error. | ★ | | main `tr-container-logs` |
| T17 | Add a sidecar container that streams a log file the main container writes to an emptyDir, so `kubectl logs` shows it. | ★★ | || main `tr-sidecar-logs` |
| T18 | A control-plane component's container is gone from `kubectl`. Find its logs with `crictl logs` or `/var/log/pods` on the node. | ★★ | || main `tr-etcd` |
| T19 | Write the events for a failing Pod, sorted by time, to a file. | ★ | | main `tr-events` |
| T28 | A static Pod never appears. Find the YAML parse error in `journalctl -u kubelet` on the node and fix the manifest. | ★★ | static Pods || main `tr-static-yaml` |
| T29 | Set `containerLogMaxSize` and `containerLogMaxFiles` in a node's kubelet config so container logs rotate, then check the files under `/var/log/pods`. | ★★ | systemd kubelet on node || main `tr-log-rotation` |
| T30 | Read the logs of every Pod of a Deployment at once (`-l app=x --prefix --since=1h`), and write how many error lines each Pod has. | ★ | || main `tr-logs-count` |
| T31 | Run a node log agent as a DaemonSet that mounts `/var/log/pods` read-only and tails one namespace's logs. | ★★ | || main `tr-log-agent` |
| T32 | An app writes to a file instead of stdout. Use a native sidecar (an init container with `restartPolicy: Always`) to stream it, and explain why it starts before the app and stops after it. | ★★ | || main `tr-native-sidecar` |

### Services and networking

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T20 | A Service times out. Its selector or `targetPort` doesn't match the Pods. Fix the Service, not the Deployment. | ★★ | | main `tr-service` |
| T21 | DNS lookups fail inside Pods. CoreDNS is scaled to 0, its ConfigMap is broken, or kube-dns has no endpoints. Fix it. | ★★ | || main `tr-coredns` |
| T22 | Pods on one node cannot reach Pods on another. The CNI config is missing or broken on that node. Fix it. | ★★★ | CNI on node || main `tr-cni` |
| T23 | kube-proxy is not running, or uses the wrong mode, so ClusterIPs don't route. Fix it. | ★★ | || main `tr-kube-proxy` |
| T24 | An Ingress or HTTPRoute returns 404/503. Find the wrong backend name, port or path and fix it. | ★★ | ingress/gateway controller || main `tr-ingress-backend` |

### Workload failures

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| T25 | A container keeps restarting with `Last State: OOMKilled` (exit code 137). Tell it apart from an app crash and raise the memory limit just enough for it to run. | ★★ | || main `tr-oomkilled` |
| T26 | Inspect a distroless Pod with no shell by attaching an ephemeral container (`kubectl debug -it --target`), and read a node's files with `kubectl debug node/<name>` instead of ssh. | ★★ | || main `tr-debug` |
| T27 | Pods are in `ImagePullBackOff` because the image is in a private registry. Create a `docker-registry` Secret and attach it via `imagePullSecrets` on the ServiceAccount so every Pod gets it. | ★★ | local registry with auth || main `tr-private-registry` |

---

## 2. Cluster Architecture, Installation & Configuration (25%)

### RBAC

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A1 | Create a ServiceAccount, a Role and a RoleBinding with least privilege, then check with `kubectl auth can-i --as=system:serviceaccount:...`. | ★★ | | main `ar-rbac` |
| A2 | Create a ClusterRole that can only `list` and `get` nodes and PersistentVolumes, and bind it to a user or group. | ★ | || main `ar-clusterrole` |
| A3 | Bind a **ClusterRole** with a **RoleBinding** so the subject gets its rights in one namespace only. Explain the difference in a file. | ★★ | || main `ar-rolebinding-scope` |
| A4 | Create a new user from a CSR (key → CSR object → approve → kubeconfig entry), then give that user read-only access to one namespace. | ★★★ | || main `ar-csr-user` |
| A5 | A Pod's ServiceAccount gets `forbidden` when it lists Pods. Find the missing rule and add it. | ★★ | || main `ar-sa-forbidden` |
| A28 | Create a short-lived token for a ServiceAccount (`kubectl create token --duration`), build a standalone kubeconfig file from it (cluster, user, context), and show it has only its RBAC rights. | ★★ | || main `ar-sa-kubeconfig` |
| A46 | Find every RoleBinding and ClusterRoleBinding that gives a user or group its rights, and remove the one that grants more than the user needs. | ★★ | || main `ar-rbac-audit` |
| A47 | Create a Role that can only `get` and `update` one named ConfigMap (`resourceNames`), and prove it can't touch the others. | ★★ | || main `ar-resourcenames` |
| A48 | Make the built-in `view` ClusterRole able to read a custom resource with an aggregated ClusterRole (`rbac.authorization.k8s.io/aggregate-to-view: "true"`). | ★★ | || main `ar-aggregate-view` |
| A49 | Grant rights on subresources only: let a user read `pods/log` and run `pods/exec`, or scale Deployments through `deployments/scale` without editing them. | ★★ | || main `ar-subresources` |
| A50 | Stop a ServiceAccount and a Pod from mounting an API token (`automountServiceAccountToken: false`), and check that `/var/run/secrets/kubernetes.io/serviceaccount` is gone. | ★ | || main `ar-no-automount` |
| A51 | Create a long-lived token Secret (`type: kubernetes.io/service-account-token`) for an outside system, and write how it differs from `kubectl create token`. | ★★ | || main `ar-legacy-token` |

### Authentication, TLS and kubeconfig

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A35 | Find how the API server authenticates and authorizes requests (`--client-ca-file`, `--anonymous-auth`, `--authorization-mode`) and write them to a file. | ★ | || main `ar-auth-config` |
| A36 | Write who the current kubeconfig authenticates as (`kubectl auth whoami`, user and groups) and everything a given user may do (`kubectl auth can-i --list --as`). | ★ | || main `ar-whoami` |
| A37 | Read the API server certificate with `openssl x509 -text`, write its issuer, SANs and expiry, and find the hostname that is missing from the SANs. | ★★ | || main `ar-apiserver-cert` |
| A38 | Write which CA signs which certificates in a kubeadm cluster: the cluster CA, the etcd CA and the front-proxy CA, with their files under `/etc/kubernetes/pki`. | ★★ | || main `ar-ca-map` |
| A39 | Create a user certificate by hand with `openssl` (key, CSR with CN and O, signed by the cluster CA), put it in a kubeconfig and use it. | ★★★ | || main `ar-openssl-user` |
| A40 | Two CSRs are pending. Approve the real one, and deny the one that asks for an admin group (`kubeadm:cluster-admins`; `system:masters` is refused at creation now). | ★★ | || main `ar-csr-deny` |
| A41 | A kubeconfig file at a given path doesn't work (wrong server port, a context that points to a missing user). Fix it and make its context the current one. | ★★ | || main `ar-broken-kubeconfig` |
| A42 | Merge two kubeconfig files into one (`KUBECONFIG=a:b kubectl config view --flatten`), then switch context and set a default namespace (what `kubectx` and `kubens` do). | ★★ | || main `ar-merge-kubeconfig` |
| A43 | Decode the client certificate inside a kubeconfig (`base64 -d`) and write its CN, groups and expiry date to a file. | ★ | || main `ar-kubeconfig-cert` |

### API groups

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A44 | Write the API group and version of given resources, and whether each is namespaced (`kubectl api-resources`, `kubectl api-versions`). | ★ | || main `ar-api-resources` |
| A45 | Call the API directly to list the Pods of a namespace, with `kubectl proxy` plus `curl` or with `kubectl get --raw`. | ★ | || main `ar-raw-api` |

### Infrastructure and kubeadm

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A6 | Prepare a node for kubeadm: load kernel modules, set sysctls (`ip_forward`, `bridge-nf-call-iptables`), disable swap and install the container runtime. | ★★ | bare node | |
| A7 | Join a new worker node to the cluster: create the token and CA hash on the control plane, then run `kubeadm join`. | ★★★ | extra node not joined | |
| A8 | Upgrade the control plane by one minor version with kubeadm (`plan`, `apply`, then kubelet/kubectl), then upgrade a worker. | ★★★ | versioned packages | |
| A9 | Check which kubeadm certificates expire soonest and renew them. Write their expiry dates to a file. | ★★ | || main `ar-certs` |
| A10 | Find the kubelet config file, the static Pod path and the cgroup driver on a node, and write them to a file. | ★ | || main `ar-kubelet-facts` |
| A29 | Set `evictionHard` and `systemReserved` in a worker's kubelet config, restart the kubelet, and check the new `Allocatable` in `kubectl describe node`. | ★★ | systemd kubelet on node || main `ar-kubelet-reserved` |

### Cluster lifecycle and etcd

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A11 | Back up etcd with `etcdctl snapshot save` (endpoints and certs), then restore from an older snapshot and get the cluster healthy. | ★★★ | etcd static Pod | main `ar-etcd` |
| A12 | Find etcd's version, data dir and member list, and write them to a file. | ★ | || main `ar-etcd-facts` |
| A27 | A Namespace or PVC is stuck in `Terminating`. Find the finalizer that holds it (e.g. `kubernetes.io/pvc-protection` held by a running Pod) and fix the cause instead of force-removing the finalizer. | ★★ | || main `ar-finalizer` |

### Highly-available control plane

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A13 | List the control-plane nodes and the etcd members in an HA cluster. Find the one etcd member that is unhealthy. | ★★ | 3 control planes | |
| A14 | Explain in a file what `--control-plane-endpoint` is for, and which component load-balances the API servers here. | ★ | HA cluster | |

### Helm and Kustomize

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A15 | Add a Helm repo and install a chart in a namespace with overridden values (`--set` or a values file). Then upgrade it, then roll it back. | ★★ | helm + local chart/repo || main `ar-helm-install` |
| A16 | Find a broken Helm release (`helm ls -a`, `pending-install`/`failed`) and delete it. Write the chart's default values to a file. | ★★ | helm || main `ar-helm-broken` |
| A17 | Render a chart with `helm template` and install only the manifests it creates, without Helm managing them. | ★★ | helm || main `ar-helm-template` |
| A18 | Use a Kustomize overlay to change the replicas, image tag and a label on a base, and apply it with `kubectl apply -k`. | ★★ | || main `ar-kustomize` |
| A19 | Create a ConfigMap and a Secret from files with a Kustomize generator, and roll them into a Deployment. | ★★ | || main `ar-kustomize-gen` |
| A30 | Find which chart version and values a release was installed with (`helm history`, `helm get values --all`), and write them to a file. | ★ | helm || main `ar-helm-history` |
| A31 | An upgrade lost a custom value because `--reset-values` was the default. Restore it with `--reuse-values` or a values file. | ★★ | helm || main `ar-helm-reuse` |
| A32 | `helm pull --untar` a chart, change one default in `values.yaml`, and install it from the local folder with `--create-namespace`. | ★★ | helm + local chart/repo || main `ar-helm-pull` |
| A33 | Find the Helm release that owns a given Deployment from the `meta.helm.sh/release-name` annotation, and list all releases with `helm ls -A`. | ★ | helm || main `ar-helm-owner` |
| A34 | A release is stuck in `pending-upgrade`, so every new upgrade fails. Roll it back to the last deployed revision. | ★★ | helm || main `ar-helm-stuck` |

### Extension interfaces (CNI, CSI, CRI)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A20 | Find which CNI plugin is installed, where its config lives (`/etc/cni/net.d`) and the Pod CIDR, and write them to a file. | ★ | || main `ar-cni-facts` |
| A21 | Find the container runtime and its socket, and list the running containers on a node with `crictl`. | ★ | || main `ar-cri-facts` |
| A22 | Install a CNI plugin from a given manifest on a cluster that has none, so the nodes become `Ready`. | ★★★ | cluster without CNI | |
| A23 | List the CSI drivers in the cluster and write which StorageClass uses which provisioner to a file. | ★ | CSI driver | |

### CRDs and operators

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| A24 | Install a CRD from a manifest and create a custom resource of that kind. Write its `kubectl explain` fields to a file. | ★★ | || main `ar-crd` |
| A25 | Install an operator (manifest or Helm), then create its custom resource so the operator builds the workload. | ★★★ | operator image || main `ar-operator` |
| A26 | List the CRDs of a group (e.g. `*.cert-manager.io`) and the custom resources that exist for one of them. | ★ | CRDs installed || main `ar-crd-list` |
| A52 | Write a CRD from scratch with a schema (a required field and an `enum`), a short name and `additionalPrinterColumns`, and show that an invalid custom resource is rejected. | ★★★ | || main `ar-crd-schema` |
| A53 | Serve a CRD in two versions (`v1alpha1` served, `v1` stored), and find the stored versions in `status.storedVersions`. | ★★ | || main `ar-crd-versions` |

---

## 3. Services & Networking (20%)

### Pod connectivity

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N1 | From a temporary Pod, check whether Pod A can reach Pod B by IP and by Service DNS name. Write the results to a file. | ★ | || main `nw-connect` |
| N2 | Find a Pod's IP, its node and the node's Pod CIDR, and write them to a file. | ★ | || main `nw-pod-facts` |

### Node networking and CNI

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N23 | On a node, read the routing table (`ip route`) and write which interface and gateway carry traffic to another node's Pod CIDR. | ★ | || main `nw-routes` |
| N24 | Find a node's internal IP, its default gateway, and whether `net.ipv4.ip_forward` is on, and write them to a file. | ★ | || main `nw-node-net` |
| N25 | Enter a Pod's network namespace from the node (`crictl inspect` → PID → `nsenter -t <pid> -n ip addr`), find the host side of its veth pair, and write its name to a file. | ★★ | || main `nw-veth` |
| N26 | Run a Pod with `hostNetwork: true`, show it gets the node's IP, and explain in a file why two such Pods can't listen on the same port on one node. | ★ | || main `nw-hostnetwork` |
| N27 | Expose a container on its node with `hostPort`, reach it on the node IP, and write how this differs from a NodePort Service. | ★★ | || main `nw-hostport` |
| N28 | Write each node's Pod CIDR, the cluster-wide Pod CIDR (`--cluster-cidr`) and the Service CIDR (`--service-cluster-ip-range`) to a file. | ★ | || main `nw-cidrs` |
| N29 | New Pods on one node hang in `ContainerCreating` with `failed to find plugin`. The CNI binary dir in containerd's config is wrong. Fix it. | ★★★ | CNI on node || main `nw-cni-bindir` |

### NetworkPolicies

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N3 | Allow ingress to `app=api` only from one label in the same namespace and from any Pod in another namespace, on one port. | ★★ | policy-enforcing CNI || main `nw-netpol-ingress` |
| N4 | Default-deny all ingress and egress in a namespace, then allow egress to DNS (UDP/TCP 53) only. | ★★ | policy-enforcing CNI || main `nw-netpol-lockdown` |
| N5 | Allow egress from a backend to a database Pod on one port and to nothing else. | ★★ | policy-enforcing CNI || main `nw-netpol-egress` |
| N6 | Several NetworkPolicies are given. Pick the one that meets the requirement with least privilege and apply it. | ★★ | policy-enforcing CNI || main `nw-netpol-pick` |
| N7 | Fix a policy whose `namespaceSelector` and `podSelector` are in separate list items (OR) when they should be one item (AND). | ★★ | policy-enforcing CNI || main `nw-netpol-and-or` |
| N51 | Allow ingress only from an IP range with `ipBlock` and `except`. | ★★ | policy-enforcing CNI || main `nw-netpol-ipblock` |
| N52 | Allow a range of ports with `endPort`, and a container port by its name. | ★★ | policy-enforcing CNI || main `nw-netpol-ports` |
| N53 | Given a set of NetworkPolicies, write which of a list of connections are allowed. Policies add up: any policy that allows a connection allows it. | ★★ | || main `nw-netpol-analysis` |

### Service types and endpoints

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N8 | Expose a Deployment with a ClusterIP Service on a different port than the container port. | ★ | || main `nw-clusterip` |
| N9 | Expose a Deployment with a NodePort Service on a fixed nodePort, and reach it from the node. | ★ | || main `nw-nodeport` |
| N10 | Create a LoadBalancer Service and explain why it stays `<pending>` with no cloud provider. | ★★ | LB provider optional || main `nw-loadbalancer` |
| N11 | Create a Service with no selector and a manual EndpointSlice that points to an outside IP. | ★★ | || main `nw-external` |
| N12 | A Service has no endpoints. Find out why (the Pods aren't Ready because of a failing readiness probe) and fix it. | ★★ | || main `nw-readiness` |
| N22 | A NodePort with `externalTrafficPolicy: Local` answers on one node but not another. Find which nodes run a backend Pod, explain why, and make it answer on every node (or explain the client-IP trade-off in a file). | ★★ | 2+ workers || main `nw-etp-local` |
| N30 | Find kube-proxy's mode (`iptables`, `ipvs` or `nftables`) in its ConfigMap, then find the rules for one Service on a node (`iptables-save` or `nft list ruleset`) and write the Service's chain name to a file. | ★★ | || main `nw-kube-proxy-mode` |
| N31 | Create a Service with a fixed `clusterIP` from the Service range, and write the error you get for an IP outside it. | ★ | || main `nw-fixed-clusterip` |
| N32 | Create a headless Service (`clusterIP: None`) and show from a Pod that DNS returns the Pod IPs instead of one virtual IP. | ★★ | || main `nw-headless` |
| N33 | Create an `ExternalName` Service for an outside DNS name and check the CNAME answer from a Pod. | ★ | || main `nw-externalname` |
| N34 | Create a multi-port Service that uses named `targetPort`s, so changing the container port number doesn't break it. | ★★ | || main `nw-named-port` |
| N35 | Make a Service send a client to the same Pod each time with `sessionAffinity: ClientIP`. | ★ | || later: the local lab kernel lacks the iptables `recent` match, so kube-proxy fails to sync while a ClientIP-affinity Service exists |

### Gateway API

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N13 | Create a Gateway (with a given GatewayClass) and an HTTPRoute that sends `/api` to one Service and `/` to another. | ★★ | Gateway API CRDs + controller || main `nw-gateway-routes` |
| N14 | Split traffic between two backends with weights (canary 80/20) in an HTTPRoute. | ★★ | Gateway API || main `nw-gateway-canary` |
| N15 | Migrate an existing Ingress (host, paths, TLS) to a Gateway plus HTTPRoute, keeping the same behaviour. | ★★★ | Gateway API || main `nw-gateway-migrate` |
| N16 | Add a header-based route match, or a request header modifier filter, to an HTTPRoute. | ★★ | Gateway API || main `nw-gateway-header` |
| N46 | An HTTPRoute in another namespace is not attached to a Gateway whose listener allows only `Same` namespace routes. Fix `allowedRoutes`. | ★★ | Gateway API || main `nw-gateway-allowed` |
| N47 | Let an HTTPRoute send traffic to a Service in another namespace with a `ReferenceGrant`. | ★★ | Gateway API || main `nw-gateway-refgrant` |
| N48 | Add an HTTPS listener with a TLS Secret (`certificateRefs`) to a Gateway, and a `RequestRedirect` filter that sends HTTP to HTTPS. | ★★★ | Gateway API || main `nw-gateway-https` |
| N49 | Traffic doesn't reach a backend. Read the Gateway and HTTPRoute status conditions (`Accepted`, `ResolvedRefs`, `Programmed`) and write the failing reason to a file. | ★★ | Gateway API || main `nw-gateway-status` |
| N50 | Add a `URLRewrite` filter that strips a path prefix before the request reaches the backend. | ★★ | Gateway API || main `nw-gateway-rewrite` |

### Ingress

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N17 | Create an Ingress with two host or path rules to two Services and a given `ingressClassName`, and test it with `curl`. | ★★ | ingress controller || main `nw-ingress-rules` |
| N18 | Add TLS to an Ingress using a Secret made from a given cert and key. | ★★ | ingress controller || main `nw-ingress-tls` |
| N41 | Use the `nginx.ingress.kubernetes.io/rewrite-target` annotation so `/app` reaches the backend as `/`. | ★★ | ingress controller || main `nw-ingress-rewrite` |
| N42 | Add a `spec.defaultBackend` to an Ingress for requests that match no rule. | ★ | ingress controller || main `nw-ingress-default` |
| N43 | An Ingress path doesn't match because of its `pathType` (`Exact` vs `Prefix`). Fix it. | ★★ | ingress controller || main `nw-ingress-pathtype` |
| N44 | An Ingress can't route to a Service in another namespace. Fix it by creating an Ingress in each Service's namespace. | ★★ | ingress controller || main `nw-ingress-cross` |
| N45 | An Ingress without `ingressClassName` is served by nobody. Mark one IngressClass as the default (`ingressclass.kubernetes.io/is-default-class`). | ★ | ingress controller || main `nw-ingress-class` |

### CoreDNS

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| N19 | Write the FQDN of a Service, and of a Pod by its IP, to a file, and check both with `nslookup` from a Pod. | ★ | || main `nw-dns-names` |
| N20 | Change the CoreDNS config to add a custom domain or stub zone, or forward an upstream zone, then check it resolves. | ★★★ | || main `nw-dns-forward` |
| N21 | Change the cluster domain the kubelet hands to Pods, or find it, and write it to a file. | ★ | || main `nw-cluster-domain` |
| N36 | Explain in a file why `nslookup web` works in a Pod but `nslookup web.other` needs the namespace (`search` and `ndots:5` in `/etc/resolv.conf`). Then give a Pod its own nameserver with `dnsPolicy: None` and `dnsConfig`. | ★★ | || main `nw-dns-config` |
| N37 | Add a static hosts entry to a Pod with `hostAliases`, and check it in the Pod's `/etc/hosts`. | ★ | || main `nw-hostaliases` |
| N38 | Write the CoreDNS Service IP, the name of the ConfigMap that holds the Corefile, and the upstream it forwards to, to a file. | ★ | || main `nw-coredns-facts` |
| N39 | Turn on the CoreDNS `log` plugin, find the queries for one name in the CoreDNS logs, then turn it off again. | ★★ | || main `nw-coredns-log` |
| N40 | Make one name resolve cluster-wide to a fixed IP with the CoreDNS `hosts` plugin (or `rewrite` it to a Service name), and check it from a Pod. | ★★ | || main `nw-coredns-hosts` |

---

## 4. Workloads & Scheduling (15%)

### Basics: labels and namespaces

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W27 | Select Pods with equality and set-based label selectors, write the matches to a file, add and remove labels, and annotate a Pod. | ★ | | main `wl-labels` |
| W28 | Create a labelled namespace and run workloads in it, and find which namespace a named Pod lives in. | ★ | | main `wl-namespace` |

### Deployments, rollouts and rollbacks

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W1 | Scale a Deployment and make sure all replicas are Ready. | ★ | | main `wl-scale` |
| W2 | Roll back a broken rollout to the last good revision, write the revision number to a file, and set `maxSurge`/`maxUnavailable`. | ★★ | | main `wl-rollout` |
| W3 | Update a Deployment's image and record the change cause. Pause the rollout, make two changes, then resume it. | ★★ | || main `wl-rollout-pause` |
| W4 | Change a Deployment to the `Recreate` strategy and explain when you would choose it, in a file. | ★ | || main `wl-recreate` |
| W29 | A Deployment update hangs with `ProgressDeadlineExceeded`. Find the reason in the new ReplicaSet's events (exceeded ResourceQuota) and get the rollout through without deleting the quota. | ★★ | || main `wl-quota-rollout` |
| W32 | Blue/green: run `v1` and `v2` as two Deployments and switch the Service selector, with no downtime. | ★★ | || main `wl-blue-green` |
| W33 | Canary without a service mesh: two Deployments behind one Service in a 3:1 replica ratio. | ★★ | || main `wl-canary` |

### ConfigMaps and Secrets

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W5 | Create a ConfigMap from literals and a file. Mount one key as a file at a given path, and expose another as an env var. | ★★ | || main `wl-configmap` |
| W6 | Create a Secret and use it as env vars (`envFrom`) and as a read-only volume. | ★★ | || main `wl-secret` |
| W7 | Decode a value from an existing Secret and write it to a file. | ★ | || main `wl-secret-decode` |
| W8 | Make a ConfigMap immutable, then change the value the app sees by creating a new ConfigMap and rolling the Deployment. | ★★ | || main `wl-immutable` |
| W37 | An app doesn't see a ConfigMap update because the file is mounted with `subPath`. Explain why and fix it so updates reach the app. | ★★ | || main `wl-subpath` |

### Autoscaling

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W9 | Create an HPA (`autoscaling/v2`) for a Deployment: min/max replicas and a CPU utilization target. | ★★ | metrics-server || main `wl-hpa` |
| W10 | Add a scale-down stabilization window or behaviour policy to an HPA. | ★★ | metrics-server || main `wl-hpa-behavior` |
| W11 | An HPA shows `<unknown>` targets. Find out why (missing resource requests or metrics-server) and fix it. | ★★ | metrics-server || main `wl-hpa-unknown` |
| W12 | Explain or set a VPA in recommendation-only mode, if the CRD exists. | ★★ | VPA CRDs || main `wl-vpa` |
| W38 | Resize a running Pod's CPU in place (`kubectl patch --subresource resize`, `resizePolicy`), without restarting it. | ★★ | || main `wl-resize` |
| W39 | Create an HPA on memory (`averageValue`) and CPU together, and explain which metric decides the replica count. | ★★ | metrics-server || main `wl-hpa-multi` |
| W40 | An HPA and a manifest that sets `replicas: 3` keep fighting. Find the cause and fix it so only the HPA scales. | ★★ | metrics-server || main `wl-hpa-fight` |
| W41 | An HPA shows `FailedGetScale` because `scaleTargetRef` has the wrong kind or name. Fix it. | ★ | metrics-server || main `wl-hpa-target` |
| W42 | Generate load with a busybox `wget` loop, watch the HPA scale up, then watch the stabilization window delay the scale-down. | ★★★ | metrics-server || main `wl-hpa-load` |
| W43 | Scale a StatefulSet with an HPA, and explain how scaling down affects its Pods and PVCs. | ★★ | metrics-server || main `wl-hpa-sts` |

### Self-healing primitives

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W13 | Add liveness, readiness and startup probes (HTTP, TCP and exec) to a Deployment. | ★★ | || main `wl-probes` |
| W14 | Create a DaemonSet that runs on every node, including the control plane (tolerations). | ★★ | || main `wl-daemonset` |
| W15 | Create a StatefulSet with a headless Service and `volumeClaimTemplates`, and check the stable Pod names and PVCs. | ★★★ | || main `wl-statefulset` |
| W16 | Create a Job with `completions`, `parallelism` and `backoffLimit`, and a CronJob with a schedule and history limits. | ★★ | || main `wl-jobs` |
| W17 | Add an init container that waits for a Service before the main container starts. | ★★ | || main `wl-init` |
| W18 | Create a PodDisruptionBudget so a drain never takes a Deployment below N Pods. | ★★ | 2+ workers || main `wl-pdb` |
| W31 | Requests fail during every rollout. Add a `preStop` hook and set `terminationGracePeriodSeconds` so Pods drain before they are killed. | ★★ | || main `wl-graceful` |
| W34 | A CronJob never runs (`suspend: true`, a wrong schedule, or `concurrencyPolicy: Forbid` behind a stuck Job). Fix it and trigger one run by hand with `kubectl create job --from=cronjob/x`. | ★★ | || main `wl-cronjob-fix` |
| W35 | Clean up a Job after it finishes and cap how long it may run (`ttlSecondsAfterFinished`, `activeDeadlineSeconds`). | ★ | || main `wl-job-limits` |
| W36 | Use `podFailurePolicy` so a Job fails at once on exit code 42 but retries other failures. | ★★ | || main `wl-pod-failure-policy` |

### Admission and scheduling

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| W19 | Set resource requests and limits on a Pod. Add a LimitRange and a ResourceQuota to a namespace, and show a Pod being rejected. | ★★ | || main `wl-quota` |
| W20 | Schedule a Pod only on nodes with a label, first with `nodeSelector`, then with required and preferred node affinity. | ★★ | | main `wl-node-affinity` |
| W21 | Taint a node, then run a Pod that tolerates the taint and one that doesn't. | ★★ | | main `wl-taints` |
| W22 | Spread a Deployment's replicas across nodes with pod anti-affinity or `topologySpreadConstraints`. | ★★ | 2+ workers || main `wl-spread` |
| W23 | A Pod stays `Pending`. Read the events (insufficient CPU, an unmatched affinity, an untolerated taint) and fix the Pod spec. | ★★ | | main `wl-pending` |
| W24 | Run a Pod on a specific node by name without the scheduler (`nodeName`), or as a static Pod on a worker. | ★ | || main `wl-nodename` |
| W25 | Create a PriorityClass and use it so a critical Pod evicts lower-priority Pods. | ★★ | || main `wl-priority` |
| W26 | Set a Pod `securityContext`: `runAsUser`, `readOnlyRootFilesystem` and dropped capabilities. | ★★ | || main `wl-security-context` |
| W30 | Enforce Pod Security Admission `restricted` on a namespace with a label, then fix a Deployment whose Pods are rejected (`runAsNonRoot`, `seccompProfile`, `allowPrivilegeEscalation: false`, drop `ALL`). | ★★ | || main `wl-psa` |
| W44 | Pin a Deployment's image by digest (`image@sha256:…`) instead of a tag, and set `imagePullPolicy` so a moved tag can't change what runs. | ★ | || main `wl-image-digest` |
| W45 | Set `runAsUser` at Pod level and a different one on one container, and write the effective UID of each container (`id`). | ★ | || main `wl-runas-precedence` |
| W46 | Find a container's process on the node (`crictl inspect` → PID, `ps`) and write which UID it runs as on the host. | ★★ | || main `wl-host-process` |

---

## 5. Storage (10%)

| # | Question | Level | Needs | Covered |
|---|---|---|---|---|
| S1 | Create a hostPath PV (capacity, access mode, `Retain`, storageClassName), a PVC that binds to it, and a Pod that writes to it. | ★★ | | main `st-pvc` |
| S2 | Create a StorageClass with a provisioner, `reclaimPolicy` and `volumeBindingMode: WaitForFirstConsumer`, and make it the default. | ★★ | dynamic provisioner || main `st-storageclass` |
| S3 | Create a PVC with no PV and get it bound through dynamic provisioning. Check that the PV it creates gets deleted with the PVC. | ★★ | dynamic provisioner || main `st-dynamic` |
| S4 | A PVC stays `Pending`. Find out why (wrong class, access mode, or too big a size) and fix the PVC without changing the PV. | ★★ | || main `st-pending-pvc` |
| S5 | Expand a PVC (`allowVolumeExpansion`) and check the new size inside the Pod. | ★★ | expandable provisioner | |
| S6 | Find the PV a PVC is bound to and its reclaim policy. Change the policy from `Delete` to `Retain` so data survives deleting the PVC. | ★ | || main `st-reclaim` |
| S7 | Reuse a `Released` PV: clear its `claimRef` so a new PVC can bind to it. | ★★ | || main `st-released` |
| S8 | Share data between two containers in a Pod with an `emptyDir` volume, then with a memory-backed one. | ★ | || main `st-emptydir` |
| S9 | Explain in a file the difference between `ReadWriteOnce`, `ReadOnlyMany`, `ReadWriteMany` and `ReadWriteOncePod`. | ★ | || main `st-access-modes` |
| S10 | A container writes to its own filesystem and the data disappears when it restarts. Prove it by writing a file and killing the container, then move the path to an `emptyDir` and show the data survives a container restart but not a Pod delete. | ★★ | || main `st-writable-layer` |
| S11 | On the node, find where a Pod's `emptyDir` lives (`/var/lib/kubelet/pods/<uid>/volumes/…`) and the container's mounts (`crictl inspect`), and write both paths to a file. | ★★ | || main `st-node-paths` |
| S12 | A Pod is stuck in `ContainerCreating` because its hostPath uses `type: Directory` and the path doesn't exist on the node. Fix it with the right hostPath `type`. | ★ | || main `st-hostpath-type` |
| S13 | A PVC resize is stuck at `FileSystemResizePending`. Finish the expansion and check the new size inside the Pod. | ★★ | expandable CSI driver | |
| S14 | List the `CSIDriver`, `CSINode` and `VolumeAttachment` objects, and write which node a volume is attached to and which driver serves it to a file. | ★★ | CSI driver | |
| S15 | Combine a ConfigMap, a Secret, the Pod's labels (downward API) and a ServiceAccount token with an audience into one `projected` volume. | ★★ | || main `st-projected` |
| S16 | Expose a Pod's labels, annotations and its container's memory limit as files with a `downwardAPI` volume. | ★ | || main `st-downward-api` |
| S17 | Mount only one key of a Secret under a given filename (`items`) with `defaultMode: 0400`, where the app expects it. | ★ | || main `st-secret-items` |
| S18 | Two PVs match a PVC. Make it bind to a specific one with `volumeName`, and reserve that PV for it with `claimRef`. | ★★ | || main `st-volumename` |
| S19 | A Pod using a `local` PV stays Pending with `volume node affinity conflict`. Find the node the PV is pinned to and fix the Pod, not the PV. | ★★ | || main `st-local-pv` |
| S20 | Choose between PVs with a PVC `selector.matchLabels`, then write the bound capacity to a file (a 1Gi request on a 5Gi PV reports 5Gi). | ★★ | || main `st-pvc-selector` |
| S21 | A Deployment keeps its data in an `emptyDir` and loses it when its Pod is deleted. Move it to a PVC and prove the data survives a Pod delete. | ★★ | || main `st-emptydir-to-pvc` |
| S22 | Mount a PVC read-write in one Pod and read-only (`readOnly: true`) in a second Pod on the same node, with a `subPath` per container. | ★★ | || main `st-readonly-share` |
| S23 | Two StorageClasses are both marked default. Make exactly one default, and explain in a file how `storageClassName: ""` differs from leaving the field out. | ★★ | || main `st-two-defaults` |
| S24 | A PVC is `Pending` and the team thinks it's broken. Show it is `WaitForFirstConsumer`, then get it bound by creating the Pod that uses it. | ★ | || main `st-wffc` |
| S25 | A non-root container gets `permission denied` reading its Secret volume (`defaultMode: 0440`). Fix it with `securityContext.fsGroup`, without running as root. | ★★ | || main `st-fsgroup` |

---

## Coverage so far

| Domain | Weight | Questions | Covered |
|---|---|---|---|
| Troubleshooting | 30% | 38 | 37 (main) |
| Cluster Architecture | 25% | 53 | 46 (main) |
| Services & Networking | 20% | 53 | 52 (main) |
| Workloads & Scheduling | 15% | 46 | 46 (main) |
| Storage | 10% | 25 | 22 (main) |

## Build progress

The work of turning these questions into catalog tasks, kept here so it survives between sessions.

**Rules:**
- Only questions the current cluster can host get built now.
- Each task is proven with `go test ./catalog -run 'TestSelftest/^<id>$' -selftest -count=1 -v`.
- Changes stay in `apps/platform/catalog/` and `docs/`.
- At the end, every question that wasn't built gets a `later: <reason>` note in its Covered cell.

**What the cluster can do (probed 2026-10-09):**
- kind v1.37 with 1 control plane and 2 workers, kindnet, CoreDNS, kube-proxy, and the `standard` local-path StorageClass (`WaitForFirstConsumer`, no expansion).
- Root on every node, workers included, has an admin kubeconfig, so a task can be hosted on a worker.
- kindnet's own NetworkPolicy engine needs `CONFIG_NFT_QUEUE`, which the WSL2 kernel lacks (its log says "initial nftables sync failed"). The lab adds kube-router in firewall-only mode instead, and that enforces NetworkPolicy (probed 2026-10-12: ingress, egress, AND/OR selectors, named ports, `endPort`, `ipBlock` with `except`, across nodes).
- The nodes have no `etcdctl`/`etcdutl`. `etcd_tools` in `catalog/lib.sh` copies them out of the etcd image.
- `catalog/lib.sh` gives answer files `COURSE=/opt/course/<task id>`, and `fresh_course` resets that folder.
- Lab add-ons (`apps/lab/internal/cluster/addons.go`, installed by every `lab up`):
  - metrics-server, with `--kubelet-insecure-tls`. Its manifest is at `/etc/cka-sim/addons/metrics-server.yaml` on the control plane.
  - kube-router v2.11 as a DaemonSet in `kube-system`, with only its firewall on (`/etc/cka-sim/addons/kube-router.yaml`). Tasks test policies with `connects` and `pod_ip` from `lib.sh`.
  - nftables table `ip cka-sim` on every node rejects new packets that are still addressed to the service CIDR `10.96.0.0/16` (`icmp port-unreachable`). kube-router's FORWARD/OUTPUT ACCEPT otherwise runs before kube-proxy's REJECT, so traffic to a Service with no endpoints would leave the node. On this machine Docker Desktop even answered DNS for `10.96.0.10`. Now it's refused as on a real cluster.
  - Gateway API v1.6.2 standard CRDs, and Traefik v3.7 in Namespace `traefik` (`/etc/cka-sim/addons/traefik.yaml`). Its ingress-nginx provider serves IngressClass `nginx` (controller `k8s.io/ingress-nginx`, not the default class) and honours `nginx.ingress.kubernetes.io/*` annotations such as `rewrite-target`. Its Gateway provider serves GatewayClass `traefik`. Traffic enters at Service `traefik` on ClusterIP `10.96.80.80` (ports 80 and 443) or on NodePorts `31080` and `31443`, with the host in a `Host:` header or `curl --resolve`. Gateway listeners use ports 80 and 443. Probed 2026-10-12: Prefix/Exact paths, rewrite-target, `defaultBackend`, TLS from a Secret, and the default-class annotation (it only fills `ingressClassName` on Ingresses created afterwards). On the Gateway side: weights, header matches, `URLRewrite`, `RequestRedirect`, HTTPS listeners, `allowedRoutes` (`NotAllowedByListeners`), ReferenceGrant (`RefNotPermitted` without it), and `BackendNotFound`.
  - VPA 1.8.0: the CRDs and the recommender only (`/etc/cka-sim/addons/vpa.yaml`). There's no updater or admission controller, so only `updateMode: "Off"` does anything. Probed 2026-10-12: a recommendation appears in about 20 s.
  - Operator image `cka-sim/web-operator:1.0`, preloaded on every node. It's a shell loop over `kubectl` that turns each `WebApp` (`cka-sim.io/v1`, `spec.image`/`replicas`/`port`) into an owned Deployment plus Service, and writes `status.readyReplicas`. Its install manifest (the CRD, Namespace `web-operator`, RBAC and the Deployment) sits **unapplied** at `/etc/cka-sim/addons/web-operator.yaml`. Probed: create, scale through the CR, a manual scale is undone, and deleting the WebApp garbage-collects its objects.
  - helm 3 on every node.
  - Chart repo `http://cka-sim-charts`: chart `web` 1.0.0 (nginx:1.26-alpine) and 1.1.0 (nginx:1.27-alpine, `replicaCount: 2`).
  - Registry `cka-sim-registry:5000`, login `candidate`/`cka-registry`, with `team/web:1.0` and `team/web:1.1`.
- Node convention, so tasks don't fight in one exam: taints go only on `cka-sim-worker` (`tr-kubelet` already breaks it), node labels only on `cka-sim-worker2`. A task that taints a node needs a `reset.sh` that removes the taint (see `wl-taints`).

**Pending (2026-10-12):** every built task passes its selftest. Every question the current cluster can host is built. What's left needs the add-ons still open in "What the cluster still needs".

**Waves (the build order):**

| # | Theme | Questions | State |
|---|---|---|---|
| 1 | Port the v1 tasks | T1, T7, A11, A1, S1, W2 | Done. |
| 2 | Quick troubleshooting | T4, T14, T15, T16, T19 | Done. |
| 3 | Control plane and node | T2, T6, T8, T9, T18 (T11 folded in) | Done. |
| 4 | Network plumbing | T5, T21, T22, T23, N12 | Done. |
| 5 | RBAC and certs | A2, A3, A4, A5, A9 | Done. |
| 6 | Cluster facts (answer files) | T10, A10, A12, A20, A21 | Done. |
| 7 | Kustomize and CRDs | A18, A19, A24, A26 | Done. |
| 8 | Services | N1, N8, N9, N10, N11 | Done. |
| 9 | DNS and addressing | N2, N19, N20, N21 | Done. |
| 10 | Rollouts and config | W3, W4, W5, W6, W7, W8 | Done. |
| 11 | Self-healing | W13–W18 | Done. |
| 12 | Scheduling I | W19–W23 | Done (plus basics W27, W28). |
| 13 | Scheduling II and logs | W24, W25, W26, T17 | Done. |
| 14 | Storage | S2, S3, S4, S6, S7, S8, S9 | Done (S9 rewritten as 4 scenarios). |
| 15 | Real-world failures | T25, T26, W29, W30, W31 | Done. |
| 16 | Real-world cluster care | A27, A28, A29, N22 | Done. |
| 17 | Logs on the node | T28, T29, T30, T31, T32 | Done. |
| 18 | Resources and health | T34, T35, T36, T37 | Done. |
| 19 | Application lifecycle | W32–W38 | Done. |
| 20 | Monitoring and HPA | T12, T13, T33, T38, W9–W11, W39–W43 | Done (after the lab add-ons). |
| 21 | Helm | A15–A17, A30–A34 | Done (after the lab add-ons). |
| 22 | Volumes and node storage | S10, S11, S12, S15, S16, S17, S25 | Done. |
| 23 | Binding PVs and PVCs | S18–S22 | Done. |
| 24 | StorageClass behaviour | S23, S24 | Done. |
| 25 | Node networking and CNI | N23–N29 | Done. |
| 26 | Service internals | N30–N35 | Done; N35 left for later (kernel). |
| 27 | DNS in depth | N36–N40 | Done. |
| 28 | TLS and kubeconfig | A35–A43 | Done. |
| 29 | RBAC in depth and the API | A44–A51 | Done. |
| 30 | CRDs and Pod security | A52, A53, W44, W45, W46, N53 | Done; W44 after the lab registry. |
| 31 | Private registry | T27 | Done. |
| 32 | NetworkPolicy | N3–N7, N51, N52 | Done (after kube-router). |
| 33 | Ingress | N17, N18, N41–N45, T24 | Done (after Traefik). |
| 34 | Gateway API | N13–N16, N46–N50 | Done (after Traefik). |
| 35 | VPA and operators | W12, A25 | Done (after the VPA recommender and the web-operator image). |

**Rewrites so the answer can be checked:**
- T10 becomes `component: static-pod|deployment|daemonset|systemd` lines.
- W4: the check reads `strategy.type: Recreate`.
- S9: write the access mode for 4 given scenarios.
- T11 is merged into T18.

**Left for later, and why:**

| Questions | Waiting for |
|---|---|
| A7, A8, A13, A14 | a spare node, kubeadm packages, or an HA control plane |
| A6, A22, T3 | a safe way to do them; they would damage the shared kind/docker host |
| A23, S5, S13, S14 | a CSI driver; local-path can't expand volumes |
| N35 | a lab kernel with the iptables `recent` match (WSL2's lacks it) |

## What the cluster still needs

These are grouped by how many questions each one unblocks:

1. ~~**2+ worker nodes**~~: done; the cluster has 2 workers.
2. ~~**metrics-server**~~: done; `lab up` installs it.
3. ~~**NetworkPolicy enforcement**~~: done; `lab up` installs kube-router (firewall only) next to kindnet.
4. ~~**Gateway API CRDs + a controller**~~: done (Traefik, GatewayClass `traefik`).
5. ~~**Ingress controller**~~: done (Traefik, IngressClass `nginx`).
6. ~~**helm and a local chart repo**~~: done; every node has helm, and `http://cka-sim-charts` serves chart `web`.
7. **Dynamic provisioner** (kind's `local-path` already does this; check its expansion support for S5).
8. **Spare unjoined node, and versioned kubeadm packages:** A7, A8. These are the hardest in kind;
   leave them last.
9. **3 control-plane nodes:** A13, A14.
10. ~~**A private registry with auth**~~: done; `cka-sim-registry:5000`.
11. **A CSI driver with expansion** (e.g. csi-driver-host-path): A23, S5, S13, S14.

## A suggested order for turning these into tasks

1. **Bring back the v1 tasks:** `tr-kubelet`, `tr-scheduler`, `tr-service`, `ar-etcd`, `ar-rbac`,
   `st-pvc`, `wl-rollout`, `nw-netpol`.
2. **Quick, no-extra-setup wins:** T4, T15, T19, T21, A2, A5, A10, A18, N8, N9, N12, W5, W6, W13,
   W19–W21, W23, S4, S6.
3. **High-weight troubleshooting:** T2, T8, T9, T22, T23.
4. **New curriculum topics:** Helm (A15), Kustomize (A18), CRDs/operators (A24–A25), Gateway API (N13–N15).
5. **The long ones:** kubeadm join and upgrade (A7, A8), HA (A13).
