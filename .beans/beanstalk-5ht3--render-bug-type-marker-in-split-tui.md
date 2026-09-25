---
# beanstalk-5ht3
title: Render bug type marker in split TUI
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-25T01:12:01Z"
---
The split-pane TUI type indicator omits the supported bug type and renders it as unknown.

Acceptance criteria:
- Render bug as B consistently with CLI list output.
- Add a split-pane row rendering regression test.