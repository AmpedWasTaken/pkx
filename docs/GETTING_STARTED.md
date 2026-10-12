# Getting started

pkx is a Go CLI that orchestrates **existing** package managers. It does not replace your dependency resolver or lockfiles, and does not install system runtimes automatically.

## One-line global installation

Install directly from [pkx.ampedwastaken.dev](https://pkx.ampedwastaken.dev) without needing Go or npm.

**Windows CMD / PowerShell:**

```cmd
powershell -NoProfile -ExecutionPolicy Bypass -Command "& ([scriptblock]::Create((irm 'https://pkx.ampedwastaken.dev/install.ps1'))) -AddToPath"
```

**macOS/Linux:**

```sh
curl -fsSL https://pkx.ampedwastaken.dev/install.sh | sh -s -- --add-to-path
```

The installer uses the latest published GitHub release and verifies the downloaded archive's SHA256 against the release checksum file. The scripts themselves are served over HTTPS and are open source in [`pkx-website`](https://github.com/AmpedWasTaken/pkx-website) as `install.ps1` and `install.sh`. You should review scripts before running downloaded code on untrusted hosts.

These commands opt into user-level PATH setup: on Windows pkx installs to `%USERPROFILE%\.local\bin`; on macOS/Linux to `~/.local/bin`. Close the terminal and open a **new** one, then verify:

```sh
pkx version
```

On Windows CMD use `where pkx` if it is not found. In PowerShell use `Get-Command pkx`. On Unix use `command -v pkx`. The PATH update is user-specific, doesn't elevate privileges and doesn't modify the machine-wide PATH. If you don't want a PATH change, download the versioned binary manually or use `scripts/install.ps1` / `scripts/install.sh` without `-AddToPath` / `--add-to-path`.

**Alternative via Go (1.23+):** `go install github.com/AmpedWasTaken/pkx/cmd/pkx@v0.5.0`. Make sure `$(go env GOPATH)/bin` is on PATH.

pkx orchestrates native package managers; using it inside a repository may still require npm, PHP/Composer, Python, Go or other tools and native lifecycle scripts may execute. It does **not** install missing runtimes automatically.

## First five minutes

```sh
cd path/to/your/repository
pkx scan                 # Discover projects
pkx status               # Inspect tool readiness
pkx plan                 # Preview exact native commands
pkx up                   # Check environment and install
```

If a tool is missing, use its official installation instructions and rerun `pkx doctor`. `pkx up --dry-run` previews commands but does **not** probe version readiness. For diagnostics without changes, use `pkx status` or `pkx doctor`.

```sh
pkx plan --only composer  # One manager
pkx plan --dir ./apps     # Another directory
pkx scan --depth 6        # Deeper monorepo scan
pkx doctor --json         # Structured diagnostics
pkx help bundle           # Command help
```

## Supported managers and behavior

| Marker | Selected tool | Planned operation |
| --- | --- | --- |
| `package.json` + `package-lock.json` | npm | `npm ci` |
| `package.json`, no lockfile | npm | `npm install` |
| `package.json` + `pnpm-lock.yaml` | pnpm | `pnpm install --frozen-lockfile` |
| `package.json` + `yarn.lock` | Yarn | Locked install flag varies by pinned version |
| `package.json` + `bun.lock` / `bun.lockb` | Bun | `bun install --frozen-lockfile` |
| `composer.json` | Composer | `composer install --no-interaction` |
| `requirements.txt` | pip | Create/reuse local `.venv`; install requirements |
| `pyproject.toml` + `uv.lock` | uv | `uv sync --locked` |
| `pyproject.toml` + `poetry.lock` | Poetry | `poetry install --no-interaction` |
| `Cargo.toml` | Cargo | `cargo fetch` (no build) |
| `go.mod` | Go | `go mod download` (no build) |

A supported Node `packageManager` pin overrides lockfile inference. Conflicting Node lockfiles without a pin fail rather than guessing. A standalone `pyproject.toml` with no recognized lockfile or requirements file is not yet supported.

## Bundles are opt-in

`pkx init` creates an empty `pkx.json`. Edit its `packages` array, then preview with `pkx bundle install pkx.json --dry-run`. Target project directories must **already exist** inside the workspace. Ordinary `pkx up` does **not** automatically apply bundles.

## Shell completion

```sh
# Bash (requires bash-completion)
mkdir -p ~/.local/share/bash-completion/completions
pkx completion bash > ~/.local/share/bash-completion/completions/pkx

# Zsh (add ~/.zfunc to fpath before compinit)
mkdir -p ~/.zfunc
pkx completion zsh > ~/.zfunc/_pkx

# Fish
mkdir -p ~/.config/fish/completions
pkx completion fish > ~/.config/fish/completions/pkx.fish
```

On PowerShell inspect `pkx completion powershell` and add its output to your profile. pkx never edits shell startup files itself.

## Troubleshooting

- **No projects found:** Run from the repo root or increase `--depth` (maximum 20); generated/hidden directories are skipped.
- **Missing tool or wrong version:** `pkx doctor` offers guidance but never installs the runtime itself.
- **Native install failed:** Reproduce the printed native command in its directory, review `git diff` for partial changes, and share a sanitized error.
- **Binary not found:** Confirm the pkx installation folder is on `PATH` and restart your terminal after changes.
- **Untrusted repository:** Preview the plan and use an isolated environment; native package managers can execute third-party scripts.

Questions? Read [SUPPORT.md](../SUPPORT.md) or use the [issue forms](https://github.com/AmpedWasTaken/pkx/issues/new/choose).

## Adding dependencies and running scripts

In a workspace with one supported project:

```sh
pkx add axios --dry-run
pkx add axios
pkx add vite --dev
pkx remove axios
pkx run dev -- --host 127.0.0.1
```

For monorepos with several supported projects, add `--project frontend` (relative to `--dir`) to select the target explicitly. Add/remove and declared-script run are currently supported for npm, pnpm, Yarn, Bun and Composer. `pkx run` reads script names from `package.json` or `composer.json`. Native manager side effects still apply; preview using `--dry-run`.

## Marketplace (development preview)

The [marketplace](https://pkx.ampedwastaken.dev/marketplace/) hosts a small, curated catalog of example dependency bundles. CLI commands below are available in **unreleased v0.6 development source**, not the published v0.5.0 binary:

```sh
pkx bundle search frontend
pkx bundle info pkx/frontend-toolkit@0.1.0
pkx bundle fetch pkx/frontend-toolkit@0.1.0
pkx bundle install pkx.json --dry-run
```

`search` and `info` do not modify local files. `fetch` downloads a versioned manifest over HTTPS, enforces the catalog SHA256 digest, validates schema and path constraints and **refuses to overwrite** an existing `pkx.json`. It does **not** install dependencies. Package-manager installation remains a separate explicit command; untrusted packages may run scripts. Catalog SHA256 integrity does not prove maintainer identity. See [RFC 0001](rfcs/0001-bundle-registry.md).

## Universal package installation (v0.6 development, not in v0.5.0)

A package can be added without a registry prefix. pkx detects the current project from its manifest and lockfile, then chooses the existing native manager:

```sh
pkx install http-server --dry-run  # npm/pnpm/Yarn/Bun depending on project
pkx install http-server            # run native add/install in this project
pkx install httpx --project backend --dry-run
pkx install serde --project rust-tool --dry-run
```

In a project with only `package.json`, `pkx install http-server` picks npm unless a lockfile or `packageManager` declares another manager. In a Python project it uses pip, uv or Poetry as appropriate; Composer, Cargo and Go modules use their native tools. If multiple projects match, pkx refuses to guess and explains `--project DIR` and `--manager MANAGER`.

**Backwards compatible:** `pkx install` with no package still bootstraps all detected project dependencies. **No network search is required to choose a manager**; it uses local project manifests. The target registry must actually contain the package name you request, and package-name validity differs by ecosystem. The command does not silently switch language when a package name is absent in the current project's registry. `--dry-run` performs no installation. Native dependency installation can run third-party scripts.

## Advanced capability resolution (unreleased development)

The official Marketplace has one **universal library card** for each use case, with reviewed implementation choices per language. The CLI embeds a pinned copy of the `capabilities.v1.json` catalog so mapped installs never depend on live or user-provided catalog contents.

```sh
# HTTP Client works in Node, Python, PHP, Rust and Go projects
pkx install axios --dry-run
# In Python this previews httpx; in PHP guzzlehttp/guzzle, etc.

pkx install http-client --dry-run
pkx install structured-logging --dry-run
pkx install schema-validation --dry-run

# Interactive installation: confirms before replacing a named library
pkx install axios

# Reviewed, explicit approval in non-interactive scripts
pkx install axios --yes --project backend

# Trusted private registries only (opt out of official registry check)
pkx install my-private-package --skip-registry-check
```

The command uses local manifests/lockfiles to choose the target project and **native manager**. In a monorepo with multiple projects it stops and requires `--project DIRECTORY`, not a guess. For curated aliases (such as `axios`), it selects the purpose-similar library for the project's ecosystem. **It is NOT a drop-in API substitute** and can require import/configuration changes. Real installs check that the package name exists in the official native registry before executing the package manager; this is not a security audit or maintainer verification. An unmapped name is treated as an exact package name for that project's native ecosystem, not auto-guessed across all registries. `--dry-run` does not contact registries or install anything; only actual installs query package existence. A private registry can be selected only using the native tool and explicit `--skip-registry-check`.

This functionality is **not in the published v0.5.0 binary**. Catalog mappings must be reviewed and updated together with the Marketplace's versioned capability catalog.
