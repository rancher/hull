package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func MustGetPathFromModuleRoot(path ...string) string {
	modulePath, err := GetPathFromModuleRoot(path...)
	if err != nil {
		panic(err)
	}
	return modulePath
}

func GetPathFromModuleRoot(path ...string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for wd != string(filepath.Separator) {
		modFile := filepath.Join(wd, "go.mod")
		matches, err := filepath.Glob(modFile)
		if err != nil {
			return "", err
		}
		if matches != nil {
			return filepath.Join(wd, filepath.Join(path...)), nil
		}
		wd = filepath.Dir(wd)
	}
	return "", fmt.Errorf("path must exist within a go module")
}
