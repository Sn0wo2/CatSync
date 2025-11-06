package config

import (
	"fmt"
	"os"

	"github.com/Sn0wo2/CatSync/internal/util"
)

func init() {
	util.DefaultConfigProvider = func() (any, bool) {
		return DefaultConfig, true
	}
}

func (c *Config) Validate(tag string) error {
	return util.Validate(c, tag)
}

func (c *Config) Merge(src *Config) {
	util.Merge(src, c)
}

func (c *Config) Check() {
	for i, a := range c.Actions {
		if a.Operation == "" {
			if a.Action == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Config > Action 'operation' field is empty!\n"+
					" - route: [%s]\n"+
					" - index: [%d]\n", a.Route, i)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Config > Action 'action' field is deprecated and will be removed in v1.11.0. Please use the 'operation' field instead.\n"+
					" - route: [%s]\n"+
					" - index: [%d]\n", a.Route, i)
			}

			a.Operation = a.Action.ToOperation()
		}
	}
}
