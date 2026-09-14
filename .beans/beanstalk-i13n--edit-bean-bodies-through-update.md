---
# beanstalk-i13n
title: Edit bean bodies through update
status: todo
type: feature
parent: beanstalk-mmhi
created_at: 2026-09-14T16:56:53Z
updated_at: 2026-09-14T16:56:53Z
---
Extend update with literal --body replacement. Support clearing with an empty value and combine body, status, and parent changes atomically. Preserve unknown front matter, the ID comment, permissions, archive location, and existing CRLF body bytes when the body is not changed. Keep TUI body editing, stdin/file input, append/replace operations, and etags out of scope.
