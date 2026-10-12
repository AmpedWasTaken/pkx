package main

import (
	"path/filepath"
	"testing"
)

func TestExampleNativeNpmPlan(t *testing.T) {
	steps, err := npmPlan(filepath.Join("fixture"))
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one native command, got %d", len(steps))
	}
	if steps[0].Manager != "npm" || steps[0].Binary != "npm" || steps[0].String() != "npm install" {
		t.Fatalf("unexpected native command: %#v", steps[0])
	}
}

func TestExampleIsReadOnlyWithMissingManager(t *testing.T) {
	// npmPlan produces a plan without requiring the npm executable on PATH.
	t.Setenv("PATH", t.TempDir())
	if _, err := npmPlan("fixture"); err != nil {
		t.Fatalf("planning unexpectedly required npm: %v", err)
	}
}
