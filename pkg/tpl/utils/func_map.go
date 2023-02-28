package utils

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func GetNoopHelmFuncMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	// Add extra functions from Helm
	extra := template.FuncMap{
		"toToml":        func(interface{}) string { return "" },
		"toYaml":        func(interface{}) string { return "" },
		"fromYaml":      func(string) map[string]interface{} { return nil },
		"fromYamlArray": func(string) []interface{} { return nil },
		"toJson":        func(interface{}) string { return "" },
		"fromJson":      func(string) map[string]interface{} { return nil },
		"fromJsonArray": func(string) []interface{} { return nil },
		"include":       func(string, interface{}) string { return "" },
		"tpl":           func(string, interface{}) interface{} { return "" },
		"required":      func(string, interface{}) (interface{}, error) { return "", nil },
		"lookup":        func(string, string, string, string) (map[string]interface{}, error) { return nil, nil },
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}
