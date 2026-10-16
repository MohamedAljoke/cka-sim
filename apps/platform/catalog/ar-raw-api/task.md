---
id: ar-raw-api
title: Talk to the API directly
domain: architecture
host: cka-sim-control-plane
topics: [api, rest, kubectl-proxy]
weight: 3
---
## Task

1. Get the list of Pods in Namespace `ar-raw-api` **straight from the REST API**, without
   `kubectl get pods`, and save the JSON response to `/opt/course/ar-raw-api/pods.json`.
2. Get Deployment `web` in that namespace from its REST path, and write its `spec.replicas` value
   to `/opt/course/ar-raw-api/replicas.txt`.
