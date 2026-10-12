# pkx early-adopter launch kit

**Drafts only — nothing here has been posted to Hacker News, dev.to, Discord or social media.** Review and adapt them to your own voice. Do not claim adoption numbers, production usage or features that are only in development.

- [Show HN draft](SHOW_HN.md)
- [dev.to article draft](DEVTO_ARTICLE.md)
- [Five small issue templates](ISSUE_TEMPLATES.md)
- [Honest README introduction alternatives](README_INTRO.md)
- [Discord/community setup checklist](../community/DISCORD.md)

## Ready-to-pick contributor issues

The first eight adapter/testing issues are live and labeled `good first issue`: [pnpm #26](https://github.com/AmpedWasTaken/pkx/issues/26), [Yarn #27](https://github.com/AmpedWasTaken/pkx/issues/27), [Bun #28](https://github.com/AmpedWasTaken/pkx/issues/28), [Composer #29](https://github.com/AmpedWasTaken/pkx/issues/29), [pip #30](https://github.com/AmpedWasTaken/pkx/issues/30), [uv/Poetry #31](https://github.com/AmpedWasTaken/pkx/issues/31), [Cargo/Go #32](https://github.com/AmpedWasTaken/pkx/issues/32), and [Node workspaces #33](https://github.com/AmpedWasTaken/pkx/issues/33).

## Before publishing

1. Run the latest main-branch CI; address current failures before making reliability claims.
2. Follow the [v0.5.0 release](https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0) installation steps on a fresh Windows, macOS and Linux environment.
3. Ask a handful of developers with real mixed-language repos to trial `pkx scan`, `pkx plan` and `pkx up` and report concrete bugs. Aim for around **10 voluntary early adopters**, not vanity metrics.
4. Ensure the official website and documentation display accurate version differences: **10 install managers, 5 add/remove/run managers**; no remote marketplace in v0.5.0.
5. When a real Discord server exists and has moderators, add its verified invite link to the website and CONTRIBUTING; don't publish placeholder or unsolicited invitations.

Suggested measurement: installs or trial confirmations volunteered by users, issues filed, contributors who return, and supported manager cases actually tested. Avoid fictitious stars, testimonials, usage counts or security guarantees.
