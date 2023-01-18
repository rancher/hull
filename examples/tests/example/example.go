package example

import (
	"github.com/aiyengar2/hull/pkg/utils"
)

var ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "example-chart")

type ValuesYaml struct {
	Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
}
