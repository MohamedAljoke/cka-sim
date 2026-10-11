# Next on Fly: metrics, suspend data, network lockdown

The `fly` branch runs labs from a pool of suspended Fly Machines ([HOSTING.md](HOSTING.md), flow 7 in
[FLOWS.md](FLOWS.md)). This plan covers the next three pieces of work, in this order:

1. **Metrics.** Measure labs from the start, so that every later change comes with numbers.
2. **Suspend data.** Find out how reliable resume really is on Fly.
3. **Network lockdown.** Make one lab unable to reach the site, other labs or the rest of the organisation.

Not now: automatic pool refill (a pool controller with Machine leases) and labs in several regions with `fly-replay`.
See [Later](#later).

## Where code goes

The engine must work the same locally, on Fly, or on GKE later. So:

| Goes on `main` | Goes on `fly` |
|---|---|
| Domain concepts: lab, box, start, end, script run, terminal | Machines API calls, pool metadata, suspend/resume |
| Small interfaces the engine owns (`sandbox.Provider`, `tasks.Runner`, `api.Opener`, the metrics hook below) | `internal/fly`, `deploy/fly/*`, `fly.toml` |
| The in-VM agent and its client (any platform needs them) | How a Fly VM gets its agent token |
| Provider-neutral metric names (`lab_*`, `script_*`) | Fly-only metrics (`fly_*`) |

Rules:
- No Fly word in `sandbox`, `server`, `exam`, `runner`, `terminal` or `tasks`: no "machine", "suspend", "6PN" or
  "region". A provider's name is only data (the `provider` label, `Status.Provider`).
- A concept that GKE would share, such as "this lab came from a warm pool", becomes neutral only once a second
  provider needs it. Until then it stays in `internal/fly`.
- Land the neutral part on `main` first, then bring `main` into `fly` and build the Fly part on top.

## 1. Metrics

### On `main`

- **A separate metrics listener**, `CKA_SIM_METRICS_ADDR`. It is off by default. It is not behind the site
  password and is never public.
- **A hook owned by the lab.** `sandbox` defines a small interface. `Lab` calls it when a start succeeds or fails and
  when a lab ends. A `metrics` package implements it with Prometheus, and tests use a fake. `Lab` already measures the start
  time (`Status.Started`/`Ready`), so this is mostly wiring.
- **Neutral metrics:**

| Metric | Type | Labels | Answers |
|---|---|---|---|
| `cka_sim_lab_start_seconds` | histogram | `provider`, `outcome` (ready/failed) | How long does a user wait for a lab? |
| `cka_sim_labs_active` | gauge | `provider` | What are we paying for right now? |
| `cka_sim_lab_lifetime_seconds` | histogram | `provider` | How long do people keep a lab? |
| `cka_sim_script_seconds` | histogram | `kind` (setup/check/solution), `outcome` | Are checks slow or failing? |
| `cka_sim_terminals_open` | gauge | — | Are labs in use, or idle? |

Locally this works too: run `curl localhost:<port>/metrics` after starting a lab.

### On `fly`

- Add a `[metrics]` section to `fly.toml` that points at the metrics listener. Fly scrapes it, and the graphs appear in Fly's
  hosted Grafana alongside its own per-VM metrics: CPU, memory, network. The per-VM memory graph feeds directly into
  the 2 GB question below.
- **Fly-only metrics in `internal/fly`:**

| Metric | Type | Labels | Answers |
|---|---|---|---|
| `cka_sim_fly_claim_total` | counter | `outcome` (restored/cold_boot/api_error) | How often does resume fail us? |
| `cka_sim_fly_resume_seconds` | histogram | — | `start` call → cluster healthy |
| `cka_sim_fly_pool_ready` | gauge | — | Warm VMs left, sampled on each claim and every minute |

- **Telling a restored VM from a cold boot.** A kind node shares the VM's kernel, so `/proc/uptime` read through the
  box is the VM's uptime. After a restore it is at least the time since the VM was built. After a cold boot it is a few
  seconds. This gives an exact label instead of inferring it from a failed health check.
- One Grafana dashboard: time-to-lab p50/p99, claim outcomes, active labs, warm VMs left. Its screenshot goes in
  HOSTING.md.

## 2. Suspend data (`fly` only)

Fly's [suspend docs](https://docs.fly.io/reference/suspend-resume) warn about two things we already do:
- suspend is "not currently recommended for large machine memory sizes (> 2 GB)", and our labs use 4 GB;
- "suspending many machines at once is not recommended", and `pool.sh fill` suspends in parallel.

They also say a snapshot "can be discarded, for example, during a deploy". We handle a cold boot by destroying the VM, but we
don't know how often it happens. This experiment answers that, using the metrics above:

| Variable | Values |
|---|---|
| Memory | 4 GB (today), 2 GB |
| Time parked before the claim | 5 min, 1 h, 24 h, 7 days |
| How the pool was filled | all suspended at once, staggered |
| Trials per cell | at least 10 |

Results to record for each cell: restored vs cold boot, and resume p50/p99.

**Fitting into 2 GB.** `apps/lab/internal/cluster/kind.yaml` has a control plane and two workers. Tasks such as
`wl-taints`, `wl-node-affinity` and `tr-kubelet` need workers. Try, in order:
1. Measure what the three nodes actually use, with `docker stats` and Fly's memory graph.
2. Trim it: lower the etcd and apiserver caches, drop anything kind adds that we don't use.
3. Drop to one worker, and check which tasks still make sense with only one.

The output is a results table in HOSTING.md. If 2 GB restores reliably and 4 GB doesn't, that is the case for
changing the lab size.

## 3. Network lockdown

**Premise.** The user is root in a privileged kind node, so they can be treated as root on the VM. Anything enforced
inside the VM can be undone. The walls that matter sit outside the VM.

### What a lab can reach today

| From a lab, the user can reach | Why | Severity |
|---|---|---|
| dockerd on every other lab (`:2375`, no TLS, no auth) | All labs share the organisation's private network | Root on other users' labs |
| Every other app in the organisation | Default private network | Depends on what else runs there |
| The site's internal ports | Same network | Low today, more once there's a database |
| The internet, without limits | No egress rules | Abuse: scanning, mining, spam |
| Fly's in-VM API (`/.fly/api`, `_api.internal`) | Present on every Machine | Unknown, to test |

### Steps

1. **Agent instead of dockerd over TCP** (on `main`). `cmd/agent` runs in the box and exposes `shell` and `run-script`,
   which are today's `DockerOpener` and `runner.Node` moved behind a small HTTP/websocket API. It accepts one bearer token.
   An agent client implements `tasks.Runner` and `api.Opener`, so `Box` doesn't change shape. Local mode can
   keep calling Docker directly.
2. **One token per VM** (on `fly`). Since the user is root in their VM, any secret in it is theirs. So a secret
   shared across the pool would let lab A claim lab B. Each VM is therefore created with its own random token in its env, and the same
   token is stored in its Machine metadata, where only the server reads it (the VM holds no Fly API token). dockerd then
   listens on the Unix socket only.
3. **Labs on their own network** (on `fly`). Create the labs app with a custom private network
   (`"network": "cka-sim-labs"` in the create-app call). The site has to reach the labs, so it joins the same network. Labs
   then can't reach anything else in the organisation.
4. **Lab to lab** (on `fly`). Machines in one app share a network, so step 3 alone still lets lab A reach lab B.
   The egress policy in step 5 closes it: a locked lab can't open a connection to another lab. One app per lab isn't
   needed. The per-VM token stays as a second wall, in case a policy is missing or late.
5. **Egress** (on `fly`). Fly enforces [network policies](https://docs.fly.io/machines/guides-examples/network-policies)
   outside the VM, set per app through the Machines API (`POST /v1/apps/<app>/network_policies`):
   - A `selector` picks Machines: `all`, specific IDs, or `metadata` key/value pairs, so `pool=building` and
     `pool=claimed` can get different policies.
   - A rule is only `allow`, plus `ingress` or `egress`, plus a list of `{protocol, port}`. Once a direction has a rule,
     everything else in that direction is denied.
   - There is no destination field, so "the site but not other labs" can't be expressed.
   - There is no deny-all switch, because an empty rule list means no restriction. The
     [community workaround](https://community.fly.io/t/machines-network-policies-can-egress-rules-cover-icmp-and-the-internal-dns-resolver-and-is-self-test-traffic-ok/28776)
     is a single dummy rule, such as TCP port 0.
   - ICMP and Fly's internal DNS (`fdaa::3`) can't be named in a rule.
   - Changes need a restart or redeploy to take effect.

   A lab needs the internet only while it is built (`cka-sim up` pulls kindest/node). After that, every task image
   is already in the nodes (`cluster.Images`, kept complete by `catalog_test.go`). So the plan is:

   | Machines | Policy |
   |---|---|
   | `pool=building` | egress TCP 80 and 443 only, to pull images |
   | `pool=claimed` | no egress at all (the dummy rule) |

   The cost is that `kubectl run --image=<anything not preloaded>` ends in `ImagePullBackOff`. The real exam has internet
   access. A middle ground is to keep 443 open for claimed labs too.

   **Tested 2026-10-10** in `gru`, with two throwaway apps on a custom network. A small web server on `:2375` stood in for
   dockerd, and the policy was a single dummy "allow TCP port 0" egress rule:

   | Question | Answer |
   |---|---|
   | Does a deny cover the private network, not just the internet? | **Yes.** A locked lab couldn't reach the site or other labs, over IPv4 or IPv6. |
   | Can the server still reach a locked lab? | **Yes.** The site → lab `:2375` worked, so connections are tracked. |
   | Does a new policy reach running VMs? | **Yes, with no restart**, despite what the docs say. |
   | Does it last through a stop/start, or a suspend/resume? | **Yes, both.** |
   | Does a VM created after the policy get it? | **Yes.** |
   | Does changing a VM's metadata to `claimed` apply it? | **No**, whether the VM was running or suspended. It stayed fully open even 30s later. Policies are matched when the policy is written or the VM is created or started, not when its metadata changes. |
   | Is Fly's DNS (`fdaa::3`) blocked? | **No**, it stays reachable, as the community post suspected. |

   So **the metadata selector doesn't fit our claim flow**: a pool VM is suspended as `ready` and only becomes
   `claimed` later. What works instead is **writing a policy that selects the VM by its ID at claim time**
   (`"selector": {"machines": [{"id": "<id>"}]}`). It took effect straight away on a resumed VM and lasted through
   another suspend/resume. Claiming becomes "resume → add the VM to the policy → health check", and destroying the VM
   removes its rule.

   This also means a lab-to-lab wall doesn't need one app per lab. A locked lab can't open a connection to anything, other labs included.
   Labs that aren't locked yet (`building` or `ready`) hold no user. The one remaining path is DNS to `fdaa::3`.

   The probe scripts from the test should move into `deploy/fly/` once the policy is wired in, so the claim flow can be
   re-tested.
6. **Prove it.** Write a script that runs from inside a lab: try to reach another lab's agent and dockerd, the site, another app
   in the organisation, and the in-VM Fly API. Every attempt must fail, and the output goes in HOSTING.md. This is the same test
   as step 3.3 of the GKE plan, so both platforms are judged the same way.

   With no policy, every path was open from a lab: the internet over IPv4 and IPv6, `fdaa::3`, the site's port and another
   lab's port. With an ID-selected policy, everything except `fdaa::3` was blocked (step 5).

## Later

- **Pool controller.** Replace `pool.sh` with a reconcile loop: keep N warm VMs, refill, reap. Use Machine leases
  (`POST /machines/{id}/lease`, `fly-machine-lease-nonce`) for claims, so that two servers can't take the same VM.
- **Several regions.** Keep a pool per region, give each user a lab near them, and use `fly-replay: instance=` to pin a
  user's requests to the server that owns their exam.
- **Write-up.** The cgroup/vfs/DNS quirks, the suspend results and the lockdown test as a community.fly.io post.
