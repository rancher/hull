package utils

import (
	"path/filepath"
	"runtime"

	"github.com/go-git/go-billy/v5"
	"github.com/rancher/charts-build-scripts/pkg/filesystem"
)

func GetRepoRoot() string {
	// corresponds to the root of the rancher/hull repo always
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

func GetRepoFs() billy.Filesystem {
	return filesystem.GetFilesystem(GetRepoRoot())
}
