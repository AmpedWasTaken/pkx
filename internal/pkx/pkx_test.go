package pkx

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func put(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func invoke(args ...string) (int, string, string) {
	var out, err bytes.Buffer
	code := Run(context.Background(), args, &out, &err)
	return code, out.String(), err.String()
}

func TestDiscoverMonorepo(t *testing.T) {
	root := t.TempDir()
	put(t, root, "frontend/package.json", `{"packageManager":"pnpm@9.0.0"}`)
	put(t, root, "frontend/pnpm-lock.yaml", "")
	put(t, root, "backend/composer.json", "{}")
	put(t, root, "scripts/requirements.txt", "requests")
	put(t, root, "node_modules/ignored/package.json", "{}")
	put(t, root, "nested/a/b/c/d/e/package.json", "{}")
	got, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []Project{
		{"backend", "php", "composer", "composer.json"},
		{"frontend", "node", "pnpm", "package.json"},
		{"scripts", "python", "pip", "requirements.txt"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestNodeLockfileConflict(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", "{}")
	put(t, root, "pnpm-lock.yaml", "")
	put(t, root, "yarn.lock", "")
	_, err := Discover(root, 0)
	if err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("expected conflict, got %v", err)
	}
	put(t, root, "package.json", `{"packageManager":"pnpm@9.0.0"}`)
	got, err := Discover(root, 0)
	if err != nil || len(got) != 1 || got[0].Manager != "pnpm" {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLockfilePlans(t *testing.T) {
	root := t.TempDir()
	put(t, root, "node/package.json", "{}")
	put(t, root, "node/package-lock.json", "{}")
	put(t, root, "api/go.mod", "module test")
	put(t, root, "rust/Cargo.toml", "[package]")
	projects, err := Discover(root, 2)
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
	want := []string{"go:go mod download", "npm:npm ci", "cargo:cargo fetch"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestPipUsesLocalVenv(t *testing.T) {
	root := t.TempDir()
	put(t, root, "requirements.txt", "httpx")
	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "")
	if err != nil || len(steps) != 2 || steps[0].Args[1] != "venv" {
		t.Fatalf("bad plan %v %v", steps, err)
	}
	if err := os.Mkdir(filepath.Join(root, ".venv"), 0755); err != nil {
		t.Fatal(err)
	}
	steps, err = InstallPlan(root, projects, "")
	if err != nil || len(steps) != 1 {
		t.Fatalf("expected existing virtualenv, got %v %v", steps, err)
	}
}

func TestBundlePlanAndValidation(t *testing.T) {
	root := t.TempDir()
	put(t, root, "front/package.json", "{}")
	put(t, root, "api/composer.json", "{}")
	path := filepath.Join(root, "pkx.json")
	put(t, root, "pkx.json", `{"schema":1,"name":"test","packages":[{"manager":"pnpm","directory":"front","dependencies":["vue"],"devDependencies":["vite"]},{"manager":"composer","directory":"api","dependencies":["laravel/sanctum"]}]}`)
	steps, err := BundlePlan(root, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 3 || steps[0].String() != "pnpm add vue" || steps[1].String() != "pnpm add --dev vite" {
		t.Fatalf("unexpected bundle plan %#v", steps)
	}
	put(t, root, "pkx.json", `{"schema":1,"name":"test","packages":[{"manager":"npm","directory":"../escape","dependencies":["react"]}]}`)
	if _, err = BundlePlan(root, path); err == nil {
		t.Fatal("accepted directory escape")
	}
	put(t, root, "pkx.json", `{"schema":1,"name":"test","packages":[{"manager":"npm","directory":"front","dependencies":["--ignore-scripts"]}]}`)
	if _, err = BundlePlan(root, path); err == nil {
		t.Fatal("accepted flag-shaped dependency")
	}
}

func TestSymlinksNotFollowed(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	put(t, outside, "package.json", "{}")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got, err := Discover(root, 3)
	if err != nil || len(got) != 0 {
		t.Fatalf("unexpected symlink discovery %v %v", got, err)
	}
	put(t, root, "pkx.json", `{"schema":1,"name":"test","packages":[{"manager":"npm","directory":"link","dependencies":["react"]}]}`)
	if _, err := BundlePlan(root, filepath.Join(root, "pkx.json")); err == nil {
		t.Fatal("accepted symlink escape")
	}
}

func TestScanAndDryRun(t *testing.T) {
	root := t.TempDir()
	put(t, root, "web/package.json", "{}")
	code, out, stderr := invoke("scan", "--dir", root, "--json")
	if code != 0 || !strings.Contains(out, `"manager":"npm"`) {
		t.Fatalf("code=%d out=%s err=%s", code, out, stderr)
	}
	code, out, stderr = invoke("install", "--dir", root, "--dry-run")
	if code != 0 || !strings.Contains(out, "npm install") || !strings.Contains(out, "nothing executed") {
		t.Fatalf("code=%d out=%s err=%s", code, out, stderr)
	}
}

func TestFailures(t *testing.T) {
	for _, args := range [][]string{{"unknown"}, {"scan", "--dry-run"}, {"install", "--depth", "99"}, {"bundle", "install"}} {
		if code, _, _ := invoke(args...); code == 0 {
			t.Errorf("expected failure for %v", args)
		}
	}
}

func TestFailedNativeCommandStops(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	root := t.TempDir()
	put(t, root, "package.json", "{}")
	dir := t.TempDir()
	path := filepath.Join(dir, "npm")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 9\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	code, _, stderr := invoke("install", "--dir", root)
	if code == 0 || !strings.Contains(stderr, "status 9") {
		t.Fatalf("expected exit 9: %d %s", code, stderr)
	}
}
