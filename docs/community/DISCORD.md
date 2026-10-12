# pkx community Discord — setup plan

**Status: proposed, not created.** There is no verified invite link in this repository, and no account/server is provisioned by these docs. Never paste a placeholder invite URL into README.

## Purpose

Make contributors comfortable asking questions, share real reproduction fixtures, and discuss small PRs without turning Discord into the project's sole source of truth. Design and decisions should still be summarized in public GitHub issues/PRs.

## Suggested channels

- `#welcome-and-rules` — Code of Conduct, links to docs and how to get help
- `#introductions` — who you are and what ecosystem you work in
- `#help-and-support` — usage questions with redacted logs
- `#adapter-lab` — lockfile fixtures, tests, compatibility discussions
- `#contributions` — PR reviews and good-first-issue pairing
- `#releases` — maintainer announcements (not a bot spam stream)

Start with **two moderators** who can respond to reports before publishing an invite. Require acceptance of rules, add a clear escalation path to maintainers, and avoid open webhook credentials or untrusted bot integrations.

## Suggested pinned message

Welcome to pkx — a small open-source project for simplifying native package-manager setup in mixed-language repos. Share what you build, tell us which manager cases broke, and ask for help with a first PR. Don't share secrets or proprietary source. Security reports belong in the private process described by [SECURITY.md](../../SECURITY.md); GitHub issues remain the source of record for decisions.

## Launch checklist

- [ ] Owner creates the server and confirms moderation responsibility.
- [ ] Apply the repository Code of Conduct and configure reporting.
- [ ] Invite a small initial group who have actually tested pkx.
- [ ] Publish a **real** invite link only after roles and channels are ready.
- [ ] Link the server from README and website in a reviewed change.
- [ ] Post monthly community updates and respond to unanswered beginner issues.

Don't optimize for a member count: useful reviews and repeat contributors matter more.
