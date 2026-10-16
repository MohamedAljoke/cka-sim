---
id: ar-csr-user
title: New user through the Certificates API
domain: architecture
host: cka-sim-control-plane
topics: [certificates, csr, kubeconfig, rbac]
weight: 7
---
## Task

Create a user `jane` who can read everything in Namespace `ar-csr-user` and do nothing else.

1. Create a private key `/opt/course/ar-csr-user/jane.key` and a CertificateSigningRequest named
   `jane` for the user `jane`, with signer `kubernetes.io/kube-apiserver-client`.
2. Approve it, and save the issued certificate as `/opt/course/ar-csr-user/jane.crt`.
3. Add credentials `jane` and a context `jane` (cluster `cka-sim`, Namespace `ar-csr-user`) to
   the kubeconfig of `root` on this node.
4. Give `jane` read-only access to Namespace `ar-csr-user` with the built-in ClusterRole `view`.
