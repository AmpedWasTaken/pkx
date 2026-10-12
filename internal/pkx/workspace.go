package pkx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// CollapseNodeWorkspaces avoids running a second install inside packages
// already managed by a parent package-manager workspace.
func CollapseNodeWorkspaces(root string, projects []Project) []Project {
	out := make([]Project, 0, len(projects))
	for _, child := range projects {
		if child.Ecosystem != "node" {
			out = append(out, child)
			continue
		}
		controlled := false
		childDir := filepath.Join(root, filepath.FromSlash(child.Directory))
		for _, parent := range projects {
			if parent.Ecosystem != "node" || parent.Directory == child.Directory {
				continue
			}
			parentDir := filepath.Join(root, filepath.FromSlash(parent.Directory))
			rel, err := filepath.Rel(parentDir, childDir)
			if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			// A package with its own lockfile is independently installable.
			if hasOwnNodeLockfile(childDir) {
				continue
			}
			if explicitNodeManager(childDir) != "" && explicitNodeManager(childDir) != parent.Manager {
				continue
			}
			if workspaceMatches(parentDir, filepath.ToSlash(rel)) {
				controlled = true
				break
			}
		}
		if !controlled {
			out = append(out, child)
		}
	}
	return out
}

func explicitNodeManager(dir string) string {
	var p struct {
		PackageManager string `json:"packageManager"`
	}
	if readJSON(filepath.Join(dir, "package.json"), &p) != nil {
		return ""
	}
	return strings.SplitN(p.PackageManager, "@", 2)[0]
}

func hasOwnNodeLockfile(dir string) bool {
	for _, name := range []string{"package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "bun.lock", "bun.lockb"} {
		if fileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func workspaceMatches(dir, rel string) bool {
	var p struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if readJSON(filepath.Join(dir, "package.json"), &p) != nil {
		return false
	}
	var patterns []string
	if len(p.Workspaces) > 0 {
		if err := json.Unmarshal(p.Workspaces, &patterns); err != nil {
			var obj struct {
				Packages []string `json:"packages"`
			}
			if json.Unmarshal(p.Workspaces, &obj) == nil {
				patterns = append(patterns, obj.Packages...)
			}
		}
	}
	// pnpm-workspace.yaml is read without a YAML dependency. Only the top-level
	// packages string-list syntax is supported; unusual YAML should not cause skipping.
	if raw, err := os.ReadFile(filepath.Join(dir, "pnpm-workspace.yaml")); err == nil {
		inside := false
		for _, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "packages:" {
				inside = true
				continue
			}
			if strings.TrimSpace(line) == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inside = false
			}
			if inside && strings.HasPrefix(trimmed, "- ") {
				pattern := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `'"`)
				patterns = append(patterns, pattern)
			}
		}
	}
	matched := false
	for _, pattern := range patterns {
		exclude := strings.HasPrefix(pattern, "!")
		pattern = strings.TrimPrefix(pattern, "!")
		pattern = strings.TrimSuffix(pattern, "/")
		// Support the common packages/* and apps/** patterns. Match never uses
		// prefix-only rules for plain paths (which would match unrelated siblings).
		ok := false
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			ok = rel == prefix || strings.HasPrefix(rel, prefix+"/")
		} else {
			ok, _ = filepath.Match(filepath.FromSlash(pattern), filepath.FromSlash(rel))
		}
		if ok {
			matched = !exclude
		}
	}
	return matched
}
