# TUI Guide

Launch the task browser from an initialized Beans project:

```bash
beanstalk
```

`beanstalk tui` remains available for explicit TUI launches.

The TUI loads active tasks from the configured Beans directory and uses the same task order as `beanstalk list`. It
displays the task hierarchy, with parent/child relationships indented beneath their parent. Archived tasks are hidden by
default and can be shown with `a`. On narrow terminals, each list row includes a task's ID, status, priority, type,
parent, and title.

The TUI checks for task-file changes automatically, normally showing them within one second while preserving the
selected task when it remains available.

On terminals at least 100 columns wide, the task tree and the selected task's details appear side by side in separately
bordered panes. The tree leads with titles and branch connectors that show parent and sibling relationships; compact type
and status metadata appears at the right of each row. Details include the remaining task metadata, body, parent,
children, and milestone progress when the selected task belongs to a milestone. Press Tab or Enter to open the selected
task in a full-screen detail view at any terminal width; press Tab, Enter, or Esc to return to the task list.
When details exceed the screen height, use the arrow keys or `j`/`k` to scroll; Home and End jump to the beginning and
end of the detail content.

The wide layout keeps keyboard shortcuts in a shared footer below both panes, leaving the panes for task content.
Press `?` to open keyboard help in a centered, bordered overlay. The task list remains visible behind the dimmed overlay;
press `?` or Esc to close it.

## Navigation

| Key | Action |
| --- | --- |
| Up arrow or `k` | Select the previous task, or scroll details up |
| Down arrow or `j` | Select the next task, or scroll details down |
| Home or `g` | Select the first task, or jump to the top of details |
| End or `G` | Select the last task, or jump to the bottom of details |
| `h` or Left arrow | Collapse the selected task or select its parent |
| `l` or Right arrow | Expand the selected task or select its first child |
| Tab or Enter | Toggle the selected task's full-screen detail view |
| Esc | Return from the full-screen detail view to the task list |
| `r` | Reload tasks |
| `a` | Show or hide archived tasks |
| `c` | Claim the selected `todo` task |
| `s` | Change the selected task's status |
| `?` | Show keyboard help |
| `q` or Ctrl-C | Exit |

## Status Changes

Press `s` to open the selected task's status picker. Use the arrow keys or `j`/`k` to select `draft`, `todo`,
`in-progress`, `completed`, or `scrapped`, then press Enter to save. Press Esc to cancel. The TUI reloads immediately
after a successful save and retains the selected task when it remains available. Write and reload errors remain visible
without exiting the TUI.

## Claim Tasks

Press `c` to atomically claim a selected `todo` task. Successful claims change its status to `in-progress` and reload
the task list immediately while retaining the selected task. Claim conflicts and other failures are displayed without
exiting the TUI.

## Supported Workflow

The TUI supports browsing, navigating the hierarchy, inspecting details, showing archived tasks, claiming todo tasks,
changing task status, and manually reloading the task list. It does not create tasks or edit task metadata or bodies.
Use the [CLI guide][cli-guide] for those workflow actions.

## Terminal Behavior

Run the TUI in an interactive terminal. It uses the terminal's alternate screen and restores the previous screen when
it exits.

An empty project displays a `No beans found.` message. The list scrolls to keep the selected task visible. Task-file
changes made outside the TUI load automatically; press `r` to force an immediate refresh or retry a failed load. Before
the terminal size is available it renders up to five rows; terminals five rows or shorter use compact list and status
picker views.

[cli-guide]: cli.md
