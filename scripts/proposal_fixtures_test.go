package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are proposal fixtures, not public CLI inputs. Keep their shapes
// reviewable and non-executable while the RFCs are being discussed.
func TestDraftBundleCatalogFixture(t *testing.T) {
	root := filepath.Join("..", "examples")
	catalogPath := filepath.Join(root, "registry", "catalog.v1.json")
	content, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
	// The catalog schema is documentation only; the typed runtime parser is a future issue.
	var raw struct {
		Schema int `json:"schema"`
		Bundles []struct {
			ID string `json:"id"`
			Version string `json:"version"`
			Summary string `json:"summary"`
			ManifestPath string `json:"manifest_path"`
			Status string `json:"status"`
		} `json:"bundles"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.Schema != 1 || len(raw.Bundles) == 0 {
		t.Fatalf("unexpected catalog schema or empty fixture: %+v", raw)
	}
	for _, entry := range raw.Bundles {
		if entry.ID == "" || entry.Version == "" || entry.Summary == "" || entry.Status != "example-only" {
			t.Fatalf("incomplete proposal metadata: %+v", entry)
		}
		if filepath.IsAbs(entry.ManifestPath) || strings.Contains(entry.ManifestPath, "://") {
			t.Fatalf("non-local fixture path: %q", entry.ManifestPath)
		}
		target := filepath.Clean(filepath.Join(filepath.Dir(catalogPath), filepath.FromSlash(entry.ManifestPath)))
		rel, err := filepath.Rel(root, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("catalog fixture escapes examples/: %q", entry.ManifestPath)
		}
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("missing referenced sample bundle %q: %v", target, err)
		}
	}
}

func TestRuntimeProposalFixtureCannotExecute(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "examples", "runtime", "bootstrap-plan.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Schema int `json:"schema"`
		Mode string `json:"mode"`
		ExecutionEnabled bool `json:"execution_enabled"`
		Runtime struct {
			Name string `json:"name"`
			Requested string `json:"requested"`
			PinFile string `json:"pin_file"`
		} `json:"runtime"`
		ProposedAction struct {
			Type string `json:"type"`
			Description string `json:"description"`
		} `json:"proposed_action"`
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Schema != 1 || plan.Mode != "preview" || plan.ExecutionEnabled ||
		plan.Runtime.Name == "" || plan.Runtime.Requested == "" ||
		plan.ProposedAction.Type != "manual-instructions" || plan.ProposedAction.Description == "" {
		t.Fatalf("runtime fixture must remain non-executable: %+v", plan)
	}
}
