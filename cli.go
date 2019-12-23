package gobump

import (
	"flag"
	"fmt"
	"io"
)

// Run the gobump
func Run(argv []string, outStream, errStream io.Writer) error {
	gb := &Gobump{
		OutStream: outStream,
	}
	var checkVersionUp bool
	fs := flag.NewFlagSet("gobump", flag.ContinueOnError)
	fs.BoolVar(&gb.Write, "w", false, "write result to (source) file instead of stdout")
	fs.BoolVar(&gb.Verbose, "v", false, "show the resulting version values")
	fs.BoolVar(&gb.Raw, "r", false, "output in raw text instead of JSON when output exists")
	fs.BoolVar(&checkVersionUp, "u", false, "ensure version up with the set command")
	fs.Usage = func() {
		out := errStream
		fs.SetOutput(out)
		fmt.Fprintln(out, `Usage: gobump (major|minor|patch|set <version>|show) [-w] [-v] [-u] [<path>]

Commands:
  major             bump major version up
  minor             bump minor version up
  patch             bump patch version up
  set <version>     set exact version (no increments)
  show              only show the versions (implies -v)
Flags:`)
		fs.PrintDefaults()
	}

	if len(argv) < 1 {
		return fmt.Errorf("please specify subcommand. `gobump -h` for more details")
	}

	parseOffset := 1
	switch argv[0] {
	case "major":
		gb.Config.MajorDelta = 1
	case "minor":
		gb.Config.MinorDelta = 1
	case "patch":
		gb.Config.PatchDelta = 1
	case "set":
		if len(argv) < 2 {
			return fmt.Errorf("please specify a version to set")
		}
		gb.Config.Exact = argv[1]
		parseOffset = 2
	case "show":
		gb.Show = true
	case "-h", "-help", "--help":
		parseOffset = 0
	default:
		return fmt.Errorf("unknown subcommand %q. `gobump -h` for more details", argv[0])
	}
	if err := fs.Parse(argv[parseOffset:]); err != nil {
		return err
	}
	gb.Config.CheckVersionUp = checkVersionUp

	gb.Target = fs.Arg(0)
	_, err := gb.Run()
	return err
}
