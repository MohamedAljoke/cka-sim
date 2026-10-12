---
id: nw-connect
title: Test Pod-to-Pod connectivity
domain: networking
host: cka-sim-control-plane
topics: [pods, services, dns, connectivity]
weight: 3
---
## Context

Namespace `nw-connect` runs Pod `backend`, which answers HTTP on port `8080`, and Service
`backend` in front of it on port `80`.

## Task

From a temporary Pod (image `busybox:1.36`) in Namespace `nw-connect`:

1. Request `http://<IP of Pod backend>:8080` and write the response body to
   `/opt/course/nw-connect/by-ip.txt`.
2. Request the Service by its DNS name on port 80 and write the response body to
   `/opt/course/nw-connect/by-name.txt`.

Delete the temporary Pod afterwards.
