#!/bin/sh
# Builds the cluster once dockerd is up; pool.sh runs this in each new VM before suspending it.
set -e
until docker info >/dev/null 2>&1; do sleep 1; done
cka-sim up | tee /tmp/cka-sim-up.log
# Locally a node can pull a missing task image later; a claimed VM is locked off the internet, so
# the image warning from cka-sim up fails the build here.
if grep -q '^warning:' /tmp/cka-sim-up.log || ! grep -q 'loading the task images' /tmp/cka-sim-up.log; then
  echo "task images are not in the nodes" >&2
  exit 1
fi
