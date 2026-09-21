---
# beanstalk-9vw9
title: Render TUI status cell without text replacement
status: todo
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-21T14:50:11Z"
---
TUI row styling replaces the first occurrence of a status string in the assembled row. An ID or title containing a status word can be colored instead of the status column.

Acceptance criteria:
- Render and style the status cell separately from the rest of the row.
- Add tests where IDs or titles contain status names.