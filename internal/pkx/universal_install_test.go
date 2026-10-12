package pkx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallWithoutPackageKeepsBootstrap(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", "{}")
	code, out, stderr := invoke("install", "--dir", root, "--dry-run")
	if code != 0 || stderr != "" || !strings.Contains(out, "npm install") || !strings.Contains(out, "nothing executed") {
		t.Fatalf("legacy install changed: %d %q %q", code, out, stderr)
	}
}

func TestPrefixFreeInstallPicksProjectManager(t *testing.T) {
	cases := []struct {
		name     string
		files    map[string]string
		pkg      string
		expected string
	}{
		{"npm", map[string]string{"package.json": "{}"}, "http-server", "npm install http-server"},
		{"pnpm", map[string]string{"package.json": "{\"packageManager\":\"pnpm@9.0.0\"}", "pnpm-lock.yaml": ""}, "http-server", "pnpm add http-server"},
		{"yarn", map[string]string{"package.json": "{\"packageManager\":\"yarn@4.0.0\"}", "yarn.lock": ""}, "http-server", "yarn add http-server"},
		{"bun", map[string]string{"package.json": "{\"packageManager\":\"bun@1.2.0\"}", "bun.lock": ""}, "http-server", "bun add http-server"},
		{"composer", map[string]string{"composer.json": "{}"}, "monolog/monolog", "composer require --no-interaction monolog/monolog"},
		{"pip", map[string]string{"requirements.txt": "requests\n"}, "httpx", "-m pip install httpx"},
		{"uv", map[string]string{"pyproject.toml": "[project]\nname='example'\n", "uv.lock": ""}, "httpx", "uv add httpx"},
		{"poetry", map[string]string{"pyproject.toml": "[tool.poetry]\nname='example'\n", "poetry.lock": ""}, "httpx", "poetry add httpx"},
		{"cargo", map[string]string{"Cargo.toml": "[package]\nname='example'\n"}, "serde", "cargo add serde"},
		{"go", map[string]string{"go.mod": "module example.com/test\n"}, "github.com/google/uuid", "go get github.com/google/uuid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for file, content := range tc.files {
				put(t, root, file, content)
			}
			code, out, stderr := invoke("install", tc.pkg, "--dir", root, "--dry-run")
			if code != 0 || stderr != "" || !strings.Contains(out, tc.expected) || !strings.Contains(out, "nothing executed") {
				t.Fatalf("manager selection failed: %d %q %q", code, out, stderr)
			}
		})
	}
}

func TestPrefixFreeInstallRejectsAmbiguousProjects(t *testing.T) {
	root := t.TempDir()
	put(t, root, "frontend/package.json", "{}")
	put(t, root, "backend/requirements.txt", "httpx\n")
	code, _, stderr := invoke("install", "http-server", "--dir", root, "--dry-run")
	if code == 0 || !strings.Contains(stderr, "--project") {
		t.Fatalf("ambiguous project was accepted: %d %q", code, stderr)
	}
	code, out, stderr := invoke("install", "http-server", "--project", "frontend", "--dir", root, "--dry-run")
	if code != 0 || stderr != "" || !strings.Contains(out, "npm install http-server") {
		t.Fatalf("project choice failed: %d %q %q", code, out, stderr)
	}
	code, _, stderr = invoke("install", "http-server", "--project", "../escape", "--dir", root, "--dry-run")
	if code == 0 || !strings.Contains(stderr, "outside") {
		t.Fatalf("directory escape accepted: %d %q", code, stderr)
	}
}

func TestPrefixFreeInstallRejectsInvalidPackages(t *testing.T) {
	cases := []string{"npm:axios", "--ignore-scripts", "https://example.com/archive.tgz", "../outside", "npm:../outside", "http server"}
	root := t.TempDir()
	put(t, root, "package.json", "{}")
	for _, pkg := range cases {
		code, _, _ := invoke("install", pkg, "--dir", root, "--dry-run")
		if code == 0 {
			t.Errorf("invalid dependency accepted: %q", pkg)
		}
	}
}

func TestPrefixFreeInstallPythonVenvSafety(t *testing.T) {
	root := t.TempDir()
	put(t, root, "requirements.txt", "httpx\n")
	code, out, stderr := invoke("install", "rich", "--dir", root, "--dry-run")
	if code != 0 || !strings.Contains(out, "-m venv .venv") || !strings.Contains(out, "-m pip install rich") {
		t.Fatalf("pip plan: %d %q %q", code, out, stderr)
	}
	if _, err := os.Lstat(filepath.Join(root, ".venv")); !os.IsNotExist(err) {
		t.Fatalf("dry run unexpectedly created venv: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".venv"), 0755); err != nil {
		t.Fatal(err)
	}
	code, out, stderr = invoke("install", "rich", "--dir", root, "--dry-run")
	if code != 0 || strings.Contains(out, "-m venv .venv") || !strings.Contains(out, "-m pip install rich") {
		t.Fatalf("existing venv: %d %q %q", code, out, stderr)
	}
}

func TestPrefixFreeInstallDevFlags(t *testing.T) {
	root := t.TempDir()
	put(t, root, "Cargo.toml", "[package]\nname='example'\n")
	code, out, stderr := invoke("install", "serde", "--dir", root, "--dev", "--dry-run")
	if code != 0 || !strings.Contains(out, "cargo add --dev serde") {
		t.Fatalf("cargo dev: %d %q %q", code, out, stderr)
	}
}

func TestInstallRouteDetection(t *testing.T) {
	for _, c := range []struct {
		args []string
		pkg  bool
	}{
		{[]string{}, false},
		{[]string{"--dir", "some-project", "--dry-run"}, false},
		{[]string{"--only", "npm"}, false},
		{[]string{"--depth=3"}, false},
		{[]string{"http-server"}, true},
		{[]string{"--dir=.", "http-server"}, true},
		{[]string{"http-server", "--project", "frontend"}, true},
	} {
		if got := packageInstallRequested(c.args); got != c.pkg {
			t.Errorf("args %q: got package route %v want %v", c.args, got, c.pkg)
		}
	}
}
