package pkx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Version = "0.1.0"

type Project struct {
	Directory string `json:"directory"`
	Ecosystem string `json:"ecosystem"`
	Manager   string `json:"manager"`
	Manifest  string `json:"manifest"`
}

type Step struct {
	Directory string   `json:"directory"`
	Manager   string   `json:"manager"`
	Binary    string   `json:"binary"`
	Args      []string `json:"args"`
}

func (s Step) String() string {
	parts := append([]string{s.Binary}, s.Args...)
	for i, part := range parts {
		if strings.ContainsAny(part, " \t\n\"'") {
			parts[i] = fmt.Sprintf("%q", part)
		}
	}
	return strings.Join(parts, " ")
}

func relative(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return dir
	}
	if rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func within(root, userPath string) (string, error) {
	if userPath == "" || filepath.IsAbs(userPath) {
		return "", fmt.Errorf("directory must be a relative path inside the workspace")
	}
	// Canonicalize the workspace before comparing paths: on macOS /var can
	// resolve to /private/var, and Windows may normalize junctions or casing.
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("workspace: %w", err)
	}
	abs := filepath.Join(canonicalRoot, filepath.Clean(userPath))
	rel, err := filepath.Rel(canonicalRoot, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("directory %q is outside the workspace", userPath)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("directory %q: %w", userPath, err)
	}
	check, err := filepath.Rel(canonicalRoot, resolved)
	if err != nil || check == ".." || strings.HasPrefix(check, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("directory %q resolves outside the workspace", userPath)
	}
	stat, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("%q is not a directory", userPath)
	}
	return resolved, nil
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
