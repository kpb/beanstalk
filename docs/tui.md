# TUI Guide

Launch the task browser from an initialized Beans project:

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
| `h` or Left arrow | Collapse the selected task or select its parent |
| `l` or Right arrow | Expand the selected task or select its first child |
| Tab or Enter | Toggle the detail view on narrow terminals |
| `r` | Reload tasks |
| `?` | Show keyboard help |
| `q` or Ctrl-C | Exit |

## Status Changes

Press `s` to open the selected task's status picker. Use the arrow keys or `j`/`k` to select `todo`, `in-progress`,
`completed`, or `scrapped`, then press Enter to save. Press Esc to cancel. The TUI reloads after a successful save and
retains the selected task when it remains available.

The TUI does not create, edit task metadata or bodies, or claim tasks. Use the [CLI guide][cli-guide] for those
workflow actions.

## Terminal Behavior

Run the TUI in an interactive terminal. It uses the terminal's alternate screen and restores the previous screen when
it exits.

An empty project displays a `No beans found.` message. The list scrolls to keep the selected task visible. Before the
terminal size is available it renders up to five rows; terminals five rows or shorter use compact list and status
picker views.

[cli-guide]: cli.md
