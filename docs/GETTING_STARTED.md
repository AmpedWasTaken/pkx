# Getting started

pkx is a Go CLI that orchestrates **existing** package managers. It does not replace your dependency resolver or lockfiles, and does not install system runtimes automatically.

## Install pkx globally — no Go/npm required

**Recommended Windows CMD setup (requires Git to check out the installer):**

```cmd
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\install.ps1 -Version v0.5.0 -AddToPath
```

The script downloads the published archive, verifies it using `SHA256SUMS`, places `pkx.exe` in `%USERPROFILE%\.local\bin`, and **only with `-AddToPath`** permanently adds that directory to your Windows *user* PATH. It does not need administrator privileges or modify system PATH. Close existing CMD/PowerShell windows and open a **new CMD**:

```cmd
where pkx
pkx version
```

If you cannot use Git or a PowerShell installer, use the [v0.5.0 release archive](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0) matching your architecture, verify the SHA256, place `pkx.exe` in `%USERPROFILE%\.local\bin`, and add that folder via **Edit environment variables for your account → Path → New**. Open a new terminal to pick it up.

**Linux and macOS:** from a reviewed Git checkout:

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
sh scripts/install.sh v0.5.0
```

The script verifies SHA256 and writes `~/.local/bin/pkx`. If not already on PATH, add `export PATH="$HOME/.local/bin:$PATH"` to `~/.bashrc` on Linux Bash or `~/.zprofile` on macOS Zsh, then restart the terminal. Confirm with `command -v pkx && pkx version`.

**With Go 1.23+:** `go install github.com/AmpedWasTaken/pkx/cmd/pkx@v0.5.0`, ensuring `$(go env GOPATH)/bin` (or your `GOBIN`) is on PATH.

pkx itself is a standalone CLI. Projects still need their underlying package managers to be installed for operations like `pkx up` and `pkx add`. No automatic system runtime installation exists yet.

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
