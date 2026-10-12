package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmpedWasTaken/pkx/internal/pkx"
)

// npmPlan demonstrates the exact production discovery -> native plan path.
// It never executes npm or makes a network request.
func npmPlan(root string) ([]pkx.Step, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	projects, err := pkx.Discover(absolute, 1)
	if err != nil {
		return nil, err
	}
	return pkx.InstallPlan(absolute, projects, "npm")
}

func main() {
	directory := flag.String("dir", "examples/adapter-npm/fixture", "example workspace directory")
	flag.Parse()
	steps, err := npmPlan(*directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adapter example:", err)
		os.Exit(1)
	}
	for _, step := range steps {
		fmt.Printf("%s: %s\n", step.Manager, step.String())
	}
	fmt.Println("Nothing executed.")
}
