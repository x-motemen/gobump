package gobump

import (
	"testing"

	"go/parser"
	"go/token"
)

func TestBump(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata/test1", nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	conf := Config{
		MinorDelta: 1,
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			names := conf.ProcessNode(f)
			if len(names) != 2 || names[0] != "version" || names[1] != "VERSION" {
				t.Errorf("expected %v: %v", []string{"version", "VERSION"}, names)
			}
		}
	}
}
