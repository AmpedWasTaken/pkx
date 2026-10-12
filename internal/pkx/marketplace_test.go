package pkx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func registryTestClient(base string) registryReader {
	return registryReader{
		base: base,
		client: &http.Client{
			Timeout: 2 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func TestMarketplaceCatalogAndManifest(t *testing.T) {
	manifest := []byte("{\"schema\":1,\"name\":\"test\",\"packages\":[{\"manager\":\"npm\",\"directory\":\".\",\"dependencies\":[\"axios\"]}]}\n")
	sum := sha256.Sum256(manifest)
	entry := registryEntry{
		ID:         "pkx/test-bundle",
		Version:    "0.1.0",
		Title:      "Test bundle",
		Summary:    "A test",
		Maintainer: "pkx",
		Managers:   []string{"npm"},
		Path:       "/marketplace/bundles/pkx/test-bundle/0.1.0.json",
		SHA256:     hex.EncodeToString(sum[:]),
	}
	catalog, err := json.Marshal(registryCatalog{Schema: 1, Bundles: []registryEntry{entry}})
	if err != nil {
		t.Fatal(err)
	}
	serveWrong := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/marketplace/catalog.json":
			w.Write(catalog)
		case entry.Path:
			if serveWrong {
				w.Write([]byte("tampered content"))
			} else {
				w.Write(manifest)
			}
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()
	registry := registryTestClient(server.URL)
	found, err := registry.list(context.Background())
	if err != nil || len(found) != 1 {
		t.Fatalf("catalog: %v / %v", found, err)
	}
	if _, _, err := splitRegistryReference("pkx/test-bundle@0.1.0"); err != nil {
		t.Fatal(err)
	}
	content, err := registry.verifiedManifest(context.Background(), found[0])
	if err != nil || string(content) != string(manifest) {
		t.Fatalf("verified manifest: %v / %v", err, content)
	}
	serveWrong = true
	if _, err := registry.verifiedManifest(context.Background(), entry); err == nil || !strings.Contains(err.Error(), "SHA256 mismatch") {
		t.Fatalf("expected digest failure: %v", err)
	}
}

func TestMarketplaceUntrustedInput(t *testing.T) {
	for _, ref := range []string{
		"../etc/passwd@1.0.0", "pkx/../../bad@1.0.0", "pkx/a@latest", "pkx/a",
		"pkx/a@1.0.0?download=true", "pkx/a@0.1.0/other",
	} {
		if _, _, err := splitRegistryReference(ref); err == nil {
			t.Errorf("accepted unsafe ref %q", ref)
		}
	}
	for _, dir := range []string{"../escape", "/tmp", "", "C:/Windows"} {
		raw := fmt.Sprintf(`{"schema":1,"name":"bad","packages":[{"manager":"npm","directory":%q,"dependencies":["axios"]}]}`, dir)
		if err := validateRemoteBundle([]byte(raw)); err == nil {
			t.Errorf("accepted directory %q", dir)
		}
	}
	if err := validateRemoteBundle([]byte(`{"schema":1,"name":"bad","packages":[{"manager":"npm","directory":".","dependencies":["--ignore-scripts"]}]}`)); err == nil {
		t.Fatal("accepted dependency flag injection")
	}
	dup := registryEntry{
		ID: "pkx/starter", Version: "1.0.0", Title: "Starter", Summary: "Starter",
		Maintainer: "pkx", Managers: []string{"npm"},
		Path: "/marketplace/bundles/pkx/starter/1.0.0.json",
		SHA256: strings.Repeat("a", 64),
	}
	catalog, _ := json.Marshal(registryCatalog{Schema: 1, Bundles: []registryEntry{dup, dup}})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.Write(catalog) }))
	defer server.Close()
	if _, err := registryTestClient(server.URL).list(context.Background()); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate catalog accepted: %v", err)
	}
}

func TestMarketplaceNetworkBoundaries(t *testing.T) {
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "http://127.0.0.1:1/private", http.StatusFound)
	}))
	defer redirect.Close()
	r := registryTestClient(redirect.URL)
	if _, err := r.list(context.Background()); err == nil || !strings.Contains(err.Error(), "302") {
		t.Fatalf("redirect followed or accepted: %v", err)
	}
	oversized := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(strings.Repeat("a", maxCatalogBytes+1)))
	}))
	defer oversized.Close()
	if _, err := registryTestClient(oversized.URL).list(context.Background()); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized catalog accepted: %v", err)
	}
}

func TestMarketplaceNeverOverwritesExistingBundle(t *testing.T) {
	root := t.TempDir()
	manifest := []byte("{}\n")
	file, err := saveRegistryManifest(root, manifest)
	if err != nil || filepath.Base(file) != "pkx.json" {
		t.Fatalf("failed to save bundle: %v", err)
	}
	if _, err = saveRegistryManifest(root, []byte("tamper")); !os.IsExist(err) {
		t.Fatalf("expected existing-file refusal: %v", err)
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != string(manifest) {
		t.Fatalf("original file replaced: %q, %v", data, err)
	}
}
