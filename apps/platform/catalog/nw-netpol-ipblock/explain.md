```yaml
  ingress:
  - from:
    - ipBlock:
        cidr: 10.244.2.0/24     # from /opt/course/nw-netpol-ipblock/cidr
        except:
        - 10.244.2.17/32        # from .../except, as a CIDR
    ports:
    - protocol: TCP
      port: 80
```

- **`except` takes CIDRs, not bare IPs.** Write a single address as `/32`.
- **Every `except` entry must fall inside `cidr`.** The API server rejects any that don't.
- **`ipBlock` matches the source IP the Pod sees.** Here that's the client Pod's own IP, since the
  CNI doesn't NAT Pod-to-Pod traffic. Traffic that arrives through a NodePort or a load balancer
  can be SNATed to a node IP, so `ipBlock` is meant mostly for addresses outside the cluster.
- The CIDR here is `cka-sim-worker2`'s `spec.podCIDR`. That's why `near` and `banned` (both on
  that node) are inside it, and `far` (on `cka-sim-worker`) isn't.

```bash
kubectl get node cka-sim-worker2 -o jsonpath='{.spec.podCIDR}'
kubectl -n nw-netpol-ipblock get pods -o wide
```
