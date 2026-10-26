package pkx

import (
	"path/filepath"
	"reflect"
	"testing"
)

func yarnPlan(t *testing.T, root string) []Step {
	t.Helper()
	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	return steps
}

// Yarn Classic (v1) freezes an existing lockfile.
func TestYarnClassicUsesFrozenLockfile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@1.22.22"}`)
	put(t, root, "yarn.lock", "# yarn lockfile v1\n")

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{{".", "node", "yarn", "package.json"}}
	if !reflect.DeepEqual(projects, want) {
		t.Fatalf("discover: got %#v want %#v", projects, want)
	}

	wantSteps := []Step{{root, "yarn", "yarn", []string{"install", "--frozen-lockfile"}}}
	if steps := yarnPlan(t, root); !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// Yarn Berry (v4) uses --immutable instead, and reports its version through
// yarnPath from .yarn/releases rather than through a v1 lockfile header.
func TestYarnBerryUsesImmutable(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@4.5.0","yarnPath":".yarn/releases/yarn-4.5.0.cjs"}`)
	put(t, root, "yarn.lock", "__metadata:\n  version: 8\n")
	put(t, root, ".yarnrc.yml", "nodeLinker: node-modules\n")

	wantSteps := []Step{{root, "yarn", "yarn", []string{"install", "--immutable"}}}
	if steps := yarnPlan(t, root); !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// A lockfile with no packageManager pin is treated as Classic, matching Yarn's
// own detection for a workspace that has not migrated to the version field.
func TestYarnLockfileWithoutPinDefaultsToClassic(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"name":"app"}`)
	put(t, root, "yarn.lock", "# yarn lockfile v1\n")

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "yarn" {
		t.Fatalf("discover: got %#v want yarn from the lockfile", projects)
	}
	wantSteps := []Step{{root, "yarn", "yarn", []string{"install", "--frozen-lockfile"}}}
	if steps := yarnPlan(t, root); !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// A packageManager pin selects Yarn and filters to it without a lockfile; the
// plan must not add a freezing flag when there is nothing to freeze.
func TestYarnPinWithoutLockfilePlansPlainInstall(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@4.5.0"}`)

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "yarn" {
		t.Fatalf("discover: got %#v want yarn from the pin", projects)
	}
	steps, err := InstallPlan(root, projects, "yarn")
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []Step{{root, "yarn", "yarn", []string{"install"}}}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
	if steps, err := InstallPlan(root, projects, "npm"); err == nil {
		t.Fatalf("--only npm must not plan yarn projects: %#v", steps)
	}
}

// The Yarn plan honours nested projects and --only yarn, and a lockfile next
// to an npm lockfile without a pin still fails with the conflict diagnostic.
func TestYarnNestedProjectsAndNpmConflict(t *testing.T) {
	root := t.TempDir()
	put(t, root, "web/package.json", `{"packageManager":"yarn@1.22.22"}`)
	put(t, root, "web/yarn.lock", "# yarn lockfile v1\n")
	put(t, root, "api/go.mod", "module example.com/api")

	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "yarn")
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []Step{{filepath.Join(root, "web"), "yarn", "yarn", []string{"install", "--frozen-lockfile"}}}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}

	conflict := t.TempDir()
	put(t, conflict, "package.json", `{"name":"app"}`)
	put(t, conflict, "yarn.lock", "")
	put(t, conflict, "package-lock.json", `{"lockfileVersion":3}`)
	if got, err := Discover(conflict, 0); err == nil {
		t.Fatalf("expected a lockfile conflict, got %#v", got)
	}
}
