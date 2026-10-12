#!/bin/sh
# Install a published pkx release for Linux or macOS, verifying its SHA256.
set -eu

usage() {
  printf '%s\n' 'Usage: sh install.sh [vX.Y.Z] [--add-to-path]' '' \
    'Download and verify an official pkx binary from GitHub Releases.' \
    'Default: latest published release. Installs to ~/.local/bin (override with PKX_INSTALL_DIR).' \
    'Optional: --add-to-path updates your user shell profile; restart the terminal afterward.'
}
fail() { printf 'pkx installer: %s\n' "$*" >&2; exit 1; }

case "${1:-}" in
  --help|-h) usage; exit 0 ;;
esac
[ "$#" -le 2 ] || { usage >&2; exit 2; }
add_to_path=0
version=""
for arg do
  case "$arg" in
    --add-to-path) add_to_path=1 ;;
    v[0-9]*) [ -z "$version" ] || fail 'specify only one release version'; version="$arg" ;;
    *) fail "unknown option or invalid version: $arg" ;;
  esac
done
command -v curl >/dev/null 2>&1 || fail 'curl is required to download releases'
command -v tar >/dev/null 2>&1 || fail 'tar is required to unpack releases'

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail 'only Linux and macOS are currently supported by this script' ;;
esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) fail 'unsupported CPU architecture; compile pkx from source instead' ;;
esac

if [ -z "$version" ]; then
  redirect=$(curl -fsSL -o /dev/null -w '%{url_effective}' \
    'https://github.com/AmpedWasTaken/pkx/releases/latest') || fail 'no published release found; try building from source'
  case "$redirect" in
    */releases/tag/*) version="${redirect##*/}" ;;
    *) fail 'no stable GitHub release found; specify a published version or build from source' ;;
  esac
fi
printf '%s\n' "$version" | LC_ALL=C grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$' \
  || fail 'invalid version (expected e.g. v0.4.0)'

asset="pkx_${version}_${os}_${arch}.tar.gz"
base="https://github.com/AmpedWasTaken/pkx/releases/download/${version}"
tmp=$(mktemp -d) || fail 'could not create temporary directory'
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
curl -fLsS --retry 2 -o "$tmp/$asset" "$base/$asset" || fail "could not download $asset"
curl -fLsS --retry 2 -o "$tmp/SHA256SUMS" "$base/SHA256SUMS" || fail 'checksum file unavailable'
expected=$(awk -v file="$asset" '$2 == file || $2 == "./" file { print $1 }' "$tmp/SHA256SUMS")
[ -n "$expected" ] || fail "no checksum for $asset in SHA256SUMS"
printf '%s\n' "$expected" | LC_ALL=C grep -Eq '^[0-9A-Fa-f]{64}$' || fail 'invalid checksum entry'
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$asset" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$tmp/$asset" | awk '{ print $1 }')
else
  fail 'sha256sum or shasum is required to verify the binary'
fi
[ "$actual" = "$expected" ] || fail 'SHA256 mismatch; refusing to install'

mkdir -p "$tmp/unpack"
tar -xzf "$tmp/$asset" -C "$tmp/unpack" pkx || fail 'archive extraction failed'
[ -f "$tmp/unpack/pkx" ] || fail 'archive did not contain pkx executable'
dest="${PKX_INSTALL_DIR:-${HOME:?}/.local/bin}"
mkdir -p "$dest"
install -m 0755 "$tmp/unpack/pkx" "$dest/pkx" || fail "could not install into $dest"
printf 'Installed pkx %s to %s/pkx\n' "$version" "$dest"
if [ "$add_to_path" -eq 1 ]; then
  # Profile editing is allowed only for the standard per-user install path.
  # A custom PKX_INSTALL_DIR may contain characters unsafe in shell profiles.
  [ "$dest" = "$HOME/.local/bin" ] || fail '--add-to-path requires default ~/.local/bin; configure custom PATH manually'
  case "${SHELL##*/}" in
    zsh) profile="$HOME/.zprofile" ;;
    bash) profile="$HOME/.bashrc" ;;
    *) if [ "$os" = darwin ]; then profile="$HOME/.zprofile"; else profile="$HOME/.profile"; fi ;;
  esac
  path_line='export PATH="$HOME/.local/bin:$PATH"'
  if ! grep -Fqx "$path_line" "$profile" 2>/dev/null; then
    printf '\n%s\n' "$path_line" >> "$profile" || fail "could not update $profile"
    printf 'Added ~/.local/bin to %s for future terminals.\n' "$profile"
  else
    printf 'PATH setup already exists in %s.\n' "$profile"
  fi
  printf 'Open a NEW terminal and run: pkx version\n'
else
  case ":$PATH:" in
    *":$dest:"*) : ;;
    *) printf 'Add %s to your PATH if pkx is not recognized, or use --add-to-path.\n' "$dest" ;;
  esac
fi
