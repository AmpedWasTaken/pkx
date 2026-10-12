# Security

## Reporting

Use GitHub's private **Report a vulnerability** feature if enabled in the repository Security tab. If it is unavailable, request a private reporting channel without publishing exploit steps.

## Trust model

pkx executes native package managers with the privileges of the user running it. Those managers can execute arbitrary lifecycle scripts from projects and dependencies. **A dry run previews the native commands, but does not sandbox their execution.** Treat installations from untrusted repositories and bundles as potentially dangerous; use an isolated environment for hostile code.

Bundle target directories are constrained to the selected workspace, symlink escapes are blocked, and dependencies beginning with command flags are rejected. These checks prevent some mistakes and obvious argument injections; they are not a general-purpose malicious package detector.

The CLI has no account, telemetry, or pkx-managed network service. Native managers may access the network normally.
