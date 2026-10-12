package pkx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSearchFlagParsing(t *testing.T) {
	opts, err := parsePackageSearch([]string{"http", "client", "--ecosystem=npm", "--json", "--project", "frontend"})
	if err != nil || opts.Query != "http client" || !opts.JSON || opts.Ecosystem != "npm" || opts.Project != "frontend" {
		t.Fatalf("options: %+v, %v", opts, err)
	}
	for _, args := range [][]string{
		{"x"}, {"hello", "--ecosystem", "pypi"}, {"hello", "--json=false"},
		{"hello", "--unsupported"}, {"hello", "--project"},
	} {
		if _, err := parsePackageSearch(args); err == nil {
			t.Errorf("accepted invalid flags %q", args)
		}
	}
}

func TestNativeRecommendationsAreProjectAware(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "web"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "api"), 0755); err != nil {
		t.Fatal(err)
	}
	projects := []Project{
		{Directory: "web", Ecosystem: "node", Manager: "pnpm"},
		{Directory: "api", Ecosystem: "php", Manager: "composer"},
	}
	if project, err := chooseSearchProject(root, "", projects); err != nil || project != nil {
		t.Fatalf("ambiguous project should not auto-select: %+v %v", project, err)
	}
	selected, err := chooseSearchProject(root, "web", projects)
	if err != nil || selected.Manager != "pnpm" {
		t.Fatalf("explicit project: %+v %v", selected, err)
	}
	items := []PackageResult{
		{ID: "npm:axios", Ecosystem: "npm", Manager: "npm", Name: "axios"},
		{ID: "composer:monolog/monolog", Ecosystem: "composer", Manager: "composer", Name: "monolog/monolog"},
	}
	enrichPackageResults(items, selected)
	if !reflect.DeepEqual(items[0].Command, []string{"pnpm", "add", "axios"}) || items[0].Project != "web" {
		t.Fatalf("wrong project command: %+v", items[0])
	}
	if len(items[1].Command) != 0 {
		t.Fatalf("cross-language project must not suggest install: %+v", items[1])
	}
	if _, err := chooseSearchProject(root, "../escape", projects); err == nil {
		t.Fatal("accepted directory outside workspace")
	}
}

func TestPackageSearchHTTPAndValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "http client" || r.URL.Query().Get("ecosystem") != "npm" {
			t.Errorf("request query: %s", r.URL)
		}
		_ = json.NewEncoder(w).Encode(packageSearchResponse{
			Query: "http client", Ecosystem: "npm", Source: "official-registry-search",
			Results: []PackageResult{{ID: "npm:axios", Name: "axios", Ecosystem: "npm", Manager: "npm", RegistryURL: "https://www.npmjs.com/package/axios"}},
		})
	}))
	defer server.Close()
	r := packageSearchReader{base: server.URL, client: server.Client()}
	data, err := r.search(context.Background(), packageSearchOptions{Query: "http client", Ecosystem: "npm"})
	if err != nil || len(data.Results) != 1 {
		t.Fatalf("read-only search: %+v %v", data, err)
	}
	if !validPackageResult(PackageResult{ID: "npm:@scope/pkg", Name: "@scope/pkg", Ecosystem: "npm", RegistryURL: "https://www.npmjs.com/package/%40scope%2Fpkg"}) {
		t.Fatal("scoped package URL rejected")
	}
	for _, item := range []PackageResult{
		{ID: "npm:../bad", Name: "../bad", Ecosystem: "npm", RegistryURL: "https://www.npmjs.com/package/..%2Fbad"},
		{ID: "npm:axios", Name: "axios", Ecosystem: "npm", RegistryURL: "https://evil.test/axios"},
		{ID: "composer:bad&name", Name: "bad&name", Ecosystem: "composer", RegistryURL: "https://packagist.org/packages/bad&name"},
	} {
		if validPackageResult(item) {
			t.Errorf("accepted unsafe package: %+v", item)
		}
	}
}

func TestPackageSearchRefusesRedirectsAndOversizedData(t *testing.T) {
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:5/", http.StatusFound)
	}))
	defer redir.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	if _, err := (packageSearchReader{base: redir.URL, client: client}).search(context.Background(), packageSearchOptions{Query: "hello", Ecosystem: "npm"}); err == nil || !strings.Contains(err.Error(), "302") {
		t.Fatalf("expected redirect refusal: %v", err)
	}
	big := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", packageSearchLimit+1)))
	}))
	defer big.Close()
	if _, err := (packageSearchReader{base: big.URL, client: big.Client()}).search(context.Background(), packageSearchOptions{Query: "hello", Ecosystem: "npm"}); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected bound: %v", err)
	}
}
