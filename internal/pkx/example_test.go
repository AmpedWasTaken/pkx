package pkx

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// exampleRoot locates the committed examples/mixed-language fixture from the
// package directory so the test exercises the real, checked-in example instead
// of a synthetic copy.
func exampleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "examples", "mixed-language")
	if _, err := os.Stat(filepath.Join(root, "javascript", "package.json")); err != nil {
		t.Fatalf("example fixture missing: %v", err)
	}
	return root
}

func TestExampleDiscovery(t *testing.T) {
	root := exampleRoot(t)
	got, err := Discover(root, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{
		{"javascript", "node", "npm", "package.json"},
		{"php", "php", "composer", "composer.json"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Discover() = %#v, want %#v", got, want)
	}
}

func TestExampleInstallPlan(t *testing.T) {
	root := exampleRoot(t)
	projects, err := Discover(root, 1)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, step := range steps {
		got = append(got, step.Manager+":"+step.String())
	}
	// The fixture has no lockfiles, so npm installs plainly and Composer runs
	// non-interactively. No command contacts the network on its own.
	want := []string{"npm:npm install", "composer:composer install --no-interaction"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("InstallPlan() = %v, want %v", got, want)
	}
}

func TestExamplePlanSelectionByOnly(t *testing.T) {
	root := exampleRoot(t)
	projects, err := Discover(root, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, selector := range []struct{ only, manager string }{
		{"node", "npm"},
		{"npm", "npm"},
		{"php", "composer"},
		{"composer", "composer"},
	} {
		steps, err := InstallPlan(root, projects, selector.only)
		if err != nil {
			t.Fatalf("--only %s: %v", selector.only, err)
		}
		if len(steps) != 1 || steps[0].Manager != selector.manager {
			t.Fatalf("--only %s = %#v, want manager %s", selector.only, steps, selector.manager)
		}
	}
	if _, err := InstallPlan(root, projects, "rust"); err == nil {
		t.Fatal("expected error for an unmatched --only selector")
	}
}

// The committed example must stay inspectable: planning it must not require
// running any native tool, and it must not carry generated dependency trees.
func TestExampleHasNoGeneratedDirectories(t *testing.T) {
	root := exampleRoot(t)
	for _, dir := range []string{
		filepath.Join("javascript", "node_modules"),
		filepath.Join("php", "vendor"),
		"node_modules",
		"vendor",
	} {
		if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
			t.Fatalf("example fixture contains a generated directory: %s", dir)
		}
	}
}
