# pkx

**Package eXchange** — run native package managers across a mixed-language repository with one command.

pkx doesn't replace npm, pip, Composer, or Cargo. It discovers the manifests already in your repository, chooses the appropriate native commands, and executes them from the correct directories. It also supports small, reviewable JSON dependency bundles.

> **Status:** v1 CLI. Local-first, no account, registry, telemetry, or background service. The public bundle marketplace discussed for the project is **not implemented**.

## Quick start

Requires Go 1.23+ to build from source. The native package managers for your projects must also be installed.

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
go install ./cmd/pkx

cd /path/to/your/project
pkx scan
pkx up --dry-run
pkx doctor
pkx up
```

On Windows, ensure `$(go env GOPATH)/bin` is on `PATH`; on Unix shells the usual location is `$(go env GOPATH)/bin` as well.

## Commands

| Command | What it does |
| --- | --- |
| `pkx scan [--json]` | List supported projects, managers, and manifests |
| `pkx install [--dry-run]` | Install or download project dependencies using native tools |
| `pkx install --only node` | Limit installation to an ecosystem or manager (`pnpm`, `pip`, etc.) |
| `pkx doctor` | Check which native commands are missing from PATH |
| `pkx bundle init` | Create a starter `pkx.json` without overwriting an existing file |
| `pkx bundle install pkx.json [--dry-run]` | Add packages declared in a bundle |
| `pkx version` | Print the CLI version |

`scan`, `install`, and `doctor` accept `--dir PATH` and `--depth N` (default 4, maximum 20). Output is ordered deterministically by project location. The scanner ignores common generated and dependency directories and never follows symlinks.

**Inspect with `--dry-run` first**, especially for third-party repositories. The command preview shows what will run and where. pkx invokes tools as argument vectors without a shell, but package managers may still run dependency lifecycle scripts and download code.

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
pkx bundle init
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