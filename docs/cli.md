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

If writing either file fails, initialization removes the files it created and removes `.beans/` only if it is empty.
Unrelated contents are left alone.

## Task-File Format

Each task is a `.md` file in the configured `beans.path` directory (`.beans/` by default). For example,
`.beans/project-a1b2--add-login.md` could contain:

```md
---
# project-a1b2
title: Add login
status: todo
type: task
priority: normal
tags:
  - auth
created_at: 2026-10-01T12:00:00Z
updated_at: 2026-10-01T12:00:00Z
---
Implement the login flow.
```

The opening and closing `---` delimit YAML front matter from the Markdown body. The ID comes from the filename before
`--` (or before `.md` if there is no slug); the text after `--` is the slug. The `# project-a1b2` comment is for readers,
not the ID source. `path` in JSON is the file path relative to `beans.path`, not a front-matter field. New IDs use the
configured `beans.prefix` and `beans.id_length` (four characters by default).

| Front-matter field | Supported values and defaults |
| --- | --- |
| `title` | Required, nonempty string; `create` uses `Untitled` when no title is given. |
| `status` | Required for imported files. `draft`, `todo`, `in-progress`, `completed`, `scrapped`; `create` uses `beans.default_status`, falling back to `todo`. |
| `type` | `milestone`, `epic`, `bug`, `feature`, `task`; omitted values load as `task`. `create` uses `beans.default_type`, falling back to `task`. |
| `priority` | `critical`, `high`, `normal`, `low`, `deferred`; omitted values load as `normal`, also the `create` default. |
| `tags` | List of strings; omitted values load as an empty list. |
| `parent` | Optional single parent ID; no parent if omitted. Parent IDs must exist and cannot form cycles. |
| `created_at`, `updated_at` | Timestamps; `create` sets both to the current UTC second. Missing `created_at` loads from the file's modification time (UTC, to the second); missing `updated_at` loads from `created_at`. |

Resolved tasks are moved to `beans.path/archive/<filename>` by `archive`, retaining their filename, front matter, and
body. Location, not status, determines whether a task is archived. Archived tasks remain available to `show` but are
excluded from `list` and cannot be updated or claimed.

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

Filtered text output includes a matching task's ancestors to retain its tree context. `--json` uses the
[bean object schema](#json-output-contracts) and applies the requested filters without adding ancestor rows.

Archived tasks are excluded from `list`. Use `show <id>` to inspect an archived task by ID.

Show one task, including its Markdown body:

```bash
beanstalk show project-a1b2
beanstalk show project-a1b2 --json
```

## Update And Organize Tasks

Update status, parent metadata, or the Markdown body by ID:

```bash
beanstalk update project-a1b2 --status in-progress
beanstalk update project-a1b2 --status completed --json
beanstalk update project-a1b2 --parent project-c3d4
beanstalk update project-a1b2 --parent ""
beanstalk update project-a1b2 --body "Document the required behavior."
beanstalk update project-a1b2 --body ""
```

Parents must reference an existing task and cannot form cycles. Passing an empty `--parent` removes the parent link.
`--body` replaces the body verbatim; passing an empty value clears it. Flags can be combined in one atomic update.
Updates (including `claim` and TUI status changes) parse and re-marshal the YAML front matter. Unknown metadata keys and
values remain available, but original YAML quoting, spacing, comments attached to changed values, and line endings are
not guaranteed to survive; the rewritten front matter uses LF line endings. An unchanged Markdown body retains its
original bytes, including CRLF line endings. Use `--body` when replacing the body intentionally.

Archive completed and scrapped tasks with an explicit bulk operation:

```bash
beanstalk archive
beanstalk archive --json
```

This moves resolved task files to `.beans/archive/` without changing their metadata or contents. Status updates never
archive tasks automatically. Archived tasks are read-only: `update` and `claim` reject them, even if an imported archived
task has an active status. To make an archived task active again, move its file out of `.beans/archive/` manually.

## Shell Completion

Generate a completion script for Bash, Zsh, Fish, or PowerShell:

```bash
source <(beanstalk completion bash)
beanstalk completion zsh > "${fpath[1]}/_beanstalk"
beanstalk completion fish > ~/.config/fish/completions/beanstalk.fish
beanstalk completion powershell | Out-String | Invoke-Expression
```

## Imported Metadata Validation

Beanstalk preserves unknown front-matter values so compatible Beans files retain metadata it does not use. It validates the
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

## JSON Output Contracts

Commands with `--json` return one JSON value on success. A bean object has these fields:

| Field | Meaning |
| --- | --- |
| `id`, `path`, `title`, `status`, `created_at`, `updated_at` | Always present; `path` is relative to `beans.path`, and timestamps are RFC 3339 strings. |
| `slug`, `type`, `priority`, `tags`, `parent`, `body` | Omitted when empty. Loaded beans have defaulted `type` and `priority`, so those fields are present for valid loaded tasks. `tags` is an array when nonempty; `parent` is an ID. |

`body` is the Markdown content after the front matter, not a YAML field. In the shapes below, `<bean>` stands for a bean
object and `N` for a count. Success responses vary by command:

| Command | JSON result |
| --- | --- |
| `create <title> --json` | `{"success":true,"bean":<bean>,"message":"Bean created"}`; the bean includes the effective default priority. |
| `list --json` | Array of active bean objects, sorted and filtered by `--status`/`--type`. No ancestor-only rows and **no `body` field**, even when a task has a body; no matches return `[]`. |
| `show <id> --json` | One bean object, without a wrapper. Includes `body` when nonempty, including for archived beans. |
| `update <id> ... --json` | `{"success":true,"bean":<bean>,"message":"Bean updated"}` with the updated bean and its body when nonempty. |
| `claim <id> --json` | `{"success":true,"bean":<bean>,"message":"Bean claimed"}` with `status` set to `in-progress` and its body when nonempty. |
| `archive --json` | `{"success":true,"beans":[<bean>,...],"message":"Archived N beans"}`. Each returned bean has an `archive/` path and retains its body when nonempty; with nothing to archive, `beans` is `null` and the message is `Archived 0 beans`. |
| `milestones --json` | Array of milestone progress objects (empty array if none). `--all` also includes resolved milestones. |

Each milestone progress object has `id`, `title`, `status`, `total`, `resolved`, `completed`, `scrapped`, `percent`, and
`statuses`. The `statuses` object has numeric `todo`, `draft`, `in_progress`, `completed`, and `scrapped` counts for leaf
descendants. `resolved` counts completed plus scrapped leaves, and `percent` is the integer resolved percentage (zero
when there are no leaves).

`init`, `prime`, `version`, and `completion` do not offer `--json`: they print text or a shell script. The default TUI
and `tui` are interactive rather than JSON commands. On failure, commands return a nonzero exit status and print a
diagnostic to stderr; `--json` does not provide a JSON error wrapper.

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

Beanstalk intentionally does not implement ready filtering, search, multi-ID show, relationships beyond a single parent
link, etags, GraphQL, web/server features, agent plugins, or worktree automation.
