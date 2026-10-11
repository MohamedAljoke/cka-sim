#!/bin/sh
set -e
# kind's nodes each run a kubelet; the node's defaults run out of inotify watches with three of them.
sysctl -q fs.inotify.max_user_watches=524288 fs.inotify.max_user_instances=512
# Unix socket only: the server reaches the lab through the agent, never through Docker.
dockerd-entrypoint.sh dockerd --host unix:///var/run/docker.sock &
until docker info >/dev/null 2>&1; do sleep 1; done
# The lab builds its own cluster; the agent comes last, so its health check means the lab is ready.
lab up
exec cka-sim-agent
