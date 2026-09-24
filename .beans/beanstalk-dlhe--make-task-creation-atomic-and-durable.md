---
# beanstalk-dlhe
title: Make task creation atomic and durable
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-24T03:36:58Z"
---
Task creation writes directly to the destination file and does not sync or remove partial output on a write or close failure.

Acceptance criteria:
- Write, sync, and close a temporary file in the Beans directory.
- Atomically rename it into place and sync the directory.
- Remove temporary files on error and test failure paths.