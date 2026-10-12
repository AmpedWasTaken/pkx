# Roadmap

This is a **direction, not a delivery promise**. Features and priorities can change based on testing, contributor availability and feedback. The current implementation is the source of truth.

## Shipped in v0.5.0

- Multi-ecosystem discovery and native install planning
- `status`, `plan`, `doctor`, `up`, JSON output and CLI aliases
- Basic Node workspace handling and runtime version diagnostics
- Local opt-in JSON dependency bundles
- Cross-platform release smoke tests and verified release packaging
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

## Feature proposals and community issues

The roadmap illustration in the README uses **shipped**, **next**, **planned** and **design** milestones. This is qualitative status, not percentage completion; no date is promised.

| Area | Status | Where to contribute |
| --- | --- | --- |
| Reliability and runtime pin tests | Next | [#6](https://github.com/AmpedWasTaken/pkx/issues/6) |
| Real mixed-language example | Next | [#5](https://github.com/AmpedWasTaken/pkx/issues/5) |
| Script discovery and monorepo selection | Planned | [#16](https://github.com/AmpedWasTaken/pkx/issues/16), [#17](https://github.com/AmpedWasTaken/pkx/issues/17) |
| **Remote bundle marketplace** | Static curated preview and read-only discovery/fetch in v0.6 development; publishing and remote installs still unimplemented | [RFC 0001](rfcs/0001-bundle-registry.md), [#12](https://github.com/AmpedWasTaken/pkx/issues/12), [#13](https://github.com/AmpedWasTaken/pkx/issues/13) |
| **Automatic runtime installation** | Design only; no runtime install execution | [RFC 0002](rfcs/0002-runtime-bootstrap.md), [#14](https://github.com/AmpedWasTaken/pkx/issues/14), [#15](https://github.com/AmpedWasTaken/pkx/issues/15) |

**Marketplace preview:** [browse the website](https://pkx.ampedwastaken.dev/marketplace/); installs are not automatic. **Technical starting points:** [offline registry catalog fixture](../examples/registry/catalog.v1.json) and [non-executable runtime plan fixture](../examples/runtime/bootstrap-plan.v1.json). They are examples used to discuss and validate future interfaces, **not available CLI features**.

## Potential ideas — not shipped

- Expand `pkx add` / `pkx remove` to more native ecosystems safely
- Richer `pkx run` script discovery, including nested monorepos
- `pkx doctor --fix`: opt-in, reviewable fixes with explicit approval
- A versioned, curated bundle exchange with ownership and provenance checks
- Optional dependency audit aggregation across supported ecosystems

A hosted marketplace, AI agent or desktop UI are **not prerequisites** for a useful CLI. Propose changes with a real developer workflow in a [feature issue](https://github.com/AmpedWasTaken/pkx/issues/new/choose).
