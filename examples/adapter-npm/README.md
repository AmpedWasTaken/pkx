# An adapter author's first working example

This is a **real, runnable demonstration of pkx's current manager pipeline**. It uses the existing npm adapter rather than inventing a plugin interface that the CLI does not expose yet.

From the repository root:

```sh
go run ./examples/adapter-npm
go test ./examples/adapter-npm
```

Expected output:

```text
npm: npm install
Nothing executed.
```

The code calls `pkx.Discover` and `pkx.InstallPlan` with a committed, harmless `package.json` fixture. It **does not run npm, install packages, contact the network, or modify lockfiles**.

## Turn this example into your first contribution

1. Find a small [good first issue](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22).
2. Reproduce its native-manager case with a fixture similar to `fixture/package.json` (or the appropriate manifest/lockfile).
3. Write a focused test with an explicit `Step.String()` expectation.
4. If the actual code needs changes, update detection in `internal/pkx/discover.go` and planning in `internal/pkx/plan.go`. Extend environment diagnostics or command dispatch only when required. There is **no independent adapter registration system yet**.
5. Run `gofmt`, `go test ./...`, `go vet ./...`, and open a focused PR.

**Do not** execute third-party package lifecycle scripts as part of tests. Native installations should be covered by intentional sandboxed integration tests instead of automatically running installs on an untrusted fixture.

See [the contributing guide](../../CONTRIBUTING.md#become-an-adapter-author-word-een-adapter-auteur).
