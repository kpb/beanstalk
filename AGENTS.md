# Agent Instructions

## Project Context

Beanstalk is a terminal-native implementation of the Beans task-file format.
Read the architecture decisions in `docs/adr/` before making design changes.
Current decision: `docs/adr/0001-use-go-for-the-terminal-application.md`.

## First Agent Milestone

The first agent task-tracking milestone is a CLI-only workflow that does not require agents to edit task files directly:

```bash
beanstalk prime
beanstalk list --status todo --json
beanstalk claim <id> --json
beanstalk show <id> --json
beanstalk update <id> --status completed --json
```

Implement `prime` as a zero-argument command that emits a built-in agent instruction block. It must work outside an
initialized project and only describe implemented Beanstalk behavior. Do not add `prime --json` in the first cut.

When `.beanstalk.yaml` exists in the current directory and contains `prime.instructions`, emit that string instead of
the built-in block. Do not search parent directories, merge instructions, or interpolate values. Report malformed
`.beanstalk.yaml` files as errors.

The `prime` instructions must tell agents:

- Do not create beans for questions, exploration, analysis, or planning.
- Before starting untracked implementation work, ask the user whether they want a bean.
- Create a bean when the user explicitly asks to create or track an issue, task, bug, feature, epic, or milestone.
- Treat a generic "issue" as a `task`.
- Create new beans with `todo` status, then claim them before implementation.
- Use `show` for task details and `update` to record `completed` or `scrapped` outcomes.

Do not claim support for unsupported original-Beans features, including ready filtering, search, multi-ID show, body
edits, relationships, archive, etags, or GraphQL.

## Go Development

- Require Go 1.26 or newer.
- Write idiomatic, readable, and testable Go.
- Prefer small, focused functions and clear names over clever abstractions.
- Keep command wiring in `internal/commands`; place reusable task-file behavior outside command handlers.
- Preserve compatibility with the current Beans on-disk format.
- Use Cobra for CLI commands.
- Use Bubble Tea when interactive TUI work begins.

## Testing And Validation

- Add or update tests for behavioral changes.
- Run `gofmt` on modified Go files.
- Run `go test ./...` and `go vet ./...`.
- Ensure `gofmt -l .` has no output before considering work complete.

## Documentation

- Update `README.md` when user-facing CLI behavior changes.
- Add an ADR in `docs/adr/` for durable architectural decisions.
