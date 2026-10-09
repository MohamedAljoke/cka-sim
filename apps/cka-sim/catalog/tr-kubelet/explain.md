**Concept.** A node is `Ready` only while its kubelet posts heartbeats to the API server. The
kubelet is not a Pod — it is a systemd service on the node, so when it dies the control plane
can only tell you *that* the node went silent, never *why*. The why lives on the node.

**The loop to learn:** `k get nodes` → `ssh <node>` → `sudo -i` → `systemctl status kubelet`
→ `journalctl -u kubelet -n 50` → `systemctl cat kubelet` (shows the unit **and** its drop-ins).

**What was wrong.** The kubeadm drop-in `10-kubeadm.conf` pointed `ExecStart` at
`/usr/local/bin/kubelet`; the binary is `/usr/bin/kubelet` (`which kubelet`). The service was
also disabled, so a reboot would have broken the node again.

**The fix.** Correct the path, `systemctl daemon-reload` (systemd caches unit files — editing
without reloading changes nothing), then `systemctl enable --now kubelet`.
