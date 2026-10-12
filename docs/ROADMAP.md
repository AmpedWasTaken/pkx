# Roadmap

This is a **direction, not a delivery promise**. Features and priorities can change based on testing, contributor availability and feedback. The current implementation is the source of truth.

## In the v0.4 developer preview

- Multi-ecosystem discovery and native install planning
- `status`, `plan`, `doctor`, `up`, JSON output and CLI aliases
- Basic Node workspace handling and runtime version diagnostics
- Local opt-in JSON dependency bundles
- Native `add`, `remove` and script `run` workflows for Node and Composer projects
- CI for Windows/macOS/Linux, shell completion and release installer scripts

## Before a stable v1.0

| Area | Desired outcome | Where contributors can help |
| --- | --- | --- |
| Real-world fixtures | Handle mixed-language layouts and lockfile edge cases | Minimal repros and regression tests |
| Platform reliability | Correct Windows/macOS/Linux paths and exit statuses | Platform-specific bug reports |
| Developer experience | Helpful failures and simple first-run instructions | UX feedback on fresh clones |
| Release reliability | Downloadable, verifiable binaries with documented support | Installation testing on each OS |
| Compatibility | Honest, tested version pin and workspace support | Native-manager test fixtures |

## Potential ideas — not shipped

- Expand `pkx add` / `pkx remove` to more native ecosystems safely
- Richer `pkx run` script discovery, including nested monorepos
- `pkx doctor --fix`: opt-in, reviewable fixes with explicit approval
- A versioned, curated bundle exchange with ownership and provenance checks
- Optional dependency audit aggregation across supported ecosystems

A hosted marketplace, AI agent or desktop UI are **not prerequisites** for a useful CLI. Propose changes with a real developer workflow in a [feature issue](https://github.com/AmpedWasTaken/pkx/issues/new/choose).
