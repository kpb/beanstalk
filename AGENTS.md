# Agent Instructions

## Project Context

Beanstalk is a terminal-native implementation of the Beans task-file format.
Read the architecture decisions in `docs/adr/` before making design changes.
Current decisions: `docs/adr/0001-use-go-for-the-terminal-application.md` and
`docs/adr/0002-use-beanstalk-for-project-task-tracking.md`.

**IMPORTANT**: Before you do anything else, run `beanstalk prime` and follow its output.

## Task Tracking

This repository dogfoods Beanstalk for task tracking. Follow the policy emitted by `beanstalk prime` when deciding
whether to create, claim, complete, or scrap beans. Claim an existing `todo` bean before implementation:

```bash
beanstalk list --status todo --json
beanstalk claim <id> --json
beanstalk show <id> --json
beanstalk update <id> --status completed --json
```

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
