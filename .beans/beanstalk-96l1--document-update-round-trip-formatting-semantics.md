---
# beanstalk-96l1
title: Document update round-trip formatting semantics
status: todo
type: task
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-21T14:50:11Z"
---
The documentation promises unknown front-matter keys remain intact, but task updates parse and re-marshal YAML. Formatting, line endings, and comments attached to updated values can change.

Acceptance criteria:
- Describe semantic metadata preservation and rewrite behavior accurately.
- Add compatibility fixtures and round-trip tests for supported update behavior.