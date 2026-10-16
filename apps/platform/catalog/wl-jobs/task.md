---
id: wl-jobs
title: Job and CronJob
domain: workloads
host: cka-sim-control-plane
topics: [jobs, cronjobs]
weight: 4
---
## Task

In Namespace `wl-jobs`:

1. Create a Job `batch` (image `busybox:1.36`, command `echo processed`) that must succeed `6`
   times in total, runs `2` Pods at a time, and gives up after `3` retries. Let it finish.
2. Create a CronJob `report` (image `busybox:1.36`, command `date`) that runs every 15 minutes,
   keeps the history of `2` successful and `1` failed Job, and doesn't start a new run while one
   is still running.
