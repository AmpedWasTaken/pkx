package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type problem struct {
	file    string
	line    int
	message string
}

var markdownLink = regexp.MustCompile(`\]\(([^)]*)\)`)

func main() {
	flag.Parse()
	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	problems, err := checkDocs(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, item := range problems {
		fmt.Printf("%s:%d: %s\n", item.file, item.line, item.message)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
}

func checkDocs(root string) ([]problem, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	var problems []problem
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != absRoot && entry.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".md", ".markdown":
			problems = append(problems, checkMarkdown(absRoot, rel, path)...)
		case ".svg":
			if item := checkSVG(rel, path); item != nil {
				problems = append(problems, *item)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repository: %w", err)
	}

	sort.Slice(problems, func(i, j int) bool {
		if problems[i].file == problems[j].file {
			return problems[i].line < problems[j].line
		}
		return problems[i].file < problems[j].file
	})
	return problems, nil
}

func checkMarkdown(root, relative, path string) []problem {
	content, err := os.ReadFile(path)
	if err != nil {
		return []problem{{file: relative, line: 1, message: fmt.Sprintf("read file: %v", err)}}
	}

	var problems []problem
	for _, match := range markdownLink.FindAllStringSubmatchIndex(string(content), -1) {
		destination := string(content[match[2]:match[3]])
		destination = strings.TrimSpace(destination)
		if destination == "" {
			continue
		}
		if strings.HasPrefix(destination, "<") {
			if end := strings.Index(destination, ">"); end >= 0 {
				destination = destination[1:end]
			}
		} else {
			destination = strings.Fields(destination)[0]
		}

		parsed, err := url.Parse(destination)
		if err != nil || parsed.IsAbs() || parsed.Host != "" || strings.HasPrefix(destination, "//") {
			continue
		}
		if parsed.Path == "" {
			continue
		}

		target := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(parsed.Path)))
		relTarget, err := filepath.Rel(root, target)
		if err != nil || relTarget == ".." || strings.HasPrefix(relTarget, ".."+string(filepath.Separator)) {
			problems = append(problems, problem{
				file:    relative,
				line:    lineNumber(string(content), match[0]),
				message: fmt.Sprintf("local link escapes the repository: %q", destination),
			})
			continue
		}
		if _, err := os.Stat(target); err != nil {
			problems = append(problems, problem{
				file:    relative,
				line:    lineNumber(string(content), match[0]),
				message: fmt.Sprintf("local link does not exist: %q", destination),
			})
		}
	}
	return problems
}

func checkSVG(relative, path string) *problem {
	file, err := os.Open(path)
	if err != nil {
		return &problem{file: relative, line: 1, message: fmt.Sprintf("read SVG: %v", err)}
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	seenRoot := false
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				if !seenRoot {
					return &problem{file: relative, line: 1, message: "SVG has no root element"}
				}
				return nil
			}
			return &problem{file: relative, line: 1, message: fmt.Sprintf("invalid SVG XML: %v", err)}
		}
		if start, ok := token.(xml.StartElement); ok && !seenRoot {
			seenRoot = true
			if start.Name.Local != "svg" {
				return &problem{file: relative, line: 1, message: fmt.Sprintf("SVG root element is <%s>, want <svg>", start.Name.Local)}
			}
		}
	}
}

func lineNumber(content string, offset int) int {
	return 1 + strings.Count(content[:offset], "\n")
}
