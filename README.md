<img src="beanstalk.jpg" alt="Beanstalk" width="720">

[![License][license-badge]][license]
[![Latest release][release-badge]][latest-release]
[![Tests][tests-badge]][tests-workflow]
[![Go version][go-badge]][go-mod]

# Beanstalk

Beanstalk is an independent, terminal-native implementation of the [Beans][beans] task
format. It stores tasks as Markdown files with YAML front matter, keeping project work readable, reviewable, and version
controlled alongside source code. It works equally well for an individual developer, a team, and coding agents.

Beanstalk provides a Cobra-based CLI with text and JSON output, plus a read-only Bubble Tea terminal UI. It reads and
writes the current Beans on-disk format without bundling Beans' web, server, GraphQL, agent, or worktree features.

## Architecture Decisions

See [Architecture Decision Records][architecture-decisions] for the project's durable technical decisions.

## Install

Install the latest release on Linux or macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/kpb/beanstalk/main/install.sh | sh
```

The installer verifies the release archive against `checksums.txt` and installs to `$HOME/.local/bin` by default. Ensure
that directory is on your `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Set `INSTALL_DIR` to a writable directory to install elsewhere, or `VERSION` to install a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/kpb/beanstalk/main/install.sh | INSTALL_DIR=/usr/local/bin sh
curl -fsSL https://raw.githubusercontent.com/kpb/beanstalk/main/install.sh | VERSION=v0.1.0 sh
```

The installer requires `curl`, `tar`, and either `sha256sum` or `shasum`. On Windows, download the ZIP archive for your
platform from the [latest release][latest-release], extract it, and place
`beanstalk.exe` on your `PATH`.

```bash
beanstalk version
```

## Agent Setup

Add the following instruction to your project's `AGENTS.md`, `CLAUDE.md`, or equivalent agent guidance file:

```md
**IMPORTANT**: Before you do anything else, run `beanstalk prime` and follow its output.
```

`beanstalk prime` provides the implemented task-tracking workflow for coding agents. Beanstalk does not provide agent
plugins or automatic hooks.

## Why?

I/Kenneth enjoy using the Beans project but there are some missing features that are important to me.

The [Beans][beans] TUI is not a priority, and the project has hinted that it may move in
another direction entirely. The Beanstalk TUI will be a first-class feature, allowing users to manage their tasks
outside agent workflows.

Beyond the TUI, I want task tracking to work naturally with coding agents. Beanstalk provides a local, scriptable CLI
with JSON output and atomic claims, while `prime` gives agents project-specific guidance that teams can override via
configuration. The goal is a focused native tool that keeps task history readable and reviewable alongside the code.

A fresh implementation is simpler to evolve than maintaining a long-lived fork. Beanstalk preserves compatibility with
the Beans task-file format while allowing its terminal, agent, and TUI workflows to develop independently.

## Use

### CLI

Use the CLI to create, organize, inspect, and update Beans-format tasks. It provides text and JSON output for people,
scripts, and coding agents.

```bash
beanstalk init
beanstalk list --status todo
```

[Read the CLI guide][cli-guide] for commands, task hierarchies, milestone progress, JSON output, and agent workflow.

### TUI

Use the read-only TUI to browse the same task list interactively.

```bash
beanstalk tui
```

[Read the TUI guide][tui-guide] for navigation keys, displayed task data, and terminal behavior.

## Development

Requires Go 1.26 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/beanstalk version
```

## License

Beanstalk is licensed under the GNU General Public License, version 3. See [LICENSE][license].

<!-- ref links ordered alphabetically -->
[architecture-decisions]: docs/adr/
[beans]: https://github.com/hmans/beans
[cli-guide]: docs/cli.md
[go-badge]: https://img.shields.io/badge/Go-1.26-00ADD8?logo=go
[go-mod]: go.mod
[latest-release]: https://github.com/kpb/beanstalk/releases/latest
[license-badge]: https://img.shields.io/github/license/kpb/beanstalk
[license]: LICENSE
[release-badge]: https://img.shields.io/github/v/release/kpb/beanstalk
[tests-badge]: https://github.com/kpb/beanstalk/actions/workflows/test.yml/badge.svg
[tests-workflow]: https://github.com/kpb/beanstalk/actions/workflows/test.yml
[tui-guide]: docs/tui.md
