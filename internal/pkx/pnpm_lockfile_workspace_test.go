package pkx

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// A pnpm lockfile plus an explicit packageManager pin selects pnpm without
// consulting any native install: only package.json and pnpm-lock.yaml exist.
func TestPnpmExplicitManagerLockfilePlan(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"pnpm@9.12.0"}`)
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	got, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{{".", "node", "pnpm", "package.json"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("discover: got %#v want %#v", got, want)
	}

	steps, err := InstallPlan(root, got, "")
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []Step{{root, "pnpm", "pnpm", []string{"install", "--frozen-lockfile"}}}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// Without a packageManager pin the lockfile alone must choose pnpm.
func TestPnpmLockfileInfersManager(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"name":"app"}`)
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	got, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Manager != "pnpm" {
		t.Fatalf("discover: got %#v want one pnpm project", got)
	}

	steps, err := InstallPlan(root, got, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].String() != "pnpm install --frozen-lockfile" || steps[0].Directory != root {
		t.Fatalf("plan: got %#v want pnpm install --frozen-lockfile in %s", steps, root)
	}
}

// A package-lock.json next to pnpm-lock.yaml with no pin is ambiguous and must
// fail with an actionable diagnostic instead of guessing a manager.
func TestPnpmNpmLockfileConflictNoPin(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"name":"app"}`)
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	put(t, root, "package-lock.json", `{"lockfileVersion":3}`)

	got, err := Discover(root, 0)
	if err == nil {
		t.Fatalf("expected conflict, got projects %#v", got)
	}
	if !strings.Contains(err.Error(), "conflicting Node lockfiles") || !strings.Contains(err.Error(), "packageManager") {
		t.Fatalf("diagnostic does not name the conflict and the fix: %v", err)
	}
}

// pnpm workspace packages are installed by the workspace root only: members
// without their own lockfile collapse into the root plan.
func TestPnpmWorkspaceCollapsesMemberPlans(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"pnpm@9.12.0","workspaces":["packages/*"]}`)
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	put(t, root, "packages/app/package.json", `{"name":"app"}`)
	put(t, root, "packages/ui/package.json", `{"name":"ui"}`)
	put(t, root, "node_modules/skipped/package.json", `{"name":"skipped"}`)

	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{{".", "node", "pnpm", "package.json"}}
	if !reflect.DeepEqual(projects, want) {
		t.Fatalf("discover: got %#v want only the workspace root", projects)
	}

	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []Step{{root, "pnpm", "pnpm", []string{"install", "--frozen-lockfile"}}}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// A pnpm-workspace.yaml glob controls members that have no lockfile of their
// own; a member with its own lockfile stays independently installable.
func TestPnpmWorkspaceYAMLAndOwnLockfile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"pnpm@9.12.0"}`)
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	put(t, root, "pnpm-workspace.yaml", "packages:\n  - 'apps/*'\n  - '!apps/legacy'\n")
	put(t, root, "apps/web/package.json", `{"name":"web"}`)
	put(t, root, "apps/legacy/package.json", `{"name":"legacy","packageManager":"pnpm@9.12.0"}`)
	put(t, root, "apps/standalone/package.json", `{"name":"standalone"}`)
	put(t, root, "apps/standalone/pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	// apps/web is covered by the workspace; apps/legacy is excluded by the
	// negated pattern; apps/standalone has its own lockfile.
	want := []Project{
		{".", "node", "pnpm", "package.json"},
		{"apps/legacy", "node", "pnpm", "package.json"},
		{"apps/standalone", "node", "pnpm", "package.json"},
	}
	if !reflect.DeepEqual(projects, want) {
		t.Fatalf("discover: got %#v want %#v", projects, want)
	}

	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	// Only the root has a lockfile: --frozen-lockfile is added there and not to
	// a member whose own directory contains no pnpm-lock.yaml.
	wantSteps := []Step{
		{root, "pnpm", "pnpm", []string{"install", "--frozen-lockfile"}},
		{filepath.Join(root, "apps", "legacy"), "pnpm", "pnpm", []string{"install"}},
		{filepath.Join(root, "apps", "standalone"), "pnpm", "pnpm", []string{"install", "--frozen-lockfile"}},
	}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
	// The plan is data only; pkx must not have created anything on disk.
	if fileExists(filepath.Join(root, "apps", "web", "node_modules")) {
		t.Fatal("plan touched a workspace member")
	}
}
