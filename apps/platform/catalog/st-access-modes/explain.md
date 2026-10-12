| Mode | Short | Who may mount it |
|---|---|---|
| `ReadWriteOnce` | RWO | read-write by **one node**; several Pods on that node can share it |
| `ReadOnlyMany` | ROX | read-only by many nodes |
| `ReadWriteMany` | RWX | read-write by many nodes |
| `ReadWriteOncePod` | RWOP | read-write by **one Pod** in the whole cluster |

The common confusion is RWO: "once" means one **node**, not one Pod. Two Pods on the same node can
both mount an RWO volume. Use `ReadWriteOncePod` (GA since v1.29, CSI volumes only) when exactly one
writer must be guaranteed.

What a volume supports depends on its storage: block disks (AWS EBS, GCE PD, local-path) are RWO,
and RWX needs shared file storage (NFS, CephFS, EFS, Azure Files). A PVC binds only to a PV that
lists every mode it asks for. The mode is a promise checked at bind and mount time, and the
`readOnly` flag on the volume mount is what actually makes a mount read-only.
