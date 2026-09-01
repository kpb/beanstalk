# TUI Guide

Launch the task browser from an initialized Beans project:

```bash
beanstalk tui
```

The TUI loads active tasks from the configured Beans directory and uses the same task order as `beanstalk list`. It
displays the task hierarchy, with parent/child relationships indented beneath their parent. Archived tasks are hidden
until the archive toggle workflow is added. On narrow terminals, each list row includes a task's ID, status, priority,
type, parent, and title.

On terminals at least 100 columns wide, the task tree and the selected task's details appear side by side. Tree rows
show ID, status, and title; details include the remaining task metadata, body, parent, children, and milestone progress
when the selected task belongs to a milestone. On narrower terminals, press Tab or Enter to switch between the list and
detail views.

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
| `c` | Claim the selected `todo` task |
| `s` | Change the selected task's status |
| `?` | Show keyboard help |
| `q` or Ctrl-C | Exit |

## Status Changes

Press `s` to open the selected task's status picker. Use the arrow keys or `j`/`k` to select `draft`, `todo`,
`in-progress`, `completed`, or `scrapped`, then press Enter to save. Press Esc to cancel. The TUI reloads after a
successful save and retains the selected task when it remains available. Write and reload errors remain visible without
exiting the TUI.

## Claim Tasks

Press `c` to atomically claim a selected `todo` task. Successful claims change its status to `in-progress` and reload
the task list while retaining the selected task. Claim conflicts and other failures are displayed without exiting the
TUI.

## Supported Workflow

The TUI supports browsing, navigating the hierarchy, inspecting details, claiming todo tasks, changing task status,
and manually reloading the task list. It does not create tasks or edit task metadata or bodies. Use the [CLI
guide][cli-guide] for those workflow actions.

## Terminal Behavior

Run the TUI in an interactive terminal. It uses the terminal's alternate screen and restores the previous screen when
it exits.

An empty project displays a `No beans found.` message. The list scrolls to keep the selected task visible. Press `r`
to load task-file changes made outside the TUI. Before the terminal size is available it renders up to five rows;
terminals five rows or shorter use compact list and status picker views.

[cli-guide]: cli.md
