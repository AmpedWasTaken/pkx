# RFC 0002: Opt-in runtime bootstrap (proposal)

**Status:** Draft / discussion wanted. **No automatic runtime installation is implemented.**

## Problem

`pkx doctor` detects missing or incompatible runtimes, but the user still has to set them up manually. Runtime provisioning can reduce setup friction, but also modifies the user's machine and must never happen as a surprise.

## Staged design

1. **Starter in this PR:** `examples/runtime/bootstrap-plan.v1.json` illustrates a proposed **non-executable plan**. It is documentation only and is not parsed by the current CLI.
2. **Pure planning layer:** use `CheckEnvironment` diagnostics and project pin files to produce `RuntimeRequirement` and `RuntimeAction` values without installing anything. Mark unsupported/ambiguous version specs instead of guessing a version.
3. **Read-only command:** `pkx doctor --fix --dry-run` (name subject to CLI discussion) prints affected project, requested version, selected provider, exact source/destination, verification mechanism, required permissions and rollback limitations.
4. **First opt-in provider:** support a single tested runtime-manager adapter on one OS before expanding coverage. Prefer user-owned versions, established native installers and verified upstream distribution metadata, not ad-hoc curl-to-shell.
5. **Explicit approval:** print plan, require confirmation per new machine-level action (and offer `--yes` only for explicitly documented noninteractive use, with safety constraints). Declining performs no mutations.
6. **Cross-platform, version pinning and failure recovery:** extend through adapter interfaces, tests and atomic install behavior.

## Trust boundaries

- No sudo, admin elevation, global PATH modification or background downloads by default.
- No anonymous remote bootstrap scripts or code fetched and executed from issue descriptions / registry metadata.
- Prefer pinned URLs and verifiable hashes/signatures from official distributors; reject unverifiable artifacts.
- Do not silently overwrite existing runtimes. Never call a provider if the requested version is ambiguous.
- Show the exact intended commands, filesystem modifications, source and post-install verification.
- Runtime installer behavior must be **separate** from dependency `pkx up` until compatibility and safety are proven.
- Respect developer policies, package manager license requirements and OS-specific failure modes.

## Suggested first contributor PR

Implement only `RuntimeRequirement` parsing and `RuntimeInstallPlan` formatting with fake providers injected in tests. Ensure the plan is side-effect-free on Windows, macOS and Linux. Do **not** add a download/execution path in that PR.

## Acceptance test ideas

Unknown `.nvmrc` aliases, mismatched Python pins, PHP constraint ranges, existing multiple runtimes, incorrect architecture, checksum mismatches, refusal to approve, restricted paths, offline operation and a provider reporting partial failure.

## Open questions

Which runtime-manager(s) should be recommended first? Should runtime setup be global, per-user or per-workspace? What are the expectations for rollback, caches and auditability?
