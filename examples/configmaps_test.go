package examples

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/aiyengar2/hull/pkg/checker/resource"

	"github.com/stretchr/testify/assert"
)

type Configmaps struct {
	resource.ConfigMaps
}

func checkIfConfigMapsHaveData(data map[string]string) checker.CheckFunc {
	return func(t *testing.T, cms Configmaps) {
		for _, cm := range cms.ConfigMaps {
			assert.Equal(t, data, cm.Data)
		}
	}
}
