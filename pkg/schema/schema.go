package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
)

func FromStructToSchema(schemaStruct interface{}) *jsonschema.Schema {
	r := &jsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
		Namer: func(t reflect.Type) string {
			return strcase.ToLowerCamel(t.Name())
		},
		KeyNamer: strcase.ToLowerCamel,
	}
	schema := r.Reflect(schemaStruct)
	return schema
}

func FromStructToSchemaBytes(schemaStruct interface{}) ([]byte, error) {
	schema := FromStructToSchema(schemaStruct)
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(schemaBytes, '\n'), nil
}

func FromStructToValuesJSONSchema(schemaStruct interface{}, chartPath string) error {
	schemaBytes, err := FromStructToSchemaBytes(schemaStruct)
	if err != nil {
		return err
	}
	path := filepath.Join(chartPath, "values.schema.json")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(schemaBytes)
	if err != nil {
		return err
	}
	logrus.Infof("Added or updated values.schema.json for %s based on %T", chartPath, schemaStruct)
	return nil
}

type JSONSchemaGenerateArg struct {
	ValuesStruct interface{}
	ChartPath    string
}

func MustProduceJSONSchemas(args []JSONSchemaGenerateArg) {
	for _, arg := range args {
		if err := FromStructToValuesJSONSchema(arg.ValuesStruct, arg.ChartPath); err != nil {
			panic(err)
		}
	}
}
