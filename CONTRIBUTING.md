# Contributing

Mimic is a small, single-binary Go static analyzer.

## Adding a rule

A rule is one file in `internal/rules`. Implement the `Matcher` interface and
register it in an `init`. Rules read the AST and need no type information. Add a
focused test to `internal/rules/rules_test.go` using `matchSource` — a good rule
fires on the bad shape and stays quiet on the good one.

## Before a change

- `make lint` and `make test` must pass. CI runs both on every push.
- `gofmt` is enforced; run `make fmt`.
- The `deploy/sample` package holds deliberate bugs and is excluded from the
  linter; keep new intentional-bug examples there.

## Commits

Conventional commits (`feat:`, `fix:`, `test:`, `docs:`, `build:`, `ci:`), one
logical change each.
