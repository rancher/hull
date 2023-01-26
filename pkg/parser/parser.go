package parser

import (
	"bytes"
	"io"
	"strings"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Parse parses the runtime.Objects tracked in a Kubernetes manifest (represented as a string) into an ObjectSet
func Parse(manifest string) (*objectset.ObjectSet, error) {
	var multiErr error
	var u unstructured.Unstructured

	decoder := yaml.NewYAMLOrJSONDecoder(manifestReader(manifest), 1000)
	os := objectset.NewObjectSet()
	for {
		uCopy := u.DeepCopy()
		err := decoder.Decode(uCopy)
		if err != nil {
			if err == io.EOF {
				break
			}
			if strings.HasPrefix(err.Error(), "error unmarshaling JSON: while decoding JSON: Object 'Kind' is missing in ") {
				// not a valid kubernetes object, but some valid JSON
				continue
			}
			multiErr = multierr.Append(err, err)
			continue
		}
		if uCopy.GetAPIVersion() == "" || uCopy.GetKind() == "" {
			// Encountered empty YAML document but successfully decoded, skip
			continue
		}
		os = os.Add(uCopy)
		logrus.Debugf("obj: %s, Kind=%s (%s/%s)", uCopy.GetAPIVersion(), uCopy.GetKind(), uCopy.GetName(), uCopy.GetNamespace())
	}
	if multiErr != nil {
		return nil, multiErr
	}
	return os, nil
}

func manifestReader(manifest string) io.Reader {
	standardizedManifest := strings.ReplaceAll(manifest, "\t", "    ")
	return bytes.NewReader([]byte(standardizedManifest))
}
