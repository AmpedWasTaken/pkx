# Contributing to pkx

Contributions are welcome. For changes to commands or the bundle schema, open an issue first to discuss backward compatibility.

1. Install Go 1.23+ and clone the repository.
2. Keep the core CLI dependency-free where possible.
3. Add tests for new adapters, options, and error conditions.
4. Run `gofmt -w cmd/pkx internal/pkx`, `go test -race ./...`, and `go vet ./...`.
5. Open a focused PR explaining the behavior and showing a reproducible example.

Design principles: preserve native lockfiles, avoid shell execution, offer dry-run previews before modifying projects, don't cross workspace boundaries, and don't hide native tool errors.

For bug reports include OS, `pkx version`, command, output, and non-sensitive sample manifests.
