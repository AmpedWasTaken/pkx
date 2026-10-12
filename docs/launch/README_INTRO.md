# README intro alternatives — drafts

Current README already uses a concise, accurate intro. These are optional alternatives for future experiments.

## Option A — default

**pkx is a small, open-source CLI for repositories using more than one package manager.** Run `pkx scan` to detect projects, `pkx plan` to preview native commands, and `pkx up` to set up dependencies. It keeps npm, Composer, pip, Cargo and Go modules in control—no extra lockfile or account.

**Scope:** 10 managers for installation; npm/pnpm/Yarn/Bun/Composer for add/remove/run. Free, MIT-licensed, and still evolving.

## Option B — developer pain first

Your frontend needs pnpm. Your backend needs Composer. The tooling directory uses Python. Instead of memorizing the sequence, run `pkx plan` to see what the repository needs, then `pkx up` to execute the existing native managers.

pkx doesn't replace your package managers or lockfiles. It wraps their setup workflow in one small, auditable CLI.

## Option C — security-conscious audience

**No new package resolver. No automatic runtime installer. No hidden native commands.** pkx discovers supported project manifests, previews exact dependency-install commands, and runs the native manager only when explicitly requested.

Download verified binaries from GitHub Releases. Full compatibility and install-script trust limits are stated in the README.

## Do not claim

- "Works with every package manager" — installation supports ten named managers.
- "One command to add any package" — add/remove/run covers only Node managers and Composer.
- "Secure / audited bundles" — SHA256 of a bundle against a catalog doesn't establish author identity.
- Download or star counts without a contemporaneous source.
