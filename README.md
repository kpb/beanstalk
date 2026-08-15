<img src="beanstalk.jpg" alt="Beanstalk" width="720">

[![License](https://img.shields.io/github/license/kpb/beanstalk)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/kpb/beanstalk)](https://github.com/kpb/beanstalk/releases/latest)
[![Tests](https://github.com/kpb/beanstalk/actions/workflows/test.yml/badge.svg)](https://github.com/kpb/beanstalk/actions/workflows/test.yml)
[![Go version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](go.mod)

# Beanstalk

Beanstalk is an independent, terminal-native implementation of the [Beans](https://github.com/hmans/beans) task
format. It stores tasks as Markdown files with YAML front matter, keeping project work readable, reviewable, and version
controlled alongside source code. It works equally well for an individual developer, a team, and coding agents.

The project is in its initial development stage. The first release will provide a Cobra-based CLI and a first-class
Bubble Tea terminal UI, with text and JSON output. It will read and write the current Beans on-disk format without
bundling Beans' web, server, GraphQL, agent, or worktree features.

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

## Development

Requires Go 1.26 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/beanstalk version
```

## License

Beanstalk is licensed under the GNU General Public License, version 3. See [LICENSE](LICENSE).
