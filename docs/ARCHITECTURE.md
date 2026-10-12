# Architecture and design

pkx is a **thin orchestration layer**. It discovers projects, prepares a deterministic native command plan, optionally probes the environment and executes the existing tools.

```text
repository directory
  → Discover()             detect manifests/lockfiles within scan depth
  → InstallPlan()          native executable + arguments + working directory
  → CheckEnvironment()     preflight for doctor/up/status
  → execute()              sequential native commands for install/up
```

Entry point: `cmd/pkx/main.go`. Implementation: `internal/pkx/`.

| File | Responsibility |
| --- | --- |
| `discover.go` | Find manifests and select package managers |
| `workspace.go` | Avoid unnecessary duplicate Node workspace installs |
| `plan.go` | Map projects to native install steps |
| `environment.go` | Check tool availability and pinned runtime versions |
| `cli.go`, `ux.go`, `completion.go` | CLI routing, output and tab completion |
| `bundle.go` | Validate local JSON bundles and plan package additions |
| `*_test.go` | Regression and behavior tests |

## Design principles

1. **Existing lockfiles win.** No cross-language dependency solver or proprietary lockfile.
2. **Read-only previews.** `plan` never installs; `status` checks tools but doesn't modify dependencies.
3. **Explicit execution.** No surprise runtime downloads. Native managers may still fetch code and run lifecycle scripts.
4. **Workspace containment.** Bundle paths and symlinks must not escape the chosen workspace.
5. **Deterministic execution.** Sequential install steps avoid surprising simultaneous lockfile writes.
6. **Understandable errors.** Keep native tool output and exit conditions visible.

A manager change should start with a minimal failing test, update discovery and planning, and document the exact lockfile behavior and failure cases. See [CONTRIBUTING](../CONTRIBUTING.md).
