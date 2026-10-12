---
id: wl-cronjob-fix
title: CronJob that never runs
domain: workloads
host: cka-sim-control-plane
topics: [cronjobs, jobs, troubleshooting]
weight: 4
---
## Context

CronJob `cleanup` in Namespace `wl-cronjob-fix` should run every 10 minutes, but it has never
created a Job.

## Task

1. Fix CronJob `cleanup` so that it runs every 10 minutes from now on.
2. Don't wait for the schedule: start one run right away as a Job named `cleanup-manual` created
   from the CronJob, and make sure it completes.
