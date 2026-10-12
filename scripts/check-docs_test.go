package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckMarkdownRejectsMissingLocalTarget(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("[missing](docs/not-there.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	problems, err := checkDocs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || problems[0].line != 1 {
		t.Fatalf("checkDocs() = %#v, want one line-1 problem", problems)
	}
}

func TestCheckSVGRejectsMalformedXML(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "broken.svg"), []byte("<svg><path></svg>"), 0o644); err != nil {
		t.Fatal(err)
	}

	problems, err := checkDocs(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("checkDocs() = %#v, want one problem", problems)
	}
}
