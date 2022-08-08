package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"gopkg.in/yaml.v3"
)

// FlagConverter is a cli to convert your old file to a new format
type FlagConverter struct {
	InputFile    string
	InputFormat  string
	OutputFile   string
	OutputFormat string
}

func (f *FlagConverter) Migrate() error {
	// Read content of the file
	content, err := ioutil.ReadFile(f.InputFile)
	if err != nil {
		return fmt.Errorf("file %v is impossible to find", f.InputFile)
	}

	flags, err := f.marshall(content)
	if err != nil {
		return err
	}

	newFileContent, err := f.unmarshall(f.convert(flags))
	if err != nil {
		return err
	}

	err = f.output(newFileContent)
	if err != nil {
		return err
	}

	return nil
}

func (f *FlagConverter) marshall(content []byte) (map[string]dto.DTO, error) {
	var flags map[string]dto.DTO
	switch strings.ToLower(f.InputFormat) {
	case "toml":
		err := toml.Unmarshal(content, &flags)
		if err != nil {
			return nil, err
		}
	case "json":
		err := json.Unmarshal(content, &flags)
		if err != nil {
			return nil, err
		}
	default:
		err := yaml.Unmarshal(content, &flags)
		if err != nil {
			return nil, err
		}
	}
	return flags, nil
}

func (f *FlagConverter) convert(flags map[string]dto.DTO) map[string]flag.InternalFlag {
	convertedFlags := make(map[string]flag.InternalFlag, len(flags))
	for k, v := range flags {
		convertedFlags[k] = v.Convert()
	}
	return convertedFlags
}

func (f *FlagConverter) unmarshall(convertedFlags map[string]flag.InternalFlag) ([]byte, error) {
	switch strings.ToLower(f.OutputFormat) {
	case "toml":
		return toml.Marshal(convertedFlags)
	case "json":
		return json.Marshal(convertedFlags)
	default:
		return yaml.Marshal(convertedFlags)
	}
}

func (f *FlagConverter) output(fileContent []byte) error {
	if f.OutputFile == "" {
		fmt.Println(string(fileContent))
		return nil
	}

	return ioutil.WriteFile(f.OutputFile, fileContent, 0o600)
}
