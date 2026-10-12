#!/bin/sh
set -e
# kind's nodes each run a kubelet; the node's defaults run out of inotify watches with three of them.
sysctl -q fs.inotify.max_user_watches=524288 fs.inotify.max_user_instances=512
# A restarted container keeps /var/run, and dockerd then waits for the containerd a stale pid names.
rm -f /var/run/docker.pid /var/run/docker/containerd/containerd.pid
# Unix socket only: the server reaches the lab through the agent, never through Docker.
dockerd-entrypoint.sh dockerd --host unix:///var/run/docker.sock &
dockerd=$!
until docker info >/dev/null 2>&1; do
  kill -0 "$dockerd" 2>/dev/null || { echo "dockerd exited" >&2; exit 1; }
  sleep 1
done
# The lab builds its own cluster; the agent comes last, so its health check means the lab is ready.
lab up
exec cka-sim-agent
