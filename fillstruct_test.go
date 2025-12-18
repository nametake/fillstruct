package fillstruct

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
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
				Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedFiles,
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
