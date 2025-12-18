package fillstruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"

	"golang.org/x/tools/go/packages"
)

type FormatError struct {
	Message string
	PosText string
}

func (e *FormatError) String() string {
	return fmt.Sprintf("%s:\n%s", e.PosText, e.Message)
}

type FormatResult struct {
	Path    string
	Output  []byte
	Errors  []*FormatError
	Changed bool
}

type Option struct {
}

func Format(pkg *packages.Package, file *ast.File, option *Option) (*FormatResult, error) {
	path := pkg.Fset.Position(file.Pos()).Filename
	basicLitExprs := make([]*ast.BasicLit, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		return false
	})

	errors := make([]*FormatError, 0, len(basicLitExprs))
	if len(basicLitExprs) == 0 {
		return &FormatResult{
			Path:    path,
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pkg.Fset, file); err != nil {
		return nil, fmt.Errorf("%s: failed to print AST: %v", pkg.Fset.Position(file.Pos()), err)
	}

	result, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to format source: %v", pkg.Fset.Position(file.Pos()), err)
	}

	return &FormatResult{
		Path:    path,
		Output:  result,
		Errors:  errors,
		Changed: true,
	}, nil
}
