---
# beanstalk-n9kr
title: Define semantics for archived task updates and restoration
status: todo
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-21T14:49:49Z"
---
Archived tasks can be updated to active statuses but remain in .beans/archive and are excluded from normal list and TUI views.

Acceptance criteria:
- Define whether archived tasks are immutable or explicitly restorable.
- Implement the chosen behavior atomically.
- Add an end-to-end discoverability regression test.