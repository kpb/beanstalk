---
# beanstalk-o3zf
title: Sanitize task content in terminal output
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-24T04:07:50Z"
---
Task-controlled IDs, titles, tags, and bodies are printed directly to CLI and TUI terminals. Control sequences in an untrusted task file can alter terminal state.

Acceptance criteria:
- Sanitize control characters in all human-facing CLI and TUI output.
- Preserve original values in JSON output.
- Add regression tests for escape and other control characters.