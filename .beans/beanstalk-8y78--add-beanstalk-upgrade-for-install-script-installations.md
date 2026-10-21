---
# beanstalk-8y78
title: Add beanstalk upgrade for install-script installations
status: todo
type: epic
created_at: 2026-10-17T00:51:15Z
updated_at: "2026-10-21T01:50:57Z"
parent: beanstalk-xxo5
---
Provide a built-in upgrade path for Beanstalk installations made by install.sh, including installations using a custom INSTALL_DIR. This is future work outside the v0.6.0 milestone.

Acceptance criteria:
- On supported Linux and macOS platforms, detect the latest published release and upgrade the running install-script installation in its current directory, without assuming $HOME/.local/bin.
- Verify downloaded release assets before replacing the executable; preserve the existing installation if download, verification, or replacement fails.
- Report when the installed release is already current.
- For Homebrew-managed installations, leave the binary untouched and direct users to brew upgrade beanstalk.
- On Windows, print manual upgrade instructions instead of replacing the executable.
- If installation provenance cannot be determined reliably, leave the binary untouched and explain the supported upgrade paths.
- Document the command and cover its supported outcomes with tests.
