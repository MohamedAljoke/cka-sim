---
id: tr-private-registry
title: Pull from a private registry
domain: troubleshooting
host: cka-sim-control-plane
topics: [images, secrets, serviceaccounts]
weight: 5
---
## Context

The Pods of Deployment `web` in Namespace `tr-private-registry` are stuck in `ImagePullBackOff`. The
image lives in the team's private registry `cka-sim-registry:5000`, which needs a login:

- username: `candidate`
- password: `cka-registry`

## Task

1. Create a Secret of type `docker-registry` named `registry-creds` in Namespace
   `tr-private-registry` with these credentials.
2. Make every Pod in the Namespace that runs as ServiceAccount `default` use that Secret for image
   pulls, without changing the Pod template of Deployment `web`.
3. Get both Pods of `web` running.
