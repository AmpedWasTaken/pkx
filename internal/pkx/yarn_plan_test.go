package pkx

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func yarnPlanSteps(t *testing.T, root string) []Step {
	t.Helper()
	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "yarn")
	if err != nil {
		t.Fatal(err)
	}
	return steps
}

func yarnPlanString(t *testing.T, root string) string {
	t.Helper()
	steps := yarnPlanSteps(t, root)
	if len(steps) != 1 {
		t.Fatalf("expected a single yarn step, got %#v", steps)
	}
	step := steps[0]
	if step.Manager != "yarn" || step.Binary != "yarn" {
		t.Fatalf("expected the yarn manager and binary, got %#v", step)
	}
	return step.String()
}

// Yarn Classic (v1) understands --frozen-lockfile; Berry (v2+) replaced it
// with --immutable. The adapter must pick the flag from packageManager,
// not from the lockfile shape, because the yarn.lock format is similar.
func TestYarnClassicPlanUsesFrozenLockfile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@1.22.19"}`)
	put(t, root, "yarn.lock", "# yarn lockfile v1\n")
	got := yarnPlanString(t, root)
	want := "yarn install --frozen-lockfile"
	if got != want {
		t.Fatalf("yarn classic plan: got %q want %q", got, want)
	}
}

func TestYarnBerryPlanUsesImmutable(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@4.5.3"}`)
	put(t, root, "yarn.lock", "__metadata:\n  version: 8\n")
	got := yarnPlanString(t, root)
	want := "yarn install --immutable"
	if got != want {
		t.Fatalf("yarn berry plan: got %q want %q", got, want)
	}
}

// A yarn.lock with no packageManager field is ambiguous: pkx keeps the
// Yarn 1 flag, the only mode that reads that lockfile header format.
func TestYarnMissingPackageManagerKeepsClassicFlag(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	put(t, root, "yarn.lock", "# yarn lockfile v1\n")
	if got, want := yarnPlanString(t, root), "yarn install --frozen-lockfile"; got != want {
		t.Fatalf("classic default plan: got %q want %q", got, want)
	}
}

// No lockfile: install may resolve freely, so no freeze flag is passed.
// Detection still reports yarn from the packageManager field.
func TestYarnMissingLockfilePlansPlainInstall(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@4.5.3"}`)
	steps := yarnPlanSteps(t, root)
	if len(steps) != 1 || steps[0].String() != "yarn install" {
		t.Fatalf("expected a plain yarn install, got %#v", steps)
	}
}

// The exact lockfile-only spellings that Discover accepts must keep
// choosing yarn so the plan above is reachable for real repositories.
func TestYarnDetectionFromPackageManager(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@4.5.3"}`)
	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "yarn" || projects[0].Ecosystem != "node" {
		t.Fatalf("expected one yarn project, got %#v", projects)
	}
	dir := t.TempDir()
	put(t, dir, "package.json", `{}`)
	put(t, dir, "yarn.lock", "# yarn lockfile v1\n")
	projects, err = Discover(dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "yarn" {
		t.Fatalf("expected yarn from yarn.lock, got %#v", projects)
	}
	if filepath.IsAbs(projects[0].Directory) || strings.ContainsAny(projects[0].Directory, "/\\") {
		t.Fatalf("expected a workspace-relative directory, got %q", projects[0].Directory)
	}
}

// Steps must carry args as an argv slice, never a shell string, so the
// plan can be shown read-only and executed without interpolation.
func TestYarnPlanArgsAreArgv(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"yarn@1.22.19"}`)
	put(t, root, "yarn.lock", "# yarn lockfile v1\n")
	steps := yarnPlanSteps(t, root)
	want := []string{"install", "--frozen-lockfile"}
	if !reflect.DeepEqual(steps[0].Args, want) {
		t.Fatalf("argv: got %#v want %#v", steps[0].Args, want)
	}
	if strings.ContainsAny(strings.Join(steps[0].Args, " "), "&|;") {
		t.Fatalf("shell metacharacters in plan args: %#v", steps[0].Args)
	}
}
