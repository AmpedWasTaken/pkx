# Publishing pkx releases

This checklist is for maintainers. A successful CI run is necessary, but does not prove every combination of native package manager and platform works.

## Preflight

1. Merge the release-preparation PR into `main`. All Linux, macOS, and Windows CI jobs must pass, including the real npm smoke test.
2. Confirm `internal/pkx/types.go` has the version to publish, e.g. `0.5.0`, and `docs/releases/v0.5.0.md` exists.
3. Confirm `go test ./...`, `go test -race ./...`, `go vet ./...`, and `bash scripts/smoke.sh` work on a trusted workstation. Test the examples in README against your own real project before advertising broad compatibility.
4. Review `SECURITY.md`, reported issues and third-party dependencies. Do not publish binaries from unreviewed source.
5. Make sure the GitHub Actions permission `contents: write` is allowed for the release workflow. Repository tag/ruleset restrictions may prevent workflow tag creation.

## One-click GitHub release

Open **Actions → Release binaries → Run workflow**. Choose **main**, enter `v0.5.0`, and submit. The workflow validates that it is running from main, matches source version and release notes, and checks formatting/tests, race tests, vet and a real npm smoke test before uploading anything. It builds all six archives (Linux/macOS/Windows × amd64/arm64), validates the archive layout and checksums, then creates a GitHub release and tag at the validated commit. It refuses to overwrite an existing release.

Alternatively, after successful main CI, create and push a tag pointing to `main`:

```sh
git fetch origin main
git switch main
git pull --ff-only
git tag -a v0.5.0 -m "pkx v0.5.0"
git push origin v0.5.0
```

This triggers the same release workflow using the pre-existing tag. Avoid creating both at once.

## Verify the published release

- Check `https://github.com/AmpedWasTaken/pkx/releases/tag/v0.5.0` exists and contains six archives and `SHA256SUMS`.
- On a clean Windows, Linux and macOS machine, download an asset, check its SHA256 against the published sums, extract it and run `pkx version`.
- From a temporary npm project, check `pkx plan`, `pkx up`, `pkx run` and `pkx add --dry-run`. Repeat with Composer/Python projects when available.
- Test the user-level install scripts with an explicit release version. They do not alter PATH.
- Report any failed install in a new issue; if a release is compromised, use GitHub's release deletion/incident procedures instead of silently overwriting assets.

## Next versions

Update the Go version constant, add `docs/releases/vX.Y.Z.md`, update the changelog, and repeat preflight. This project follows Semantic Versioning within its pre-1.0 compatibility scope.
