package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sn0wo2/CatSync/log"
	"gopkg.in/yaml.v3"
)

var CLIConfigPath string

const configJSONExtension = ".json"

func LoadConfig(configPath string) (*Config, string, error) {
	if configPath == "" {
		configPath = "./data/config.yml"

		if CLIConfigPath != "" {
			if _, err := os.Stat(CLIConfigPath); err == nil {
				configPath = CLIConfigPath

				goto LOAD
			}

			base := filepath.Dir(CLIConfigPath)
			for _, ext := range []string{".yml", ".yaml", configJSONExtension} {
				tryPath := base + ext
				if _, err := os.Stat(tryPath); err == nil {
					configPath = tryPath // 完成寻找, 开始LOAD

					goto LOAD
				}
			}

			configPath = CLIConfigPath // 找不到, 使用默认的开始LOAD吧

			goto LOAD
		}
	}

LOAD:
	fileCfg := &Config{}

	if err := loadFile(fileCfg, configPath); err != nil {
		return nil, configPath, fmt.Errorf("failed to load config file %s: %w", configPath, err)
	}

	fileCfg = ApplyDefaults(fileCfg, GetDefaultConfig())

	if err := fileCfg.Validate(); err != nil {
		return fileCfg, configPath, fmt.Errorf("validation failed for config file %s: %w", configPath, err)
	}

	log.Writef("Config loaded from file: %s", configPath)

	return fileCfg, configPath, nil
}

func loadFile(cfg *Config, fileName string) error {
	file, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return err
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case configJSONExtension:
		return json.Unmarshal(file, cfg)
	case ".yaml", ".yml":
		return yaml.Unmarshal(file, cfg)
	}

	var errs []error

	if err := yaml.Unmarshal(file, cfg); err == nil {
		return nil
	} else {
		errs = append(errs, fmt.Errorf("yaml: %w", err))
	}

	if err := json.Unmarshal(file, cfg); err == nil {
		return nil
	} else {
		errs = append(errs, fmt.Errorf("json: %w", err))
	}

	return errors.Join(errs...)
}

func SaveConfig(cfg *Config, fileName string) error {
	var (
		file []byte
		err  error
	)

	if strings.ToLower(filepath.Ext(fileName)) == configJSONExtension {
		file, err = json.Marshal(cfg)
	} else {
		file, err = yaml.Marshal(cfg)
	}

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fileName), 0o750); err != nil {
		return err
	}

	return os.WriteFile(fileName, file, 0o600)
}
