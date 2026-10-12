package pkx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func InstallPlan(root string, projects []Project, only string) ([]Step, error) {
	var steps []Step
	matched := false
	for _, project := range projects {
		if only != "" && only != project.Manager && only != project.Ecosystem {
			continue
		}
		matched = true
		dir := filepath.Join(root, filepath.FromSlash(project.Directory))
		switch project.Manager {
		case "npm":
			cmd := "install"
			if fileExists(filepath.Join(dir, "package-lock.json")) || fileExists(filepath.Join(dir, "npm-shrinkwrap.json")) {
				cmd = "ci"
			}
			steps = append(steps, Step{dir, "npm", "npm", []string{cmd}})
		case "pnpm":
			args := []string{"install"}
			if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
				args = append(args, "--frozen-lockfile")
			}
			steps = append(steps, Step{dir, "pnpm", "pnpm", args})
		case "yarn":
			args := []string{"install"}
			if fileExists(filepath.Join(dir, "yarn.lock")) {
				var data struct {
					PackageManager string `json:"packageManager"`
				}
				if err := readJSON(filepath.Join(dir, "package.json"), &data); err != nil {
					return nil, err
				}
				if strings.HasPrefix(data.PackageManager, "yarn@1.") || data.PackageManager == "" {
					args = append(args, "--frozen-lockfile")
				} else {
					args = append(args, "--immutable")
				}
			}
			steps = append(steps, Step{dir, "yarn", "yarn", args})
		case "bun":
			args := []string{"install"}
			if fileExists(filepath.Join(dir, "bun.lock")) || fileExists(filepath.Join(dir, "bun.lockb")) {
				args = append(args, "--frozen-lockfile")
			}
			steps = append(steps, Step{dir, "bun", "bun", args})
		case "composer":
			steps = append(steps, Step{dir, "composer", "composer", []string{"install", "--no-interaction"}})
		case "cargo":
			steps = append(steps, Step{dir, "cargo", "cargo", []string{"fetch"}})
		case "go":
			steps = append(steps, Step{dir, "go", "go", []string{"mod", "download"}})
		case "uv":
			steps = append(steps, Step{dir, "uv", "uv", []string{"sync", "--locked"}})
		case "poetry":
			steps = append(steps, Step{dir, "poetry", "poetry", []string{"install", "--no-interaction"}})
		case "pip":
			python := pythonCommand()
			venv := filepath.Join(dir, ".venv")
			exists, err := readDir(venv)
			if err != nil {
				return nil, err
			}
			if !exists {
				steps = append(steps, Step{dir, "python", python, []string{"-m", "venv", ".venv"}})
			}
			bin := filepath.Join(venv, "bin", "python")
			if runtime.GOOS == "windows" {
				bin = filepath.Join(venv, "Scripts", "python.exe")
			}
			steps = append(steps, Step{dir, "pip", bin, []string{"-m", "pip", "install", "-r", "requirements.txt"}})
		default:
			return nil, fmt.Errorf("unsupported manager %q", project.Manager)
		}
	}
	if only != "" && !matched {
		return nil, fmt.Errorf("no projects found for %q", only)
	}
	return steps, nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular()
}

func readDir(path string) (bool, error) {
	st, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return st.IsDir(), nil
}

func pythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

// MissingCommands checks all external executables before any installation starts.
// The project-local venv interpreter is deliberately skipped because pkx creates it.
func MissingCommands(steps []Step, lookup func(string) (string, error)) []string {
	missing := []string{}
	seen := map[string]bool{}
	for _, step := range steps {
		if strings.ContainsAny(step.Binary, "/\\") {
			continue
		}
		if seen[step.Binary] {
			continue
		}
		seen[step.Binary] = true
		if _, err := lookup(step.Binary); err != nil {
			missing = append(missing, step.Binary)
		}
	}
	return missing
}
