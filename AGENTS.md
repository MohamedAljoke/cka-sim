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

## 3. Interfaces belong to the consumer

The package that calls the methods defines the interface, with only the methods it calls. The package that
implements it returns a concrete type and never declares the interface itself.

- Inside an app, each domain is a folder under `internal/domain` (`internal/domain/tasks`). Its root holds the
  entity and its rules (`tasks.Task`, `Validate`). Its `usecases` folder holds one use case per action
  (`ListTasks`), built with `New…` and run with `Execute`. A use case defines the repository interface it
  needs (`usecases.TaskRepository`) and gets it injected.
- `internal/infra/adapters` implements those repositories, one folder per technology and never a file at its
  root: `adapters/files` (`files.TaskRepository`) today, `adapters/postgres` later. `internal/api` is the
  HTTP layer and takes the use cases as concrete types (`*usecases.ListTasks`). `cmd` wires them together.
- `domain` never imports `infra` or `api`; both of those import `domain`.
- Abstract only when needed: add an interface where a real second implementation exists or is planned (a
  repository behind files or Postgres), not as a reflex on every layer boundary.
- Keep interfaces small: one or two methods is normal. Add a method only when a consumer calls it.

## 4. Errors: the domain names them, the HTTP layer answers them

- A failure the domain can name is a domain error, declared with `go-shared/domainerr` next to its domain
  (`tasks.ErrUnavailable` in `internal/domain/tasks/errors.go`). Its kind says what went wrong
  (`NotFound`, `Conflict`, `Unavailable`…), never an HTTP status. Mark it `.AsFault()` when it should
  also be logged.
- A use case wraps the cause so the log keeps it: `fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)`.
- Only the HTTP layer turns errors into responses: `httpx.Problems` maps each kind to a status and shows
  the domain error's message. Anything that isn't a domain error is a 500 whose cause stays in the log.
