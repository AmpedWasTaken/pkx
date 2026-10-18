package pkx

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:embed capabilities.v1.json
var builtInCapabilities []byte

type capabilityCatalog struct {
	Schema   int                `json:"schema"`
	Families []capabilityFamily `json:"families"`
}

type capabilityFamily struct {
	ID       string                       `json:"id"`
	Title    string                       `json:"title"`
	Summary  string                       `json:"summary"`
	Aliases  []string                     `json:"aliases"`
	Variants map[string]capabilityVariant `json:"variants"`
}

type capabilityVariant struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
	URL      string `json:"url"`
}

type resolvedPackage struct {
	Requested   string
	Package     string
	Capability  string
	Alternative bool
}

// loadCapabilities verifies the entire embedded catalog, including aliases
// and native registry names. Registry metadata does not execute any code.
func loadCapabilities(data []byte) (capabilityCatalog, error) {
	var catalog capabilityCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return catalog, fmt.Errorf("capability catalog: %w", err)
	}
	if catalog.Schema != 1 || len(catalog.Families) == 0 {
		return catalog, fmt.Errorf("unsupported or empty capability catalog")
	}
	seenID := map[string]bool{}
	seenAlias := map[string]bool{}
	managerFor := map[string]string{
		"node": "npm", "python": "pip", "php": "composer", "rust": "cargo", "go": "go",
	}
	for _, family := range catalog.Families {
		if !commonPackageName.MatchString(family.ID) || seenID[family.ID] || family.Title == "" {
			return catalog, fmt.Errorf("invalid or duplicate capability id %q", family.ID)
		}
		seenID[family.ID] = true
		if len(family.Aliases) == 0 || len(family.Variants) == 0 {
			return catalog, fmt.Errorf("capability %s has no aliases or variants", family.ID)
		}
		for _, alias := range family.Aliases {
			key := strings.ToLower(alias)
			if key == "" || seenAlias[key] {
				return catalog, fmt.Errorf("duplicate or missing capability alias %q", alias)
			}
			seenAlias[key] = true
		}
		for ecosystem, target := range family.Variants {
			manager := managerFor[ecosystem]
			if manager == "" {
				return catalog, fmt.Errorf("unknown ecosystem %q", ecosystem)
			}
			if err := validateLibraryName(manager, target.Name); err != nil {
				return catalog, fmt.Errorf("capability %s: %w", family.ID, err)
			}
			if !strings.HasPrefix(target.URL, "https://") {
				return catalog, fmt.Errorf("invalid official registry URL for %s", target.Name)
			}
		}
	}
	return catalog, nil
}

func resolveCapability(catalog capabilityCatalog, request, ecosystem string) (resolvedPackage, error) {
	for _, family := range catalog.Families {
		for _, alias := range family.Aliases {
			if !strings.EqualFold(request, alias) {
				continue
			}
			variant, exists := family.Variants[ecosystem]
			if !exists {
				return resolvedPackage{}, fmt.Errorf("capability %q has no reviewed %s alternative; refusing to guess", family.ID, ecosystem)
			}
			return resolvedPackage{
				Requested:   request,
				Package:     variant.Name,
				Capability:  family.Title,
				Alternative: !strings.EqualFold(request, variant.Name),
			}, nil
		}
	}
	return resolvedPackage{Requested: request, Package: request}, nil
}

// Confirm when a capability alias maps to a DIFFERENT native package.
// An explicit --yes permits automation, but an EOF never means yes.
func approveAlternative(in io.Reader, out io.Writer) bool {
	fmt.Fprint(out, "This is a functionally similar library, NOT an API-compatible drop-in. Install the suggested alternative? [y/N]: ")
	scan := bufio.NewScanner(io.LimitReader(in, 32))
	if !scan.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scan.Text()))
	return answer == "y" || answer == "yes"
}

var verificationHTTP = &http.Client{
	Timeout: 7 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Registry existence checks avoid accidental installation of unknown typo
// names or packages from the wrong ecosystem. Package metadata is not audited.
func verifyPackageRegistry(ctx context.Context, client *http.Client, manager, name string) error {
	if err := validateLibraryName(manager, name); err != nil {
		return err
	}
	var address string
	encoded := url.PathEscape(name)
	switch manager {
	case "npm", "pnpm", "yarn", "bun":
		address = "https://registry.npmjs.org/" + encoded
	case "pip", "uv", "poetry":
		address = "https://pypi.org/pypi/" + encoded + "/json"
	case "composer":
		// Packagist provider metadata lists exact Composer package names.
		address = "https://repo.packagist.org/p2/" + name + ".json"
	case "cargo":
		address = "https://crates.io/api/v1/crates/" + encoded
	case "go":
		address = "https://proxy.golang.org/" + name + "/@v/list"
	default:
		return fmt.Errorf("no verification provider for %q", manager)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pkx/0.6 (https://github.com/AmpedWasTaken/pkx)")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not check the official %s registry: %w", manager, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return fmt.Errorf("%q was not found in the official %s registry; no installation started", name, manager)
	}
	return fmt.Errorf("official %s registry returned HTTP %d for %q; no installation started", manager, resp.StatusCode, name)
}

func resolveAndCheckPackages(ctx context.Context, project Project, requested []string, checkRegistry bool) ([]string, []resolvedPackage, error) {
	catalog, err := loadCapabilities(builtInCapabilities)
	if err != nil {
		return nil, nil, err
	}
	var packages []string
	var resolved []resolvedPackage
	seen := map[string]bool{}
	for _, name := range requested {
		if err := validateSpec(name); err != nil {
			return nil, nil, err
		}
		match, err := resolveCapability(catalog, name, project.Ecosystem)
		if err != nil {
			return nil, nil, err
		}
		if err = validateLibraryName(project.Manager, match.Package); err != nil {
			return nil, nil, err
		}
		if checkRegistry {
			if err = verifyPackageRegistry(ctx, verificationHTTP, project.Manager, match.Package); err != nil {
				return nil, nil, err
			}
		}
		if seen[match.Package] {
			continue
		}
		seen[match.Package] = true
		packages = append(packages, match.Package)
		resolved = append(resolved, match)
	}
	return packages, resolved, nil
}
