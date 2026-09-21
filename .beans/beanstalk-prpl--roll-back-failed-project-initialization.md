---
# beanstalk-prpl
title: Roll back failed project initialization
status: todo
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-21T14:50:11Z"
---
init creates .beans before writing .gitignore and .beans.yml. A later failure leaves partial state that blocks a retry.

Acceptance criteria:
- Make initialization retryable after any write failure.
- Roll back only artifacts created by the current invocation, or stage initialization atomically.
- Add tests for failures after the directory is created.