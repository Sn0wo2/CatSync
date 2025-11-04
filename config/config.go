package config

import (
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
