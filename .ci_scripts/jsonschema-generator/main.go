package main

import (
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
)

func main() {
	fmt.Print("yo")

	a := dto.DTOv1{}
	d := jsonschema.Reflect(a)
	c, err := d.MarshalJSON()
	fmt.Println(string(c), err)
}
