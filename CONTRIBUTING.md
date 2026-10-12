# Contributing to pkx

Thanks for considering a contribution. pkx is small on purpose: a reliable, understandable CLI is more useful than a long list of half-working integrations. Code, tests, docs, accessibility changes, reproduction fixtures and platform reports all count.

## Find something to work on

- [Open issues](https://github.com/AmpedWasTaken/pkx/issues) — bugs, ideas and starter tasks.
- [Good first issues](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) — scoped work when that label is applied.
- [Roadmap](docs/ROADMAP.md) — priorities and unshipped ideas.

Fix a typo or add a failing regression test without asking first. For a new manager, CLI behavior or bundle schema change, start with an issue to discuss compatibility before writing a large patch.

## Local development

Install Go 1.23+ and work in a topic branch:

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
git checkout -b fix/describe-the-issue
go test ./...
go run ./cmd/pkx help
```

Before a pull request:

```sh
gofmt -w cmd/pkx internal/pkx
go test ./...
go test -race ./... # requires a race-enabled Go environment
go vet ./...
go build ./cmd/pkx
```

CI checks formatting, tests, vet, build and installer help/syntax on Linux, macOS and Windows. You don't need access to all three OSes locally: state which platforms you tested.

## What helps reviewers

1. **Reproduce:** add a small non-sensitive fixture or describe OS, `pkx version`, the command and sanitized output.
2. **Fix one thing:** preserve existing native manager and lockfile behavior.
3. **Test failures:** include conflicting lockfiles, invalid paths, failed installers and relevant platform behavior.
4. **Document it:** update README and guides for user-facing changes.
5. **Explain trade-offs:** include compatibility and security implications in the PR.

CLI rules: plans stay read-only, execution must be explicit, don't interpolate shell strings, avoid hidden network work, keep native tool errors visible, and don't write outside the selected workspace.

### Commits and PRs

Use short and accurate messages such as `fix(doctor): handle missing Python executable` or `test(pnpm): cover workspace lockfile conflict`. No fabricated commit times or artificial history. Focused commits are easier to review and revert.

Fork the repository, push your branch and open a PR against `main` using the provided template. Link related issues (`Closes #123` only when it genuinely resolves them). Do not submit credentials, private registries, caches, build artifacts or huge generated files. Reviews are volunteer effort and no guaranteed response time is promised.

Security problems should **not** be reported with public exploit details: see [SECURITY.md](SECURITY.md). Be respectful in issues and reviews: see [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

Contributed code is licensed under the repository's [MIT license](LICENSE). Contributors retain copyright in their work.
