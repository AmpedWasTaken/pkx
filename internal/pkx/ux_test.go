package pkx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// uxFixture is deliberately independent of the other tests' helpers.
func uxFixture(t *testing.T, root, rel, contents string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPlanShowsMixedLanguageCommands(t *testing.T) {
	root := t.TempDir()
	uxFixture(t, root, "frontend/package.json", `{}`)
	uxFixture(t, root, "frontend/package-lock.json", `{}`)
	uxFixture(t, root, "api/composer.json", `{}`)
	code, out, errOut := invoke("plan", "--dir", root)
	if code != 0 || errOut != "" || !strings.Contains(out, "npm ci") || !strings.Contains(out, "composer install") || !strings.Contains(out, "Nothing executed") {
		t.Fatalf("plan: exit=%d out=%q err=%q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(root, "frontend", "node_modules")); !os.IsNotExist(err) {
		t.Fatal("plan must never install dependencies")
	}
}

func TestPlanJSONUsesRelativeDirectoriesAndSelectedManager(t *testing.T) {
	root := t.TempDir()
	uxFixture(t, root, "web/package.json", `{}`)
	uxFixture(t, root, "api/composer.json", `{}`)
	code, out, errOut := invoke("plan", "--dir", root, "--only", "node", "--json")
	if code != 0 || errOut != "" {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	var doc struct {
		Root  string `json:"root"`
		Steps []struct {
			Directory string   `json:"directory"`
			Manager   string   `json:"manager"`
			Command   []string `json:"command"`
		} `json:"steps"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	canonicalRoot, err := rootPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Root != canonicalRoot || len(doc.Steps) != 1 || doc.Steps[0].Directory != "web" || doc.Steps[0].Manager != "npm" || strings.Join(doc.Steps[0].Command, " ") != "npm install" {
		t.Fatalf("bad JSON plan: %#v", doc)
	}
}

func TestStatusDoesNotModifyProject(t *testing.T) {
	root := t.TempDir()
	uxFixture(t, root, "api/composer.json", `{}`)
	uxFixture(t, root, "api/composer.lock", `{}`)
	code, out, errOut := invoke("status", "--dir", root, "--json")
	if code != 0 || errOut != "" {
		t.Fatalf("exit=%d out=%q err=%q", code, out, errOut)
	}
	var doc struct {
		Projects []struct {
			Directory string `json:"directory"`
			Lockfile  string `json:"lockfile"`
		} `json:"projects"`
		Diagnostics []map[string]string `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Projects) != 1 || doc.Projects[0].Lockfile != "composer.lock" || doc.Projects[0].Directory != "api" {
		t.Fatalf("bad status: %+v", doc)
	}
	if _, err := os.Stat(filepath.Join(root, "api", "vendor")); !os.IsNotExist(err) {
		t.Fatal("status must not install dependencies")
	}
}

func TestHelpAliasesAndTypoSuggestions(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"help", "plan"}, "pkx plan"},
		{[]string{"up", "--help"}, "pkx up"},
		{[]string{"bundle", "--help"}, "pkx bundle"},
		{[]string{"help", "i"}, "pkx install"},
		{[]string{"help", "ls"}, "pkx scan"},
	} {
		status, out, errOut := invoke(tc.args...)
		if status != 0 || errOut != "" || !strings.Contains(out, tc.want) {
			t.Fatalf("%v: code %d out=%q err=%q", tc.args, status, out, errOut)
		}
	}
	status, _, errOut := invoke("doctro")
	if status != 2 || !strings.Contains(errOut, "Did you mean 'pkx doctor'?") {
		t.Fatalf("typo hint missing: %d %q", status, errOut)
	}
}

func TestShortcutsAndEmptyWorkspaceHints(t *testing.T) {
	root := t.TempDir()
	if status, out, errOut := invoke("ls", "--dir", root); status != 0 || errOut != "" || !strings.Contains(out, "increasing --depth") {
		t.Fatalf("empty workspace: status=%d out=%q err=%q", status, out, errOut)
	}
	uxFixture(t, root, "package.json", `{}`)
	if status, out, _ := invoke("i", "--dir", root, "--dry-run"); status != 0 || !strings.Contains(out, "npm install") {
		t.Fatalf("install shortcut: status=%d out=%q", status, out)
	}
}

func TestInitStartsWithNoSurpriseDependencies(t *testing.T) {
	root := t.TempDir()
	status, out, errOut := invoke("init", "--dir", root)
	if status != 0 || errOut != "" || !strings.Contains(out, "Created") {
		t.Fatalf("init failed: %d %q %q", status, out, errOut)
	}
	var bundle Bundle
	if err := readJSON(filepath.Join(root, "pkx.json"), &bundle); err != nil {
		t.Fatal(err)
	}
	if len(bundle.Packages) != 0 {
		t.Fatalf("new bundle must not add arbitrary packages: %+v", bundle)
	}
}

func TestScanEmptyWorkspaceHasJSONArray(t *testing.T) {
	root := t.TempDir()
	status, out, errOut := invoke("scan", "--dir", root, "--json")
	if status != 0 || errOut != "" || strings.TrimSpace(out) != "[]" {
		t.Fatalf("empty scan: code=%d out=%q err=%q", status, out, errOut)
	}
}
