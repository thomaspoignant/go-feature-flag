package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"gopkg.in/yaml.v3"
)

var ConfigFileDefaultLocations = []string{
	"./",
	"/goff/",
	"/etc/opt/goff/",
}

func LoadConfigFile(
	inputFilePath string,
	configFormat string,
	defaultLocations []string,
) (map[string]dto.DTO, error) {
	filename := "flags.goff"
	if defaultLocations == nil {
		defaultLocations = ConfigFileDefaultLocations
	}
	supportedExtensions := []string{
		"yaml",
		"toml",
		"json",
		"yml",
	}

	if inputFilePath != "" {
		if _, err := os.Stat(inputFilePath); err != nil {
			return nil, fmt.Errorf("impossible to find config file %s", inputFilePath)
		}
		return readConfigFile(inputFilePath, configFormat)
	}
	for _, location := range defaultLocations {
		for _, ext := range supportedExtensions {
			configFile := fmt.Sprintf("%s%s.%s", location, filename, ext)
			if _, err := os.Stat(configFile); err != nil {
				continue
			}
			return readConfigFile(configFile, ext)
		}
	}
	return nil, fmt.Errorf(
		"impossible to find config file in the default locations [%s]",
		strings.Join(defaultLocations, ","),
	)
}

func readConfigFile(configFile, configFormat string) (map[string]dto.DTO, error) {
	dat, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var flags map[string]dto.DTO
	switch strings.ToLower(configFormat) {
	case "toml":
		err := toml.Unmarshal(dat, &flags)
		if err != nil {
			return nil, fmt.Errorf("%s: could not parse file (toml): %w", configFile, err)
		}
		return flags, nil
	case "json":
		err := json.Unmarshal(dat, &flags)
		if err != nil {
			return nil, fmt.Errorf("%s: could not parse file (json): %w", configFile, err)
		}
		return flags, nil
	default:
		// default is YAML
		err := yaml.Unmarshal(dat, &flags)
		if err != nil {
			return nil, fmt.Errorf("%s: could not parse file (yaml): %w", configFile, err)
		}
		return flags, nil
	}
}
