When a PVC bound to a `Retain` PV is deleted, the PV moves to **Released**. It keeps its data, and
its `spec.claimRef` still points at the old claim (with that claim's UID). A Released PV is never
bound again automatically, because Kubernetes can't know whether the data is safe to hand to
someone else. An admin decides.

To hand it to a new claim, clear the stale `claimRef`. The PV becomes `Available`, and the
pending PVC binds:

```sh
kubectl get pv pv-archive                                   # Released, CLAIM st-released/archive
kubectl patch pv pv-archive --type=json -p '[{"op":"remove","path":"/spec/claimRef"}]'
kubectl get pv pv-archive                                   # Available, then Bound to archive-v2
```

Alternatively, set `claimRef` to the new claim's namespace and name (without a uid) to **reserve**
the PV for exactly that claim. Then no other matching PVC can grab it first.

Deleting and recreating the PV would work for hostPath, but on real storage, deleting a `Retain`
PV object leaves the disk behind with nothing pointing at it. Fixing `claimRef` is the clean way.
