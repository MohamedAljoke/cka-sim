# Agent Guide

## 1. Apps never import each other

Every project under `apps/` is independent. An app may import `apps/go-shared`, and nothing else from `apps/`.

- `platform` must not import `web`, `web` must not import `platform`, and so on for any app added later.
- When two apps need to agree on something (request/response types, events, API shapes), put that contract in `apps/go-shared/contract/` and have both sides import it from there.
- `go-shared` itself never imports any app.

## 2. Comments only when truly needed

The code explains itself. Use clear names and small functions instead of comments.

- Do not add comments that restate what the code does.
- Write a comment only for a non-obvious *why*: a workaround, a hidden constraint, a surprising decision.
- When in doubt, leave it out.
