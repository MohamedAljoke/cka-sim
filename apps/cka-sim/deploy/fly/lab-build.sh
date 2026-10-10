#!/bin/sh
# Builds the cluster once dockerd is up; pool.sh runs this in each new VM before suspending it.
set -e
until docker info >/dev/null 2>&1; do sleep 1; done
cka-sim up
