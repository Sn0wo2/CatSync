package cli

import (
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	ConfigPath string
	CheckOnly  bool
}

func Execute() *Options {
	var opts Options
	ran := false

	root := &cobra.Command{
		Use:     "CatSync",
		Short:   "Sync the cat config backend server",
		Version: GetCLIVersion(),
		Run: func(cmd *cobra.Command, _ []string) {
			ran = true
			opts.ConfigPath, _ = cmd.Flags().GetString("config")
			opts.CheckOnly, _ = cmd.Flags().GetBool("check")
		},
	}
	root.SetVersionTemplate("{{.Version}}\n")

	flags := root.Flags()

	flags.StringP("config", "cfg", "", "Path to the configuration file")
	flags.BoolP("check", "c", false, "Validate configuration and exit")
	_ = root.MarkFlagFilename("config", "yaml", "yml", "json")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	if !ran {
		return nil
	}

	return &opts
}
