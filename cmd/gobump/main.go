package main

import (
	"flag"
	"fmt"
	"os"

	"go/parser"
	"go/printer"
	"go/token"

	"github.com/motemen/gobump"
)

func main() {
	var (
		write      = flag.Bool("w", false, "write result to (source) file instead of stdout")
		bumpMajor  = flag.Bool("major", false, "bump major version up")
		bumpMinor  = flag.Bool("minor", false, "bump minor version up")
		bumpPatch  = flag.Bool("patch", false, "bump patch version up")
		setVersion = flag.String("set", "", "set exact version (no bump)")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gobump (-major|-minor|-patch) [-w] [path]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		target = "."
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.Mode(0))
	dieIf(err)

	conf := gobump.Config{}
	if *setVersion != "" {
		conf.Exact = *setVersion
	} else if *bumpMajor {
		conf.MajorDelta = 1
	} else if *bumpMinor {
		conf.MinorDelta = 1
	} else if *bumpPatch {
		conf.PatchDelta = 1
	} else {
		flag.Usage()
	}

	found := false
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			names := conf.ProcessNode(f)
			if names != nil {
				out := os.Stdout
				if *write {
					file, err := os.Create(fset.File(f.Pos()).Name())
					dieIf(err)

					out = file
				}

				conf := &printer.Config{
					Mode:     printer.UseSpaces | printer.TabIndent,
					Tabwidth: 8,
				}
				conf.Fprint(out, fset, f)

				found = true
			}
		}
	}

	if found == false {
		os.Exit(1)
	}
}

func dieIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}