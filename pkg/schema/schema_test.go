package schema

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/utils"
)

type ValuesYaml struct {
	Data map[string]interface{} `jsonschema:"description=Data to be inserted into a ConfigMap"`
}

func TestMustProduceJSONSchemas(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Error(err)
		}
	}()
	chartPath := utils.MustGetPathFromModuleRoot("testdata", "charts", "with-schema")
	MustProduceJSONSchemas([]JSONSchemaGenerateArg{
		{
			ValuesStruct: ValuesYaml{},
			ChartPath:    chartPath,
		},
	})
}
