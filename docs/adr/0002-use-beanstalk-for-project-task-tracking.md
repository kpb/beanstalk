# ADR 0002: Use Beanstalk for Project Task Tracking

## Status

Accepted

## Date

2026-08-19

## Context

Beanstalk needs a practical task-tracking workflow for coding agents before this
repository can use it for its own work. The first agent CLI milestone provides
that workflow without requiring agents to edit task files directly.

## Decision

Use Beanstalk to track work in this repository once the first agent CLI
milestone is complete.

The milestone is complete when Beanstalk provides `prime`, JSON output for
`list`, `claim`, and `show`, and status updates through `update`. Agents use
the documented list, claim, show, and update workflow to work on tracked tasks.

## Consequences

- Project tasks are stored as version-controlled files in the compatible Beans
  on-disk format.
- Contributors and agents use Beanstalk's CLI workflow to create, claim, and
  complete tracked work.
- The repository does not adopt Beanstalk for task tracking until the first
  agent CLI milestone is complete.
