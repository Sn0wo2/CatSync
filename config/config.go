package config

import (
	"github.com/Sn0wo2/CatSync/internal/util"
)

func (c *Config) Validate() error {
	return util.Validate(c)
}

func (c *Config) Merge(src *Config) {
	util.Merge(c, src)
}
