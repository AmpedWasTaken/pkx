---
title: "I built pkx: one command to set up mixed-language repos without replacing npm or Composer"
published: false
description: "How a small Go CLI detects Node, PHP, Python, Rust and Go projects, previews native commands, and welcomes adapter contributors."
tags: opensource, go, devtools, cli
---

# I built pkx: one command, every ecosystem

I work with repositories where the setup instructions aren't a single command. A frontend may use pnpm, the API Composer, one utility Python and another Go modules. Each native manager already works. The friction is knowing **which commands to run in which directories**, and catching missing runtimes before you start.

That's why I started [pkx — Package eXchange](https://github.com/AmpedWasTaken/pkx).

pkx is a small, MIT-licensed Go CLI. Instead of inventing another package resolver, it detects the tools a repository already uses and prepares native commands.

## A preview-first workflow

From the repository root:

```sh
pkx scan
pkx status
pkx plan
pkx up
```

`scan` finds manifests. `status` explains the environment and missing tools. `plan` shows the exact commands and **does not install anything**. `up` performs preflight checks, then delegates to native package managers.

That means a repository with `package.json` and `composer.json` can produce an npm and Composer install plan in one place. Lockfiles remain native. pkx does not create its own cross-language lockfile.

For a safe, committed demo, see [examples/mixed-language](https://github.com/AmpedWasTaken/pkx/tree/main/examples/mixed-language).

## What does pkx currently support?

The first published release is v0.5.0. The exact support split matters:

| Workflow | Supported today |
| --- | --- |
| Native dependency installation | npm, pnpm, Yarn, Bun, pip, uv, Poetry, Composer, Cargo and Go modules |
| `add`, `remove`, and declared-script `run` | npm, pnpm, Yarn, Bun and Composer |

That second row is the current scope gap. Python/Cargo/Go package editing and script execution aren't covered yet, and some workspace configurations need better tests. If your repository hits a mismatch, I'd rather see an issue with a minimal fixture than say everything works.

There's also a bundle marketplace **preview on the website**, but the corresponding `bundle search/info/fetch` commands live in unreleased v0.6 development code, not v0.5.0. A preview isn't a shipped CLI feature.

## What about installer security?

The [release page](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0) provides platform archives and `SHA256SUMS`. You can download and verify them without running a shell script. A one-line convenience installer exists too, but running code fetched from the network requires trusting the host. The archive checksum is an integrity check against the published sums; it isn't a substitute for validating the source and release owner.

Package-manager execution carries its usual risks as well. npm/Composer lifecycle scripts can run code. On an unfamiliar project, use `pkx plan` and read the manifests before `pkx up`.

## How it works

The core is intentionally straightforward:

```text
Discover repo manifests/lockfiles
            ↓
Build native command plan
            ↓
Check runtimes and binaries
            ↓
Execute only when requested
```

Go code in `internal/pkx/discover.go` selects projects and managers; `plan.go` produces executable and argument arrays; `environment.go` checks availability and version pins.

I'm keeping the core small so it's practical for contributors to test and extend.

## Want to write an adapter?

There's a **working, no-network npm example**:

```sh
go run ./examples/adapter-npm
go test ./examples/adapter-npm
```

It demonstrates `Discover` and `InstallPlan` using a tiny checked-in manifest. Nothing gets installed. The [contributing guide](https://github.com/AmpedWasTaken/pkx/blob/main/CONTRIBUTING.md) explains where a manager behavior change belongs and what tests to write.

The best first contributions aren't giant integrations. They're small lockfile edge cases, manager detection tests and useful diagnostics.

## What I want to learn next

I'm looking for a small group of early adopters with genuine mixed-language projects. Try the current release, tell me where it gives the wrong plan, and share a sanitized fixture if you can.

If you want to help, visit the [GitHub issues](https://github.com/AmpedWasTaken/pkx/issues) or [good first issues](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22).

I'm not trying to replace every package manager. I want setup to feel less fragmented, while leaving developers in control of what runs.
