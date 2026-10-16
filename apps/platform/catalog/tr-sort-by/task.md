---
id: tr-sort-by
title: Sort Pods by age and restarts
domain: troubleshooting
host: cka-sim-control-plane
topics: [kubectl, sort-by, jsonpath]
weight: 3
---
## Context

Namespace `tr-sort-by` runs four Pods. Some of them keep restarting.

## Task

1. Write a command to `/opt/course/tr-sort-by/by-age.sh` that lists the Pods of **all** Namespaces
   sorted by their creation time, oldest first. Running `bash /opt/course/tr-sort-by/by-age.sh`
   must print that list.
2. Write the name of the Pod in Namespace `tr-sort-by` with the most container restarts to
   `/opt/course/tr-sort-by/restarts.txt`.
