package config

import "github.com/Sn0wo2/CatSync/internal/util"

func (c *Config) validate() error {
	return util.Validate(c)
}

func (c *Config) merge(src *Config) {
	util.Merge(c, src)
}
