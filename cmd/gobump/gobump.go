package main

import (
	"flag"
	"fmt"
	"github.com/blang/semver"
	"log"
	"os"
	"regexp"
	"strconv"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

var rxVersionName = regexp.MustCompile(`^(?i)version$`)

const defaultVersion = "0.0.0"

type Config struct {
	MajorDelta uint64
	MinorDelta uint64
	PatchDelta uint64
}

func main() {
	var (
		write     = flag.Bool("w", false, "write result to (source) file instead of stdout")
		bumpMajor = flag.Bool("major", false, "bump major version up")
		bumpMinor = flag.Bool("minor", false, "bump minor version up")
		bumpPatch = flag.Bool("patch", false, "bump patch version up")
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

	conf := Config{}
	if *bumpMajor {
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
			names := conf.bumpNode(f)
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

// bumpNode finds and bump ups "version" value inside given node.
// returns the rewrote identifiers inside node.
func (c Config) bumpNode(node ast.Node) (names []string) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.Package:
			return true

		case *ast.File:
			return true

		case *ast.GenDecl:
			return true

		case *ast.ValueSpec:
			for i, ident := range decl.Names {
				if !rxVersionName.MatchString(ident.Name) {
					continue
				}

				if decl.Values == nil {
					decl.Values = make([]ast.Expr, len(decl.Names))
				}

				currentVersion := defaultVersion

				if decl.Values[i] != nil {
					lit, ok := decl.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						log.Fatalf("expected string literal")
					}

					v, err := strconv.Unquote(lit.Value)
					if err != nil {
						log.Fatal("could not parse: %v", lit.Value)
					}

					currentVersion = v
				}

				ver, err := c.bumpedVersion(currentVersion)
				if err != nil {
					log.Fatalf("version bump failed: %s: %q", err, currentVersion)
				}

				decl.Values[i] = &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(ver),
				}

				if names == nil {
					names = []string{}
				}
				names = append(names, ident.Name)
			}
		}

		return false
	})

	return
}

// bumpedVersion returns new bumped-up version according to given spec.
func (c Config) bumpedVersion(version string) (string, error) {
	v, err := semver.Parse(version)
	if err != nil {
		return "", err
	}

	if c.MajorDelta > 0 {
		v.Major = v.Major + c.MajorDelta
		v.Minor = 0
		v.Patch = 0
	} else if c.MinorDelta > 0 {
		v.Minor = v.Minor + c.MinorDelta
		v.Patch = 0
	} else if c.PatchDelta > 0 {
		v.Patch = v.Patch + c.PatchDelta
	}

	return v.String(), nil
}

func dieIf(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
