/*
gobump bumps up program version by rewriting `version`-like variable/constant values in the Go source code.

Usage:
	gobump (major|minor|patch|set <version>) [-w] [-v] [<path>]

Commands:
	major             bump major version up
	minor             bump minor version up
	patch             bump patch version up
	set <version>     set exact version (no increments)
	show              only show the versions (implies -v)

Flags:
	  -v=false: show the resulting version values
	  -w=false: write result to (source) file instead of stdout
*/
package main

import (
	"encoding/json"
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
		write   = flag.Bool("w", false, "write result to (source) file instead of stdout")
		verbose = flag.Bool("v", false, "show the resulting version values")
		raw     = flag.Bool("r", false, "outputs in raw text instead of JSON when output exists")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: gobump (major|minor|patch|set <version>) [-w] [-v] [<path>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, `Commands:
  major             bump major version up
  minor             bump minor version up
  patch             bump patch version up
  set <version>     set exact version (no increments)
  show              only show the versions (implies -v)
`)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(2)
	}

	shift := func() string {
		if len(os.Args) < 2 {
			flag.Usage()
		}

		arg := os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)

		return arg
	}

	conf := gobump.Config{}

	var noWrite bool

	command := shift()
	switch command {
	case "major":
		conf.MajorDelta = 1
	case "minor":
		conf.MinorDelta = 1
	case "patch":
		conf.PatchDelta = 1
	case "set":
		conf.Exact = shift()
	case "show":
		noWrite = true
		*verbose = true
	default:
		flag.Usage()
	}

	flag.Parse()

	target := flag.Arg(0)
	if target == "" {
		target = "."
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, target, nil, parser.ParseComments)
	dieIf(err)

	found := false
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			vers, err := conf.ProcessNode(fset, f)
			dieIf(err)

			// rewrote successfully
			if vers != nil {
				found = true

				if *verbose {
					if *raw {
						for _, v := range vers {
							fmt.Println(v)
						}
					} else {
						json.NewEncoder(os.Stdout).Encode(vers)
					}
				}

				if noWrite {
					continue
				}

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
