# Five small good-first-issue templates — drafts

These are reusable text templates, not automatically posted from this file. Adjust exact labels and paths before creating issues. They intentionally focus on a single supported manager and offline fixture instead of assigning a beginner an entire ecosystem.

## 1. pnpm workspace precedence test

**Labels:** `good first issue`, `enhancement`

**Goal:** Test the current `packageManager: pnpm@...` and `pnpm-lock.yaml` precedence in a minimal workspace.

- [ ] Add a small package.json + lockfile fixture.
- [ ] Test `Discover` selects pnpm when pinned.
- [ ] Test `InstallPlan` selects `pnpm install --frozen-lockfile` when the lockfile is present.
- [ ] Add a negative test for conflicting lockfile metadata.
- [ ] Run `go test ./...` and `gofmt`.

No native pnpm execution in unit tests. Scope: tests and minimal bug fix if needed.

## 2. Yarn Classic vs modern Yarn plan

**Labels:** `good first issue`, `enhancement`

**Goal:** Verify `yarn.lock` plus `packageManager` chooses the right native immutable flag.

- [ ] One fixture for Yarn 1 and one for Yarn 3/4.
- [ ] Assert exact frozen / immutable argument arrays.
- [ ] Test missing lockfile and conflicting lockfile detection.
- [ ] Document any unsupported Yarn variants in the PR.

No running Yarn scripts; the test checks the plan.

## 3. Bun lockfile behavior

**Labels:** `good first issue`, `enhancement`

**Goal:** Assert correct detection for `bun.lock` and `bun.lockb`.

- [ ] Cover each file type in a separate temporary fixture.
- [ ] Verify Bun install plan uses frozen lockfile only when appropriate.
- [ ] Test the `packageManager` pin and mixed npm/Bun conflict.
- [ ] Do not add or commit `node_modules`.

## 4. Composer read-only plan regression

**Labels:** `good first issue`, `enhancement`

**Goal:** Extend Composer planning tests without network, PHP runtime or credentials.

- [ ] Add minimal `composer.json` and optional `composer.lock` fixtures.
- [ ] Assert `composer install --no-interaction` exactly.
- [ ] Confirm a plan preview creates no `vendor` or lockfile.
- [ ] Cover selection with `--only composer`.

No installer execution; don't rely on a Composer installation on test machines.

## 5. Python pip local virtualenv plan

**Labels:** `good first issue`, `enhancement`

**Goal:** Cover first and subsequent `requirements.txt` install plans for Linux/macOS/Windows.

- [ ] First plan creates `.venv` and invokes pip in that venv.
- [ ] When `.venv` exists, plan excludes the setup step.
- [ ] Compare platform-specific `bin/python` vs `Scripts/python.exe` logically.
- [ ] Never download Python packages in unit tests.

See [adapter example](../../examples/adapter-npm/README.md) for a runnable test pattern.
