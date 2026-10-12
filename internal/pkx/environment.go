package pkx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Diagnostic struct {
	Directory string `json:"directory"`
	Tool      string `json:"tool"`
	Problem   string `json:"problem"`
	Fix       string `json:"fix"`
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s: %s — %s (fix: %s)", d.Directory, d.Tool, d.Problem, d.Fix)
}

var versionNumbers = regexp.MustCompile(`\d+(?:\.\d+){0,2}`)

// ProbeVersion has a timeout so a broken manager cannot hang `pkx doctor`.
func ProbeVersion(binary string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CheckEnvironment reports actionable diagnostics without modifying the machine.
// `probe` is injected to make version checks reproducible in tests.
func CheckEnvironment(root string, projects []Project, steps []Step, lookup func(string) (string, error), probe func(string, ...string) (string, error)) []Diagnostic {
	var issues []Diagnostic
	missing := map[string]bool{}
	for _, step := range steps {
		bin := step.Binary
		if strings.ContainsAny(bin, `/\`) {
			continue
		} // project-local venv interpreter
		if missing[bin] {
			continue
		}
		if _, err := lookup(bin); err != nil {
			missing[bin] = true
			issues = append(issues, Diagnostic{".", bin, "command not found in PATH", installHint(bin)})
		}
	}
	checked := map[string]bool{}
	for _, project := range projects {
		dir := filepath.Join(root, filepath.FromSlash(project.Directory))
		switch project.Ecosystem {
		case "node":
			if _, err := lookup("node"); err != nil {
				if !checked["missing:node"] {
					issues = append(issues, Diagnostic{project.Directory, "node", "Node.js runtime missing", installHint("node")})
					checked["missing:node"] = true
				}
				continue
			}
			desired := readVersionPin(dir, ".node-version", ".nvmrc")
			if desired == "" {
				desired = readVersionPin(root, ".node-version", ".nvmrc")
			}
			if desired == "" {
				desired = toolVersion(dir, "nodejs")
			}
			if desired == "" {
				desired = toolVersion(root, "nodejs")
			}
			if desired != "" {
				issues = appendVersionIssue(issues, project.Directory, "node", desired, "node", []string{"--version"}, probe)
			}
			var manifest struct {
				PackageManager string `json:"packageManager"`
			}
			if readJSON(filepath.Join(dir, "package.json"), &manifest) == nil {
				parts := strings.SplitN(manifest.PackageManager, "@", 2)
				if len(parts) == 2 && parts[0] == project.Manager && !missing[parts[0]] {
					issues = appendVersionIssue(issues, project.Directory, parts[0], parts[1], parts[0], []string{"--version"}, probe)
				}
			}
		case "python":
			if project.Manager != "pip" {
				continue
			} // uv/poetry manage their own interpreters
			bin := pythonCommand()
			desired := readVersionPin(dir, ".python-version")
			if desired == "" {
				desired = readVersionPin(root, ".python-version")
			}
			if desired == "" {
				desired = toolVersion(dir, "python")
			}
			if desired == "" {
				desired = toolVersion(root, "python")
			}
			if desired != "" && !missing[bin] {
				issues = appendVersionIssue(issues, project.Directory, "python", desired, bin, []string{"--version"}, probe)
			}
		case "php":
			if _, err := lookup("php"); err != nil {
				if !checked["missing:php"] {
					issues = append(issues, Diagnostic{project.Directory, "php", "PHP runtime missing", installHint("php")})
					checked["missing:php"] = true
				}
				continue
			}
			var manifest struct {
				Require map[string]string `json:"require"`
			}
			if readJSON(filepath.Join(dir, "composer.json"), &manifest) == nil && manifest.Require["php"] != "" {
				constraint := manifest.Require["php"]
				if supportsSimpleConstraint(constraint) {
					issues = appendConstraintIssue(issues, project.Directory, "php", constraint, "php", []string{"-r", "echo PHP_VERSION;"}, probe)
				}
			}
		case "go":
			if missing["go"] {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[0] == "go" {
					issues = appendConstraintIssue(issues, project.Directory, "go", ">="+fields[1], "go", []string{"version"}, probe)
					break
				}
			}
		}
	}
	return issues
}

func installHint(bin string) string {
	switch bin {
	case "node", "npm", "npx":
		return "install Node.js from https://nodejs.org"
	case "pnpm", "yarn":
		return "enable Corepack or install " + bin + " using your Node.js installation"
	case "bun":
		return "install Bun from https://bun.sh"
	case "composer", "php":
		return "install PHP and Composer from https://getcomposer.org"
	case "python", "python3":
		return "install Python from https://python.org and add it to PATH"
	case "uv":
		return "install uv from https://docs.astral.sh/uv/"
	case "poetry":
		return "install Poetry from https://python-poetry.org"
	case "cargo", "rustc":
		return "install Rust using https://rustup.rs"
	case "go":
		return "install Go from https://go.dev/dl/"
	default:
		return "install " + bin + " and add it to PATH"
	}
}

func readVersionPin(dir string, files ...string) string {
	for _, name := range files {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		s := strings.TrimSpace(string(b))
		if s != "" && !strings.HasPrefix(s, "#") {
			return strings.Fields(s)[0]
		}
	}
	return ""
}

func toolVersion(dir, tool string) string {
	b, err := os.ReadFile(filepath.Join(dir, ".tool-versions"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == tool {
			return parts[1]
		}
	}
	return ""
}

func appendVersionIssue(issues []Diagnostic, dir, tool, want, bin string, args []string, probe func(string, ...string) (string, error)) []Diagnostic {
	actual, err := probe(bin, args...)
	if err != nil {
		return append(issues, Diagnostic{dir, tool, "could not determine version: " + err.Error(), "check `" + bin + " --version`"})
	}
	if !matchesPinnedVersion(want, actual) {
		return append(issues, Diagnostic{dir, tool, fmt.Sprintf("expected %s, found %s", want, actual), "select the project's pinned version or update its version file"})
	}
	return issues
}

func appendConstraintIssue(issues []Diagnostic, dir, tool, constraint, bin string, args []string, probe func(string, ...string) (string, error)) []Diagnostic {
	actual, err := probe(bin, args...)
	if err != nil {
		return append(issues, Diagnostic{dir, tool, "could not determine version: " + err.Error(), "check the installed runtime"})
	}
	if !matchesConstraint(constraint, actual) {
		return append(issues, Diagnostic{dir, tool, fmt.Sprintf("requires %s, found %s", constraint, actual), "select a compatible runtime version"})
	}
	return issues
}

func matchesPinnedVersion(pin, output string) bool {
	if strings.EqualFold(pin, "system") || strings.EqualFold(pin, "latest") || strings.EqualFold(pin, "lts/*") {
		return true
	}
	want := versionNumbers.FindString(pin)
	got := versionNumbers.FindString(output)
	if want == "" || got == "" {
		return true
	} // unknown syntax: do not claim incompatibility
	wp := strings.Split(want, ".")
	gp := strings.Split(got, ".")
	if len(gp) < len(wp) {
		return false
	}
	for i, v := range wp {
		if gp[i] != v {
			return false
		}
	}
	return true
}

func supportsSimpleConstraint(c string) bool {
	c = strings.TrimSpace(c)
	return regexp.MustCompile(`^(>=|\^|~)?\s*\d+(?:\.\d+){0,2}(?:\.\*)?$`).MatchString(c)
}

func matchesConstraint(constraint, output string) bool {
	want := versionNumbers.FindString(constraint)
	got := versionNumbers.FindString(output)
	if want == "" || got == "" {
		return true
	}
	a := semverParts(got)
	b := semverParts(want)
	cmp := compareParts(a, b)
	c := strings.TrimSpace(constraint)
	switch {
	case strings.HasPrefix(c, ">="):
		return cmp >= 0
	case strings.HasPrefix(c, "^"):
		return cmp >= 0 && a[0] == b[0]
	case strings.HasPrefix(c, "~"):
		return cmp >= 0 && a[0] == b[0] && a[1] == b[1]
	default:
		return matchesPinnedVersion(want, got)
	}
}
func semverParts(s string) [3]int {
	var v [3]int
	for i, p := range strings.Split(s, ".") {
		if i >= 3 {
			break
		}
		v[i], _ = strconv.Atoi(p)
	}
	return v
}
func compareParts(a, b [3]int) int {
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
