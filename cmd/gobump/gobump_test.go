package main

import (
	"os"
	"testing"

	"go/parser"
	"go/printer"
	"go/token"
)

func TestBump(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../../testdata/test1", nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	conf := Config{
		MinorDelta: 1,
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			idents := conf.bumpNode(f)
			if idents != nil {
				printer.Fprint(os.Stdout, fset, f)
			}
		}
	}
}
