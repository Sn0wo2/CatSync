//go:build !catsync_all && !feature_config_json

package loader

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/config"
)

type JSONLoader struct{}

func NewJSONLoader() *JSONLoader {
	return &JSONLoader{}
}

func (j *JSONLoader) GetTag() string {
	return "json"
}

func (j *JSONLoader) Load(_ *config.Config, fileName string) error {
	return fmt.Errorf("json loader is not enabled: %s (rebuild with -tags catsync_all or -tags feature_config_json)", fileName)
}

func (j *JSONLoader) Save(_ *config.Config, fileName string) error {
	return fmt.Errorf("json loader is not enabled: %s (rebuild with -tags catsync_all or -tags feature_config_json)", fileName)
}

func (j *JSONLoader) GetAllowFileExtensions() []string {
	return []string{"json"}
}
