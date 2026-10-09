# cka-sim

## Explaining a flow

When the dev asks how something works ("what happens when…", "big picture", "how does X reach Y"), read
`docs/FLOWS.md` first and answer in its shape. Don't answer with a list of components.

1. **Trigger:** the one thing the dev does (`make dev`, opening the page, clicking a task).
2. **Hops, in order:** what runs next and where it goes. Name the command, file (`path:line` when useful),
   endpoint or container at each step, and give the reason behind any surprising step in half a sentence.
3. **What they see:** the result on screen or in the terminal, including the failure case.

Use a few short paragraphs, one per stage. Read the code to confirm each hop before you state it, and
say plainly which parts are not built yet.

If the flow isn't in `docs/FLOWS.md` yet, or the code no longer matches it, offer to add or fix it there in
the same shape after answering.

## Architecture docs (EventCatalog)

`docs/eventcatalog` documents how the parts talk to each other: services, data stores, and flows such as
`QuizAttemptLifecycle`. The prose behind it lives in `docs/HOSTING.md` and `docs/ORCHESTRATION.md`.

When a change significantly alters a flow (a new step or component, a different path between parts, a
moved responsibility), then once the work is done and the dev has approved it, **ask the dev whether to
update the EventCatalog docs**. Don't update them unprompted, and don't ask mid-task.

Before editing the catalog, read `docs/eventcatalog/AGENTS.md`. Afterwards, check with `npm run lint` and
`npm run build` in `docs/eventcatalog`.
