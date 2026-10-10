#!/bin/sh
set -e
# Fly boots a hybrid cgroup layout with no name=systemd hierarchy, so systemd in the kind nodes can't start.
if [ ! -d /sys/fs/cgroup/systemd ]; then
  mkdir /sys/fs/cgroup/systemd
  mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd
fi
# The root disk is an overlay, where Docker falls back to the slow vfs driver.
if ! mountpoint -q /var/lib/docker; then
  truncate -s 15G /docker.img && mkfs.ext4 -q /docker.img
  mkdir -p /var/lib/docker && mount -o loop /docker.img /var/lib/docker
fi
sysctl -q fs.inotify.max_user_watches=524288 fs.inotify.max_user_instances=512
# Fly's resolver is IPv6 only (fdaa::3), which image builds can't reach.
# The TCP socket is for the cka-sim server, over Fly's private network only: the VM has no public address.
exec dockerd-entrypoint.sh dockerd --dns 8.8.8.8 --dns 1.1.1.1 \
  --host unix:///var/run/docker.sock --host tcp://[::]:2375 --tls=false
