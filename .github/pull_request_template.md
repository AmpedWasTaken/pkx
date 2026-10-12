## Why

<!-- Link an issue and explain the developer problem. -->

## What changed

<!-- Explain user behavior, compatibility and docs changes. -->

## Testing

- [ ] `gofmt` checked (Go changes)
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pkx`
- [ ] Relevant regression tests or reproduction details included

**Platforms tested:** <!-- Windows / macOS / Linux / not tested -->

## Safety and compatibility

- [ ] Native managers and lockfiles remain the source of truth
- [ ] No unexpected execution, network changes or out-of-workspace writes
- [ ] User-facing docs updated (or not applicable)
- [ ] No secrets, private logs, package caches or generated binaries

**Known limitations / follow-ups:**
