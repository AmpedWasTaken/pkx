# Mixed-language example

A tiny, self-contained fixture that shows what `pkx` detects in a repository
with more than one ecosystem. It is safe to inspect: it has **no credentials**,
**no generated dependency directories** (no `node_modules/`, no `vendor/`) and
**no lockfiles**, so nothing is downloaded until you explicitly run an installer.

```
examples/mixed-language/
  javascript/        # Node project (npm)
    package.json
  php/               # PHP project (Composer)
    composer.json
```

Both manifests are intentionally minimal: `pkx` reads them and delegates to the
native package managers, so the fixture stays small and readable.

## Try it

Run these from the fixture root, `examples/mixed-language`. The output below is
what pkx v0.5.0 prints on this fixture; if your installed version differs, trust
your version and open an issue. The default scan depth (`4`) is enough to reach
both subdirectories.

```sh
pkx scan     # List detected projects
pkx status   # Inspect tool readiness
pkx plan     # Preview the exact install commands
```

### `pkx scan`

`scan` only lists what was found. It reads manifests and detects the package
manager for each directory:

```
node       npm        javascript (package.json)
php        composer   php (composer.json)
```

Add `--json` for a machine-readable list.

### `pkx status`

`status` adds lockfile detection and checks whether the required tools are on
`PATH`. Because this fixture ships no lockfiles, both projects report
`no lockfile detected`. On a machine with Node and Composer installed:

```
Workspace: /path/to/examples/mixed-language
2 projects, 2 planned native commands
  javascript     node       npm        no lockfile detected
  php            php        composer   no lockfile detected
Environment checks passed. Next: 'pkx plan' or 'pkx up'.
```

If a tool is missing it is listed instead, and `status` exits without running
anything, for example on a machine without PHP:

```
Environment: 2 issues to resolve:
  ! .: composer — command not found in PATH (fix: install PHP and Composer from https://getcomposer.org)
  ! php: php — PHP runtime missing (fix: install PHP and Composer from https://getcomposer.org)
Next: fix issues above, then run 'pkx up'.
```

### `pkx plan`

`plan` prints the exact native commands `pkx` would run, then stops:

```
Install plan (2 commands across 2 projects):
  javascript   npm            npm install
  php          composer       composer install --no-interaction
Nothing executed. Run 'pkx up' to check your environment and install.
```

Add `--json` for a machine-readable report, or `--only node` / `--only php` to
narrow the plan to a single ecosystem.

## Why `pkx plan` is safe, and why `pkx up` deserves review

`pkx scan`, `pkx status` and `pkx plan` are **read-only**. They walk the
directory tree, read manifests and inspect the environment; they never launch a
package manager and never write files. Using them first is how you find out what
an unfamiliar repository would do before anything runs.

`pkx up` is different: it runs the native installers (`npm install`,
`composer install`, ...) in each project directory. Native package managers may
execute third-party lifecycle scripts, for example npm `preinstall` /
`postinstall` hooks and Composer scripts. Those scripts run with your user
account and can do anything your user can do. That is a property of npm and
Composer, not something `pkx` adds, but it means you should:

1. Run `pkx plan` (and read the manifests) before `pkx up` on code you do not
   trust.
2. Prefer running `pkx up` inside a container, VM or throwaway checkout when the
   repository is untrusted.

This fixture deliberately declares no lifecycle scripts, so `pkx plan` on it
cannot surprise you.

## No network, no credentials

Nothing here contacts a registry, download server or private service. `pkx scan`
and `pkx plan` install nothing, and `pkx plan` never runs a lifecycle script.
The committed manifests intentionally omit lockfiles so the example does not
imply a resolved dependency set; installing is always an explicit, separate
step.