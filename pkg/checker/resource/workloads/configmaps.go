package workloads

import (
	"testing"

	"github.com/aiyengar2/hull/pkg/checker"
	"github.com/stretchr/testify/assert"
)

func EnsureNumConfigMaps(numConfigMaps int) checker.CheckFunc {
	return func(t *testing.T, cms struct{ ConfigMaps }) {
		assert.Equal(t, numConfigMaps, len(cms.ConfigMaps))
	}
}

func EnsureConfigMapsHaveData(data map[string]string) checker.CheckFunc {
	return func(t *testing.T, cms struct{ ConfigMaps }) {
		for _, cm := range cms.ConfigMaps {
			assert.Equal(t, data, cm.Data)
		}
	}
}
