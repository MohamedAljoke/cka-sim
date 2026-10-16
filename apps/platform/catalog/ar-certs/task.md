---
id: ar-certs
title: Check and renew kubeadm certificates
domain: architecture
host: cka-sim-control-plane
topics: [kubeadm, certificates, tls, kube-apiserver]
weight: 5
---
## Task

1. Write the expiry date of the cluster CA (`ca`) to `/opt/course/ar-certs/ca-expiry`, in any
   date format.
2. Renew the kube-apiserver's serving certificate with kubeadm.
3. Make the running kube-apiserver serve the renewed certificate.
