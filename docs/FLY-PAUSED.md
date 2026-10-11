# Fly is paused

Since 2026-10-11, cka-sim on Fly is switched off to save credit. Nothing runs and nothing is deleted.

## What is left on Fly

| Thing | State | Costs while paused |
|---|---|---|
| App `cka-sim-maljoke` (the site) | Machine `8576106f422d08` **stopped, autostart off**, so a visit doesn't wake it | its small root disk only |
| Its secrets (`FLY_API_TOKEN` for the labs app, `CKA_SIM_PASSWORD`) and IPs (shared IPv4, IPv6) | kept | nothing |
| App `cka-sim-labs-maljoke` (exam VMs) | **no Machines**; the pool is empty | nothing |
| Images `registry.fly.io/cka-sim-labs-maljoke:lab` and `registry.fly.io/cka-sim-maljoke:latest` | kept | nothing |

The site password is still in `~/.config/cka-sim/fly-password`.

## Turn it back on

All commands run from `apps/cka-sim/deploy/fly` on this branch.

1. **Log in.** Run `fly auth login`, then `fly apps list` should show both apps.
2. **Rebuild the lab image if the code changed since 2026-10-11** (the `cka-sim` CLI or `lab-build.sh`):
   `make image`. Skip this if nothing changed; `pool.sh` always uses the digest the tag has now.
3. **Fill the pool:** `make pool N=1`. This takes about 2.5 minutes per VM, in parallel. Each warm VM is suspended and
   costs about $3/month for its 20 GB disk. `make list` should show it `suspended  ready`.
4. **Start the site:** `make deploy`. This redeploys from `fly.toml`, which turns autostart back on. If you don't need a new
   build, `fly machine update 8576106f422d08 -a cka-sim-maljoke --autostart=true --yes` followed by
   `fly machine start 8576106f422d08 -a cka-sim-maljoke` is enough.
5. **Check it.** Open https://cka-sim-maljoke.fly.dev with the password and start a lab. The topbar should say
   "Lab ready in ~5s · Fly". `make probe ID=<the claimed VM>` should show every line blocked except `dns`.

## Pause it again

1. End any running exam or lab on the site. That destroys its VM and lock.
2. Destroy any warm VMs: `./pool.sh empty`.
3. Stop the site and keep visits from waking it:

   ```
   fly machine update 8576106f422d08 -a cka-sim-maljoke --autostart=false --yes
   fly machine stop 8576106f422d08 -a cka-sim-maljoke
   ```

   If `make deploy` created a new site Machine, take its ID from `fly machine list -a cka-sim-maljoke`.

## Remove it for good

`make down` destroys both apps, every VM, the secrets and the images. Coming back after that starts again from
`make apps` (flow 7 in `FLOWS.md`).
