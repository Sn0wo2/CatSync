package loader

import (
	"os"
	"path/filepath"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/internal/util"
	"gopkg.in/yaml.v3"
)

type YAMLLoader struct{}

func NewYAMLLoader() *YAMLLoader {
	return &YAMLLoader{}
}

func (y *YAMLLoader) GetTag() string {
	return "yaml"
}

func (y *YAMLLoader) Load(cfg *config.Config, fileName string) error {
	file, err := os.ReadFile(fileName) //nolint:gosec
	if err != nil {
		return err
	}

	node := &yaml.Node{}
	if err := yaml.Unmarshal(file, node); err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, cfg); err != nil {
		return err
	}

	util.MergeYamlNode(node, cfg)

	cd, err := yaml.Marshal(node)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fileName), 0o750); err != nil {
		return err
	}

	return os.WriteFile(fileName, cd, 0o600)
}

func (y *YAMLLoader) Save(cfg *config.Config, fileName string) error {
	file, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fileName), 0o750); err != nil {
		return err
	}

	return os.WriteFile(fileName, file, 0o600)
}

func (y *YAMLLoader) GetAllowFileExtensions() []string {
	return []string{"yaml", "yml"}
}
