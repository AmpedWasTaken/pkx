package pkx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManageAddAndRemoveDryRun(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"name":"example","scripts":{"dev":"vite"}}`)

	code, out, errOut := invoke("add", "axios", "zod", "--dev", "--dir", root, "--dry-run")
	if code != 0 || errOut != "" || !strings.Contains(out, "npm install --save-dev axios zod") || !strings.Contains(out, "nothing executed") {
		t.Fatalf("add: %d %q %q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(root, "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("dry run changed dependencies: %v", err)
	}
	code, out, errOut = invoke("remove", "axios", "--dir", root, "--dry-run")
	if code != 0 || errOut != "" || !strings.Contains(out, "npm uninstall axios") {
		t.Fatalf("remove: %d %q %q", code, out, errOut)
	}
}

func TestManageComposerCommands(t *testing.T) {
	root := t.TempDir()
	put(t, root, "composer.json", `{"scripts":{"test":"phpunit"}}`)
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"add", "laravel/sanctum", "--dev"}, "composer require --no-interaction --dev laravel/sanctum"},
		{[]string{"remove", "laravel/sanctum"}, "composer remove --no-interaction laravel/sanctum"},
		{[]string{"run", "test"}, "composer run-script test"},
	}
	for _, tc := range cases {
		args := append(append([]string{}, tc.args...), "--dir", root, "--dry-run")
		code, out, errOut := invoke(args...)
		if code != 0 || errOut != "" || !strings.Contains(out, tc.want) {
			t.Fatalf("%v: status=%d out=%q err=%q", args, code, out, errOut)
		}
	}
}

func TestManageAmbiguousProjectsRequireSelection(t *testing.T) {
	root := t.TempDir()
	put(t, root, "frontend/package.json", `{"scripts":{"dev":"vite"}}`)
	put(t, root, "backend/composer.json", `{"scripts":{"test":"phpunit"}}`)
	code, _, errOut := invoke("add", "axios", "--dir", root, "--dry-run")
	if code == 0 || !strings.Contains(errOut, "--project") {
		t.Fatalf("ambiguous add: status=%d stderr=%q", code, errOut)
	}
	code, out, errOut := invoke("add", "axios", "--project", "frontend", "--dir", root, "--dry-run")
	if code != 0 || errOut != "" || !strings.Contains(out, "npm install axios") {
		t.Fatalf("selected add: status=%d out=%q err=%q", code, out, errOut)
	}
	code, out, errOut = invoke("run", "dev", "--dir", root, "--dry-run")
	if code != 0 || errOut != "" || !strings.Contains(out, "npm run dev") {
		t.Fatalf("unique run: status=%d out=%q err=%q", code, out, errOut)
	}
}

func TestManageRunPassesArgumentsWithoutShell(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"vite","build":"vite build"}}`)
	code, out, errOut := invoke("run", "dev", "--dir", root, "--dry-run", "--", "--host", "127.0.0.1")
	if code != 0 || errOut != "" || !strings.Contains(out, "npm run dev -- --host 127.0.0.1") {
		t.Fatalf("run: status=%d out=%q err=%q", code, out, errOut)
	}
	code, _, errOut = invoke("run", "missing", "--dir", root, "--dry-run")
	if code == 0 || !strings.Contains(errOut, "available: build (.), dev (.)") {
		t.Fatalf("missing script: %d %q", code, errOut)
	}
}

func TestManageValidatesInputs(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	for _, args := range [][]string{
		{"add", "--ignore-scripts", "--dir", root},
		{"add", "", "--dir", root},
		{"add", "axios", "--project", "../escape", "--dir", root, "--dry-run"},
		{"remove", "axios", "--dev", "--dir", root},
		{"run", "dev", "--manager", "composer", "--dir", root, "--dry-run"},
		{"run", "dev", "extra", "--dir", root},
		{"add", "axios", "--", "--ignore-scripts", "--dir", root},
	} {
		code, _, _ := invoke(args...)
		if code == 0 {
			t.Errorf("expected failure for %q", args)
		}
	}
}

func TestManageMultipleMatchingScriptsRequireProject(t *testing.T) {
	root := t.TempDir()
	put(t, root, "one/package.json", `{"scripts":{"dev":"vite"}}`)
	put(t, root, "two/package.json", `{"scripts":{"dev":"vite"}}`)
	code, _, errOut := invoke("run", "dev", "--dir", root, "--dry-run")
	if code == 0 || !strings.Contains(errOut, "multiple projects match") {
		t.Fatalf("expected ambiguity, got %d %q", code, errOut)
	}
}
