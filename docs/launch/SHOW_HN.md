# Show HN draft — NOT POSTED

**Suggested title:** Show HN: pkx – one command to set up mixed-language repos using native package managers

**Suggested URL:** https://github.com/AmpedWasTaken/pkx

**Proposed first comment:**

Hi HN, I built pkx, a small, MIT-licensed Go CLI for developers who switch between Node, PHP, Python, Rust and Go projects.

In a repository, `pkx scan` detects supported manifests, `pkx plan` shows the exact native commands, and `pkx up` checks the environment and installs with the tools you already use. It doesn't replace npm, Composer, pip, Cargo or Go modules, and it doesn't introduce a pkx lockfile.

I made this because getting a multi-project repo running often means reading docs to figure out which managers and lockfiles are in play before you can start coding.

The v0.5.0 release supports install orchestration for 10 managers. Adding/removing dependencies and running declared scripts is narrower: npm, pnpm, Yarn, Bun and Composer only. That's a real gap, and I'm hoping other developers can help improve the adapters and tests.

The source, issues and a working adapter example are available here: https://github.com/AmpedWasTaken/pkx

You can download a platform binary from GitHub Releases and verify the published SHA256 sums; there's also an optional one-line installer, but I know some developers prefer not to execute a remotely hosted shell script.

I'm specifically looking for feedback from people working with mixed-language monorepos: what did it misdetect? Which manager behavior or version pin caused trouble? Does a preview-first setup tool meaningfully help your workflow?

This is still a 0.x project. I would rather fix real compatibility reports than add a long list of speculative features. Thanks for taking a look.

## Before posting

- Check current GitHub CI and release assets.
- Actually try the linked commands on one or more real-world repositories.
- Reply personally to concrete bug reports rather than posting marketing replies.
- Do not claim enterprise adoption, proven security or completed marketplace features.
