package chart

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	multierr "github.com/hashicorp/go-multierror"
	helmLintRules "helm.sh/helm/v3/pkg/lint/rules"
	helmLintSupport "helm.sh/helm/v3/pkg/lint/support"
)

func (t *template) runDefaultRules(linter *helmLintSupport.Linter) {
	helmLintRules.Chartfile(linter)
	helmLintRules.ValuesWithOverrides(linter, t.Values)
	helmLintRules.Templates(linter, t.Values, t.Options.Release.Namespace, true)
	helmLintRules.Dependencies(linter)
}

func (t *template) runCustomRules(linter *helmLintSupport.Linter) {
	linter.RunLinterRule(helmLintSupport.ErrorSev, "values.schema.json", t.validateValuesSchemaExists())
}

func (t *template) validateValuesSchemaExists() error {
	if t.Chart.Schema == nil {
		return errors.New("no values.schema.json found")
	}
	return nil
}

func (t *template) runRancherRules(linter *helmLintSupport.Linter) {
	linter.RunLinterRule(helmLintSupport.ErrorSev, "Chart.yaml", t.validateRancherAnnotations())
}

func (t *template) validateRancherAnnotations() error {
	meta := t.Chart.Metadata
	if meta.Annotations == nil {
		return errors.New("missing required Rancher annotations: no annotations found")
	}
	annotations := meta.Annotations
	var err error

	// Required Annotations
	for _, a := range []string{"catalog.cattle.io/display-name", "catalog.cattle.io/namespace", "catalog.cattle.io/release-name"} {
		_, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
		}
	}

	// Required Annotations With Semver Values
	for _, a := range []string{"catalog.cattle.io/kube-version", "catalog.cattle.io/rancher-version"} {
		val, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
		}
		_, constraintErr := semver.NewConstraint(val)
		if constraintErr != nil {
			err = multierr.Append(err, fmt.Errorf("chart has an invalid semver constraint string for annotation '%s': %s", a, constraintErr))
		}
	}

	// Required Annotations With Enum Values
	for a, possibleV := range map[string][]string{
		"catalog.cattle.io/permits-os": {"linux", "windows", "linux,windows", "windows,linux"},
	} {
		val, ok := annotations[a]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("chart missing required annotation '%s'", a))
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
		}
	}

	return err
}
