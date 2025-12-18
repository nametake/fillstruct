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
		filePath   string
		goldenFile string
		option     *Option
		want       *FormatResult
	}{
		{
			filePath:   "simple/input.go",
			goldenFile: "simple/golden.go",
			option:     &Option{},
			want: &FormatResult{
				Path:    addDirPrefix("simple.go"),
				Changed: true,
				Errors:  []*FormatError{},
			},
		},
	}

	for _, test := range tests {
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
			continue
		}

		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("Format(%q) returned unexpected result (-want +got):\n%s", test.filePath, diff)
		}
	}
}
