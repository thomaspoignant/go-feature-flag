package configfile

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
	flags, err := ParseConfig(dat, configFormat)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	return flags, nil
}

// ParseConfig parses the content of a flag file in the given format (toml, json or yaml).
// Any format other than toml or json is parsed as YAML.
func ParseConfig(data []byte, configFormat string) (map[string]dto.DTO, error) {
	var flags map[string]dto.DTO
	switch strings.ToLower(configFormat) {
	case "toml":
		if err := toml.Unmarshal(data, &flags); err != nil {
			return nil, fmt.Errorf("could not parse file (toml): %w", err)
		}
	case "json":
		if err := json.Unmarshal(data, &flags); err != nil {
			return nil, fmt.Errorf("could not parse file (json): %w", err)
		}
	default:
		// default is YAML
		if err := yaml.Unmarshal(data, &flags); err != nil {
			return nil, fmt.Errorf("could not parse file (yaml): %w", err)
		}
	}
	return flags, nil
}

// FlagError is the validation error of a single flag.
type FlagError struct {
	Flag string
	Err  error
}

// Error implements the error interface.
func (e FlagError) Error() string {
	return fmt.Sprintf("invalid flag %s: %v", e.Flag, e.Err)
}

// Unwrap returns the underlying validation error.
func (e FlagError) Unwrap() error { return e.Err }

// ValidateFlags converts and validates every flag, it returns the validation errors sorted by flag key.
func ValidateFlags(flags map[string]dto.DTO) []FlagError {
	keys := make([]string, 0, len(flags))
	for key := range flags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	errs := make([]FlagError, 0)
	for _, key := range keys {
		if err := validateFlag(flags[key]); err != nil {
			errs = append(errs, FlagError{Flag: key, Err: err})
		}
	}
	return errs
}

// validateFlag converts and validates a single flag.
// A flag definition that makes the validation panic is reported as an invalid flag.
func validateFlag(flagDto dto.DTO) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid flag definition: %v", r)
		}
	}()
	flag := flagDto.Convert()
	return flag.IsValid()
}
