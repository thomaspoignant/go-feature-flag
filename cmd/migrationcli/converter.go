package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"

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
	content, err := os.ReadFile(f.InputFile)
	if err != nil {
		return fmt.Errorf("file %v is impossible to find", f.InputFile)
	}

	flags, err := f.unmarshall(content)
	if err != nil {
		return err
	}

	convertedFlag := f.convert(flags)
	newFileContent, err := f.marshall(convertedFlag)
	if err != nil {
		return err
	}

	err = f.output(newFileContent)
	if err != nil {
		return err
	}

	return nil
}

func (f *FlagConverter) unmarshall(content []byte) (map[string]dto.DTO, error) {
	var flags map[string]dto.DTO
	var err error
	switch strings.ToLower(f.InputFormat) {
	case "toml":
		err = toml.Unmarshal(content, &flags)
	case "json":
		err = json.Unmarshal(content, &flags)
	default:
		// default unmarshaller is YAML
		err = yaml.Unmarshal(content, &flags)
	}
	if err != nil {
		return nil, err
	}

	return flags, nil
}

func (f *FlagConverter) convert(flags map[string]dto.DTO) map[string]dto.DTO {
	convertedFlags := make(map[string]dto.DTO, len(flags))
	for k, v := range flags {
		// convert to internal Flag
		convertedFlags[k] = convertToDto(v.Convert())
	}
	return convertedFlags
}

func (f *FlagConverter) marshall(convertedFlags map[string]dto.DTO) ([]byte, error) {
	switch strings.ToLower(f.OutputFormat) {
	case "toml":
		buf := new(bytes.Buffer)
		_ = toml.NewEncoder(buf).Encode(convertedFlags)
		return buf.Bytes(), nil
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

	return os.WriteFile(f.OutputFile, fileContent, os.ModePerm)
}

func convertToDto(internalFlag flag.InternalFlag) dto.DTO {
	var experimentation *dto.ExperimentationDto
	if internalFlag.Experimentation != nil {
		experimentation = &dto.ExperimentationDto{
			Start: internalFlag.Experimentation.Start,
			End:   internalFlag.Experimentation.End,
		}
	}

	return dto.DTO{
		TrackEvents: internalFlag.TrackEvents,
		Disable:     internalFlag.Disable,
		Version:     internalFlag.Version,
		DTOv1: dto.DTOv1{
			Variations:      internalFlag.Variations,
			Rules:           internalFlag.Rules,
			DefaultRule:     internalFlag.DefaultRule,
			Scheduled:       internalFlag.Scheduled,
			Experimentation: experimentation,
		},
	}
}
