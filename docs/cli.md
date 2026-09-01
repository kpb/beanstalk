# CLI Guide

Beanstalk stores tasks as Markdown files with YAML front matter. The CLI reads and writes this format from the current
project directory.

## Initialize

Create the Beans directory and configuration:

```bash
beanstalk init
```

This creates `.beans/`, `.beans/.gitignore`, and `.beans.yml`. Initialization refuses to overwrite existing paths. The
generated configuration retains the Beans configuration structure for compatibility; Beanstalk currently ignores the
worktree, agent, and server settings.

## Create And Inspect Tasks

Create a task using the configured defaults:

```bash
beanstalk create "Add login"
```

Use `--status`, `--type`, `--priority`, `--body`, repeatable `--tag`, and `--parent <bean-id>` to set metadata. Pass
`--json` to return the created task for scripts.

List tasks, optionally filtering by status or type:

```bash
beanstalk list --status todo
beanstalk list --type bug --json
```

Without `--json`, list output is a sorted task tree with these columns:

| Column | Meaning |
| --- | --- |
| `ID` | Task ID |
| `S` | Status marker: `?` draft, `o` todo, `>` in-progress, `x` completed, or `-` scrapped |
| `T` | Type marker: `M` milestone, `E` epic, `B` bug, `F` feature, or `T` task |
| `TITLE` | Task title, indented with `|-` and `` `-`` tree connectors |

Filtered text output includes a matching task's ancestors to retain its tree context. `--json` preserves the existing
task schema and applies the requested filters without adding ancestor rows.

Archived tasks are excluded from `list`. Use `show <id>` to inspect an archived task by ID.

Show one task, including its Markdown body:

```bash
beanstalk show project-a1b2
beanstalk show project-a1b2 --json
```

## Update And Organize Tasks

Update status or parent metadata by ID:

```bash
beanstalk update project-a1b2 --status in-progress
beanstalk update project-a1b2 --status completed --json
beanstalk update project-a1b2 --parent project-c3d4
beanstalk update project-a1b2 --parent ""
```

Parents must reference an existing task and cannot form cycles. Passing an empty `--parent` removes the parent link.

Archive completed and scrapped tasks with an explicit bulk operation:

```bash
beanstalk archive
beanstalk archive --json
```

This moves resolved task files to `.beans/archive/` without changing their metadata or contents. Status updates never
archive tasks automatically.

## Imported Metadata Validation

Beanstalk preserves unknown front-matter keys so compatible Beans files retain metadata it does not use. It validates the
status, type, and priority values it supports, plus duplicate IDs and parent links. Invalid values, missing parents,
and parent cycles prevent `list`, `milestones`, and the TUI from loading the project with the same diagnostic.

Claim a todo task atomically before working on it:

```bash
beanstalk claim project-a1b2 --json
```

## Milestone Progress

Show active milestones and progress from their leaf descendants:

```bash
beanstalk milestones
beanstalk milestones --all --json
```

The default includes milestones with `todo`, `draft`, or `in-progress` status. `--all` also includes completed and
scrapped milestones. Completed and scrapped leaf tasks are both resolved work; JSON output reports their counts
separately.

## Agent Workflow

Print the built-in instructions for coding agents:

```bash
beanstalk prime
```

This also works outside an initialized project. To provide project-specific instructions, create `.beanstalk.yaml` in
the current directory with a `prime.instructions` string.

An agent can list available work, claim it, inspect it, and record the outcome:

```bash
beanstalk list --status todo --json
beanstalk claim project-a1b2 --json
beanstalk show project-a1b2 --json
beanstalk update project-a1b2 --status completed --json
```

## Supported Scope

Beanstalk intentionally does not implement ready filtering, search, multi-ID show, body edits, relationships beyond a
single parent link, etags, GraphQL, web/server features, agent plugins, or worktree automation.
