package pkx

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

var commandHelp = map[string]string{
	"status":     "pkx status [--dir PATH] [--depth N] [--json]\n  Show detected projects, lockfiles, planned steps, and environment issues.\n  Reads project files and checks installed tool versions; changes nothing.\n",
	"plan":       "pkx plan [--dir PATH] [--depth N] [--only MANAGER] [--json]\n  Preview the exact native install commands, without installing anything.\n  Use --json to integrate with scripts and editors.\n",
	"scan":       "pkx scan [--dir PATH] [--depth N] [--json]\n  List projects detected under PATH. 'pkx ls' is a shortcut.\n",
	"up":         "pkx up [--dir PATH] [--depth N] [--only MANAGER] [--dry-run]\n  Check the environment, then run native dependency installers.\n  Use 'pkx plan' first to see what will run.\n",
	"install":    "pkx install [--dir PATH] [--depth N] [--only MANAGER] [--dry-run]\n  Run native package managers without checking runtime version pins.\n  'pkx i' is a shortcut; 'pkx up' also checks the environment.\n",
	"doctor":     "pkx doctor [--dir PATH] [--depth N] [--only MANAGER] [--json]\n  Diagnose missing tools or incompatible pinned runtime versions.\n  Use --json for automation (exit code 1 when issues are found).\n  Does not install runtimes or dependencies.\n",
	"bundle":     "pkx bundle init [--dir PATH]\npkx bundle install FILE [--dir PATH] [--dry-run]\n  Create a local bundle manifest, or apply one after reviewing the plan.\n  'pkx init' is shorthand for 'pkx bundle init'.\n",
	"init":       "pkx init [--dir PATH]\n  Shortcut for 'pkx bundle init'. Creates a starter pkx.json.\n",
	"version":    "pkx version\n  Print the version of the installed CLI.\n",
	"completion": "pkx completion <bash|zsh|fish|powershell>\n  Print tab completion for your shell; see the README for installation.\n",
}

func canonicalCommand(command string) string {
	switch command {
	case "i":
		return "install"
	case "ls", "list":
		return "scan"
	default:
		return command
	}
}

func showHelp(command string, stdout, stderr io.Writer) int {
	if command == "help" {
		fmt.Fprint(stdout, help)
		return 0
	}
	if command == "--version" || command == "-v" {
		command = "version"
	}
	message, exists := commandHelp[command]
	if !exists {
		fmt.Fprintf(stderr, "pkx: unknown help topic %q\n", command)
		if suggestion := closestCommand(command); suggestion != "" {
			fmt.Fprintf(stderr, "Did you mean 'pkx help %s'?\n", suggestion)
		}
		return 2
	}
	fmt.Fprint(stdout, message)
	return 0
}

func closestCommand(value string) string {
	value = strings.ToLower(value)
	closest, score := "", 3
	for _, name := range []string{"scan", "status", "plan", "up", "install", "doctor", "bundle", "init", "version", "completion"} {
		distance := editDistance(value, name)
		if distance < score {
			closest, score = name, distance
		}
	}
	return closest
}

// Bounded Levenshtein distance is adequate for our small, fixed command list.
func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(current[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = current
	}
	return prev[len(b)]
}

func minInt(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

type planStep struct {
	Directory string   `json:"directory"`
	Manager   string   `json:"manager"`
	Command   []string `json:"command"`
}

type planReport struct {
	Root     string     `json:"root"`
	Projects []Project  `json:"projects"`
	Steps    []planStep `json:"steps"`
}

func showPlan(root string, projects []Project, steps []Step, depth int, outputJSON bool, stdout, stderr io.Writer) int {
	report := planReport{Root: root, Projects: projects, Steps: make([]planStep, 0, len(steps))}
	if report.Projects == nil {
		report.Projects = []Project{}
	}
	for _, step := range steps {
		report.Steps = append(report.Steps, planStep{Directory: relative(root, step.Directory), Manager: step.Manager, Command: append([]string{step.Binary}, step.Args...)})
	}
	if outputJSON {
		return writeJSON(report, stdout, stderr)
	}
	if len(report.Projects) == 0 {
		printNoProjects(stdout, root, depth)
		return 0
	}
	fmt.Fprintf(stdout, "Install plan (%d %s across %d %s):\n", len(report.Steps), countNoun(len(report.Steps), "command"), len(report.Projects), countNoun(len(report.Projects), "project"))
	for _, step := range steps {
		fmt.Fprintf(stdout, "  %-12s %-14s %s\n", relative(root, step.Directory), step.Manager, step.String())
	}
	fmt.Fprintln(stdout, "Nothing executed. Run 'pkx up' to check your environment and install.")
	return 0
}

type statusProject struct {
	Directory string `json:"directory"`
	Ecosystem string `json:"ecosystem"`
	Manager   string `json:"manager"`
	Manifest  string `json:"manifest"`
	Lockfile  string `json:"lockfile,omitempty"`
}

type statusReport struct {
	Root        string          `json:"root"`
	Projects    []statusProject `json:"projects"`
	Planned     int             `json:"plannedCommands"`
	Diagnostics []Diagnostic    `json:"diagnostics"`
}

func showStatus(root string, projects []Project, steps []Step, depth int, outputJSON bool, stdout, stderr io.Writer) int {
	report := statusReport{Root: root, Projects: []statusProject{}, Planned: len(steps), Diagnostics: []Diagnostic{}}
	for _, project := range projects {
		p := statusProject{Directory: project.Directory, Ecosystem: project.Ecosystem, Manager: project.Manager, Manifest: project.Manifest}
		p.Lockfile = projectLockfile(filepath.Join(root, filepath.FromSlash(project.Directory)), project.Manager)
		report.Projects = append(report.Projects, p)
	}
	if len(projects) > 0 {
		report.Diagnostics = append(report.Diagnostics, CheckEnvironment(root, projects, steps, exec.LookPath, ProbeVersion)...)
	}
	if outputJSON {
		return writeJSON(report, stdout, stderr)
	}
	if len(projects) == 0 {
		printNoProjects(stdout, root, depth)
		return 0
	}
	fmt.Fprintf(stdout, "Workspace: %s\n%d %s, %d planned native %s\n", root, len(report.Projects), countNoun(len(report.Projects), "project"), len(steps), countNoun(len(steps), "command"))
	for _, project := range report.Projects {
		lock := "no lockfile detected"
		if project.Lockfile != "" {
			lock = project.Lockfile
		}
		fmt.Fprintf(stdout, "  %-14s %-10s %-10s %s\n", project.Directory, project.Ecosystem, project.Manager, lock)
	}
	if len(report.Diagnostics) > 0 {
		fmt.Fprintf(stdout, "Environment: %d %s to resolve:\n", len(report.Diagnostics), countNoun(len(report.Diagnostics), "issue"))
		for _, issue := range report.Diagnostics {
			fmt.Fprintf(stdout, "  ! %s\n", issue.String())
		}
		fmt.Fprintln(stdout, "Next: fix issues above, then run 'pkx up'.")
	} else {
		fmt.Fprintln(stdout, "Environment checks passed. Next: 'pkx plan' or 'pkx up'.")
	}
	return 0
}

// A package may have no lockfile (especially a new repository).
func projectLockfile(dir, manager string) string {
	var candidates []string
	switch manager {
	case "npm":
		candidates = []string{"npm-shrinkwrap.json", "package-lock.json"}
	case "pnpm":
		candidates = []string{"pnpm-lock.yaml"}
	case "yarn":
		candidates = []string{"yarn.lock"}
	case "bun":
		candidates = []string{"bun.lock", "bun.lockb"}
	case "composer":
		candidates = []string{"composer.lock"}
	case "cargo":
		candidates = []string{"Cargo.lock"}
	case "go":
		candidates = []string{"go.sum"}
	case "uv":
		candidates = []string{"uv.lock"}
	case "poetry":
		candidates = []string{"poetry.lock"}
	}
	for _, name := range candidates {
		if fileExists(filepath.Join(dir, name)) {
			return name
		}
	}
	return ""
}

func printNoProjects(out io.Writer, root string, depth int) {
	fmt.Fprintf(out, "No supported projects found under %s (depth %d).\n", root, depth)
	fmt.Fprintln(out, "Try running from the repository root or increasing --depth. See 'pkx help scan'.")
}

func writeJSON(value any, stdout, stderr io.Writer) int {
	if err := json.NewEncoder(stdout).Encode(value); err != nil {
		fmt.Fprintf(stderr, "pkx: failed to write JSON: %v\n", err)
		return 1
	}
	return 0
}

func countNoun(n int, noun string) string {
	if n == 1 {
		return noun
	}
	return noun + "s"
}

// filterProjects makes --only apply consistently to both the plan and diagnostics.
func filterProjects(projects []Project, only string) []Project {
	selected := make([]Project, 0, len(projects))
	for _, project := range projects {
		if only == "" || project.Ecosystem == only || project.Manager == only {
			selected = append(selected, project)
		}
	}
	return selected
}

type doctorReport struct {
	Root        string       `json:"root"`
	Ready       bool         `json:"ready"`
	Projects    []Project    `json:"projects"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}
