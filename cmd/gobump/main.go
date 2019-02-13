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
	"flag"
	"log"
	"os"

	"github.com/motemen/gobump"
)

func main() {
	log.SetFlags(0)
	err := gobump.Run(os.Args[1:])
	if err != nil && err != flag.ErrHelp {
		log.Println(err)
		os.Exit(1)
	}
}
