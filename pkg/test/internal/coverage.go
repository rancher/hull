package internal

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
)

func CalculateCoverage(values map[string]interface{}, valuesStructType reflect.Type) (float64, string) {
	setKeys := getSetKeysFromMapInterface(values)
	allKeys := getAllKeysFromStructType(valuesStructType)

	numSetKeys := 0
	var setKeySlice []string
	var unsetKeySlice []string
	for k := range allKeys {
		if setKeys[k] {
			numSetKeys += 1
			setKeySlice = append(setKeySlice, k)
			continue
		}
		unsetKeySlice = append(unsetKeySlice, k)
	}

	sort.Strings(setKeySlice)
	sort.Strings(unsetKeySlice)

	var report string
	if len(unsetKeySlice) == 0 {
		report = fmt.Sprintf("All keys in struct are fully covered: %v", allKeys)
	} else {
		report = fmt.Sprintf("The following keys are not set: %v\nOnly the following keys are covered: %v", unsetKeySlice, setKeySlice)
	}

	return float64(numSetKeys) / float64(len(allKeys)), report
}

func getSetKeysFromMapInterface(values map[string]interface{}) map[string]bool {
	setKeys := make(map[string]bool)
	if values == nil {
		return setKeys
	}
	var collectSetKeys func(string, interface{})
	collectSetKeys = func(prefix string, valuesInterface interface{}) {
		switch val := valuesInterface.(type) {
		case map[string]interface{}:
			for k, v := range val {
				// if key represents struct key
				collectSetKeys(prefix+"."+k, v)
				if len(prefix) > 0 {
					// if key represents map key; cannot be at root
					collectSetKeys(prefix+"[]", v)
				}
			}
		case []interface{}:
			for _, v := range val {
				collectSetKeys(prefix+"[]", v)
			}
		default:
			for strings.HasSuffix(prefix, "[]") {
				prefix = strings.TrimSuffix(prefix, "[]")
			}
			if valuesInterface != nil && len(prefix) != 0 {
				setKeys[prefix] = true
			}
		}
	}
	collectSetKeys("", values)
	return setKeys
}

func getAllKeysFromStructType(valuesStructType reflect.Type) map[string]bool {
	allKeys := make(map[string]bool)
	if len(valuesStructType.Name()) > 0 {
		structName := valuesStructType.Name()
		// struct must be public to get any keys
		if string(structName[0]) == strings.ToLower(string(structName[0])) {
			return allKeys
		}
	}

	var collectAllKeys func(string, reflect.Type)
	collectAllKeys = func(prefix string, valuesType reflect.Type) {
		if valuesType.Kind() == reflect.Ptr {
			valuesType = valuesType.Elem()
		}
		switch valuesType.Kind() {
		case reflect.Struct:
			for i := 0; i < valuesType.NumField(); i++ {
				field := valuesType.Field(i)
				if string(field.Name[0]) == strings.ToLower(string(field.Name[0])) {
					// ignore unexported fields
					continue
				}
				fieldType := field.Type
				jsonFieldVal, ok := field.Tag.Lookup("json")
				var jsonFieldName string
				if ok && len(jsonFieldVal) > 0 {
					jsonFieldValSplit := strings.SplitN(jsonFieldVal, ",", 2) // ignore other comma-delimited args, like ',omitempty'
					jsonFieldName = jsonFieldValSplit[0]
				}
				if len(jsonFieldName) == 0 {
					if field.Anonymous && fieldType.Kind() == reflect.Struct {
						// special case for anonymous field; if there's no JSON tag, infer that the fields are recorded inline
						collectAllKeys(jsonFieldName, fieldType)
						continue
					}
					jsonFieldName = strcase.ToLowerCamel(field.Name)
				}
				collectAllKeys(prefix+"."+jsonFieldName, fieldType)
			}
		case reflect.Slice, reflect.Map:
			collectAllKeys(prefix+"[]", valuesType.Elem())
		default:
			for strings.HasSuffix(prefix, "[]") {
				prefix = strings.TrimSuffix(prefix, "[]")
			}
			if len(prefix) > 0 {
				allKeys[prefix] = true
			}
		}
	}
	collectAllKeys("", valuesStructType)
	return allKeys
}
