package pkx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReviewedCapabilityCatalogHasDistinctAliases(t *testing.T) {
	catalog, err := loadCapabilities(builtInCapabilities)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Families) < 3 {
		t.Fatalf("expected at least three reviewed capability families, got %d", len(catalog.Families))
	}
	for _, family := range catalog.Families {
		for _, ecosystem := range []string{"node", "python", "php", "rust", "go"} {
			if _, ok := family.Variants[ecosystem]; !ok {
				t.Errorf("%s: no %s variant", family.ID, ecosystem)
			}
		}
	}
}

func TestOneAliasResolvesToDifferentNativeLibraries(t *testing.T) {
	catalog, _ := loadCapabilities(builtInCapabilities)
	cases := []struct{ ecosystem, name string }{
		{"node", "axios"},
		{"python", "httpx"},
		{"php", "guzzlehttp/guzzle"},
		{"rust", "reqwest"},
		{"go", "github.com/go-resty/resty/v2"},
	}
	for _, c := range cases {
		resolved, err := resolveCapability(catalog, "axios", c.ecosystem)
		if err != nil || resolved.Package != c.name {
			t.Errorf("%s: want %q got %+v / %v", c.ecosystem, c.name, resolved, err)
		}
		if resolved.Alternative != (c.ecosystem != "node") {
			t.Errorf("%s: mismatched replacement flag %+v", c.ecosystem, resolved)
		}
	}
}

func TestCapabilityStableFamilyKeyAndUnmappedName(t *testing.T) {
	catalog, _ := loadCapabilities(builtInCapabilities)
	for ecosystem, target := range map[string]string{"node": "axios", "python": "httpx", "php": "guzzlehttp/guzzle", "rust": "reqwest", "go": "github.com/go-resty/resty/v2"} {
		got, err := resolveCapability(catalog, "http-client", ecosystem)
		if err != nil || got.Package != target {
			t.Fatalf("%s: %q / %v", ecosystem, got.Package, err)
		}
	}
	got, err := resolveCapability(catalog, "http-server", "node")
	if err != nil || got.Package != "http-server" || got.Alternative {
		t.Fatalf("unmapped native name changed: %+v / %v", got, err)
	}
}

func TestCapabilityResolverIsOfflineForDryRun(t *testing.T) {
	project := Project{Directory: ".", Ecosystem: "python", Manager: "uv"}
	steps, resolved, err := resolveAndCheckPackages(context.Background(), project, []string{"axios", "http-client"}, false)
	if err != nil || len(steps) != 1 || steps[0] != "httpx" || len(resolved) != 1 {
		t.Fatalf("dedup resolution failed: %+v %+v %v", steps, resolved, err)
	}
}

func TestExplicitAlternativeApprovalIsRequired(t *testing.T) {
	for _, c := range []struct {
		input    string
		approved bool
	}{
		{"y\n", true}, {"yes\n", true}, {"YES\n", true}, {"n\n", false}, {"\n", false}, {"", false},
	} {
		var output bytes.Buffer
		got := approveAlternative(strings.NewReader(c.input), &output)
		if got != c.approved || !strings.Contains(output.String(), "NOT an API-compatible") {
			t.Errorf("input %q: approval=%v output=%q", c.input, got, output.String())
		}
	}
}

type mockRegistryTransport func(*http.Request) (*http.Response, error)

func (fn mockRegistryTransport) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestExactOfficialRegistryPreflightIsBoundedToTrustedHost(t *testing.T) {
	cases := []struct{ manager, name, hostname string }{
		{"npm", "axios", "registry.npmjs.org"},
		{"pnpm", "axios", "registry.npmjs.org"},
		{"pip", "httpx", "pypi.org"},
		{"uv", "httpx", "pypi.org"},
		{"poetry", "httpx", "pypi.org"},
		{"composer", "guzzlehttp/guzzle", "repo.packagist.org"},
		{"cargo", "reqwest", "crates.io"},
		{"go", "github.com/go-resty/resty/v2", "proxy.golang.org"},
	}
	for _, c := range cases {
		t.Run(c.manager, func(t *testing.T) {
			client := &http.Client{Transport: mockRegistryTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || r.URL.Host != c.hostname || r.Method != "GET" {
					return nil, fmt.Errorf("unexpected registry request: %s", r.URL)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
			})}
			if err := verifyPackageRegistry(context.Background(), client, c.manager, c.name); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRegistryNotFoundRefusesInstallation(t *testing.T) {
	client := &http.Client{Transport: mockRegistryTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	err := verifyPackageRegistry(context.Background(), client, "npm", "axios")
	if err == nil || !strings.Contains(err.Error(), "no installation started") {
		t.Fatalf("missing package was allowed: %v", err)
	}
}

func TestCatalogDuplicateAliasRejected(t *testing.T) {
	input := []byte(`{"schema":1,"families":[{"id":"one","title":"One","aliases":["dup"],"variants":{"node":{"name":"axios","url":"https://npmjs.com"}}},{"id":"two","title":"Two","aliases":["dup"],"variants":{"node":{"name":"zod","url":"https://npmjs.com"}}}]}`)
	if _, err := loadCapabilities(input); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate aliases not rejected: %v", err)
	}
}
