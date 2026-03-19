package cli

import (
	"flag"
	"os"
)

func Execute() {
	fs := flag.NewFlagSet("CatSync", flag.ContinueOnError)

	versionFlag := fs.Bool("version", false, "Print version information")
	vFlag := fs.Bool("v", false, "Print version information")

	_ = fs.Parse(os.Args[1:])

	if version := *versionFlag || *vFlag; version {
		handleVersion()

		return
	}
}
