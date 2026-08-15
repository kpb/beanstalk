# ADR 0001: Use Go for the Terminal Application

## Status

Accepted

## Date

2026-08-15

## Context

Beanstalk is a terminal-native implementation of the Beans task-file format.
It needs a portable CLI, an interactive terminal UI, and straightforward
distribution as a single executable.

## Decision

Build Beanstalk in Go.

Use Cobra for the command-line interface. Use Bubble Tea for the interactive
terminal UI when that UI is introduced.

Require Go 1.26 or newer.

## Consequences

- The CLI and TUI can be distributed as a single native binary.
- The project uses Go's standard tooling for formatting, testing, and static
  analysis (`gofmt`, `go test`, and `go vet`).
- Contributors need a Go 1.26+ development environment.
- UI work follows Bubble Tea's event-driven terminal model.
