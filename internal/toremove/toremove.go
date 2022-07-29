package main

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"gopkg.in/yaml.v2"
)

func main() {
	t := map[string]dto.DTO{
		"test-flag": {
			DTOv0: dto.DTOv0{
				Disable:     nil,
				Rule:        testconvert.String("key eq \"random-key\""),
				Percentage:  testconvert.Float64(100),
				True:        testconvert.Interface(true),
				False:       testconvert.Interface(false),
				Default:     testconvert.Interface(false),
				TrackEvents: testconvert.Bool(false),
			},
		},
		"test-flag2": {
			DTOv0: dto.DTOv0{
				Disable:     nil,
				Rule:        testconvert.String("key eq \"random-key\""),
				Percentage:  testconvert.Float64(0),
				True:        testconvert.Interface("true"),
				False:       testconvert.Interface("false"),
				Default:     testconvert.Interface("false"),
				TrackEvents: testconvert.Bool(false),
			},
		},
	}

	for k, v := range t {
		fmt.Println("--- " + k)
		y, _ := yaml.Marshal(v.Convert())
		fmt.Println(string(y))

	}
}
