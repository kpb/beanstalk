---
# beanstalk-4o42
title: Validate configured prefixes prevent path traversal
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-24T03:24:58Z"
---
Configured beans.prefix values are incorporated into the task filename without path validation. A prefix containing traversal or separators can place created files outside the configured Beans directory.

Acceptance criteria:
- Reject unsafe prefixes before file creation.
- Ensure generated task paths remain beneath the Beans directory.
- Cover traversal and separator cases with tests.