package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"strings"
	"sync"

	"github.com/nametake/fillstruct"
	"golang.org/x/tools/go/packages"
)

type arrayFlags []string

func (a *arrayFlags) String() string {
	return fmt.Sprint(*a)
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	var typeFlags arrayFlags
	var defaultFlags arrayFlags
	flag.Var(&typeFlags, "type", "target type (importpath.TypeName), can be specified multiple times")
	flag.Var(&defaultFlags, "default", "custom default value (format: TypeSpec=ConstantName), can be specified multiple times")
	flag.Parse()

	// If no --type flag is specified, do nothing
	if len(typeFlags) == 0 {
		os.Exit(0)
	}

	args := flag.Args()
	pattern := "./..."
	if len(args) > 0 {
		pattern = args[0]
	}

	// Extract directory from pattern for resolving target types
	dir := "."
	if pattern != "./..." && pattern != "." {
		dir = pattern
		if len(dir) >= 4 && dir[len(dir)-4:] == "/..." {
			dir = dir[:len(dir)-4]
		}
		if dir == "" {
			dir = "."
		}
	}

	// Resolve target types
	targetTypes, err := fillstruct.ResolveTargetTypes(typeFlags, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving target types: %v\n", err)
		os.Exit(1)
	}

	// Parse default values
	customDefaults, err := parseDefaultValues(defaultFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing default values: %v\n", err)
		os.Exit(1)
	}

	option := &fillstruct.Option{
		TargetTypes:    targetTypes,
		CustomDefaults: customDefaults,
	}

	if err := run(pattern, option); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// parseDefaultValues parses default value specifications
// Format: "TypeSpec=ConstantName"
// TypeSpec can be:
//   - Basic type name (e.g., "int", "string", "bool")
//   - Fully qualified type name (e.g., "github.com/example/domain.Status")
func parseDefaultValues(specs []string) (map[string]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	defaults := make(map[string]string)
	for _, spec := range specs {
		parts := strings.SplitN(spec, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %q (expected TypeSpec=ConstantName)", spec)
		}

		typeSpec := strings.TrimSpace(parts[0])
		constantName := strings.TrimSpace(parts[1])

		// Basic validation
		if typeSpec == "" || constantName == "" {
			return nil, fmt.Errorf("type and constant cannot be empty in %q", spec)
		}

		defaults[typeSpec] = constantName
	}

	return defaults, nil
}

func run(dir string, option *fillstruct.Option) error {
	waitGroup := sync.WaitGroup{}

	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedImports,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return fmt.Errorf("failed to load packages: path = %s: %v", dir, err)
	}

	errCount := 0
	format := func(pkg *packages.Package, file *ast.File, wg *sync.WaitGroup) {
		defer func() {
			wg.Done()
		}()

		result, err := fillstruct.Format(pkg, file, option)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				errCount += 1
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
		if !result.Changed {
			return
		}

		if err := os.WriteFile(result.Path, result.Output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			waitGroup.Add(1)
			go format(pkg, file, &waitGroup)
		}
	}

	waitGroup.Wait()

	if errCount > 0 {
		return fmt.Errorf("failed to format %d files", errCount)
	}

	return nil
}
