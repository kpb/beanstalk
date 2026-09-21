---
# beanstalk-1n3l
title: Return default priority from create JSON
status: todo
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-21T14:49:49Z"
---
create --json returns an empty priority when no priority flag is supplied, but reloads default the persisted task to normal.

Acceptance criteria:
- Apply the normal priority before rendering and returning a created bean.
- Ensure create JSON and subsequent show JSON agree.
- Add a default-priority regression test.