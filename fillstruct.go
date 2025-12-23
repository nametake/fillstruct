package fillstruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"unicode"
	"unicode/utf8"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
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
	TargetTypes []*types.Named
}

// ResolveTargetTypes resolves type specifications to *types.Named
// typeSpecs format: "importpath.TypeName" (e.g., "github.com/example/foo.Bar")
// dir is the directory to resolve packages from (e.g., "." or "./...")
func ResolveTargetTypes(typeSpecs []string, dir string) ([]*types.Named, error) {
	if len(typeSpecs) == 0 {
		return nil, nil
	}

	var targetTypes []*types.Named

	for _, spec := range typeSpecs {
		// Parse "importpath.TypeName"
		lastDot := -1
		for i := len(spec) - 1; i >= 0; i-- {
			if spec[i] == '.' {
				lastDot = i
				break
			}
		}

		if lastDot == -1 || lastDot == 0 || lastDot == len(spec)-1 {
			return nil, fmt.Errorf("invalid type specification format %q: expected 'importpath.TypeName'", spec)
		}

		importPath := spec[:lastDot]
		typeName := spec[lastDot+1:]

		// Load the package
		cfg := &packages.Config{
			Mode:  packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
			Dir:   dir,
			Tests: true,
		}
		pkgs, err := packages.Load(cfg, importPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load package %q: %w", importPath, err)
		}

		if len(pkgs) == 0 {
			return nil, fmt.Errorf("no packages found for %q", importPath)
		}

		// Try to find the type in all loaded packages (including test packages)
		var obj types.Object
		var foundPkg *packages.Package
		for _, pkg := range pkgs {
			if len(pkg.Errors) > 0 {
				continue
			}
			obj = pkg.Types.Scope().Lookup(typeName)
			if obj != nil {
				foundPkg = pkg
				break
			}
		}

		if obj == nil {
			return nil, fmt.Errorf("type %q not found in package %q", typeName, importPath)
		}

		if foundPkg != nil && len(foundPkg.Errors) > 0 {
			return nil, fmt.Errorf("errors in package %q: %v", importPath, foundPkg.Errors)
		}

		typeNameObj, ok := obj.(*types.TypeName)
		if !ok {
			return nil, fmt.Errorf("%q is not a type in package %q", typeName, importPath)
		}

		named, ok := typeNameObj.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("%q is not a named type in package %q", typeName, importPath)
		}

		// Check if underlying type is a struct
		if _, ok := named.Underlying().(*types.Struct); !ok {
			return nil, fmt.Errorf("type %q in package %q is not a struct (underlying type: %T)", typeName, importPath, named.Underlying())
		}

		targetTypes = append(targetTypes, named)
	}

	return targetTypes, nil
}

func Format(pkg *packages.Package, file *ast.File, option *Option) (*FormatResult, error) {
	path := pkg.Fset.Position(file.Pos()).Filename
	errors := make([]*FormatError, 0)

	// Convert ast.File to dst.File
	dec := decorator.NewDecorator(pkg.Fset)
	dstFile, err := dec.DecorateFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decorate file: %w", err)
	}

	changed := false

	// Inspect and modify composite literals
	dst.Inspect(dstFile, func(n dst.Node) bool {
		lit, ok := n.(*dst.CompositeLit)
		if !ok {
			return true
		}

		// Get corresponding ast.Node to access type information
		astNode := dec.Ast.Nodes[lit]
		astLit, ok := astNode.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Get type information
		tv, ok := pkg.TypesInfo.Types[astLit]
		if !ok {
			return true
		}

		// Get the underlying struct type and check if it matches target types
		var structType *types.Struct
		var namedType *types.Named

		switch t := tv.Type.(type) {
		case *types.Named:
			if s, ok := t.Underlying().(*types.Struct); ok {
				structType = s
				namedType = t
			}
		case *types.Pointer:
			if named, ok := t.Elem().(*types.Named); ok {
				if s, ok := named.Underlying().(*types.Struct); ok {
					structType = s
					namedType = named
				}
			}
		case *types.Struct:
			structType = t
		}

		if structType == nil {
			return true
		}

		// If target types are specified, check if this type matches
		if len(option.TargetTypes) > 0 {
			if namedType == nil {
				// Skip anonymous structs when target types are specified
				return true
			}

			matched := false
			for _, targetType := range option.TargetTypes {
				// Compare by package path and type name instead of types.Identical
				// because they may be from different package loads
				if namedType.Obj().Pkg().Path() == targetType.Obj().Pkg().Path() &&
					namedType.Obj().Name() == targetType.Obj().Name() {
					matched = true
					break
				}
			}

			if !matched {
				return true
			}
		}

		// Check if all elements are keyed
		if !isAllKeyed(lit.Elts) {
			return true
		}

		// Collect present fields
		presentFields := make(map[string]bool)
		for _, elt := range lit.Elts {
			if kv, ok := elt.(*dst.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*dst.Ident); ok {
					presentFields[ident.Name] = true
				}
			}
		}

		// Rebuild elements in struct field order
		type fieldInfo struct {
			index     int
			name      string
			fieldType types.Type
		}

		var allFields []fieldInfo
		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)
			if !isExportedField(field.Name()) {
				continue
			}
			allFields = append(allFields, fieldInfo{
				index:     i,
				name:      field.Name(),
				fieldType: field.Type(),
			})
		}

		// Check if any fields are missing
		hasMissing := false
		for _, field := range allFields {
			if !presentFields[field.name] {
				hasMissing = true
				break
			}
		}

		if !hasMissing {
			return true
		}

		// Build new elements list in struct field order
		var newElts []dst.Expr
		existingKVs := make(map[string]*dst.KeyValueExpr)
		var sampleKV *dst.KeyValueExpr

		for _, elt := range lit.Elts {
			if kv, ok := elt.(*dst.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*dst.Ident); ok {
					existingKVs[ident.Name] = kv
					if sampleKV == nil {
						sampleKV = kv
					}
				}
			}
		}

		for _, field := range allFields {
			if kv, ok := existingKVs[field.name]; ok {
				// Use existing KeyValueExpr
				newElts = append(newElts, kv)
			} else {
				// Create new KeyValueExpr for missing field
				zeroValue := generateZeroValue(field.fieldType, pkg)
				newKV := &dst.KeyValueExpr{
					Key:   &dst.Ident{Name: field.name},
					Value: zeroValue,
				}

				// Copy decorations from existing element if available
				if sampleKV != nil {
					newKV.Decs.Before = sampleKV.Decs.Before
					newKV.Decs.After = sampleKV.Decs.After
				} else {
					newKV.Decs.Before = dst.NewLine
					newKV.Decs.After = dst.NewLine
				}

				newElts = append(newElts, newKV)
			}
		}

		lit.Elts = newElts

		changed = true
		return true
	})

	if !changed {
		return &FormatResult{
			Path:    path,
			Output:  nil,
			Errors:  errors,
			Changed: false,
		}, nil
	}

	// Print dst.File with decorations preserved
	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, dstFile); err != nil {
		return nil, fmt.Errorf("failed to print dst file: %w", err)
	}

	// Format the output
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format source: %w", err)
	}

	return &FormatResult{
		Path:    path,
		Output:  formatted,
		Errors:  errors,
		Changed: true,
	}, nil
}

// isAllKeyed checks if all elements in the composite literal are keyed
func isAllKeyed(elts []dst.Expr) bool {
	if len(elts) == 0 {
		return true
	}

	for _, elt := range elts {
		if _, ok := elt.(*dst.KeyValueExpr); !ok {
			return false
		}
	}
	return true
}

// isExportedField checks if a field name is exported
func isExportedField(name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// generateZeroValue generates a zero value expression for the given type
func generateZeroValue(t types.Type, pkg *packages.Package) dst.Expr {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return &dst.Ident{Name: "false"}
		case types.String:
			return &dst.BasicLit{Kind: token.STRING, Value: `""`}
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
			types.Uintptr, types.Float32, types.Float64, types.Complex64, types.Complex128:
			return &dst.BasicLit{Kind: token.INT, Value: "0"}
		default:
			return &dst.Ident{Name: "nil"}
		}

	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return &dst.Ident{Name: "nil"}

	case *types.Struct:
		return &dst.CompositeLit{}

	case *types.Named:
		underlying := t.Underlying()
		// Check if the underlying type is an interface
		if _, ok := underlying.(*types.Interface); ok {
			return &dst.Ident{Name: "nil"}
		}
		// If underlying type is a basic type, return its zero value
		if basic, ok := underlying.(*types.Basic); ok {
			return generateZeroValue(basic, pkg)
		}
		// For named types with struct underlying, get the type name and create a composite literal
		typeName := t.Obj().Name()
		if pkgPath := t.Obj().Pkg(); pkgPath != nil && pkgPath.Path() != pkg.Types.Path() {
			// Need to qualify with package name
			return &dst.CompositeLit{
				Type: &dst.SelectorExpr{
					X:   &dst.Ident{Name: pkgPath.Name()},
					Sel: &dst.Ident{Name: typeName},
				},
			}
		}
		return &dst.CompositeLit{
			Type: &dst.Ident{Name: typeName},
		}

	case *types.Array:
		return &dst.CompositeLit{
			Type: &dst.ArrayType{
				Len: &dst.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Len())},
				Elt: typeToExpr(t.Elem()),
			},
		}

	default:
		return &dst.Ident{Name: "nil"}
	}
}

// typeToExpr converts a types.Type to a dst.Expr for use in array type expressions
func typeToExpr(t types.Type) dst.Expr {
	switch t := t.(type) {
	case *types.Basic:
		return &dst.Ident{Name: t.Name()}
	case *types.Named:
		return &dst.Ident{Name: t.Obj().Name()}
	case *types.Pointer:
		return &dst.StarExpr{X: typeToExpr(t.Elem())}
	case *types.Slice:
		return &dst.ArrayType{Elt: typeToExpr(t.Elem())}
	case *types.Array:
		return &dst.ArrayType{
			Len: &dst.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", t.Len())},
			Elt: typeToExpr(t.Elem()),
		}
	default:
		return &dst.Ident{Name: "interface{}"}
	}
}
