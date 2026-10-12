package pkx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletionScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell", "pwsh"} {
		t.Run(shell, func(t *testing.T) {
			status, output, errOut := invoke("completion", shell)
			if status != 0 || errOut != "" || !strings.Contains(output, "pkx") || !strings.Contains(output, "doctor") || !strings.HasSuffix(output, "\n") {
				t.Fatalf("completion for %s: exit=%d output=%q err=%q", shell, status, output, errOut)
			}
		})
	}
	for _, args := range [][]string{{"completion"}, {"completion", "invalid"}, {"completion", "bash", "extra"}} {
		if status, _, _ := invoke(args...); status != 2 {
			t.Errorf("expected usage error for %v, got %d", args, status)
		}
	}
}

func TestDoctorJSONReportsMissingTools(t *testing.T) {
	root := t.TempDir()
	uxFixture(t, root, "api/composer.json", `{}`)
	// Use an empty PATH so behavior never depends on a CI runner's PHP install.
	t.Setenv("PATH", t.TempDir())
	status, output, errOut := invoke("doctor", "--dir", root, "--json")
	if status != 1 || errOut != "" {
		t.Fatalf("doctor status=%d stdout=%q stderr=%q", status, output, errOut)
	}
	var doc doctorReport
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Ready || len(doc.Projects) != 1 || len(doc.Diagnostics) == 0 {
		t.Fatalf("bad diagnostic report: %+v", doc)
	}
	if filepath.Base(doc.Root) != filepath.Base(root) {
		t.Fatal("wrong workspace root")
	}
}

func TestDoctorJSONOnlyFiltersProjects(t *testing.T) {
	root := t.TempDir()
	uxFixture(t, root, "api/composer.json", `{}`)
	uxFixture(t, root, "client/package.json", `{}`)
	t.Setenv("PATH", t.TempDir())
	status, output, errOut := invoke("doctor", "--dir", root, "--only", "php", "--json")
	if status != 1 || errOut != "" {
		t.Fatalf("doctor exit=%d out=%q err=%q", status, output, errOut)
	}
	var doc doctorReport
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Projects) != 1 || doc.Projects[0].Manager != "composer" {
		t.Fatalf("--only did not narrow results: %+v", doc.Projects)
	}
	for _, issue := range doc.Diagnostics {
		if issue.Tool == "node" || issue.Tool == "npm" {
			t.Fatalf("unexpected Node diagnostic in php-only check: %+v", issue)
		}
	}
}

func TestDoctorJSONEmptyWorkspace(t *testing.T) {
	root := t.TempDir()
	status, output, errOut := invoke("doctor", "--dir", root, "--json")
	if status != 0 || errOut != "" {
		t.Fatalf("empty doctor: %d %q %q", status, output, errOut)
	}
	var doc doctorReport
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatal(err)
	}
	if !doc.Ready || len(doc.Projects) != 0 || len(doc.Diagnostics) != 0 {
		t.Fatalf("unexpected data: %+v", doc)
	}
}

func TestCompletionNoFilesystemChanges(t *testing.T) {
	root := t.TempDir()
	before, _ := os.ReadDir(root)
	if status, _, _ := invoke("completion", "bash"); status != 0 {
		t.Fatal("failed to generate bash completion")
	}
	after, _ := os.ReadDir(root)
	if len(before) != len(after) {
		t.Fatal("completion generator unexpectedly changed filesystem")
	}
}
