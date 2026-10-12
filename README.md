# pkx

**Package eXchange** — run native package managers across a mixed-language repository with one command.

pkx doesn't replace npm, pip, Composer, or Cargo. It discovers the manifests already in your repository, chooses the appropriate native commands, and executes them from the correct directories. It also supports small, reviewable JSON dependency bundles.

> **Status:** v0.4.0 developer preview. Local-first, no account, registry, telemetry, or background service. The planned public bundle marketplace is **not implemented**.

## Quick start

pkx only requires its own CLI plus the package managers your project uses. To build from source, install Go 1.23 or later:

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
go install ./cmd/pkx
```

Add `$(go env GOPATH)/bin` to your `PATH` if `pkx` isn't found. Versioned release archives are produced by GitHub Actions when a release tag is published. After a release exists, you can install the latest published binary **without Go** using the checked-in installer scripts (review the scripts first):

```sh
# Linux / macOS; from the pkx repository checkout
sh scripts/install.sh
# Or request a specific published release: sh scripts/install.sh v0.4.0
```

```powershell
# Windows PowerShell; from the pkx repository checkout
.\scripts\install.ps1
# Or: .\scripts\install.ps1 -Version v0.4.0
```

Both scripts verify the published SHA256 checksum and install to your user directory (`~/.local/bin`); they do not modify PATH automatically. If there is no published GitHub release yet, build from source using the Go command above.

Then, from **any project directory**:

```sh
pkx status                 # What projects, tools and lockfiles do I have?
pkx plan                   # What commands would pkx run? No side effects
pkx up                     # Validate runtimes, then install dependencies
```

If a tool is missing, `pkx status` and `pkx doctor` show guidance. pkx doesn't install Node, Python, PHP or other system tools behind your back.

## Everyday commands

| Command | What it does |
| --- | --- |
| `pkx status` | Show detected projects, lockfiles and environment warnings |
| `pkx status --json` | Structured status for scripts and editor integrations |
| `pkx plan` | Preview the native install commands; never execute anything |
| `pkx plan --json` | Structured command plan using workspace-relative directories |
| `pkx up` | Check required tools and version pins, then install dependencies |
| `pkx scan [--json]` | List supported projects and selected package managers |
| `pkx install --dry-run` | Older preview syntax, kept for compatibility |
| `pkx install --only node` | Install only one ecosystem or manager (`pnpm`, `pip`, etc.) |
| `pkx doctor` | Check for missing tools and incompatible runtime pins |
| `pkx doctor --json` | Machine-readable `ready`, `projects`, and diagnostics; exits 1 on issues |
| `pkx completion bash` | Generate tab completion for your shell (also Zsh, Fish, PowerShell) |
| `pkx init` | Create a local `pkx.json` starter bundle (no dependencies added) |
| `pkx bundle install pkx.json --dry-run` | Preview applying a dependency bundle |
| `pkx version` | Show CLI version |

### Short commands and help

```sh
pkx i --dry-run       # 'install' shortcut
pkx ls                # 'scan' shortcut
pkx help plan         # Command-specific examples and flags
pkx up --help         # The same help through a command
pkx plan --dir ./apps # Plan for another workspace
```

`scan`, `status`, `plan`, `up`, `install`, and `doctor` accept `--dir PATH` and `--depth N` (default 4, maximum 20). `plan`, `up`, and `install` accept `--only MANAGER` to select one ecosystem/manager. `--json` is available on `scan`, `plan`, `status`, and `doctor`. `doctor` additionally accepts `--only` to focus on a single manager or ecosystem.

### Terminal tab completion

pkx generates shell completion scripts locally; it never edits shell startup files:

```sh
# Bash (if bash-completion is enabled)
mkdir -p ~/.local/share/bash-completion/completions
pkx completion bash > ~/.local/share/bash-completion/completions/pkx

# Zsh (add the directory to fpath before compinit)
mkdir -p ~/.zfunc
pkx completion zsh > ~/.zfunc/_pkx

# Fish
mkdir -p ~/.config/fish/completions
pkx completion fish > ~/.config/fish/completions/pkx.fish
```

For PowerShell, inspect `pkx completion powershell` then add its output to your PowerShell profile. Completion suggestions are static and require no network or background process.

**Inspect a plan before running a third-party repository.** pkx invokes package managers with argument arrays (not shell strings), but those tools may download and execute untrusted lifecycle scripts.

## Supported project types

| Project marker | Manager | Install behavior |
| --- | --- | --- |
| `package.json` + `package-lock.json` | npm | `npm ci` |
| `package.json` (no lockfile) | npm | `npm install` |
| `package.json` + `pnpm-lock.yaml` | pnpm | `pnpm install --frozen-lockfile` |
| `package.json` + `yarn.lock` | Yarn | Yarn 1: `--frozen-lockfile`; Yarn 2+: `--immutable` when specified in `packageManager` |
| `package.json` + `bun.lock` / `bun.lockb` | Bun | `bun install --frozen-lockfile` |
| `composer.json` | Composer | `composer install --no-interaction` |
| `Cargo.toml` | Cargo | `cargo fetch` (downloads dependencies, does not build) |
| `go.mod` | Go | `go mod download` |
| `uv.lock` + `pyproject.toml` | uv | `uv sync --locked` |
| `poetry.lock` + `pyproject.toml` | Poetry | `poetry install --no-interaction` |
| `requirements.txt` | pip | create/reuse local `.venv`, then install requirements |

If `package.json` defines `packageManager`, that supported manager takes precedence over lockfile inference. If different Node lockfiles disagree and no manager is specified, pkx stops instead of guessing. An unsupported Python `pyproject.toml` without a supported lockfile or requirements file is ignored.

Python installs with pip use a **project-local `.venv`**, not the global interpreter. On Windows, pkx looks for `python`; on Unix, `python3`.

## Bundles

Bundles are **a separate, explicit operation**. Ordinary `pkx install` never installs arbitrary bundle contents. Start with:

```sh
pkx init                     # Generates a starter with an empty packages array
# Edit pkx.json to add the dependencies you actually want
pkx bundle install pkx.json --dry-run
pkx bundle install pkx.json
```

A `pkx.json` bundle has this form:

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

Supported bundle managers: npm, pnpm, Yarn, Bun, Composer, and pip. Bundles add dependencies through their native manager and **can modify the project's manifests and lockfiles**; inspect the resulting diff. pip bundle dependencies install into the target directory's `.venv`, which pkx creates if necessary. Bundles can't contain arbitrary shell commands. Bundle directories must exist inside the workspace and cannot escape it through `..` or symlinks. Paths are relative to `--dir` (default: current directory).

## Design decisions and limits

- No separate cross-ecosystem lockfile or dependency solver. Existing lockfiles remain the source of truth.
- No automatic downloading of runtimes. `pkx doctor` reports missing commands so you can install them deliberately.
- A deterministic, sequential plan avoids partially parallel writes to shared lockfiles. A failed command stops remaining steps.
- Project discovery is bounded by `--depth`; generated folders and hidden directories are ignored.
- No automatic rollback: installers can change files before failing. Review `git diff` and native tool output.
- No registry, signatures, audit command, or package publishing in this version.
- `status` checks environment readiness, not whether all dependencies have already been installed.
- `plan` previews native commands; it does not guarantee the native installer will succeed.

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/pkx
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
