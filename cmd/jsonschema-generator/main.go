package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/jessevdk/go-flags"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"log"
	"os"
)

func main() {
	var opts struct {
		SchemaLocation string `long:"schema-location" description:"Location where the schema will be saved." required:"true"` //nolint: lll
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("impossible to parse command line parameters", err)
	}

	if opts.SchemaLocation == "" {
		log.Fatal("schema-location is required")
	}

	a := map[string]dto.DTOv1{}
	d := jsonschema.Reflect(a)
	jsonSchema, err := d.MarshalJSON()
	if err != nil {
		log.Fatal("impossible to parse jsonschema", err)
	}

	prettyJSONSchema, err := PrettyString(string(jsonSchema))
	if err != nil {
		log.Fatal("impossible to prettify jsonschema", err)
	}

	fmt.Println("Writing jsonschema to", opts.SchemaLocation)
	err = os.WriteFile(opts.SchemaLocation, prettyJSONSchema, 0600)
	if err != nil {
		log.Fatalf("impossible to write jsonschema file to %s: %s", opts.SchemaLocation, err)
	}
}

// PrettyString will prettify the JSON string
func PrettyString(str string) ([]byte, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return []byte(""), err
	}
	return prettyJSON.Bytes(), nil
}
