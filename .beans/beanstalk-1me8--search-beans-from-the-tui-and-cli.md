---
# beanstalk-1me8
title: Search beans from the TUI and CLI
status: todo
type: epic
created_at: 2026-10-17T01:01:10Z
updated_at: "2026-10-21T01:50:57Z"
parent: beanstalk-xxo5
---
Add search as a future feature outside the v0.6.0 milestone. Use shared matching behavior for interactive TUI search and a scriptable JSON CLI command; no server API is required.

Acceptance criteria:
- Pressing / in the TUI opens a search dialog for entering a query; users can browse matches and navigate to a selected bean. Escape cancels without losing the prior selection.
- Provide beanstalk search "<text>" with human-readable and --json output for scripts and agents.
- Match bean IDs, titles, and Markdown bodies case-insensitively using the same rules in both interfaces.
- Exclude archived beans by default; provide a clear way to include them when needed.
- Handle empty queries, zero matches, and task reloads predictably.
- Document search behavior and add tests for matching, CLI output, and TUI navigation.
