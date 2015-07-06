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
			vers, err := conf.ProcessNode(fset, f)
			if err != nil {
				t.Errorf("got error: %s", err)
			}
			if _, ok := vers["version"]; !ok {
				t.Errorf("should detect `version`")
			}
			if _, ok := vers["VERSION"]; !ok {
				t.Errorf("should detect `VERSION`")
			}
			if vers["version"] != "1.1.0" {
				t.Errorf("expected %v: got %v", "1.1.0", vers["version"])
			}
			if vers["VERSION"] != "2.1.0" {
				t.Errorf("expected %v: got %v", "2.1.0", vers["VERSION"])
			}
		}
	}
}
