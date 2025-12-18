package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
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
	flag.Var(&typeFlags, "type", "target type (importpath.TypeName), can be specified multiple times")
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

	// Resolve target types
	targetTypes, err := fillstruct.ResolveTargetTypes(typeFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving target types: %v\n", err)
		os.Exit(1)
	}

	option := &fillstruct.Option{
		TargetTypes: targetTypes,
	}

	if err := run(pattern, option); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(dir string, option *fillstruct.Option) error {
	waitGroup := sync.WaitGroup{}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedImports,
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
