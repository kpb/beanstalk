# Agent Instructions

## Project Context

Beanstalk is a terminal-native implementation of the Beans task-file format.
Read the architecture decisions in `docs/adr/` before making design changes.
Current decision: `docs/adr/0001-use-go-for-the-terminal-application.md`.

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
