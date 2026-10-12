---
id: ar-csr-deny
title: Approve one CSR, deny the other
domain: architecture
host: cka-sim-control-plane
topics: [certificates, csr, security]
weight: 4
---
## Context

Two CertificateSigningRequests are waiting for approval: `dev-anna` and `ci-runner`. Before
approving anything, check what identity each one asks for.

## Task

Approve the CSR that asks for an ordinary user, and **deny** the one whose certificate would make
its holder a cluster administrator.
