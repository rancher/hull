package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGetPathFromModuleRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testPath := filepath.Join("hello", "world")
	expectedModulePath := filepath.Join(filepath.Dir(filepath.Dir(wd)), testPath)
	defer func() {
		err := recover()
		if err != nil {
			t.Error(err)
		}
	}()
	assert.Equal(t, expectedModulePath, MustGetPathFromModuleRoot(testPath))
}

func TestGetPathFromModuleRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testPath := filepath.Join("hello", "world")
	testCases := []struct {
		Name             string
		FromDir          string
		Expect           string
		ShouldThrowError bool
	}{
		{
			Name:    "Current",
			FromDir: ".",
			Expect:  filepath.Join(filepath.Dir(filepath.Dir(wd)), testPath),
		},
		{
			Name:             "Outside Module",
			FromDir:          filepath.Join("..", "..", ".."),
			ShouldThrowError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				if err := os.Chdir(wd); err != nil {
					t.Error(err)
					return
				}
			}()
			if err := os.Chdir(filepath.Join(wd, tc.FromDir)); err != nil {
				t.Error(err)
				return
			}
			modulePath, err := GetPathFromModuleRoot(testPath)
			if err != nil {
				assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
			}
			if err == nil {
				assert.False(t, tc.ShouldThrowError, "expected error to be thrown, found modulePath %s", modulePath)
			}
			if t.Failed() {
				return
			}
			assert.Equal(t, tc.Expect, modulePath, "did not find expected path")
		})
	}
}
