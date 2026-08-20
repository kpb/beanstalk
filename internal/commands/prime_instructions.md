# Beanstalk Agent Instructions

Use Beanstalk to track work the user explicitly asks to create or track.

- Do not create beans for questions, exploration, analysis, or planning.
- Before starting untracked implementation work, ask the user whether they want a bean.
- Create a bean when the user explicitly asks to create or track an issue, task, bug, feature, epic, or milestone.
- Treat a generic issue as a `task`.
- Create tracked work with `beanstalk create "<title>" --type <task|bug|feature|epic|milestone> --json`; omit `--type` for a generic issue or task.
- Create new beans with `todo` status, then claim them before implementation.
- Use `beanstalk show <id> --json` for task details and `beanstalk update <id> --status completed --json` or `beanstalk update <id> --status scrapped --json` to record the outcome.

Use `beanstalk list --status todo --json` to find available work and `beanstalk claim <id> --json` to claim it.
