<img src="beanstalk.jpg" alt="Beanstalk" width="720">

# Beanstalk

Beanstalk is an independent, terminal-native implementation of the [Beans](https://github.com/hmans/beans) task
format. It stores tasks as Markdown files with YAML front matter, keeping project work readable, reviewable, and version
controlled alongside source code. I works equally well for an individual developer, a team, and coding agents.

The project is in its initial development stage. The first release will provide a Cobra-based CLI and a first-class
Bubble Tea terminal UI, with text and JSON output. It will read and write the current Beans on-disk format without
bundling Beans' web, server, GraphQL, agent, or worktree features.

## Why?

I enjoy using the Beans project. But, the TUI is missing some features that I'd like and the project has hinted that the
TUI is not a priority and it may move in another direction entirely. Rather than fork the project, I decided to create a
TUI forward version.

## Development

Requires Go 1.26 or newer.

```bash
go test ./...
go vet ./...
go run ./cmd/beanstalk version
```

## License

Beanstalk is licensed under the GNU General Public License, version 3. See [LICENSE](LICENSE).
