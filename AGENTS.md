# Agent Guide

These rules win over the patterns on `main`. Use `main` as a reference for behaviour, not for structure.

## 1. Apps never import each other

Every project under `apps/` is independent. An app may import `apps/go-shared`, and nothing else from `apps/`.

- When two apps must agree on something (API shapes, events), the contract goes in `apps/go-shared/contract/`
  as a `.proto` file served and called with Connect. `make proto` lints it and regenerates the Go code, which is
  committed. Generated types stay in `api/` and adapters; a domain may use a contract's enum, not its messages.
- `go-shared` never imports any app.

## 2. Comments only when truly needed

Clear names and small functions instead of comments. Comment only a non-obvious *why* (a workaround, a hidden
constraint). When in doubt, leave it out.

## 3. App layout

```
internal/domain/<name>/           entity and its rules (tasks.Task, Validate), domain errors
internal/domain/<name>/usecases/  one use case per action: NewListTasks(repo), Execute(ctx)
internal/infra/adapters/<tech>/   repository implementations: files/ today, postgres/ later; no files at adapters/
internal/api/                     HTTP layer; takes use cases as concrete types
cmd/                              wiring
e2e/                              end-to-end tests
```

`domain` never imports `infra` or `api`.

## 4. Interfaces: at the consumer, and only when needed

- The package that calls the methods defines the interface, with only those methods (`usecases.TaskRepository`).
  Implementations return concrete types.
- Add an interface only where a second implementation exists or is planned (files or Postgres behind a
  repository), not at every layer boundary.

## 5. Errors: the domain names them, the HTTP layer answers them

- A failure the domain can name is a sentinel in its domain's `errors.go`:
  `domainerr.New("task_not_found", "no task has that id")`. Just a code and a message, never a status.
- Use cases keep the cause: `fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)`.
- Each handler maps its own use case's errors right after the call, with `errors.Is`:
  `return httpx.NotFound(err)`, `httpx.Unavailable(err)`; anything else is returned as is.
- `httpx.Problems` shows a domain error's code and message only when a handler gave it a status. Everything
  else is a 500 that never describes itself, and every 5xx is logged with its cause.

## 6. Tests

- Unit tests per layer. Fake only what sits behind an interface (repositories); use the real use case
  everywhere else.
- `e2e/` builds the real server binary, runs it on a free port and calls it over HTTP. Every endpoint gets e2e
  cases for its most important behaviour, including how it refuses bad requests.
- Test functions first, helpers and fakes at the end of the file. `make test` runs everything, e2e uncached.
