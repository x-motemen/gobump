package gobump

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func Run(argv []string) error {
	fs := flag.NewFlagSet("gobump", flag.ContinueOnError)
	var (
		write   = fs.Bool("w", false, "write result to (source) file instead of stdout")
		verbose = fs.Bool("v", false, "show the resulting version values")
		raw     = fs.Bool("r", false, "output in raw text instead of JSON when output exists")
	)
	fs.Usage = func() {
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
		fs.PrintDefaults()
	}

	if len(argv) < 1 {
		return fmt.Errorf("please specify subcommand. `gobump -h` for more details")
	}

	conf := Config{}
	var noWrite bool
	parseOffset := 1
	switch argv[0] {
	case "major":
		conf.MajorDelta = 1
	case "minor":
		conf.MinorDelta = 1
	case "patch":
		conf.PatchDelta = 1
	case "set":
		if len(argv) < 2 {
			return fmt.Errorf("please specify a version to set")
		}
		conf.Exact = argv[1]
		parseOffset = 2
	case "show":
		noWrite = true
		*verbose = true
	case "-h", "-help", "--help":
		parseOffset = 0
	default:
		return fmt.Errorf("unknown subcommand %q. `gobump -h` for more details", argv[0])
	}
	if err := fs.Parse(argv[parseOffset:]); err != nil {
		return err
	}

	target := fs.Arg(0)
	if target == "" {
		target = "."
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, target, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	found := false
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			vers, err := conf.ProcessNode(fset, f)
			if err != nil {
				return err
			}

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
					if err != nil {
						return err
					}
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
		return fmt.Errorf("version not found")
	}
	return nil
}
