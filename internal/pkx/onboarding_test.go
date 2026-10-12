package pkx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpDryRunShowsPlan(t *testing.T) {
	root := t.TempDir()
	put(t, root, "backend/composer.json", `{"require":{"php":"^8.2"}}`)
	put(t, root, "frontend/package.json", `{}`)
	code, out, stderr := invoke("up", "--dir", root, "--dry-run")
	if code != 0 || !strings.Contains(out, "composer install") || !strings.Contains(out, "npm install") || stderr != "" {
		t.Fatalf("code %d out %q stderr %q", code, out, stderr)
	}
}

func TestWorkspaceCollapseOnlyOwnedDirectories(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"pnpm@9.0.0"}`)
	put(t, root, "pnpm-workspace.yaml", "packages:\n  - 'packages/*'\n  - '!packages/excluded'\n")
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: 9.0\n")
	put(t, root, "packages/core/package.json", `{}`)
	put(t, root, "packages/excluded/package.json", `{}`)
	put(t, root, "apps/api/package.json", `{}`)
	projects, err := Discover(root, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 3 {
		t.Fatalf("want 3 root/excluded/api projects, got %+v", projects)
	}
	for _, p := range projects {
		if p.Directory == "packages/core" {
			t.Fatalf("workspace dependency repeated: %+v", projects)
		}
	}
}

func TestWorkspaceRetainsIndependentLockfiles(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"workspaces":["packages/*"]}`)
	put(t, root, "packages/a/package.json", `{}`)
	put(t, root, "packages/b/package.json", `{}`)
	put(t, root, "packages/b/package-lock.json", `{}`)
	projects, err := Discover(root, 4)
	if err != nil || len(projects) != 2 {
		t.Fatalf("got %+v %v", projects, err)
	}
}

func TestVersionPins(t *testing.T) {
	tests := []struct {
		pin, actual string
		want        bool
	}{
		{"22", "v22.20.0", true}, {"22.2", "v22.20.0", false},
		{"22.20.0", "v22.20.0", true}, {"20", "v22.20.0", false},
		{"lts/*", "v20.0.0", true},
	}
	for _, tt := range tests {
		if got := matchesPinnedVersion(tt.pin, tt.actual); got != tt.want {
			t.Errorf("%s against %s = %v", tt.pin, tt.actual, got)
		}
	}
	if !matchesConstraint("^8.2", "8.4.2") || matchesConstraint("^8.2", "9.0.0") {
		t.Fatal("unexpected composer constraint check")
	}
	if !matchesConstraint(">=1.23", "go version go1.23.2 linux/amd64") {
		t.Fatal("go version check failed")
	}
}

func TestEnvironmentMissingToolsAndVersionMismatch(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	put(t, root, ".nvmrc", "20\n")
	projects := []Project{{Directory: ".", Ecosystem: "node", Manager: "npm", Manifest: "package.json"}}
	steps := []Step{{Directory: root, Manager: "npm", Binary: "npm", Args: []string{"install"}}}
	diagnostics := CheckEnvironment(root, projects, steps, func(bin string) (string, error) {
		if bin == "npm" {
			return "", os.ErrNotExist
		}
		return "/usr/bin/" + bin, nil
	}, func(bin string, args ...string) (string, error) { return "v22.20.0", nil })
	if len(diagnostics) != 2 {
		t.Fatalf("expected missing npm and runtime mismatch, got %+v", diagnostics)
	}
	if !strings.Contains(diagnostics[1].Problem, "expected 20") {
		t.Fatalf("unexpected %+v", diagnostics)
	}
}

func TestEnvironmentChecksAreNonDestructive(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	projects := []Project{{Directory: ".", Ecosystem: "node", Manager: "npm", Manifest: "package.json"}}
	steps := []Step{{Directory: root, Manager: "npm", Binary: "npm"}}
	problems := CheckEnvironment(root, projects, steps, func(s string) (string, error) { return filepath.Join(root, s), nil }, func(s string, args ...string) (string, error) { return "v22.0.0", nil })
	if len(problems) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", problems)
	}
	if _, err := os.Stat(filepath.Join(root, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("doctor should not install dependencies")
	}
}
