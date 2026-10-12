package pkx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const registryOrigin = "https://pkx.ampedwastaken.dev"
const maxCatalogBytes = 512 << 10
const maxManifestBytes = 256 << 10

var bundleReference = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,49}/[a-z0-9][a-z0-9-]{0,49}$`)
var bundleVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
var bundleDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

type registryEntry struct {
	ID         string   `json:"id"`
	Version    string   `json:"version"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Ecosystem  string   `json:"ecosystem"`
	Managers   []string `json:"managers"`
	Summary    string   `json:"summary"`
	Maintainer string   `json:"maintainer"`
	Path       string   `json:"path"`
	SHA256     string   `json:"sha256"`
	Packages   []string `json:"packages"`
	Kind       string   `json:"kind"`
}

type registryCatalog struct {
	Schema  int             `json:"schema"`
	Bundles []registryEntry `json:"bundles"`
}

type registryReader struct {
	base   string
	client *http.Client
}

func defaultRegistry() registryReader {
	return registryReader{
		base: registryOrigin,
		client: &http.Client{
			Timeout: 12 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (r registryReader) get(ctx context.Context, resource string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.base+resource, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("marketplace request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketplace responded with HTTP %d", resp.StatusCode)
	}
	bytes, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(bytes)) > maxBytes {
		return nil, fmt.Errorf("marketplace response exceeds %d bytes", maxBytes)
	}
	return bytes, nil
}

func validEntry(entry registryEntry) bool {
	if !bundleReference.MatchString(entry.ID) || !bundleVersion.MatchString(entry.Version) || !bundleDigest.MatchString(entry.SHA256) {
		return false
	}
	if entry.Path != "/marketplace/bundles/"+entry.ID+"/"+entry.Version+".json" {
		return false
	}
	return entry.Title != "" && entry.Summary != "" && entry.Maintainer != "" && len(entry.Managers) > 0
}

func (r registryReader) list(ctx context.Context) ([]registryEntry, error) {
	bytes, err := r.get(ctx, "/marketplace/catalog.json", maxCatalogBytes)
	if err != nil {
		return nil, err
	}
	var catalog registryCatalog
	if err := json.Unmarshal(bytes, &catalog); err != nil {
		return nil, fmt.Errorf("invalid marketplace catalog: %w", err)
	}
	if catalog.Schema != 1 || catalog.Bundles == nil {
		return nil, fmt.Errorf("unsupported marketplace catalog schema")
	}
	seen := map[string]bool{}
	for _, entry := range catalog.Bundles {
		key := entry.ID + "@" + entry.Version
		if !validEntry(entry) || seen[key] {
			return nil, fmt.Errorf("invalid or duplicate marketplace catalog entry %q", key)
		}
		seen[key] = true
	}
	sort.Slice(catalog.Bundles, func(i, j int) bool {
		a, b := catalog.Bundles[i], catalog.Bundles[j]
		if a.ID == b.ID {
			return a.Version < b.Version
		}
		return a.ID < b.ID
	})
	return catalog.Bundles, nil
}

func splitRegistryReference(input string) (string, string, error) {
	id, version, ok := strings.Cut(input, "@")
	if !ok || !bundleReference.MatchString(id) || !bundleVersion.MatchString(version) {
		return "", "", errors.New("expected OWNER/NAME@VERSION (example: pkx/frontend-toolkit@0.1.0)")
	}
	return id, version, nil
}

func findRegistryEntry(items []registryEntry, id, version string) (registryEntry, error) {
	for _, entry := range items {
		if entry.ID == id && entry.Version == version {
			return entry, nil
		}
	}
	return registryEntry{}, fmt.Errorf("bundle %s@%s not found in catalog", id, version)
}

func validateRemoteBundle(data []byte) error {
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("invalid bundle JSON: %w", err)
	}
	if bundle.Schema != 1 || strings.TrimSpace(bundle.Name) == "" || len(bundle.Packages) == 0 {
		return errors.New("unsupported or empty bundle manifest")
	}
	for _, group := range bundle.Packages {
		switch group.Manager {
		case "npm", "pnpm", "yarn", "bun", "composer", "pip":
		default:
			return fmt.Errorf("unsupported bundle manager %q", group.Manager)
		}
		dir := filepath.Clean(filepath.FromSlash(group.Directory))
		if group.Directory == "" || filepath.IsAbs(dir) || dir == ".." || strings.HasPrefix(dir, ".."+string(filepath.Separator)) || strings.Contains(group.Directory, ":") {
			return fmt.Errorf("unsafe bundle directory %q", group.Directory)
		}
		if len(group.Dependencies)+len(group.Dev) == 0 {
			return errors.New("bundle group contains no dependencies")
		}
		for _, spec := range append(append([]string{}, group.Dependencies...), group.Dev...) {
			if err := validateSpec(spec); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r registryReader) verifiedManifest(ctx context.Context, item registryEntry) ([]byte, error) {
	if !validEntry(item) {
		return nil, errors.New("invalid marketplace entry")
	}
	data, err := r.get(ctx, item.Path, maxManifestBytes)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != item.SHA256 {
		return nil, errors.New("SHA256 mismatch: refusing unverified bundle")
	}
	if err := validateRemoteBundle(data); err != nil {
		return nil, fmt.Errorf("bundle validation failed: %w", err)
	}
	return data, nil
}

func runMarketplaceBundle(ctx context.Context, name string, args []string, stdout, stderr io.Writer) int {
	var selector string
	if name == "info" || name == "fetch" {
		if len(args) == 0 || strings.HasPrefix(args[0], "-") {
			fmt.Fprintf(stderr, "pkx: usage: pkx bundle %s OWNER/NAME@VERSION [--dir PATH]\n", name)
			return 2
		}
		selector, args = args[0], args[1:]
		if _, _, err := splitRegistryReference(selector); err != nil {
			fmt.Fprintf(stderr, "pkx: %v\n", err)
			return 2
		}
	}
	fs := flag.NewFlagSet("bundle "+name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "machine-readable output (search/info)")
	dir := fs.String("dir", ".", "workspace to save pkx.json (fetch only)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if (name != "search" && fs.NArg() > 0) || (name == "fetch" && *jsonOutput) || (name != "fetch" && *dir != ".") {
		fmt.Fprintln(stderr, "pkx: unsupported arguments; run 'pkx help bundle' for usage")
		return 2
	}
	registry := defaultRegistry()
	entries, err := registry.list(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	if name == "search" {
		query := strings.ToLower(strings.Join(fs.Args(), " "))
		matches := make([]registryEntry, 0)
		for _, entry := range entries {
			haystack := strings.ToLower(entry.Title+" "+entry.Summary+" "+entry.ID+" "+strings.Join(entry.Packages, " "))
			if strings.Contains(haystack, query) {
				matches = append(matches, entry)
			}
		}
		if *jsonOutput {
			return writeJSON(matches, stdout, stderr)
		}
		if len(matches) == 0 {
			fmt.Fprintln(stdout, "No bundles matched. Browse https://pkx.ampedwastaken.dev/marketplace/")
			return 0
		}
		for _, entry := range matches {
			fmt.Fprintf(stdout, "%-39s %s\n", entry.ID+"@"+entry.Version, entry.Title)
		}
		fmt.Fprintln(stdout, "Preview with 'pkx bundle info OWNER/NAME@VERSION'; nothing installed.")
		return 0
	}
	id, version, _ := splitRegistryReference(selector)
	entry, err := findRegistryEntry(entries, id, version)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	data, err := registry.verifiedManifest(ctx, entry)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	if name == "info" {
		if *jsonOutput {
			return writeJSON(entry, stdout, stderr)
		}
		fmt.Fprintf(stdout, "%s@%s\n%s\nMaintainer: %s\nManagers: %s\nDependencies: %s\nSHA256: %s (verified artifact integrity, not authorship)\n",
			entry.ID, entry.Version, entry.Summary, entry.Maintainer, strings.Join(entry.Managers, ", "), strings.Join(entry.Packages, ", "), entry.SHA256)
		fmt.Fprintln(stdout, "Preview download: pkx bundle fetch "+selector)
		return 0
	}
	root, err := rootPath(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: invalid workspace: %v\n", err)
		return 1
	}
	dst := filepath.Join(root, "pkx.json")
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: cannot create %s (existing bundles are never overwritten): %v\n", dst, err)
		return 1
	}
	if _, err = file.Write(data); err != nil {
		file.Close()
		os.Remove(dst)
		fmt.Fprintf(stderr, "pkx: writing bundle failed: %v\n", err)
		return 1
	}
	if err := file.Close(); err != nil {
		fmt.Fprintf(stderr, "pkx: saving bundle failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Saved verified bundle %s to %s\n", selector, dst)
	fmt.Fprintln(stdout, "Nothing installed. Review pkx.json, then run 'pkx bundle install pkx.json --dry-run'.")
	return 0
}

