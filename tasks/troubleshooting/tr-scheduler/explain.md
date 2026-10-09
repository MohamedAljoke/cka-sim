**Concept.** `Pending` with no node assigned means *nothing scheduled it*. Scheduling is the
kube-scheduler's only job, and on a kubeadm cluster it runs as a **static Pod**: the kubelet on
the control plane node starts it straight from `/etc/kubernetes/manifests/kube-scheduler.yaml`.
The API object you see in `kube-system` is only a mirror — editing it with `kubectl edit` does
nothing. Edit the file; the kubelet notices within seconds and recreates the Pod.

**How to find it.** `k describe pod probe` shows no `Scheduled` event at all → look at the
scheduler. `k -n kube-system get pods` shows it crash-looping; its logs (or `crictl logs` on the
node when the API is unusable) say the command `kube-schedulerr` does not exist.

**Same pattern, other components:** kube-apiserver, kube-controller-manager and etcd are static
Pods in the same directory. A typo there is a classic exam fault.
