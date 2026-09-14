---
# beanstalk-kpm7
title: Implement atomic body replacement in update
status: todo
type: task
parent: beanstalk-i13n
created_at: 2026-09-14T16:57:10Z
updated_at: 2026-09-14T16:57:10Z
---
Add the --body flag and the corresponding update field. Replace the body verbatim, distinguish an omitted flag from an empty body, and include body changes in the existing atomic update path.
