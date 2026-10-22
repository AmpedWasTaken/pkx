package pkx

import (
	"path/filepath"
	"reflect"
	"testing"
)

func cargoGoPlan(t *testing.T, root, only string) []Step {
	t.Helper()
	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, only)
	if err != nil {
		t.Fatal(err)
	}
	return steps
}

// assertOnlySkips proves a filter that matches nothing fails as a diagnostic
// instead of silently planning the other ecosystem.
func assertOnlySkips(t *testing.T, root, only string) {
	t.Helper()
	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	if steps, err := InstallPlan(root, projects, only); err == nil {
		t.Fatalf("--only %s must not plan other projects: %#v", only, steps)
	}
}

func TestCargoFetchPlan(t *testing.T) {
	root := t.TempDir()
	put(t, root, "Cargo.toml", "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n")
	put(t, root, "src/main.rs", "fn main() {}\n")

	got := cargoGoPlan(t, root, "")
	want := []Step{{root, "cargo", "cargo", []string{"fetch"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan: got %#v want %#v", got, want)
	}

	if got := cargoGoPlan(t, root, "cargo"); !reflect.DeepEqual(got, want) {
		t.Fatalf("--only cargo: got %#v want %#v", got, want)
	}
	if got := cargoGoPlan(t, root, "rust"); !reflect.DeepEqual(got, want) {
		t.Fatalf("--only rust ecosystem: got %#v want %#v", got, want)
	}
	assertOnlySkips(t, root, "go")
}

func TestGoModDownloadPlan(t *testing.T) {
	root := t.TempDir()
	put(t, root, "api/go.mod", "module example.com/api\n\ngo 1.23\n")
	put(t, root, "api/main.go", "package main\n\nfunc main() {}\n")
	dir := filepath.Join(root, "api")

	got := cargoGoPlan(t, root, "")
	want := []Step{{dir, "go", "go", []string{"mod", "download"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan: got %#v want %#v", got, want)
	}

	if got := cargoGoPlan(t, root, "go"); !reflect.DeepEqual(got, want) {
		t.Fatalf("--only go: got %#v want %#v", got, want)
	}
	assertOnlySkips(t, root, "cargo")
}

// A workspace with both manifest kinds keeps the two plans separate and in
// directory order, and each step carries its own directory.
func TestCargoAndGoPlansStaySeparate(t *testing.T) {
	root := t.TempDir()
	put(t, root, "crates/core/Cargo.toml", "[package]\nname = \"core\"\n")
	put(t, root, "api/go.mod", "module example.com/api\n")

	steps := cargoGoPlan(t, root, "")
	want := []Step{
		{filepath.Join(root, "api"), "go", "go", []string{"mod", "download"}},
		{filepath.Join(root, "crates", "core"), "cargo", "cargo", []string{"fetch"}},
	}
	if !reflect.DeepEqual(steps, want) {
		t.Fatalf("plan: got %#v want %#v", steps, want)
	}
	for _, step := range steps {
		if step.Binary != "go" && step.Binary != "cargo" {
			t.Fatalf("unexpected binary in %#v", step)
		}
	}
}

// Planning is data only: a failing --only match is a diagnostic, not a command.
func TestCargoGoOnlyWithoutMatch(t *testing.T) {
	root := t.TempDir()
	put(t, root, "Cargo.toml", "[package]\nname = \"demo\"\n")

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InstallPlan(root, projects, "go"); err == nil {
		t.Fatal("expected a diagnostic when no project matches --only go")
	}
}
