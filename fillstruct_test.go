package fillstruct

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/tools/go/packages"
)

func TestFormat(t *testing.T) {
	// for cloud.google.com/go/spanner module
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir("testdata"); err != nil {
		t.Fatalf("failed to change directory to testdata: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatalf("failed to change directory to %q: %v", currentDir, err)
		}
	})

	testdataDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	addDirPrefix := func(s string) string {
		return fmt.Sprintf("%s/%s", testdataDir, s)
	}

	tests := []struct {
		name       string
		filePath   string
		goldenFile string
		option     *Option
		want       *FormatResult
	}{
		{
			name:       "single missing field is filled with zero value",
			filePath:   "simple/input.go",
			goldenFile: "simple/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("simple/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "multiple target types are specified, missing fields are added to each type",
			filePath:   "multiple_types/input.go",
			goldenFile: "multiple_types/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("multiple_types/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "pointer type is handled correctly",
			filePath:   "pointer/input.go",
			goldenFile: "pointer/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("pointer/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "nested struct field is filled with empty composite literal",
			filePath:   "nested_struct/input.go",
			goldenFile: "nested_struct/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("nested_struct/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "unexported field is not added",
			filePath:   "unexported_field/input.go",
			goldenFile: "unexported_field/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("unexported_field/input.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "position-based literal is skipped",
			filePath:   "position_based/input.go",
			goldenFile: "position_based/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("position_based/input.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "string pointer field is filled with nil",
			filePath:   "string_pointer/input.go",
			goldenFile: "string_pointer/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("string_pointer/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "test file is handled correctly",
			filePath:   "test_file/input.go",
			goldenFile: "test_file/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("test_file/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "defined types are handled correctly",
			filePath:   "defined_types/input.go",
			goldenFile: "defined_types/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("defined_types/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "external package type is handled correctly",
			filePath:   "external_package/input.go",
			goldenFile: "external_package/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("external_package/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "all fields are specified, no changes",
			filePath:   "complete/input.go",
			goldenFile: "complete/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("complete/input.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "custom default values are used for enum types",
			filePath:   "custom_default/input.go",
			goldenFile: "custom_default/golden.go",
			option: &Option{
				// Note: In test environment, package path is "command-line-arguments" because
				// we load files directly. In real usage, it would be the actual import path
				// like "github.com/example/domain.Status".
				CustomDefaults: map[string]string{
					"command-line-arguments.Status": "StatusUnknown",
					"command-line-arguments.Role":   "RoleGuest",
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("custom_default/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "custom default and zero values are mixed correctly",
			filePath:   "custom_default_mixed/input.go",
			goldenFile: "custom_default_mixed/golden.go",
			option: &Option{
				// Note: In test environment, package path is "command-line-arguments".
				// In real usage, it would be the actual import path.
				CustomDefaults: map[string]string{
					"command-line-arguments.Status": "StatusUnknown",
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("custom_default_mixed/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "custom default values are used for basic types",
			filePath:   "custom_default_basic/input.go",
			goldenFile: "custom_default_basic/golden.go",
			option: &Option{
				// Note: Basic types like "int", "string", "bool" can also have custom defaults.
				// All fields of the same basic type will use the same default value.
				CustomDefaults: map[string]string{
					"int":  "8080",
					"bool": "true",
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("custom_default_basic/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "external package enum with fully qualified constant name",
			filePath:   "external_enum/input.go",
			goldenFile: "external_enum/golden.go",
			option: &Option{
				// Note: For external package types, the constant name must include
				// the package qualifier (e.g., "otherpkg.StatusUnknown").
				CustomDefaults: map[string]string{
					"github.com/nametake/fillstruct/testdata/external_enum/otherpkg.Status": "otherpkg.StatusUnknown",
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("external_enum/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "target fields are not specified, all missing fields are filled",
			filePath:   "target_fields_unspecified/input.go",
			goldenFile: "target_fields_unspecified/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_unspecified/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "target fields entry exists only for another type, type without entry is fully filled and type with entry fills only listed fields",
			filePath:   "target_fields_other_type/input.go",
			goldenFile: "target_fields_other_type/golden.go",
			option: &Option{
				// Note: In test environment, package path is "command-line-arguments".
				// In real usage, it would be the actual import path.
				TargetFields: map[string][]string{
					"command-line-arguments.Config": {"Port"},
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_other_type/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "one target field is listed and missing, only the listed field is filled",
			filePath:   "target_fields_single/input.go",
			goldenFile: "target_fields_single/golden.go",
			option: &Option{
				TargetFields: map[string][]string{
					"command-line-arguments.User": {"Age"},
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_single/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "multiple target fields are listed and missing, all listed fields are filled and unlisted fields are not",
			filePath:   "target_fields_multiple/input.go",
			goldenFile: "target_fields_multiple/golden.go",
			option: &Option{
				TargetFields: map[string][]string{
					"command-line-arguments.User": {"Age", "Email"},
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_multiple/input.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "listed target field is already present, no changes are made",
			filePath:   "target_fields_already_present/input.go",
			goldenFile: "target_fields_already_present/golden.go",
			option: &Option{
				TargetFields: map[string][]string{
					"command-line-arguments.User": {"Name"},
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_already_present/input.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
		{
			name:       "target fields entry is empty, no fields are filled",
			filePath:   "target_fields_empty/input.go",
			goldenFile: "target_fields_empty/golden.go",
			option: &Option{
				TargetFields: map[string][]string{
					"command-line-arguments.User": {},
				},
			},
			want: &FormatResult{
				Path:    addDirPrefix("target_fields_empty/input.go"),
				Changed: false,
				Errors:  []*FormatError{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want.Changed {
				golden, err := os.ReadFile(test.goldenFile)
				if err != nil {
					t.Errorf("failed to read golden file %q: %v", test.goldenFile, err)
				}
				test.want.Output = golden
			}

			cfg := &packages.Config{
				Mode:  packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles,
				Tests: true,
			}
			pkgs, err := packages.Load(cfg, test.filePath)
			if err != nil {
				t.Errorf("failed to load packages: path = %s: %v", test.filePath, err)
			}
			if len(pkgs) != 1 {
				t.Errorf("expected exactly one package: %s", test.filePath)
			}

			pkg := pkgs[0]

			if len(pkg.Syntax) != 1 {
				t.Errorf("expected exactly one file: %s", test.filePath)
			}

			file := pkg.Syntax[0]

			got, err := Format(pkg, file, test.option)
			if err != nil {
				t.Errorf("Format(%q) returned unexpected error: %v", test.filePath, err)
				return
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Format(%q) returned unexpected result (-want +got):\n%s", test.filePath, diff)
			}
		})
	}
}

func TestParseTypeSpec(t *testing.T) {
	type args struct {
		spec string
	}
	type want struct {
		importPath string
		typeName   string
		fieldName  string
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "importpath.TypeName format, import path and type name are returned",
			args: args{spec: "github.com/example/domain.Foo"},
			want: want{
				importPath: "github.com/example/domain",
				typeName:   "Foo",
				fieldName:  "",
				err:        nil,
			},
		},
		{
			name: "importpath.TypeName.FieldName format, field name is also returned",
			args: args{spec: "github.com/example/domain.Foo.Name"},
			want: want{
				importPath: "github.com/example/domain",
				typeName:   "Foo",
				fieldName:  "Name",
				err:        nil,
			},
		},
		{
			name: "import path without slash, package name and type name are returned",
			args: args{spec: "domain.Foo"},
			want: want{
				importPath: "domain",
				typeName:   "Foo",
				fieldName:  "",
				err:        nil,
			},
		},
		{
			name: "type name only without import path, an error is returned",
			args: args{spec: "Foo"},
			want: want{
				importPath: "",
				typeName:   "",
				fieldName:  "",
				err:        cmpopts.AnyError,
			},
		},
		{
			name: "more than three dot-separated elements after the last slash, an error is returned",
			args: args{spec: "github.com/example/domain.Foo.Name.Extra"},
			want: want{
				importPath: "",
				typeName:   "",
				fieldName:  "",
				err:        cmpopts.AnyError,
			},
		},
		{
			name: "empty element between dots, an error is returned",
			args: args{spec: "github.com/example/domain..Foo"},
			want: want{
				importPath: "",
				typeName:   "",
				fieldName:  "",
				err:        cmpopts.AnyError,
			},
		},
	}

	cmpOpts := []cmp.Option{cmp.AllowUnexported(want{}), cmpopts.EquateErrors()}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			importPath, typeName, fieldName, err := parseTypeSpec(test.args.spec)
			got := want{
				importPath: importPath,
				typeName:   typeName,
				fieldName:  fieldName,
				err:        err,
			}
			if diff := cmp.Diff(test.want, got, cmpOpts...); diff != "" {
				t.Errorf("parseTypeSpec(%q) returned unexpected result (-want +got):\n%s", test.args.spec, diff)
			}
		})
	}
}
