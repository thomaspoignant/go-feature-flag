package helper

import (
	"fmt"
	"reflect"

	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest/model"
)

func FlagTypeFromVariations(variations map[string]*interface{}) (model.FlagType, error) {
	if variations == nil {
		return "", fmt.Errorf("impossible to find type, no variations found")
	}
	variationTypes := make(map[model.FlagType]interface{}, len(variations))
	for _, val := range variations {
		if val == nil {
			// we skip if value is nil
			continue
		}
		vv := *val
		switch vv.(type) {
		case bool:
			variationTypes[model.FlagTypeBoolean] = interface{}(nil)
		case string:
			variationTypes[model.FlagTypeString] = interface{}(nil)
		case int:
			variationTypes[model.FlagTypeInteger] = interface{}(nil)
		case float64:
			variationTypes[model.FlagTypeFloat] = interface{}(nil)
		case map[string]interface{}:
			variationTypes[model.FlagTypeObject] = interface{}(nil)
		default:
			// do nothing here
			continue
		}
	}

	// we found the type and return it
	if len(variationTypes) == 1 {
		for key := range variationTypes {
			return key, nil
		}
	}
	_, okFloat := variationTypes[model.FlagTypeFloat]
	_, okInteger := variationTypes[model.FlagTypeInteger]
	if len(variationTypes) == 2 && okInteger && okFloat {
		// we need to check if it is a float or an integer
		for _, v := range variations {
			if v == nil {
				// we skip if value is nil
				continue
			}
			numberType, err := numberType(*v)
			if err != nil {
				return "", err
			}
			if numberType == "float" {
				return model.FlagTypeFloat, nil
			}
		}
		return model.FlagTypeInteger, nil
	}
	return "", fmt.Errorf("impossible to find type")
}

func numberType(value interface{}) (string, error) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer", nil
	case reflect.Float32, reflect.Float64:
		// Check if the float has a whole number value.
		floatVal := val.Float()
		if floatVal == float64(int64(floatVal)) {
			return "integer", nil
		}
		return "float", nil
	default:
		return "", fmt.Errorf("unknown type %v", val.Kind())
	}
}
