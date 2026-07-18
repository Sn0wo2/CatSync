package cli

import (
	"fmt"
	"os"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/spf13/pflag"
)

type Options struct {
	ConfigPath string
	CheckOnly  bool
}

func Execute() *Options {
	args := preprocessArgs(os.Args[1:])

	fs := pflag.NewFlagSet("CatSync", pflag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "CatSync - Sync the cat config backend server\n\nUsage:\n  CatSync [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}

	configPath := fs.String("config", "", "Path to the configuration file")
	check := fs.BoolP("check", "c", false, "Validate configuration and exit")
	version := fs.BoolP("version", "v", false, "Print version information")
	help := fs.BoolP("help", "h", false, "help for CatSync")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *help {
		fs.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Println(GetCLIVersion())
		os.Exit(0)
	}

	if *configPath != "" {
		config.SetConfigPath(*configPath)
	}

	return &Options{
		ConfigPath: *configPath,
		CheckOnly:  *check,
	}
}

// preprocessArgs converts non-standard short flags before pflag parsing.
//
//	-cfg -> --config  (user-friendly shorthand)
func preprocessArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for i := range args {
		arg := args[i]
		if arg == "-cfg" {
			result = append(result, "--config")

			continue
		}

		if arg == "--cfg" { // normalize --cfg to --config
			result = append(result, "--config")

			continue
		}

		result = append(result, arg)
	}

	return result
}
