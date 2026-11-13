package configfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
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
	flags, err := ParseConfig(dat, configFormat)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	return flags, nil
}

// ParseConfig parses the content of a flag file in the given format (toml, json or yaml).
// Any format other than toml or json is parsed as YAML.
// It uses the same parser as the retrievers.
func ParseConfig(data []byte, configFormat string) (map[string]dto.DTO, error) {
	flags, err := cache.ConvertToFlagStruct(data, configFormat)
	if err != nil {
		format := strings.ToLower(configFormat)
		if format != "toml" && format != "json" {
			format = "yaml"
		}
		return nil, fmt.Errorf("could not parse file (%s): %w", format, err)
	}
	return flags, nil
}
