# pkx

**Package eXchange** — one command to manage dependencies in mixed-language repositories.

pkx uses the **native** package managers already present in your project. It doesn't replace npm, pip, Composer, Cargo, or Go modules, and it does not invent a new cross-language dependency resolver.

> **Status:** v0.1.0 developer preview. The local CLI works without a hosted registry or account. A community marketplace is a possible future feature, not part of this release.

## Install

Build with Go 1.23 or later:

```sh
git clone https://github.com/AmpedWasTaken/pkx.git
cd pkx
go install ./cmd/pkx
```

Ensure Go's binary installation directory (`$(go env GOPATH)/bin`) is on your `PATH`.

## Quick start

```sh
cd your-project
pkx scan
pkx install --dry-run
pkx doctor
pkx install
```

By default pkx discovers supported projects up to four directory levels below your current directory. It ignores generated directories such as `node_modules`, `vendor`, `.venv`, `target`, and `dist`, and doesn't follow symlinks.

**Review the dry run before installing third-party projects.** Native package managers can execute lifecycle scripts and download dependencies.

## Commands

| Command | Description |
| --- | --- |
| `pkx scan --json` | Discover supported ecosystems in machine-readable form |
| `pkx install --dry-run` | Preview all native install/download commands |
| `pkx install` | Execute the plan, stopping on the first error |
| `pkx install --only node` | Install only one manager or ecosystem |
| `pkx doctor` | Report missing package-manager executables |
| `pkx bundle init` | Create a local `pkx.json` without overwriting one |
| `pkx bundle install pkx.json --dry-run` | Preview adding a dependency bundle |
| `pkx bundle install pkx.json` | Apply a bundle using native managers |
| `pkx version` | Print the CLI version |

`scan`, `install`, and `doctor` also accept `--dir PATH` and `--depth N` (default 4, range 0–20). The bundle commands accept `--dir PATH`.

## Native manager support

| Manifest / lockfile | Selected manager | Command |
| --- | --- | --- |
| `package.json` + `package-lock.json` | npm | `npm ci` |
| `package.json` without a lockfile | npm | `npm install` |
| `package.json` + `pnpm-lock.yaml` | pnpm | `pnpm install --frozen-lockfile` |
| `package.json` + `yarn.lock` | Yarn | `yarn install` with matching locked-install flag |
| `package.json` + `bun.lock` / `bun.lockb` | Bun | `bun install --frozen-lockfile` |
| `composer.json` | Composer | `composer install --no-interaction` |
| `Cargo.toml` | Cargo | `cargo fetch` |
| `go.mod` | Go | `go mod download` |
| `pyproject.toml` + `uv.lock` | uv | `uv sync --locked` |
| `pyproject.toml` + `poetry.lock` | Poetry | `poetry install --no-interaction` |
| `requirements.txt` | pip | Create/reuse local `.venv` and install there |

When the Node `packageManager` property is set, it takes precedence. If conflicting Node lockfiles exist with no explicit manager, pkx stops rather than guessing.

A plain `pyproject.toml` without a recognized lockfile or `requirements.txt` is not currently supported. Cargo and Go download dependencies without building the project.

## Dependency bundles

A bundle is an explicit, reviewable JSON document. It never runs automatically as part of `pkx install`:

```json
{
  "schema": 1,
  "name": "my-fullstack-bundle",
  "packages": [
    {"manager": "pnpm", "directory": "frontend", "dependencies": ["vue", "axios"], "devDependencies": ["vite"]},
    {"manager": "composer", "directory": "backend", "dependencies": ["laravel/sanctum"]}
  ]
}
```

Supported bundle managers are npm, pnpm, Yarn, Bun, Composer, and pip. The declared directories must already exist inside the workspace. Bundles use native package managers, so they may modify manifests and lockfiles. Python dependencies use a local `.venv`, created when missing.

## Limitations and security

- pkx **does not automatically install missing runtimes**. Use `pkx doctor` to see what is missing.
- There is no automatic rollback after a failed command. Inspect `git diff` if a native manager modified your project before failing.
- Install plans run sequentially. No background daemon or telemetry is used.
- Bundles cannot contain shell commands, and package names that look like CLI options are rejected. This is **not a security sandbox**.
- There is no remote bundle registry, signing service, AI dependency fixer, or built-in vulnerability scanner in this version.

For contribution guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md). For the threat model, see [SECURITY.md](SECURITY.md).

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/pkx
```

MIT licensed.
