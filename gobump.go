package gobump

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/blang/semver"
)

var defaultNamePattern = regexp.MustCompile(`^(?i)version$`)

// Config is the entrypoint of gobump.
type Config struct {
	// Increments major version. Precedes MinorDelta and PatchDelta.
	MajorDelta uint64
	// Increments minor version. Precedes PatchDelta.
	MinorDelta uint64
	// Increments patch version.
	PatchDelta uint64
	// Sets the version to exact version (no bump). Precedes all of above delta's.
	Exact string
	// The pattern of "version" variable/constants. Defaults to /^(?i)version$/.
	NamePattern *regexp.Regexp
	// Default version in the case none was set. Defaults to "0.0.0".
	Default string
}

// Process takes a Go source file and bumps version declaration according to conf.
// Returns the modified code and version identifier names and an error, if any.
func (conf Config) Process(filename string, src interface{}) ([]byte, []string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	names, err := conf.ProcessNode(fset, file)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	if err != nil {
		return nil, nil, err
	}

	out := buf.Bytes()

	out, err = format.Source(out)
	if err != nil {
		return nil, nil, err
	}

	return out, names, nil
}

// NodeErr represents for a ProcessNode error.
type NodeErr struct {
	Fset *token.FileSet
	Pos  token.Pos
	Msg  string
}

func (e NodeErr) Error() string {
	return e.Fset.Position(e.Pos).String() + ": " + e.Msg
}

// ProcessNode finds and bumps up "version" value found inside given node.
// returns the rewrote identifiers inside node.
func (conf Config) ProcessNode(fset *token.FileSet, node ast.Node) (names []string, nodeErr error) {
	namePattern := defaultNamePattern
	if conf.NamePattern != nil {
		namePattern = conf.NamePattern
	}

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
				if !namePattern.MatchString(ident.Name) {
					continue
				}

				if decl.Values == nil {
					decl.Values = make([]ast.Expr, len(decl.Names))
				}

				currentVersion := "0.0.0"
				if conf.Default != "" {
					currentVersion = conf.Default
				}

				if decl.Values[i] != nil {
					lit, ok := decl.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						nodeErr = NodeErr{
							Fset: fset,
							Pos:  decl.Values[i].Pos(),
							Msg:  "expected string literal",
						}
						return false
					}

					v, err := strconv.Unquote(lit.Value)
					if err != nil {
						nodeErr = NodeErr{
							Fset: fset,
							Pos:  lit.Pos(),
							Msg:  fmt.Sprintf("could not parse: %s", lit.Value),
						}
						return false
					}

					currentVersion = v
				}

				ver, err := conf.bumpedVersion(currentVersion)
				if err != nil {
					nodeErr = NodeErr{
						Fset: fset,
						Pos:  decl.Pos(),
						Msg:  fmt.Sprintf("version bump failed: %s: %q", err, currentVersion),
					}
					return false
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
func (conf Config) bumpedVersion(version string) (string, error) {
	if conf.Exact != "" {
		exact, err := semver.New(conf.Exact)
		if err != nil {
			return "", err
		}

		return exact.String(), nil
	}

	v, err := semver.Parse(version)
	if err != nil {
		return "", err
	}

	if conf.MajorDelta > 0 {
		v.Major = v.Major + conf.MajorDelta
		v.Minor = 0
		v.Patch = 0
	} else if conf.MinorDelta > 0 {
		v.Minor = v.Minor + conf.MinorDelta
		v.Patch = 0
	} else if conf.PatchDelta > 0 {
		v.Patch = v.Patch + conf.PatchDelta
	}

	return v.String(), nil
}
