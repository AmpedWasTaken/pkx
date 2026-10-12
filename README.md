<p align="center"><img src="assets/pkx-mark.svg" width="88" height="88" alt="pkx logo" /></p>
<h1 align="center">pkx — Package eXchange</h1>
<p align="center"><strong>One command. Every ecosystem. Your existing package managers.</strong></p>
<p align="center">
  <a href="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml"><img alt="CI status" src="https://github.com/AmpedWasTaken/pkx/actions/workflows/ci.yml/badge.svg" /></a>
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-blue.svg" /></a>
  <a href="CONTRIBUTING.md"><img alt="Contributions welcome" src="https://img.shields.io/badge/contributions-welcome-58b7a6.svg" /></a>
  <a href="https://github.com/AmpedWasTaken/pkx/issues"><img alt="Open issues" src="https://img.shields.io/github/issues/AmpedWasTaken/pkx" /></a>
</p>
<p align="center">An open-source CLI for setting up mixed-language repositories.<br />No new dependency resolver. No replacement lockfiles. No required account.</p>
<p align="center"><img src="assets/social-preview.svg" width="100%" alt="pkx: one command for every ecosystem" /></p>

> **Version:** `v0.5.0` — first public **0.x** release candidate. The CLI is usable but its interfaces may change before v1.0. A remote bundle marketplace and automatic runtime installation do **not** exist yet.

## Install once, use `pkx` anywhere

**pkx is a standalone executable:** you do **not** need Go, Node.js, npm or administrator rights to install the [published binaries](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0). To install dependencies *inside a project*, the relevant native manager (npm, Composer, etc.) must still be installed.

### Windows — works globally in CMD and PowerShell

**Recommended (automated, with checksum verification):** Open **CMD** in any directory and run these commands. This method needs Git to fetch the installer script, but does not need Go or npm:

```cmd
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\install.ps1 -Version v0.5.0 -AddToPath
```

The script verifies the GitHub release archive's SHA256, installs `pkx.exe` into **`%USERPROFILE%\.local\bin`** and (only with `-AddToPath`) permanently adds that folder to **your user PATH**. It doesn't require admin permissions or change the system PATH. You can inspect [the installer source](scripts/install.ps1) before running it.

**Important: close CMD / PowerShell and open a NEW terminal** after installation. From **any directory**, run:

```cmd
where pkx
pkx version
cd C:\path\to\your-project
pkx status
pkx plan
pkx up
```

**Without Git or PowerShell scripts:** Download `pkx_v0.5.0_windows_amd64.zip` (x64 PCs) or `pkx_v0.5.0_windows_arm64.zip` (Windows ARM) from [GitHub Releases](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0). Verify it against `SHA256SUMS`, unzip `pkx.exe` into `%USERPROFILE%\.local\bin`, then add that folder to your *user* `Path` in **Edit environment variables for your account → Path → New**. Open a new CMD and run `pkx version`.

### macOS or Linux — works in every terminal

From a reviewed checkout (Git required, but no Go/npm):

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
sh scripts/install.sh v0.5.0
```

The checksum-verifying script installs `pkx` to `~/.local/bin`. If `pkx version` returns **command not found**, add that directory to your shell PATH **once**:

**macOS (default Zsh):**

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zprofile
source ~/.zprofile
pkx version
```

**Linux (Bash):**

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
pkx version
```

You can instead download your matching `.tar.gz` and `SHA256SUMS` from [GitHub Releases](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0), verify it, and manually place the executable in a directory already on your PATH.

### Alternative: install with Go

If you already have Go 1.23+ and prefer `go install`:

```sh
go install github.com/AmpedWasTaken/pkx/cmd/pkx@v0.5.0
```

Go installs the binary in `$(go env GOPATH)/bin` (unless `GOBIN` is configured). **That directory must also be on your PATH** for `pkx` to work globally.

**Troubleshooting:** On Windows CMD use `where pkx`, in PowerShell use `Get-Command pkx`, and on macOS/Linux use `command -v pkx`. If the binary is installed but not detected, check the PATH entry and restart your terminal (on Windows, signing out/in may be necessary for apps that kept an old environment). For the full guide see [Getting started](docs/GETTING_STARTED.md).

## Development status

<p align="center">
  <a href="https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0"><img src="assets/badge-core.svg" alt="Core CLI: v0.5.0 shipped" /></a>
  <a href="https://github.com/AmpedWasTaken/pkx/issues/12"><img src="assets/badge-marketplace.svg" alt="Bundle marketplace: design proposal" /></a>
  <a href="https://github.com/AmpedWasTaken/pkx/issues/14"><img src="assets/badge-runtimes.svg" alt="Automatic runtime bootstrap: design proposal" /></a>
</p>
<p align="center"><a href="docs/ROADMAP.md"><img src="assets/roadmap.svg" width="100%" alt="pkx roadmap: core CLI shipped, reliability next, developer UX planned, remote marketplace and runtime bootstrap under design" /></a></p>

**The graphic shows milestone states, not measured percentages or a release schedule.** It is updated manually as work is completed. Follow [the roadmap](docs/ROADMAP.md), [open feature issues](https://github.com/AmpedWasTaken/pkx/issues) and [the bundle marketplace](docs/rfcs/0001-bundle-registry.md) / [runtime bootstrap](docs/rfcs/0002-runtime-bootstrap.md) proposals. The marketplace and automatic runtime installation are **not implemented yet**.

## Why pkx?

A repo can have a JavaScript frontend, a PHP backend and Python scripts. Getting started often means figuring out which package manager belongs in which folder, matching runtime versions and remembering install commands. **pkx finds those projects and delegates to their native tools.**

```sh
cd your-project
pkx status  # See detected projects, lockfiles and missing tools
pkx plan    # Preview native commands; makes no changes
pkx up      # Check the environment, then install dependencies
```

`pkx up` runs native package managers with your permissions; their lifecycle scripts may execute third-party code. Review untrusted repositories and use `pkx plan` first.

## Everyday commands

| Command | What it does |
| --- | --- |
| `pkx status` | Explain what's in this workspace and which tools are missing |
| `pkx plan` | Preview native install commands **without** running them |
| `pkx up` | Check the environment and install via native managers |
| `pkx doctor` | Diagnose missing/incompatible runtimes (`--json` supported) |
| `pkx scan` / `pkx ls` | Discover existing projects (`--json` supported) |
| `pkx install` / `pkx i` | Install without runtime-version preflight (`--dry-run` supported) |
| `pkx add axios` | Add packages via the detected native manager (`--dev`, `--dry-run`) |
| `pkx remove axios` / `pkx rm axios` | Remove packages through the native manager |
| `pkx run dev` | Run a declared Node/Composer script and forward optional arguments after `--` |
| `pkx plan --only node` | Preview one ecosystem/manager |
| `pkx init` | Create an empty local `pkx.json` bundle manifest |
| `pkx bundle install pkx.json --dry-run` | Preview **explicit** dependency additions |
| `pkx completion bash` | Generate shell completion (`zsh`, `fish`, `powershell` also supported) |
| `pkx help plan` | Show help for one command |

Supports npm, pnpm, Yarn, Bun, pip, uv, Poetry, Composer, Cargo and Go modules. See [exact compatibility details](docs/GETTING_STARTED.md#supported-managers-and-behavior). Existing lockfiles remain authoritative.

## Manage packages and run scripts

These commands use the project’s **existing** package manager. Supported for npm, pnpm, Yarn, Bun and Composer; they do not create a separate pkx lockfile.

```sh
pkx add axios --dry-run          # Preview first; nothing executes
pkx add axios zod                # Add dependencies
pkx add vite --dev               # Add development dependencies
pkx remove axios                 # Remove native dependencies
pkx run dev                      # Run a declared project script
pkx run dev -- --host 127.0.0.1 # Forward arguments to the script
```

With multiple matching projects, pkx **refuses to guess** where packages should go. Choose a project explicitly:

```sh
pkx add axios --project frontend
pkx run test --project backend
pkx add laravel/sanctum --project backend --manager composer
```

`--project` is relative to `--dir` (default `.`); `--manager` filters the ecosystem. `pkx run` only runs a script defined in `package.json` or `composer.json`, and suggests available scripts if its name is unknown. Add/remove for pip, Cargo and Go are not supported yet. All changes go through native package managers, which may execute untrusted installation hooks or scripts. Preview with `--dry-run` before working with unfamiliar repositories.

## Local dependency bundles

Bundles share a *list of dependencies*, not a new package format. `pkx init` creates a manifest with **no packages**. Add entries for existing project directories, preview them, and install explicitly:

```json
{
  "schema": 1,
  "name": "web-starter",
  "packages": [
    { "manager": "pnpm", "directory": "frontend", "dependencies": ["vue", "axios"], "devDependencies": ["vite"] },
    { "manager": "composer", "directory": "backend", "dependencies": ["laravel/sanctum"] }
  ]
}
```

```sh
pkx bundle install pkx.json --dry-run
pkx bundle install pkx.json
```

Directories must exist inside the workspace. Bundle installation may modify manifests and native lockfiles. There is **no hosted registry** in this preview.

## Built with contributors

pkx is MIT licensed and developed in the open. Bug fixes, tests, documentation, platform reports and new adapter proposals are welcome. You can fix a typo immediately; for larger behavior changes, start with an issue so the design stays maintainable.

- **[Contributing](CONTRIBUTING.md)** — how to set up, test and submit a PR
- **[Good first issues](https://github.com/AmpedWasTaken/pkx/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)** · **[All issues](https://github.com/AmpedWasTaken/pkx/issues)** — choose a task or report friction
- **[Roadmap](docs/ROADMAP.md)** · **[Architecture](docs/ARCHITECTURE.md)** — project scope and decisions
- **[Changelog](CHANGELOG.md)** · **[Release notes](docs/releases/v0.5.0.md)** · **[Release process](docs/RELEASING.md)** — versions, packages and verification
- **[Support](SUPPORT.md)** · **[Code of conduct](CODE_OF_CONDUCT.md)** · **[Security](SECURITY.md)**

If pkx helps in your own repo, **try it and open an issue with a reproducible case**. Real-world layouts and platform-specific issues are especially useful. Stars are appreciated, but good reports and PRs matter more.

## Scope and safety

pkx does **not** install system runtimes, run a daemon, collect telemetry, guarantee native installation success or automatically roll back file changes. `status` checks tool readiness rather than installed dependencies; `plan` previews commands, not their downstream effects. Cargo and Go steps download dependencies without building. Read [design principles](docs/ARCHITECTURE.md) and [security notes](SECURITY.md).

---
<sub>Maintained by <a href="https://github.com/AmpedWasTaken">AmpedWasTaken</a> with contributions welcome. MIT licensed.</sub>
