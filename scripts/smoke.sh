#!/usr/bin/env bash
# A real, dependency-free npm project verifies more than dry-run command rendering.
set -euo pipefail

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/web"
cat > "$work/web/package.json" <<'JSON'
{
  "name": "pkx-smoke",
  "version": "0.0.0",
  "private": true,
  "scripts": { "check": "node -e \"console.log('pkx-script-ok')\"" }
}
JSON

suffix=""
if [[ "${RUNNER_OS:-}" == "Windows" ]]; then suffix=".exe"; fi
go build -o "$work/pkx$suffix" ./cmd/pkx
bin="$work/pkx$suffix"
root="$work"
if [[ "${RUNNER_OS:-}" == "Windows" ]]; then root="$(cygpath -m "$work")"; fi

"$bin" version | grep -Fx "pkx 0.5.0"
"$bin" scan --dir "$root" --json | grep -q '"manager":"npm"'
"$bin" plan --dir "$root" | grep -q 'npm install'
"$bin" add axios --dir "$root" --project web --dry-run | grep -q 'nothing executed'
"$bin" run check --dir "$root" --project web --dry-run | grep -q 'nothing executed'

# No external dependencies; disable npm's telemetry-like update, audit and funding requests.
export npm_config_audit=false npm_config_fund=false npm_config_update_notifier=false
"$bin" up --dir "$root"
"$bin" up --dir "$root" # Lockfile now exists: npm ci should be selected.
"$bin" run check --dir "$root" --project web | grep -q 'pkx-script-ok'
echo "pkx native npm smoke passed"
