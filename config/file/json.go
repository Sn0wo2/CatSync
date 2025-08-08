package file

import (
	"encoding/json"
	"os"

	"github.com/Sn0wo2/CatSync/config"
)

type JSONLoader struct{}

func NewJSONLoader() *JSONLoader {
	return &JSONLoader{}
}

func (j *JSONLoader) Load(fileName string) (*config.Config, error) {
	file, err := os.ReadFile(fileName) //nolint:gosec
	if err != nil {
		return nil, err
	}

	var cfg config.Config

	return &cfg, json.Unmarshal(file, &cfg)
}

func (j *JSONLoader) GetAllowFileExtensions() []string {
	return []string{"json"}
}
