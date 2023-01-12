//go:generate go run ./codegen.go

package main

import (
	"github.com/aiyengar2/hull/examples/tests/example"
	"github.com/aiyengar2/hull/pkg/schema"
	"github.com/sirupsen/logrus"
)

var (
	schemas = []schema.JSONSchemaGenerateArg{
		{
			ValuesStruct: example.ValuesYaml{},
			ChartPath:    example.ChartPath,
		},
	}
)

func main() {
	logrus.Infof("Generating JSON Schemas...")
	schema.MustProduceJSONSchemas(schemas)
}
