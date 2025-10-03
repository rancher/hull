package checker

import (
	"fmt"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/rancher/wrangler/v3/pkg/relatedresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Key[O metav1.Object](obj O) relatedresource.Key {
	return relatedresource.NewKey(obj.GetNamespace(), obj.GetName())
}

func Select[O metav1.Object](name string, namespace string, obj O) bool {
	return obj.GetName() == name && obj.GetNamespace() == namespace
}

func hasEntries[O metav1.Object](obj O, expected, actual map[string]string, match bool) error {
	var err error
	for k, v := range expected {
		val, ok := actual[k]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("%T %s is missing label %s", obj, Key(obj), k))
			continue
		}
		if val != v {
			err = multierr.Append(err, fmt.Errorf("%T %s is has wrong value for label %s, expected %s found %s", obj, Key(obj), k, v, val))
			continue
		}
	}
	if !match {
		return err
	}
	for k := range actual {
		_, ok := expected[k]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("%T %s has additional label %s", obj, Key(obj), k))
			continue
		}
	}
	return err
}

func HasLabels[O metav1.Object](obj O, labels map[string]string) error {
	return hasEntries(obj, labels, obj.GetLabels(), false)
}

func HasAnnotations[O metav1.Object](obj O, labels map[string]string) error {
	return hasEntries(obj, labels, obj.GetAnnotations(), false)
}

func MatchLabels[O metav1.Object](obj O, labels map[string]string) error {
	return hasEntries(obj, labels, obj.GetLabels(), true)
}

func MatchAnnotations[O metav1.Object](obj O, labels map[string]string) error {
	return hasEntries(obj, labels, obj.GetAnnotations(), true)
}
