package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"go.yaml.in/yaml/v4"
)

func SetupConfig(profile string) error {
	configMap, err := OpenConfigMap()
	if err != nil {
		return err
	}

	config, ok := configMap[profile]

	if !ok {
		fmt.Printf("Profile does not exists. Creating new profile [%s].\n", profile)
		config = Config{}
	} else {
		fmt.Printf("Profile [%s] found. Updating existing profile.\n", profile)
	}

	urlPrompt := promptui.Prompt{
		Label: "Immich Server URL",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("server URL cannot be empty")
			}
			if u, err := url.Parse(s); err == nil {
				if u.Scheme != "http" && u.Scheme != "https" {
					return fmt.Errorf("invalid URL scheme: %s", u.Scheme)
				}
			} else {
				return fmt.Errorf("invalid server URL: %w", err)
			}
			return nil
		},
		Default:   config.ImmichURL,
		AllowEdit: true,
	}

	result, err := urlPrompt.Run()
	if err != nil {
		err = fmt.Errorf("prompt failed: %w", err)
		return err
	}
	config.ImmichURL = result

	apiKeyPrompt := promptui.Prompt{
		Label: "Immich API Key",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("API key cannot be empty")
			}
			return nil
		},
		Default:   config.ImmichAPIKey,
		AllowEdit: true,
	}

	result, err = apiKeyPrompt.Run()
	if err != nil {
		err = fmt.Errorf("prompt failed: %w", err)
		return err
	}
	config.ImmichAPIKey = result

	configMap[profile] = config

	err = SaveConfigMap(configMap)
	if err != nil {
		return err
	}

	return nil
}

func CreateConfigPath() (path string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error(
			"failed to get user home directory",
			slog.String("error", err.Error()),
		)
		return
	}
	path = filepath.Join(homeDir, ".immich-importer", "config.yaml")

	return
}

func OpenConfigMap() (configMap map[string]Config, err error) {
	path, err := CreateConfigPath()
	if err != nil {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if errors.Is(err, os.ErrNotExist) {
		slog.Warn(
			"Configuration file does not exist. Creating empty configuration map.",
			slog.String("path", path),
		)
		configMap = make(map[string]Config)
		return
	} else if err != nil {
		err = fmt.Errorf("failed to open configuration file: %w", err)
		return
	}
	defer file.Close()

	// Parse the YAML file into the Config struct
	configMap = make(map[string]Config)
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&configMap)
	if errors.Is(err, io.EOF) {
		slog.Warn(
			"Configuration file is empty. Creating empty configuration map.",
			slog.String("path", path),
		)
		configMap = make(map[string]Config)
		return
	} else if err != nil {
		err = fmt.Errorf("failed to parse existing configuration file: %w", err)
		return
	}
	return
}

func SaveConfigMap(configMap map[string]Config) error {
	path, err := CreateConfigPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Base(path), 0777)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to save configuration file: %w", err)
	}
	defer file.Close()

	yamlEncoder := yaml.NewEncoder(file)
	defer yamlEncoder.Close()

	err = yamlEncoder.Encode(configMap)
	if err != nil {
		return fmt.Errorf("failed to save configuration file: %w", err)
	}

	return nil
}
