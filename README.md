<img src="beanstalk.jpg" alt="Beanstalk" width="720">

[![License](https://img.shields.io/github/license/kpb/beanstalk)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/kpb/beanstalk)](https://github.com/kpb/beanstalk/releases/latest)
[![Tests](https://github.com/kpb/beanstalk/actions/workflows/test.yml/badge.svg)](https://github.com/kpb/beanstalk/actions/workflows/test.yml)
[![Go version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)

# Beanstalk

Beanstalk is an independent, terminal-native implementation of the [Beans](https://github.com/hmans/beans) task
format. It stores tasks as Markdown files with YAML front matter, keeping project work readable, reviewable, and version
controlled alongside source code. It works equally well for an individual developer, a team, and coding agents.

Beanstalk currently provides a Cobra-based CLI with text and JSON output. An interactive Bubble Tea terminal UI is
planned. It reads and writes the current Beans on-disk format without bundling Beans' web, server, GraphQL, agent, or
worktree features.

## Install

Download the archive for your platform from the [latest release](https://github.com/kpb/beanstalk/releases/latest),
extract it, and place `beanstalk` on your `PATH`. Verify the download with the release's `checksums.txt` file.

```bash
beanstalk version
```

## Why?

I enjoy using the Beans project. But, the TUI is missing some features that I'd like and the project has hinted that the
TUI is not a priority and it may move in another direction entirely. Rather than fork the project, I decided to create a
TUI forward version.

## Use

Initialize a project with the current Beans on-disk layout:

```bash
beanstalk init
```

This creates `.beans/`, `.beans/.gitignore`, and `.beans.yml`. It refuses to overwrite an existing path.

The generated `.beans.yml` includes the current Beans configuration structure for compatibility. Beanstalk currently
ignores the worktree, agent, and server settings.

Print the built-in instructions for coding agents:

```bash
beanstalk prime
```

This also works outside an initialized project. To use project-specific instructions, create `.beanstalk.yaml` in the
current directory with a `prime.instructions` string; Beanstalk emits that string instead of the built-in instructions.

Create a task using the configured defaults:

```bash
beanstalk create "Add login"
```

Use `--status`, `--type`, `--priority`, `--body`, and repeatable `--tag` flags to set initial metadata. Pass `--json`
to emit the created bean as JSON for scripts and coding agents.

List tasks, optionally filtering by status or type:

```bash
beanstalk list --status todo
```

Pass `--json` to emit a JSON array for scripts and coding agents.

Show a task, including its Markdown body:

```bash
beanstalk show project-a1b2
beanstalk show project-a1b2 --json
```

Update a task's status by ID:

```bash
beanstalk update project-a1b2 --status in-progress
beanstalk update project-a1b2 --status completed --json
```

An agent can list available work with `beanstalk list --status todo --json`, then atomically claim it before working:

```bash
beanstalk claim project-a1b2 --json
beanstalk show project-a1b2 --json
beanstalk update project-a1b2 --status completed
```

## Development

Requires Go 1.26 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/beanstalk version
```

## License

Beanstalk is licensed under the GNU General Public License, version 3. See [LICENSE](LICENSE).
