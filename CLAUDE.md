# cka-sim

## Architecture docs (EventCatalog)

`docs/eventcatalog` documents how the parts talk to each other: services, data stores, and flows such as
`QuizAttemptLifecycle`. The prose behind it lives in `docs/HOSTING.md` and `docs/ORCHESTRATION.md`.

When a change significantly alters a flow (a new step or component, a different path between parts, a
moved responsibility), then once the work is done and the dev has approved it, **ask the dev whether to
update the EventCatalog docs**. Don't update them unprompted, and don't ask mid-task.

Before editing the catalog, read `docs/eventcatalog/AGENTS.md`. Afterwards, check with `npm run lint` and
`npm run build` in `docs/eventcatalog`.
