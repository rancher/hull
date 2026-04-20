package chart

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	multierr "github.com/hashicorp/go-multierror"
)

func (t *template) validateRancherAnnotations(opts RancherHelmLintOptions) error {
	if !opts.Enabled {
		// nothing to check
		return nil
	}

	ignoreAnnotations := make(map[string]bool)
	for _, a := range opts.IgnoreAnnotations {
		ignoreAnnotations[a] = true
	}

	meta := t.Chart.Metadata
	if meta.Annotations == nil {
		return errors.New("missing required Rancher annotations: no annotations found")
	}
	if _, hasHiddenAnnotation := meta.Annotations["catalog.cattle.io/hidden"]; hasHiddenAnnotation {
		// no need to check for annotations
		return nil
	}
	annotations := meta.Annotations
	var err error

	// Required Annotations
	for _, a := range []string{"catalog.cattle.io/display-name", "catalog.cattle.io/namespace", "catalog.cattle.io/release-name"} {
		_, ignored := ignoreAnnotations[a]
		if ignored {
			continue
		}
		_, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
			continue
		}
	}

	// Required Annotations With Semver Values
	for _, a := range []string{"catalog.cattle.io/kube-version", "catalog.cattle.io/rancher-version"} {
		_, ignored := ignoreAnnotations[a]
		if ignored {
			continue
		}
		val, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
			continue
		}
		_, constraintErr := semver.NewConstraint(val)
		if constraintErr != nil {
			err = multierr.Append(err, fmt.Errorf("chart has an invalid semver constraint string for annotation '%s': %s", a, constraintErr))
			continue
		}
	}

	// Required Annotations With Enum Values
	for a, possibleV := range map[string][]string{
		"catalog.cattle.io/permits-os": {"linux", "windows", "linux,windows", "windows,linux"},
	} {
		_, ignored := ignoreAnnotations[a]
		if ignored {
			continue
		}
		val, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
			continue
		}
		validVal := false
		for _, v := range possibleV {
			if val == v {
				validVal = true
				break
			}
		}
		if !validVal {
			err = multierr.Append(err, fmt.Errorf("chart has an invalid value for '%s': must be one of %s", a, possibleV))
			continue
		}
	}

	return err
}
