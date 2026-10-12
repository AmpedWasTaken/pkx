# RFC 0001: Remote bundle exchange (proposal)

**Status:** Draft / discussion wanted. **Not implemented in pkx v0.5.0.**

## Problem

Local `pkx.json` bundles are useful, but there is no way to discover a versioned bundle maintained by another contributor. A public catalog can help developers find bundles while keeping native lockfiles as the source of truth.

## Scope and incremental delivery

1. **Starter included in this PR:** a deliberately offline fixture at `examples/registry/catalog.v1.json` and a documented catalog shape. It is **not** a hosted service, a public registry, or an installable remote bundle.
2. **Read-only MVP:** `pkx bundle search QUERY` and `pkx bundle info OWNER/NAME@VERSION` query a pinned HTTPS catalog with timeouts, response-size limits and an explicit cache policy. Network results are untrusted.
3. **Verified retrieval:** download an immutable, versioned `pkx.json` artifact; compare content digest with the catalog record; validate the existing schema-1 bundle, workspace-boundary rules and dependency specifiers before showing a dry-run plan. A digest alone is **not** identity or author authenticity.
4. **Explicit install:** only after showing owner, URL, version, digest and native commands, require explicit opt-in. Never automatically install a remote bundle during `pkx up`. No arbitrary scripts embedded in the registry format.
5. **Publishing / moderation (later):** only if discovery works reliably should we introduce authenticated publishing, namespace ownership, revocation, signatures, abuse reports and a hosted registry UI.

## Catalog draft

The first sample index has `schema`, `bundles[]`, and per-entry `id`, `version`, `summary`, `manifest_path`, `status`. The `manifest_path` in this fixture points to a **local example only**. It is not a portable remote URL and cannot be used to install anything through the current CLI.

Future network catalogs must use an explicit server-controlled schema rather than blindly trusting paths or arbitrary URLs from this sample.

## Security acceptance criteria

- Never run downloaded shell commands or scripts as part of catalog parsing.
- TLS with redirect restrictions; reject local/private/loopback, file:// and unsafe URL schemes to prevent SSRF.
- Enforce max catalog/artifact bytes and request timeouts. Fail closed on invalid JSON, digests, unsupported schemas and incompatible versions.
- Record artifact digest in any future bundle lock/pin; users can see provenance details before an explicit install.
- Native packages can execute lifecycle scripts. Show this limitation prominently, even when catalog metadata is verified.
- Unit tests for malformed catalog, duplicate owners/versions, redirect abuse, invalid digest, size limits and approval flow.

## Non-goals for initial PRs

A hosted SaaS, paid marketplace, package publishing CLI, automatic runtime downloads, remote code execution, star counts, ratings, and unverified popularity claims.

## Suggested first contributor PR

Implement a **pure, offline** parser/validator for the catalog fixture, returning deterministic typed metadata with no network access. Add tests for duplicate IDs, invalid versions and path traversal. Then propose the read-only search interface separately.
