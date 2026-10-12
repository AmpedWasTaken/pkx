package pkx

import (
	"errors"
	"strings"
	"testing"
)

// okLookup reports every tool as present.
func okLookup(string) (string, error) { return "/usr/bin/tool", nil }

// stubProbe returns a fixed version string for every binary.
func stubProbe(output string) func(string, ...string) (string, error) {
	return func(string, ...string) (string, error) { return output, nil }
}

func TestVersionPinReadsProjectFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantPin string
	}{
		{"node-version", map[string]string{".node-version": "20.11.1\n"}, "20.11.1"},
		{"nvmrc", map[string]string{".nvmrc": "v18\n"}, "v18"},
		{"python-version", map[string]string{".python-version": "3.11.4\n"}, "3.11.4"},
		{"tool-versions", map[string]string{".tool-versions": "nodejs 20.11.1\npython 3.11.4\n"}, "20.11.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for name, content := range tt.files {
				put(t, root, name, content)
			}
			var got string
			switch tt.name {
			case "node-version", "nvmrc":
				got = readVersionPin(root, ".node-version", ".nvmrc")
			case "python-version":
				got = readVersionPin(root, ".python-version")
			case "tool-versions":
				got = toolVersion(root, "nodejs")
			}
			if got != tt.wantPin {
				t.Fatalf("read pin = %q, want %q", got, tt.wantPin)
			}
		})
	}
}

// A commented or blank version file must not be treated as a pin.
func TestVersionPinIgnoresCommentsAndBlankFiles(t *testing.T) {
	root := t.TempDir()
	put(t, root, ".node-version", "# managed by asdf\n\n")
	if got := readVersionPin(root, ".node-version"); got != "" {
		t.Fatalf("expected empty pin for comment-only file, got %q", got)
	}
}

func TestNodePinDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		pin        string
		actual     string
		wantIssue  bool
		wantSubstr string
	}{
		{"valid match", "22.20.0", "v22.20.0", false, ""},
		{"major match", "22", "v22.20.0", false, ""},
		{"mismatch", "20", "v22.20.0", true, "expected 20, found v22.20.0"},
		{"malformed pin does not guess", "not-a-version", "v22.20.0", false, ""},
		{"malformed output reports unknown", "22", "weird", false, ""},
		{"lts alias", "lts/*", "v20.11.1", false, ""},
		{"system alias", "system", "v20.11.1", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "package.json", `{}`)
			put(t, root, ".node-version", tt.pin+"\n")
			projects := []Project{{Directory: ".", Ecosystem: "node", Manager: "npm", Manifest: "package.json"}}
			steps := []Step{{Directory: root, Manager: "npm", Binary: "npm", Args: []string{"install"}}}
			issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe(tt.actual))
			got := len(issues) == 1
			if got != tt.wantIssue {
				t.Fatalf("pin %q vs %q: issues=%+v, want issue=%v", tt.pin, tt.actual, issues, tt.wantIssue)
			}
			if tt.wantIssue && !strings.Contains(issues[0].Problem, tt.wantSubstr) {
				t.Fatalf("problem = %q, want substring %q", issues[0].Problem, tt.wantSubstr)
			}
		})
	}
}

func TestPythonPinDiagnostics(t *testing.T) {
	root := t.TempDir()
	put(t, root, "requirements.txt", "httpx\n")
	put(t, root, ".python-version", "3.11.4\n")
	projects := []Project{{Directory: ".", Ecosystem: "python", Manager: "pip", Manifest: "requirements.txt"}}
	steps := []Step{{Directory: root, Manager: "pip", Binary: pythonCommand(), Args: []string{"-m", "pip", "install", "-r", "requirements.txt"}}}

	// Mismatch: the pinned patch must be reported against the actual runtime.
	issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("Python 3.12.5"))
	if len(issues) != 1 || !strings.Contains(issues[0].Problem, "expected 3.11.4") {
		t.Fatalf("mismatch diagnostics = %+v", issues)
	}
	if !strings.Contains(issues[0].Fix, "pinned version") {
		t.Fatalf("fix should be actionable, got %q", issues[0].Fix)
	}

	// Match: the exact pinned patch release is accepted.
	if issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("Python 3.11.4")); len(issues) != 0 {
		t.Fatalf("expected no issues for matching pin, got %+v", issues)
	}

	// A different patch of the same minor is still a mismatch: an exact pin is
	// compared part-by-part rather than accepting any compatible release.
	if issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("Python 3.11.9")); len(issues) != 1 {
		t.Fatalf("expected a mismatch for a different patch, got %+v", issues)
	}
}

// uv and poetry manage their own interpreters, so a Python pin is not checked.
func TestPythonPinSkippedForUvAndPoetry(t *testing.T) {
	root := t.TempDir()
	put(t, root, "pyproject.toml", "[project]\n")
	put(t, root, "uv.lock", "")
	put(t, root, ".python-version", "3.11.4\n")
	projects := []Project{{Directory: ".", Ecosystem: "python", Manager: "uv", Manifest: "pyproject.toml"}}
	steps := []Step{{Directory: root, Manager: "uv", Binary: "uv", Args: []string{"sync", "--locked"}}}
	if issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("Python 3.12.5")); len(issues) != 0 {
		t.Fatalf("uv should not be checked for python pin, got %+v", issues)
	}
}

func TestPhpConstraintDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		actual     string
		wantIssue  bool
	}{
		{"caret match", "^8.2", "8.4.2", false},
		{"caret cross-major", "^8.2", "9.0.0", true},
		{"gte match", ">=8.1", "8.3.0", false},
		{"gte fail", ">=8.4", "8.1.0", true},
		{"tilde cross-minor", "~8.2", "8.3.0", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "composer.json", `{"require":{"php":"`+tt.constraint+`"}}`)
			projects := []Project{{Directory: ".", Ecosystem: "php", Manager: "composer", Manifest: "composer.json"}}
			steps := []Step{{Directory: root, Manager: "composer", Binary: "composer", Args: []string{"install", "--no-interaction"}}}
			issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe(tt.actual))
			if got := len(issues) == 1; got != tt.wantIssue {
				t.Fatalf("%s vs %s: issues=%+v, want issue=%v", tt.constraint, tt.actual, issues, tt.wantIssue)
			}
		})
	}
}

// An unsupported composer constraint must be documented, not silently guessed:
// supportsSimpleConstraint gates the check, so complex constraints are skipped.
func TestUnsupportedPhpConstraintIsSkipped(t *testing.T) {
	for _, constraint := range []string{"^8.1 || ^9.0", ">=8.1 <8.5", "8.*.*", "*"} {
		if supportsSimpleConstraint(constraint) {
			t.Errorf("supportsSimpleConstraint(%q) = true, want false", constraint)
		}
	}
	for _, constraint := range []string{"8.2", ">=8.1", "^8.2", "~8.2", "8.2.*"} {
		if !supportsSimpleConstraint(constraint) {
			t.Errorf("supportsSimpleConstraint(%q) = false, want true", constraint)
		}
	}
	root := t.TempDir()
	put(t, root, "composer.json", `{"require":{"php":"^8.1 || ^9.0"}}`)
	projects := []Project{{Directory: ".", Ecosystem: "php", Manager: "composer", Manifest: "composer.json"}}
	steps := []Step{{Directory: root, Manager: "composer", Binary: "composer", Args: []string{"install"}}}
	if issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("7.4.0")); len(issues) != 0 {
		t.Fatalf("complex constraint should be skipped, got %+v", issues)
	}
}

func TestGoModVersionDiagnostics(t *testing.T) {
	root := t.TempDir()
	put(t, root, "go.mod", "module example\n\ngo 1.23\n")
	projects := []Project{{Directory: ".", Ecosystem: "go", Manager: "go", Manifest: "go.mod"}}

	steps := []Step{{Directory: root, Manager: "go", Binary: "go", Args: []string{"mod", "download"}}}
	if issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("go version go1.23.2 linux/amd64")); len(issues) != 0 {
		t.Fatalf("expected matching go version, got %+v", issues)
	}

	issues := CheckEnvironment(root, projects, steps, okLookup, stubProbe("go version go1.21.0 linux/amd64"))
	if len(issues) != 1 || !strings.Contains(issues[0].Problem, "requires >=1.23") {
		t.Fatalf("expected go version mismatch, got %+v", issues)
	}
}

// When a version probe fails, the diagnostic must say so and point at a manual check.
func TestVersionProbeFailureIsActionable(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{}`)
	put(t, root, ".node-version", "22\n")
	projects := []Project{{Directory: ".", Ecosystem: "node", Manager: "npm", Manifest: "package.json"}}
	steps := []Step{{Directory: root, Manager: "npm", Binary: "npm", Args: []string{"install"}}}
	probe := func(string, ...string) (string, error) { return "", errors.New("exit status 1") }
	issues := CheckEnvironment(root, projects, steps, okLookup, probe)
	if len(issues) != 1 || !strings.Contains(issues[0].Problem, "could not determine version") {
		t.Fatalf("expected probe failure diagnostic, got %+v", issues)
	}
	if !strings.Contains(issues[0].Fix, "node --version") {
		t.Fatalf("fix should name the manual command, got %q", issues[0].Fix)
	}
}

// A missing runtime is reported once with an install hint, before any version
// comparison is attempted.
func TestMissingRuntimeReportedOnce(t *testing.T) {
	root := t.TempDir()
	put(t, root, "composer.json", `{"require":{"php":">=8.1"}}`)
	projects := []Project{{Directory: ".", Ecosystem: "php", Manager: "composer", Manifest: "composer.json"}}
	steps := []Step{{Directory: root, Manager: "composer", Binary: "composer", Args: []string{"install"}}}
	lookup := func(bin string) (string, error) {
		if bin == "php" || bin == "composer" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + bin, nil
	}
	issues := CheckEnvironment(root, projects, steps, lookup, stubProbe("8.4.0"))
	if len(issues) != 2 {
		t.Fatalf("expected php and composer missing, got %+v", issues)
	}
	for _, issue := range issues {
		if !strings.Contains(issue.Fix, "getcomposer.org") {
			t.Fatalf("missing-tool fix should be actionable, got %q", issue.Fix)
		}
	}
}
