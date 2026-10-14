---
# beanstalk-n9kr
title: Define semantics for archived task updates and restoration
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-10-14T02:59:35Z"
---
Archived tasks can be updated to active statuses but remain in .beans/archive and are excluded from normal list and TUI views.

Acceptance criteria:
- Define whether archived tasks are immutable or explicitly restorable.
- Implement the chosen behavior atomically.
- Add an end-to-end discoverability regression test.

Decision: Archived tasks are read-only. Updates (status, parent, and body) and claims fail without changing the file; show and the TUI archived view remain available. Beanstalk provides no restore command. To make a task active again, move its file out of the archive manually. The archive lock and per-bean lock guard the update/claim checks against concurrent archive operations.