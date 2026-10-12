package pkx

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var ignoredDirectories = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, ".venv": true,
	"venv": true, "target": true, "dist": true, "build": true,
	".next": true, ".turbo": true, ".cache": true, "__pycache__": true,
	".pnpm": true, ".idea": true, ".tox": true,
}

func Discover(root string, maxDepth int) ([]Project, error) {
	if maxDepth < 0 || maxDepth > 20 {
		return nil, fmt.Errorf("depth must be between 0 and 20")
	}
	dirs := map[string]map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (ignoredDirectories[entry.Name()] || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			if rel != "." && strings.Count(filepath.ToSlash(rel), "/")+1 > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		switch entry.Name() {
		case "package.json", "pnpm-lock.yaml", "package-lock.json", "npm-shrinkwrap.json", "yarn.lock", "bun.lock", "bun.lockb", "composer.json", "Cargo.toml", "go.mod", "requirements.txt", "pyproject.toml", "uv.lock", "poetry.lock":
			dir := filepath.Dir(path)
			if dirs[dir] == nil {
				dirs[dir] = map[string]bool{}
			}
			dirs[dir][entry.Name()] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var projects []Project
	for dir, files := range dirs {
		rel := relative(root, dir)
		if files["package.json"] {
			manager, err := nodeManager(dir, files)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", rel, err)
			}
			projects = append(projects, Project{rel, "node", manager, "package.json"})
		}
		if files["composer.json"] {
			projects = append(projects, Project{rel, "php", "composer", "composer.json"})
		}
		if files["Cargo.toml"] {
			projects = append(projects, Project{rel, "rust", "cargo", "Cargo.toml"})
		}
		if files["go.mod"] {
			projects = append(projects, Project{rel, "go", "go", "go.mod"})
		}
		if files["uv.lock"] || files["poetry.lock"] || files["requirements.txt"] {
			manager, manifest, err := pythonManager(files)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", rel, err)
			}
			projects = append(projects, Project{rel, "python", manager, manifest})
		}
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Directory != projects[j].Directory {
			return projects[i].Directory < projects[j].Directory
		}
		return projects[i].Ecosystem < projects[j].Ecosystem
	})
	return CollapseNodeWorkspaces(root, projects), nil
}

func nodeManager(dir string, files map[string]bool) (string, error) {
	var p struct {
		PackageManager string `json:"packageManager"`
	}
	if err := readJSON(filepath.Join(dir, "package.json"), &p); err != nil {
		return "", err
	}
	explicit := strings.SplitN(p.PackageManager, "@", 2)[0]
	if explicit != "" {
		switch explicit {
		case "npm", "pnpm", "yarn", "bun":
			return explicit, nil
		default:
			return "", fmt.Errorf("unsupported packageManager %q", p.PackageManager)
		}
	}
	locks := []struct{ name, manager string }{
		{"pnpm-lock.yaml", "pnpm"}, {"yarn.lock", "yarn"},
		{"bun.lock", "bun"}, {"bun.lockb", "bun"},
		{"package-lock.json", "npm"}, {"npm-shrinkwrap.json", "npm"},
	}
	found := map[string]bool{}
	for _, lock := range locks {
		if files[lock.name] {
			found[lock.manager] = true
		}
	}
	if len(found) > 1 {
		return "", fmt.Errorf("conflicting Node lockfiles: specify packageManager in package.json")
	}
	for name := range found {
		return name, nil
	}
	return "npm", nil
}

func pythonManager(files map[string]bool) (string, string, error) {
	if files["uv.lock"] && files["poetry.lock"] {
		return "", "", fmt.Errorf("both uv.lock and poetry.lock found")
	}
	if files["uv.lock"] {
		if !files["pyproject.toml"] {
			return "", "", fmt.Errorf("uv.lock requires pyproject.toml")
		}
		return "uv", "pyproject.toml", nil
	}
	if files["poetry.lock"] {
		if !files["pyproject.toml"] {
			return "", "", fmt.Errorf("poetry.lock requires pyproject.toml")
		}
		return "poetry", "pyproject.toml", nil
	}
	if files["requirements.txt"] {
		return "pip", "requirements.txt", nil
	}
	return "", "", fmt.Errorf("pyproject.toml has no supported lockfile or requirements.txt; use uv, Poetry, or requirements.txt")
}
