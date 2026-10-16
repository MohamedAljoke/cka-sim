# Agent Guide

These rules win over the patterns on `main`. Use `main` as a reference for behaviour, not for structure.

## 1. Apps never import each other

Every project under `apps/` is independent. An app may import `apps/go-shared`, and nothing else from `apps/`.

- When two apps must agree on something (API shapes, events), the contract goes in `apps/go-shared/contract/`.
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

- A failure the domain can name is a `go-shared/domainerr` error in its domain's `errors.go`
  (`tasks.ErrUnavailable`). Its kind says what went wrong, never an HTTP status. `.AsFault()` also logs it.
- Use cases keep the cause: `fmt.Errorf("%w: %w", tasks.ErrUnavailable, err)`.
- Only `httpx.Problems` turns errors into responses: kind → status, with the domain error's message. Anything
  else is a 500 whose cause stays in the log.

## 6. Tests

- Unit tests per layer. Fake only what sits behind an interface (repositories); use the real use case
  everywhere else.
- `e2e/` builds the real server binary, runs it on a free port and calls it over HTTP. Every endpoint gets e2e
  cases for its most important behaviour, including how it refuses bad requests.
- Test functions first, helpers and fakes at the end of the file. `make test` runs everything, e2e uncached.
