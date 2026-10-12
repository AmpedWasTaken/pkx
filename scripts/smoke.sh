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
scan_out="$("$bin" scan --dir "$root" --json)"
grep -Fq '"manager":"npm"' <<<"$scan_out"
plan_out="$("$bin" plan --dir "$root")"
grep -Fq 'npm install' <<<"$plan_out"
add_out="$("$bin" add axios --dir "$root" --project web --dry-run)"
grep -Fq 'nothing executed' <<<"$add_out"
dry_out="$("$bin" run check --dir "$root" --project web --dry-run)"
grep -Fq 'nothing executed' <<<"$dry_out"

# No external dependencies; disable npm's telemetry-like update, audit and funding requests.
export npm_config_audit=false npm_config_fund=false npm_config_update_notifier=false
"$bin" up --dir "$root"
"$bin" up --dir "$root" # Lockfile now exists: npm ci should be selected.
script_out="$("$bin" run check --dir "$root" --project web)"
grep -Fq 'pkx-script-ok' <<<"$script_out"
echo "pkx native npm smoke passed"
