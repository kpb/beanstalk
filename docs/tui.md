# TUI Guide

Launch the read-only task browser from an initialized Beans project:

```bash
beanstalk tui
```

The TUI loads the configured Beans directory and uses the same task order as `beanstalk list`. It displays each task's
ID, status, priority, type, parent, and title.

## Navigation

| Key | Action |
| --- | --- |
| Up arrow or `k` | Select the previous task |
| Down arrow or `j` | Select the next task |
| Home or `g` | Select the first task |
| End or `G` | Select the last task |
| `q` or Ctrl-C | Exit |

The TUI does not create, edit, claim, or update tasks. Use the [CLI guide][cli-guide] for workflow actions.

## Terminal Behavior

Run the TUI in an interactive terminal. It uses the terminal's alternate screen and restores the previous screen when
it exits.

An empty project displays a `No beans found.` message. The list scrolls to keep the selected task visible. Before the
terminal size is available it renders up to five rows; terminals five rows or shorter use a compact view.

[cli-guide]: cli.md
