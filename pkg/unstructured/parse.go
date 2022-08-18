package unstructured

import (
	"fmt"
	"regexp"

	"github.com/lrills/helm-unittest/unittest/valueutils"
)

var (
	pathRegex = regexp.MustCompile(`^([^\.^\[^\]]+(\[(([1-9]+[\d]*)|_)\])*\.)*[^\.^\[^\]]+(\[(([1-9]+[\d]*)|_)\])*`)
)

func isValidPath(path string) bool {
	if len(path) == 0 {
		return true
	}
	for _, match := range pathRegex.FindAllString(path, -1) {
		if match == path {
			return true
		}
	}
	return false
}

func Parse(path string, unstructured Unstructured) (interface{}, error) {
	if !isValidPath(path) {
		return nil, fmt.Errorf("%s is an invalid path", path)
	}
	return valueutils.GetValueOfSetPath(unstructured.Object, path)
}
