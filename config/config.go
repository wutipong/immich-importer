package config

import (
	"fmt"
)

type Config struct {
	ImmichURL    string `yaml:"immich_url"`
	ImmichAPIKey string `yaml:"immich_api_key"`
}

func LoadConfig(profile string, path string) (config Config, err error) {
	configMap, err := OpenConfigMap(path)
	if err != nil {
		err = fmt.Errorf("unable to open configuration file: %w", err)
		return
	}

	config, ok := configMap[profile]

	if !ok {
		err = fmt.Errorf("profile not found: %s", profile)
		return
	}

	return
}
