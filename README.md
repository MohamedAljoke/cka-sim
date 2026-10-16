# cka-sim

A CKA practice simulator: pick a task from the catalog, start a real Kubernetes cluster, solve it.

`make` lists the commands; `cp .env.example .env` first. Rules for working in this repo are in [AGENTS.md](AGENTS.md).

## Docs

How the apps fit together and a deep dive per part (starting with the lab container) live in
[cka-sim docs](https://claude.ai/code/artifact/5f47a2bd-5b68-4d15-b90f-1d76ae14750d).

The rebuild follows [docs/REBUILD-IN-STEPS.md](docs/REBUILD-IN-STEPS.md), brought over from `main`;
[docs/REBUILD-STATUS.md](docs/REBUILD-STATUS.md) tracks which steps have landed on this branch.

When a change alters what they describe (an app, a port, a make target, the lab's image or startup), update
the docs in the same piece of work.
