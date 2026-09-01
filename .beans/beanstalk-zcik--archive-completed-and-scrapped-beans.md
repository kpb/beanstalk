---
# beanstalk-zcik
title: Archive completed and scrapped beans
status: completed
type: feature
parent: beanstalk-uibo
created_at: 2026-08-26T23:51:47Z
updated_at: "2026-09-01T01:38:47Z"
---
Add an explicit archive workflow that moves completed and scrapped Beans-format task files into `.beans/archive/`.

Archiving must not happen automatically when a task status changes. The archive operation must preserve each task's
filename, contents, and metadata, skip tasks in other statuses, and retain reliable lookup of archived tasks by ID.
Archived tasks should be excluded from the default active task views; `beanstalk-hk9k` will add the TUI control for
showing them.
