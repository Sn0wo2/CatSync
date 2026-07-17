package cli

import (
	"errors"
	"flag"
	"os"
)

func Execute() {
	fs := flag.NewFlagSet("CatSync", flag.ContinueOnError)

	versionFlag := fs.Bool("version", false, "Print version information")
	vFlag := fs.Bool("v", false, "Print version information")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}

		os.Exit(2)
	}

	if *versionFlag || *vFlag {
		handleVersion()

		return
	}
}
