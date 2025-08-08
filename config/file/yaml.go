package file

import (
	"os"

	"github.com/Sn0wo2/CatSync/config"
	"gopkg.in/yaml.v3"
)

type YAMLLoader struct{}

func NewYAMLLoader() *YAMLLoader {
	return &YAMLLoader{}
}

func (y *YAMLLoader) Load(fileName string) (*config.Config, error) {
	file, err := os.ReadFile(fileName) //nolint:gosec
	if err != nil {
		return nil, err
	}

	var cfg config.Config

	return &cfg, yaml.Unmarshal(file, &cfg)
}

func (y *YAMLLoader) GetAllowFileExtensions() []string {
	return []string{"yaml", "yml"}
}
