package pkx

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// A Bun workspace with both lockfile markers picks Bun and plans one frozen
// install. bun.lock is plain JSON text; bun.lockb is binary, so a small marker
// that is not valid JSON stands in for a real downloaded lockfile.
func TestBunLockfilesPlanFrozenInstall(t *testing.T) {
	markers := map[string]string{
		"bun.lock":  "{\"lockfileVersion\":1,\"workspaces\":{\"\":{}}}\n",
		"bun.lockb": "\x00\x01BUNBINARYMARKER\xff",
	}
	for marker, content := range markers {
		t.Run(marker, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "package.json", `{"name":"app"}`)
			put(t, root, marker, content)

			projects, err := Discover(root, 0)
			if err != nil {
				t.Fatal(err)
			}
			want := []Project{{".", "node", "bun", "package.json"}}
			if !reflect.DeepEqual(projects, want) {
				t.Fatalf("discover: got %#v want %#v", projects, want)
			}

			steps, err := InstallPlan(root, projects, "")
			if err != nil {
				t.Fatal(err)
			}
			wantSteps := []Step{{root, "bun", "bun", []string{"install", "--frozen-lockfile"}}}
			if !reflect.DeepEqual(steps, wantSteps) {
				t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
			}
		})
	}
}

// An explicit packageManager pin selects Bun even without a lockfile, and the
// plan then omits --frozen-lockfile because there is nothing to freeze.
func TestBunExplicitPinWithoutLockfile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"bun@1.2.0"}`)

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "bun" {
		t.Fatalf("discover: got %#v want one bun project", projects)
	}
	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []Step{{root, "bun", "bun", []string{"install"}}}
	if !reflect.DeepEqual(steps, wantSteps) {
		t.Fatalf("plan: got %#v want %#v", steps, wantSteps)
	}
}

// A package-lock.json next to bun.lock with no pin is ambiguous: the existing
// lockfile conflict diagnostic must win instead of silently choosing Bun.
func TestBunAndNpmLockfileConflictNoPin(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"name":"app"}`)
	put(t, root, "bun.lock", "")
	put(t, root, "package-lock.json", `{"lockfileVersion":3}`)

	got, err := Discover(root, 0)
	if err == nil {
		t.Fatalf("expected conflict, got projects %#v", got)
	}
	if msg := err.Error(); !strings.Contains(msg, "conflicting Node lockfiles") || !strings.Contains(msg, "packageManager") {
		t.Fatalf("diagnostic does not name the conflict and the fix: %v", err)
	}

	// An explicit pin resolves the ambiguity in Bun's favour.
	put(t, root, "package.json", `{"packageManager":"bun@1.2.0"}`)
	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	steps, err := InstallPlan(root, projects, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].String() != "bun install --frozen-lockfile" || steps[0].Directory != root {
		t.Fatalf("plan: got %#v want a pinned frozen bun install in %s", steps, root)
	}
}

// A manifest packageManager pin that names a different manager must not be
// overridden by a stray Bun lockfile, and an unsupported pin is still rejected.
func TestBunLockfileDoesNotOverrideExplicitManager(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"packageManager":"pnpm@9.12.0"}`)
	put(t, root, "bun.lockb", "")
	put(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")

	projects, err := Discover(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Manager != "pnpm" {
		t.Fatalf("discover: got %#v want pnpm to win", projects)
	}

	put(t, root, "package.json", `{"packageManager":"deno@2.0.0"}`)
	if _, err := Discover(root, 0); err == nil {
		t.Fatal("expected an unsupported packageManager diagnostic for deno")
	}
}

// --only must accept both the Bun manager name and the node ecosystem, and a
// filter that matches nothing must fail rather than plan something else.
func TestBunOnlyFilter(t *testing.T) {
	root := t.TempDir()
	put(t, root, "app/package.json", `{"packageManager":"bun@1.2.0"}`)
	put(t, root, "app/bun.lockb", "")
	put(t, root, "other/go.mod", "module example.com/other")

	projects, err := Discover(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, only := range []string{"bun", "node"} {
		steps, err := InstallPlan(root, projects, only)
		if err != nil {
			t.Fatalf("--only %s: %v", only, err)
		}
		if len(steps) != 1 || steps[0].Directory != filepath.Join(root, "app") {
			t.Fatalf("--only %s: got %#v want only the bun project", only, steps)
		}
	}
	if steps, err := InstallPlan(root, projects, "pnpm"); err == nil {
		t.Fatalf("--only pnpm must not plan bun projects: %#v", steps)
	}
}
