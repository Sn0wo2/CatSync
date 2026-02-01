//go:build !catsync_all && !feature_config_yaml

package loader

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/config"
)

type YAMLLoader struct{}

func NewYAMLLoader() *YAMLLoader {
	return &YAMLLoader{}
}

func (y *YAMLLoader) GetTag() string {
	return "yaml"
}

func (y *YAMLLoader) Load(_ *config.Config, fileName string) error {
	return fmt.Errorf("yaml loader is not enabled: %s (rebuild with -tags catsync_all or -tags feature_config_yaml)", fileName)
}

func (y *YAMLLoader) Save(_ *config.Config, fileName string) error {
	return fmt.Errorf("yaml loader is not enabled: %s (rebuild with -tags catsync_all or -tags feature_config_yaml)", fileName)
}

func (y *YAMLLoader) GetAllowFileExtensions() []string {
	return []string{"yaml", "yml"}
}
