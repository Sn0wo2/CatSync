package config

import (
	"encoding/json"
	"os"
)

type JSONLoader struct{}

func NewJSONLoader() *JSONLoader {
	return &JSONLoader{}
}
func (j *JSONLoader) Load(fileName string) (*Config, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, json.Unmarshal(file, &cfg)
}

func (j *JSONLoader) GetAllowFileExtensions() []string {
	return []string{"json"}
}
