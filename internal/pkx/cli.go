package pkx

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const help = `pkx — Package eXchange

Get started:
  pkx status               See what's in this repository
  pkx plan                 Preview installs without running anything
  pkx up                   Check your environment, then install dependencies
  pkx install http-server     Detect project and install its package with the native manager
  pkx run dev               Run a project-defined script

Commands:
  status   Overview of projects, lockfiles and required tools
  plan     Show the exact install plan (--json supported)
  up       Check environment and install dependencies
  scan     List detected projects (--json supported)
  install  Install project dependencies, or install PACKAGE into a detected project
  add      Add one or more dependencies to a selected project
  remove   Remove dependencies from a selected project
  run      Execute a declared Node/Composer project script
  doctor   Diagnose missing tools and pinned versions
  bundle   Local bundles plus read-only marketplace search/info/fetch
  search   Find packages in official registries and suggest native commands
  version  Show the installed pkx version
  completion  Generate shell tab-completion (bash, zsh, fish, powershell)

Shortcuts: pkx ls = scan, pkx i = install, pkx rm = remove, pkx init = bundle init
More help: pkx help <command> or pkx <command> --help

Native package managers may run third-party scripts. Review untrusted
repositories and run 'pkx plan' before installing.
`

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, help)
		return 0
	}
	command := canonicalCommand(args[0])
	if command == "help" || command == "-h" || command == "--help" {
		if len(args) == 1 {
			fmt.Fprint(stdout, help)
			return 0
		}
		if len(args) > 2 {
			fmt.Fprintln(stderr, "pkx: usage: pkx help [command]")
			return 2
		}
		return showHelp(canonicalCommand(args[1]), stdout, stderr)
	}
	if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
		return showHelp(command, stdout, stderr)
	}
	switch command {
	case "completion":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "pkx: usage: pkx completion <bash|zsh|fish|powershell>")
			return 2
		}
		return writeCompletion(args[1], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, "pkx", Version)
		return 0
	case "scan", "status", "plan", "doctor", "up":
		return runWorkspace(ctx, command, args[1:], stdout, stderr)
	case "install":
		if packageInstallRequested(args[1:]) {
			return runUniversalInstall(ctx, args[1:], stdout, stderr)
		}
		return runWorkspace(ctx, command, args[1:], stdout, stderr)
	case "search":
		return runPackageSearch(ctx, args[1:], stdout, stderr)
	case "add", "remove", "run":
		return runManage(ctx, command, args[1:], stdout, stderr)
	case "bundle":
		if len(args) > 2 && (args[2] == "--help" || args[2] == "-h") {
			return showHelp("bundle", stdout, stderr)
		}
		return runBundle(ctx, args[1:], stdout, stderr)
	case "init":
		return runBundle(ctx, append([]string{"init"}, args[1:]...), stdout, stderr)
	default:
		fmt.Fprintf(stderr, "pkx: unknown command %q.\n", args[0])
		if suggestion := closestCommand(args[0]); suggestion != "" {
			fmt.Fprintf(stderr, "Did you mean 'pkx %s'?\n", suggestion)
		}
		fmt.Fprintln(stderr, "Run 'pkx help' for available commands.")
		return 2
	}
}

func workspaceFlags(name string, args []string, stderr io.Writer) (*flag.FlagSet, *string, *int) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", ".", "project directory")
	depth := fs.Int("depth", 4, "maximum subdirectory depth (0-20)")
	return fs, dir, depth
}

func rootPath(dir string) (string, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(root)
}

func runWorkspace(ctx context.Context, command string, args []string, stdout, stderr io.Writer) int {
	fs, dir, depth := workspaceFlags(command, args, stderr)
	only := fs.String("only", "", "select package manager or ecosystem")
	dry := fs.Bool("dry-run", false, "show commands without executing them")
	outputJSON := fs.Bool("json", false, "output machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "pkx: unexpected positional arguments; run 'pkx help "+command+"' for usage")
		return 2
	}
	if (command != "install" && command != "up" && command != "plan" && command != "doctor" && *only != "") || (command != "install" && command != "up" && *dry) || (command != "scan" && command != "status" && command != "plan" && command != "doctor" && *outputJSON) {
		fmt.Fprintf(stderr, "pkx: unsupported flag for %s; run 'pkx help %s' for usage\n", command, command)
		return 2
	}
	root, err := rootPath(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: couldn't open workspace %q: %v\n", *dir, err)
		return 1
	}
	projects, err := Discover(root, *depth)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: couldn't scan %s: %v\n", root, err)
		return 1
	}
	if command == "scan" {
		if *outputJSON {
			if projects == nil {
				projects = []Project{}
			}
			if err := json.NewEncoder(stdout).Encode(projects); err != nil {
				fmt.Fprintf(stderr, "pkx: %v\n", err)
				return 1
			}
			return 0
		}
		if len(projects) == 0 {
			printNoProjects(stdout, root, *depth)
			return 0
		}
		for _, p := range projects {
			fmt.Fprintf(stdout, "%-10s %-10s %s (%s)\n", p.Ecosystem, p.Manager, p.Directory, p.Manifest)
		}
		return 0
	}
	steps, err := InstallPlan(root, projects, *only)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	if command == "plan" {
		return showPlan(root, filterProjects(projects, *only), steps, *depth, *outputJSON, stdout, stderr)
	}
	if command == "status" {
		return showStatus(root, projects, steps, *depth, *outputJSON, stdout, stderr)
	}
	if len(steps) == 0 {
		if command == "doctor" && *outputJSON {
			return writeJSON(doctorReport{Root: root, Ready: true, Projects: []Project{}, Diagnostics: []Diagnostic{}}, stdout, stderr)
		}
		printNoProjects(stdout, root, *depth)
		return 0
	}
	if command == "doctor" || (command == "up" && !*dry) {
		checkedProjects := projects
		if *only != "" {
			checkedProjects = filterProjects(projects, *only)
		}
		diagnostics := CheckEnvironment(root, checkedProjects, steps, exec.LookPath, ProbeVersion)
		if command == "doctor" && *outputJSON {
			code := writeJSON(doctorReport{Root: root, Ready: len(diagnostics) == 0, Projects: checkedProjects, Diagnostics: diagnostics}, stdout, stderr)
			if code != 0 {
				return code
			}
			if len(diagnostics) > 0 {
				return 1
			}
			return 0
		}
		if command == "doctor" {
			for _, p := range checkedProjects {
				fmt.Fprintf(stdout, "✓ %s (%s) at %s\n", p.Manager, p.Ecosystem, p.Directory)
			}
		}
		for _, d := range diagnostics {
			fmt.Fprintf(stderr, "pkx: %s\n", d.String())
		}
		if len(diagnostics) > 0 {
			fmt.Fprintf(stderr, "pkx: %d environment issue(s); no installation started\n", len(diagnostics))
			return 1
		}
		if command == "doctor" {
			fmt.Fprintln(stdout, "Environment checks passed.")
			return 0
		}
		fmt.Fprintln(stdout, "Environment checks passed; installing dependencies...")
	}
	return execute(ctx, root, steps, *dry, stdout, stderr)
}

func runBundle(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "pkx: use 'pkx bundle init' or 'pkx bundle install FILE'")
		return 2
	}
	name := args[0]
	if name == "search" || name == "info" || name == "fetch" {
		return runMarketplaceBundle(ctx, name, args[1:], stdout, stderr)
	}
	fs := flag.NewFlagSet("bundle "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", ".", "workspace directory")
	dry := fs.Bool("dry-run", false, "show commands without executing")
	// For bundle install, flags before and after the filename are accepted.
	var file string
	rest := args[1:]
	if name == "install" && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		file, rest = rest[0], rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return 2
	}
	root, err := rootPath(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	switch name {
	case "init":
		if fs.NArg() != 0 || *dry {
			fmt.Fprintln(stderr, "pkx: usage: pkx bundle init [--dir PATH]")
			return 2
		}
		path := filepath.Join(root, "pkx.json")
		if err := BundleInit(path); err != nil {
			fmt.Fprintf(stderr, "pkx: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Created", path)
		return 0
	case "install":
		if file == "" && fs.NArg() == 1 {
			file = fs.Arg(0)
		}
		if file == "" || (file != "" && fs.NArg() > 0) {
			fmt.Fprintln(stderr, "pkx: usage: pkx bundle install FILE [--dir PATH] [--dry-run]")
			return 2
		}
		// Bundle path is relative to the workspace, not the caller's working directory.
		if !filepath.IsAbs(file) {
			file = filepath.Join(root, file)
		}
		steps, err := BundlePlan(root, file)
		if err != nil {
			fmt.Fprintf(stderr, "pkx: %v\n", err)
			return 1
		}
		return execute(ctx, root, steps, *dry, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "pkx: unknown bundle command %q\n", name)
		return 2
	}
}

func execute(ctx context.Context, root string, steps []Step, dry bool, stdout, stderr io.Writer) int {
	if len(steps) == 0 {
		fmt.Fprintln(stdout, "No dependencies to install.")
		return 0
	}
	for _, step := range steps {
		fmt.Fprintf(stdout, "[%s] %s: %s\n", step.Manager, relative(root, step.Directory), step.String())
	}
	if dry {
		fmt.Fprintf(stdout, "Dry run: %d command(s); nothing executed.\n", len(steps))
		return 0
	}
	missing := MissingCommands(steps, exec.LookPath)
	if len(missing) > 0 {
		fmt.Fprintf(stderr, "pkx: required tools not found in PATH: %s\n", strings.Join(missing, ", "))
		return 1
	}
	began := time.Now()
	completed := 0
	for _, step := range steps {
		started := time.Now()
		fmt.Fprintf(stdout, "→ %s (%s)\n", step.Manager, relative(root, step.Directory))
		cmd := exec.CommandContext(ctx, step.Binary, step.Args...)
		cmd.Dir = step.Directory
		cmd.Stdin = os.Stdin
		cmd.Stdout, cmd.Stderr = stdout, stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(stderr, "pkx: failed after %s (%d/%d steps completed)\n", time.Since(began).Round(time.Millisecond), completed, len(steps))
			fmt.Fprintf(stderr, "pkx: retry in %s with: %s\n", step.Directory, step.String())
			fmt.Fprintf(stderr, "pkx: run `pkx doctor` to check the local environment; inspect lockfile changes before retrying\n")
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				fmt.Fprintf(stderr, "pkx: %s exited with status %d\n", step.Manager, exit.ExitCode())
			}
			if !errors.As(err, &exit) {
				fmt.Fprintf(stderr, "pkx: failed to run %s: %v\n", step.Manager, err)
			}
			return 1
		}
		completed++
		fmt.Fprintf(stdout, "✓ %s in %s\n", step.Manager, time.Since(started).Round(time.Millisecond))
	}
	fmt.Fprintf(stdout, "Done: %d command(s) completed successfully in %s.\n", len(steps), time.Since(began).Round(time.Millisecond))
	return 0
}
