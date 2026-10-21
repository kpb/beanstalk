---
# beanstalk-368x
title: Validate statuses in the beans update API
status: completed
type: bug
parent: beanstalk-gf92
created_at: 2026-09-21T14:42:58Z
updated_at: "2026-09-24T03:41:08Z"
---
beans.Update and UpdateStatus accept any status from internal callers, while only the CLI validates status values. Invalid data then prevents project loading.

Acceptance criteria:
- Validate statuses in the reusable beans API.
- Return ErrInvalidBeanStatus for unsupported values.
- Test direct API calls with invalid statuses.