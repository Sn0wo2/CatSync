package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/common/version"
)

type Options struct {
	ConfigPath string
	CheckOnly  bool
}

func Execute() *Options {
	flags := flag.NewFlagSet("CatSync", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(os.Stdout, "Usage: CatSync [options]")

		flags.PrintDefaults()
	}

	opts := &Options{}

	flags.StringVar(&opts.ConfigPath, "config", "", "Path to the configuration file")
	flags.BoolVar(&opts.CheckOnly, "check", false, "Validate configuration and exit")
	flags.BoolVar(&opts.CheckOnly, "c", false, "Validate configuration and exit")
	showVersion := flags.Bool("version", false, "Print version and exit")

	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		os.Exit(1)
	}

	if *showVersion {
		_, _ = fmt.Fprintln(os.Stdout, version.Print("CatSync"))

		return nil
	}

	return opts
}
