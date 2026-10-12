#!/usr/bin/env bash
# Build all published binaries and verify each archive and checksum before release.
set -euo pipefail
version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Usage: bash scripts/package-release.sh X.Y.Z" >&2
  exit 2
fi
if [[ -d dist && -n "$(find dist -mindepth 1 -maxdepth 1 -print -quit)" ]]; then
  echo "dist/ must be empty before preparing a release" >&2
  exit 1
fi

mkdir -p dist
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

for os in linux darwin windows; do
  for arch in amd64 arm64; do
    executable=pkx
    if [[ "$os" == "windows" ]]; then executable=pkx.exe; fi
    GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X github.com/AmpedWasTaken/pkx/internal/pkx.Version=$version" \
      -o "$stage/$executable" ./cmd/pkx
    name="pkx_v${version}_${os}_${arch}"
    if [[ "$os" == "linux" && "$arch" == "amd64" && "$(uname -s)" == "Linux" && "$(uname -m)" == "x86_64" ]]; then
      [[ "$("$stage/pkx" version)" == "pkx $version" ]]
    fi
    if [[ "$os" == "windows" ]]; then
      (cd "$stage" && zip -q "$OLDPWD/dist/${name}.zip" "$executable")
    else
      tar -czf "dist/${name}.tar.gz" -C "$stage" "$executable"
    fi
    rm -f "$stage/$executable"
  done
done

(
  cd dist
  sha256sum ./* > SHA256SUMS
  sha256sum -c SHA256SUMS
  for asset in *.tar.gz; do
    [[ "$(tar -tzf "$asset")" == "pkx" ]]
  done
  for asset in *.zip; do
    [[ "$(unzip -Z -1 "$asset")" == "pkx.exe" ]]
  done
  [[ "$(find . -maxdepth 1 -name 'pkx_*' -type f | wc -l)" -eq 6 ]]
)
echo "Release packages and SHA256SUMS verified for v$version"
