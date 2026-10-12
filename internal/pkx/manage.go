package pkx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

type manageOptions struct {
	Dir     string
	Project string
	Manager string
	Dev     bool
	DryRun  bool
}

// parseManageArgs accepts flags before or after package names. Arguments after
// -- are forwarded only to 'run', never interpreted as pkx options.
func parseManageArgs(command string, args []string) (manageOptions, []string, []string, error) {
	opts := manageOptions{Dir: "."}
	var positional, forwarded []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			forwarded = append(forwarded, args[i+1:]...)
			break
		}
		key, value, hasValue := strings.Cut(arg, "=")
		switch key {
		case "--dir", "--project", "--manager":
			if !hasValue {
				i++
				if i >= len(args) || args[i] == "" || strings.HasPrefix(args[i], "-") {
					return opts, nil, nil, fmt.Errorf("%s needs a value", key)
				}
				value = args[i]
			}
			if value == "" {
				return opts, nil, nil, fmt.Errorf("%s needs a value", key)
			}
			switch key {
			case "--dir":
				opts.Dir = value
			case "--project":
				opts.Project = value
			case "--manager":
				opts.Manager = value
			}
		case "--dev":
			if hasValue || command != "add" {
				return opts, nil, nil, fmt.Errorf("--dev is only supported by 'pkx add'")
			}
			opts.Dev = true
		case "--dry-run":
			if hasValue {
				return opts, nil, nil, fmt.Errorf("--dry-run does not take a value")
			}
			opts.DryRun = true
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, nil, nil, fmt.Errorf("unknown option %q; see 'pkx help %s'", arg, command)
			}
			positional = append(positional, arg)
		}
	}
	if command == "run" && len(positional) != 1 {
		return opts, nil, nil, fmt.Errorf("usage: pkx run SCRIPT [--project DIR] [--dry-run] [-- ARGS...]")
	}
	if (command == "add" || command == "remove") && len(positional) == 0 {
		return opts, nil, nil, fmt.Errorf("usage: pkx %s PACKAGE... [--project DIR] [--dry-run]", command)
	}
	if command != "run" && len(forwarded) > 0 {
		return opts, nil, nil, fmt.Errorf("arguments after -- are only supported by 'pkx run'")
	}
	return opts, positional, forwarded, nil
}

func runManage(ctx context.Context, command string, args []string, stdout, stderr io.Writer) int {
	opts, positional, forwarded, err := parseManageArgs(command, args)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 2
	}
	for _, value := range positional {
		if err := validateSpec(value); err != nil {
			fmt.Fprintf(stderr, "pkx: %v\n", err)
			return 2
		}
	}
	root, err := rootPath(opts.Dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: workspace: %v\n", err)
		return 1
	}
	projects, err := Discover(root, 4)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: couldn't scan workspace: %v\n", err)
		return 1
	}
	project, err := selectManageProject(root, projects, opts, command, first(positional))
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	step, err := manageStep(root, project, command, positional, forwarded, opts.Dev)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	return execute(ctx, root, []Step{step}, opts.DryRun, stdout, stderr)
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func manageable(manager, command string) bool {
	switch manager {
	case "npm", "pnpm", "yarn", "bun", "composer":
		return true
	default:
		return false
	}
}

func selectManageProject(root string, projects []Project, opts manageOptions, command, script string) (Project, error) {
	var candidates []Project
	var selection string
	if opts.Project != "" {
		dir, err := within(root, opts.Project)
		if err != nil {
			return Project{}, err
		}
		selection = relative(root, dir)
	}
	for _, project := range projects {
		if selection != "" && project.Directory != selection {
			continue
		}
		if opts.Manager != "" && project.Manager != opts.Manager && project.Ecosystem != opts.Manager {
			continue
		}
		if !manageable(project.Manager, command) {
			continue
		}
		if command == "run" {
			scripts, err := projectScripts(filepath.Join(root, filepath.FromSlash(project.Directory)), project.Manager)
			if err != nil {
				return Project{}, err
			}
			if !hasScript(scripts, script) {
				continue
			}
		}
		candidates = append(candidates, project)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) == 0 {
		if command == "run" {
			var available []string
			for _, project := range projects {
				if !manageable(project.Manager, command) || (selection != "" && project.Directory != selection) {
					continue
				}
				scripts, err := projectScripts(filepath.Join(root, filepath.FromSlash(project.Directory)), project.Manager)
				if err == nil {
					for _, name := range scripts {
						available = append(available, fmt.Sprintf("%s (%s)", name, project.Directory))
					}
				}
			}
			sort.Strings(available)
			if len(available) > 0 {
				return Project{}, fmt.Errorf("script %q not found; available: %s", script, strings.Join(available, ", "))
			}
			return Project{}, fmt.Errorf("no supported scripts found; define a script in package.json or composer.json")
		}
		return Project{}, fmt.Errorf("no supported package project found; run 'pkx scan' or choose --project DIR (add/remove supports npm, pnpm, yarn, bun and Composer)")
	}
	var paths []string
	for _, project := range candidates {
		paths = append(paths, fmt.Sprintf("%s (%s)", project.Directory, project.Manager))
	}
	return Project{}, fmt.Errorf("multiple projects match: %s; use --project DIR to choose one", strings.Join(paths, ", "))
}

func hasScript(scripts []string, name string) bool {
	for _, script := range scripts {
		if script == name {
			return true
		}
	}
	return false
}

func projectScripts(dir, manager string) ([]string, error) {
	filename := "package.json"
	if manager == "composer" {
		filename = "composer.json"
	}
	var manifest struct {
		Scripts map[string]json.RawMessage `json:"scripts"`
	}
	if err := readJSON(filepath.Join(dir, filename), &manifest); err != nil {
		return nil, err
	}
	var names []string
	for name := range manifest.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func manageStep(root string, project Project, command string, packages, forwarded []string, dev bool) (Step, error) {
	dir := filepath.Join(root, filepath.FromSlash(project.Directory))
	manager := project.Manager
	var args []string
	switch command {
	case "add":
		switch manager {
		case "npm":
			args = []string{"install"}
			if dev {
				args = append(args, "--save-dev")
			}
		case "composer":
			args = []string{"require", "--no-interaction"}
			if dev {
				args = append(args, "--dev")
			}
		default:
			args = []string{"add"}
			if dev {
				args = append(args, "--dev")
			}
		}
		args = append(args, packages...)
	case "remove":
		switch manager {
		case "npm":
			args = []string{"uninstall"}
		case "composer":
			args = []string{"remove", "--no-interaction"}
		default:
			args = []string{"remove"}
		}
		args = append(args, packages...)
	case "run":
		if len(packages) != 1 {
			return Step{}, fmt.Errorf("run expects one script name")
		}
		if manager == "composer" {
			args = []string{"run-script", packages[0]}
		} else {
			args = []string{"run", packages[0]}
		}
		if len(forwarded) > 0 {
			args = append(args, "--")
			args = append(args, forwarded...)
		}
	default:
		return Step{}, fmt.Errorf("unsupported action %q", command)
	}
	return Step{Directory: dir, Manager: manager, Binary: manager, Args: args}, nil
}
