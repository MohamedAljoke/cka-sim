#!/bin/sh
set -e
caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
exec cka-sim-server
