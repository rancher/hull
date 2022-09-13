package chart

import (
	"errors"

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
