package pkx

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var npmPackageName = regexp.MustCompile(`^(?:@[a-z0-9][a-z0-9._~-]*/)?[a-z0-9][a-z0-9._~-]*$`)
var commonPackageName = regexp.MustCompile(`(?i)^[a-z0-9][a-z0-9._-]*$`)
var composerPackageName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*/[a-z0-9][a-z0-9._-]*$`)
var goModuleName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*\.[a-z]{2,}(?:/[a-z0-9][a-z0-9._~+-]*)+$`)

// A package argument activates package-aware install. No positional arguments
// retain the original pkx install project-bootstrap behavior.
func packageInstallRequested(args []string) bool {
	for i := 0; i < len(args); i++ {
		key, _, equals := strings.Cut(args[i], "=")
		switch key {
		case "--dir", "--depth", "--only":
			if !equals && i+1 < len(args) {
				i++
			}
		case "--project", "--manager", "--dev":
			return true
		case "--dry-run", "--json", "-h", "--help":
		default:
			if !strings.HasPrefix(args[i], "-") {
				return true
			}
		}
	}
	return false
}

func selectLibraryProject(root string, projects []Project, opts manageOptions) (Project, error) {
	var candidates []Project
	var selected string
	if opts.Project != "" {
		dir, err := within(root, opts.Project)
		if err != nil {
			return Project{}, err
		}
		selected = relative(root, dir)
	}
	for _, project := range projects {
		if selected != "" && project.Directory != selected {
			continue
		}
		if opts.Manager != "" && opts.Manager != project.Manager && opts.Manager != project.Ecosystem {
			continue
		}
		candidates = append(candidates, project)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) == 0 {
		return Project{}, fmt.Errorf("no matching package project found; run 'pkx scan' from your project or use --project DIRECTORY")
	}
	var choices []string
	for _, project := range candidates {
		choices = append(choices, fmt.Sprintf("%s (%s)", project.Directory, project.Manager))
	}
	sort.Strings(choices)
	return Project{}, fmt.Errorf("multiple projects detected: %s; use --project DIRECTORY to choose safely", strings.Join(choices, ", "))
}

func validateLibraryName(manager, name string) error {
	var valid bool
	switch manager {
	case "npm", "pnpm", "yarn", "bun":
		valid = npmPackageName.MatchString(name)
	case "composer":
		valid = composerPackageName.MatchString(name)
	case "pip", "uv", "poetry", "cargo":
		valid = commonPackageName.MatchString(name)
	case "go":
		valid = goModuleName.MatchString(name)
	default:
		return fmt.Errorf("unsupported package manager %q", manager)
	}
	if !valid {
		return fmt.Errorf("package %q is not a valid %s registry package name (no ecosystem prefix or URL required)", name, manager)
	}
	return nil
}

func universalInstallSteps(root string, project Project, packages []string, dev bool) ([]Step, error) {
	dir := filepath.Join(root, filepath.FromSlash(project.Directory))
	for _, name := range packages {
		if err := validateSpec(name); err != nil {
			return nil, err
		}
		if err := validateLibraryName(project.Manager, name); err != nil {
			return nil, err
		}
	}
	switch project.Manager {
	case "npm", "pnpm", "yarn", "bun", "composer":
		step, err := addStep(dir, project.Manager, packages, dev)
		if err != nil {
			return nil, err
		}
		return []Step{step}, nil
	case "uv", "poetry", "cargo":
		args := []string{"add"}
		if dev {
			if project.Manager == "poetry" {
				args = append(args, "--group", "dev")
			} else {
				args = append(args, "--dev")
			}
		}
		return []Step{{Directory: dir, Manager: project.Manager, Binary: project.Manager, Args: append(args, packages...)}}, nil
	case "go":
		if dev {
			return nil, fmt.Errorf("Go modules do not support --dev")
		}
		return []Step{{Directory: dir, Manager: "go", Binary: "go", Args: append([]string{"get"}, packages...)}}, nil
	case "pip":
		if dev {
			return nil, fmt.Errorf("pip requirements projects do not support --dev; use a native dependency group manager")
		}
		venv := filepath.Join(dir, ".venv")
		var steps []Step
		st, err := os.Lstat(venv)
		if os.IsNotExist(err) {
			steps = append(steps, Step{Directory: dir, Manager: "python", Binary: pythonCommand(), Args: []string{"-m", "venv", ".venv"}})
		} else if err != nil {
			return nil, err
		} else if !st.IsDir() {
			return nil, fmt.Errorf(".venv must be an ordinary directory, not a file or symlink")
		}
		binary := filepath.Join(venv, "bin", "python")
		if filepath.Separator == '\\' {
			binary = filepath.Join(venv, "Scripts", "python.exe")
		}
		steps = append(steps, Step{Directory: dir, Manager: "pip", Binary: binary, Args: append([]string{"-m", "pip", "install"}, packages...)})
		return steps, nil
	default:
		return nil, fmt.Errorf("no package-install adapter for %q", project.Manager)
	}
}

func runUniversalInstall(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts, packages, _, err := parseManageArgs("add", args)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: install PACKAGE... [--dir PATH] [--project DIR] [--manager MANAGER] [--dev] [--dry-run]: %v\n", err)
		return 2
	}
	root, err := rootPath(opts.Dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: workspace: %v\n", err)
		return 1
	}
	projects, err := Discover(root, 4)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: discovery: %v\n", err)
		return 1
	}
	project, err := selectLibraryProject(root, projects, opts)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	steps, err := universalInstallSteps(root, project, packages, opts.Dev)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "Detected %s project using %s at %s.\n", project.Ecosystem, project.Manager, project.Directory)
	if project.Manager == "pip" {
		fmt.Fprintln(stdout, "Note: pip install does not update requirements.txt; record the dependency there for reproducible setup.")
	}
	return execute(ctx, root, steps, opts.DryRun, stdout, stderr)
}
