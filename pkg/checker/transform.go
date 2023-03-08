package checker

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v2"
)

func ToYAML[O interface{}](obj O) string {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func FromYAML[O interface{}](str string) O {
	obj := new(O)
	err := yaml.Unmarshal([]byte(str), obj)
	if err != nil {
		return *new(O)
	}
	return *obj
}

func ToJSON[O interface{}](obj O) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(data)
}

func FromJSON[O interface{}](str string) O {
	obj := new(O)
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		return *new(O)
	}
	return *obj
}
