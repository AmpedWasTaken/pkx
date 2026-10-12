package pkx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const packageAPI = "https://pkx.ampedwastaken.dev/api/packages"
const packageSearchLimit = 1024 * 1024

type PackageResult struct {
	ID          string   `json:"id"`
	Ecosystem   string   `json:"ecosystem"`
	Manager     string   `json:"manager"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     *string  `json:"version"`
	RegistryURL string   `json:"registryUrl"`
	Downloads   *int64   `json:"downloads"`
	Command     []string `json:"command,omitempty"`
	Project     string   `json:"project,omitempty"`
}

type packageSearchResponse struct {
	Query     string          `json:"query"`
	Ecosystem string          `json:"ecosystem"`
	Results   []PackageResult `json:"results"`
	Warnings  []string        `json:"warnings"`
	Source    string          `json:"source"`
}

type packageSearchOptions struct {
	Query     string
	Ecosystem string
	Dir       string
	Project   string
	JSON      bool
}

func parsePackageSearch(args []string) (packageSearchOptions, error) {
	opts := packageSearchOptions{Dir: ".", Ecosystem: "all"}
	var parts []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		key, value, has := strings.Cut(arg, "=")
		switch key {
		case "--json":
			if has {
				return opts, errors.New("--json does not take a value")
			}
			opts.JSON = true
		case "--dir", "--project", "--ecosystem":
			if !has {
				i++
				if i == len(args) || strings.HasPrefix(args[i], "-") || args[i] == "" {
					return opts, fmt.Errorf("%s requires a value", key)
				}
				value = args[i]
			}
			if value == "" {
				return opts, fmt.Errorf("%s requires a value", key)
			}
			switch key {
			case "--dir":
				opts.Dir = value
			case "--project":
				opts.Project = value
			case "--ecosystem":
				opts.Ecosystem = value
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, fmt.Errorf("unknown option %s", arg)
			}
			parts = append(parts, arg)
		}
	}
	opts.Query = strings.TrimSpace(strings.Join(parts, " "))
	if len(opts.Query) < 2 || len(opts.Query) > 72 || strings.ContainsAny(opts.Query, "\n\r\x00") {
		return opts, errors.New("search query must contain 2–72 printable characters")
	}
	switch opts.Ecosystem {
	case "all", "npm", "composer", "cargo":
	default:
		return opts, errors.New("ecosystem must be all, npm, composer, or cargo")
	}
	return opts, nil
}

type packageSearchReader struct {
	base   string
	client *http.Client
}

func defaultPackageReader() packageSearchReader {
	return packageSearchReader{
		base: packageAPI,
		client: &http.Client{
			Timeout: 8 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func validPackageResult(p PackageResult) bool {
	if len(p.Name) == 0 || len(p.Name) > 180 || strings.ContainsAny(p.Name, " \n\r\t\x00;|&$\\") || strings.Contains(p.Name, "..") || strings.HasPrefix(p.Name, "-") {
		return false
	}
	if p.ID != p.Ecosystem+":"+p.Name {
		return false
	}
	switch p.Ecosystem {
	case "npm":
		return p.RegistryURL == "https://www.npmjs.com/package/"+escapePackageName(p.Name)
	case "composer":
		return p.RegistryURL == "https://packagist.org/packages/"+p.Name
	case "cargo":
		return p.RegistryURL == "https://crates.io/crates/"+escapePackageName(p.Name)
	default:
		return false
	}
}

func escapePackageName(s string) string {
	return url.QueryEscape(s)
}

func (r packageSearchReader) search(ctx context.Context, opts packageSearchOptions) (packageSearchResponse, error) {
	endpoint := r.base + "?q=" + url.QueryEscape(opts.Query) + "&ecosystem=" + opts.Ecosystem
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return packageSearchResponse{}, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return packageSearchResponse{}, fmt.Errorf("package API unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return packageSearchResponse{}, fmt.Errorf("package API returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, packageSearchLimit+1))
	if err != nil {
		return packageSearchResponse{}, err
	}
	if len(data) > packageSearchLimit {
		return packageSearchResponse{}, errors.New("package search response too large")
	}
	var result packageSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return packageSearchResponse{}, fmt.Errorf("invalid package API response: %w", err)
	}
	if result.Source != "official-registry-search" || result.Query != opts.Query || len(result.Results) > 50 {
		return packageSearchResponse{}, errors.New("unexpected package API response")
	}
	for _, item := range result.Results {
		if !validPackageResult(item) {
			return packageSearchResponse{}, errors.New("invalid package data returned by registry")
		}
	}
	return result, nil
}

func projectPackageEcosystem(manager string) string {
	switch manager {
	case "npm", "pnpm", "yarn", "bun":
		return "npm"
	case "composer":
		return "composer"
	case "cargo":
		return "cargo"
	default:
		return ""
	}
}

func chooseSearchProject(root, target string, projects []Project) (*Project, error) {
	if target == "" {
		if len(projects) == 1 {
			return &projects[0], nil
		}
		return nil, nil
	}
	selected, err := within(root, target)
	if err != nil {
		return nil, err
	}
	rel := relative(root, selected)
	for i := range projects {
		if projects[i].Directory == rel {
			return &projects[i], nil
		}
	}
	return nil, fmt.Errorf("no detected project at %q", target)
}

func enrichPackageResults(results []PackageResult, project *Project) {
	for i := range results {
		entry := &results[i]
		manager := entry.Manager
		if project != nil {
			if projectPackageEcosystem(project.Manager) != entry.Ecosystem {
				continue
			}
			manager = project.Manager
			entry.Project = project.Directory
		}
		switch manager {
		case "npm":
			entry.Command = []string{"npm", "install", entry.Name}
		case "pnpm", "yarn", "bun":
			entry.Command = []string{manager, "add", entry.Name}
		case "composer":
			entry.Command = []string{"composer", "require", entry.Name}
		case "cargo":
			entry.Command = []string{"cargo", "add", entry.Name}
		}
	}
}

func runPackageSearch(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts, err := parsePackageSearch(args)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 2
	}
	root, err := rootPath(opts.Dir)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: directory: %v\n", err)
		return 1
	}
	// Discovery is read-only. Ambiguous multi-project repos never cause an
	// arbitrary manager to be selected; use --project explicitly.
	projects, err := Discover(root, 4)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: discovery: %v\n", err)
		return 1
	}
	selected, err := chooseSearchProject(root, opts.Project, projects)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: project: %v\n", err)
		return 1
	}
	resp, err := defaultPackageReader().search(ctx, opts)
	if err != nil {
		fmt.Fprintf(stderr, "pkx: %v\n", err)
		return 1
	}
	enrichPackageResults(resp.Results, selected)
	if opts.JSON {
		return writeJSON(resp, stdout, stderr)
	}
	if selected != nil {
		fmt.Fprintf(stdout, "Project: %s (%s)\n", selected.Directory, selected.Manager)
	} else if len(projects) > 1 {
		fmt.Fprintln(stdout, "Multiple projects detected; use --project DIR for project-specific commands.")
	}
	if len(resp.Results) == 0 {
		fmt.Fprintln(stdout, "No packages found. Try another search term or ecosystem.")
		return 0
	}
	for _, p := range resp.Results {
		ver := ""
		if p.Version != nil {
			ver = " v" + *p.Version
		}
		fmt.Fprintf(stdout, "%s%s — %s\n", p.ID, ver, p.Description)
		if len(p.Command) > 0 {
			fmt.Fprintf(stdout, "  Suggestion (not run): %s\n", strings.Join(p.Command, " "))
		}
	}
	for _, warning := range resp.Warnings {
		fmt.Fprintf(stderr, "pkx: warning: %s\n", warning)
	}
	fmt.Fprintln(stdout, "Nothing installed. Review packages before using pkx add or the native manager.")
	return 0
}

