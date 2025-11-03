package cli

import (
	"flag"
)

func Execute() {
	versionFlag := flag.Bool("version", false, "Print version information")
	vFlag := flag.Bool("v", false, "Print version information")

	flag.Parse()

	if version := *versionFlag || *vFlag; version {
		handleVersion()

		return
	}
}
