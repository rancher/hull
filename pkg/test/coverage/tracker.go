package coverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/test/coverage/internal"
	"github.com/rancher/hull/pkg/tpl"
	"github.com/rancher/hull/pkg/tpl/parse"
)

type Tracker struct {
	FieldUsage FieldTracker
}

func NewTracker(usage *tpl.TemplateUsage, includeSubcharts bool) *Tracker {
	if usage == nil {
		return nil
	}

	var trackFieldsFromResult func(*parse.Result, string, []string)

	fieldUsage := NewFieldTracker()
	trackFieldsFromResult = func(result *parse.Result, templatePath string, withinTemplates []string) {
		for _, field := range result.Fields {
			if !strings.HasPrefix(field, ".Values") {
				continue
			}
			fieldUsage.Track(field, withinTemplates, templatePath)
		}
		for _, templateCall := range result.TemplateCalls {
			newWithinTemplates := append([]string{templateCall}, withinTemplates...)
			trackFieldsFromResult(usage.NamedTemplates[templateCall], templatePath, newWithinTemplates)
		}
	}

	for templatePath, result := range usage.Files {
		if !includeSubcharts && strings.HasPrefix(templatePath, "charts/") {
			continue
		}
		trackFieldsFromResult(result, templatePath, nil)
	}
	return &Tracker{
		FieldUsage: fieldUsage,
	}
}

func (t *Tracker) Record(templateOptions *chart.TemplateOptions, fieldOrNamedTemplates []string) error {
	if templateOptions == nil {
		// nothing to track, nothing is modified
		return nil
	}
	if len(fieldOrNamedTemplates) == 0 {
		// nothing to track; nothing is covered
		return nil
	}

	var setFields []string
	// setFields := []string{
	// 	".Chart.Name",
	// 	".Chart.Home",
	// 	".Chart.Sources",
	// 	".Chart.Version",
	// 	".Chart.Description",
	// 	".Chart.Keywords",
	// 	".Chart.Maintainers",
	// 	".Chart.Icon",
	// 	".Chart.APIVersion",
	// 	".Chart.Condition",
	// 	".Chart.Tags",
	// 	".Chart.AppVersion",
	// 	".Chart.Deprecated",
	// 	".Chart.Annotations",
	// 	".Chart.KubeVersion",
	// 	".Chart.Dependencies",
	// 	".Chart.Type",
	// 	".Chart.IsRoot",
	// }

	// // Get setFields under .Release.*
	// if len(templateOptions.Release.Name) > 0 {
	// 	setFields = append(setFields, ".Release.Name")
	// }
	// if len(templateOptions.Release.Namespace) > 0 {
	// 	setFields = append(setFields, ".Release.Namespace")
	// }
	// if templateOptions.Release.Revision > 0 {
	// 	setFields = append(setFields, ".Release.Revision")
	// }
	// if templateOptions.Release.IsUpgrade {
	// 	setFields = append(setFields, ".Release.IsUpgrade")
	// }
	// if templateOptions.Release.IsInstall {
	// 	setFields = append(setFields, ".Release.IsInstall")
	// }

	// // Get setFields under .Capabilities.*
	// if templateOptions.Capabilities != nil {
	// 	if len(templateOptions.Capabilities.KubeVersion.String()) > 0 {
	// 		setFields = append(setFields,
	// 			".Capabilities.KubeVersion",
	// 			".Capabilities.KubeVersion.GitVersion",
	// 			".Capabilities.KubeVersion.Version",
	// 			".Capabilities.KubeVersion.Major",
	// 			".Capabilities.KubeVersion.Minor",
	// 		)
	// 	}
	// 	if len(templateOptions.Capabilities.APIVersions) > 0 {
	// 		setFields = append(setFields, ".Capabilties.APIVersions", ".Capabilties.APIVersions.Has")
	// 	}
	// 	if len(templateOptions.Capabilities.HelmVersion.Version) > 0 {
	// 		setFields = append(setFields, ".Capabilties.HelmVersion.Version")
	// 	}
	// 	if len(templateOptions.Capabilities.HelmVersion.GitCommit) > 0 {
	// 		setFields = append(setFields, ".Capabilties.HelmVersion.GitCommit")
	// 	}
	// 	if len(templateOptions.Capabilities.HelmVersion.GitTreeState) > 0 {
	// 		setFields = append(setFields, ".Capabilties.HelmVersion.GitTreeState")
	// 	}
	// 	if len(templateOptions.Capabilities.HelmVersion.GoVersion) > 0 {
	// 		setFields = append(setFields, ".Capabilties.HelmVersion.GoVersion")
	// 	}
	// }

	// Get setFields under .Values.*
	valueOpts := templateOptions.Values
	if valueOpts == nil {
		valueOpts = chart.NewValues()
	} else {
		// at least one value was overridden
		setFields = append(setFields, ".Values")
	}
	values, err := valueOpts.ToMap()
	if err != nil {
		return err
	}
	setKeysFromValues := internal.GetSetKeysFromMapInterface(values)
	for field := range setKeysFromValues {
		setFields = append(setFields, ".Values"+field)
	}

	// Add trackers
	for _, fieldOrNamedTemplate := range fieldOrNamedTemplates {
		for _, fieldSeen := range setFields {
			t.FieldUsage.Covered(fieldSeen, fieldOrNamedTemplate)
		}
	}

	return nil
}

func (t *Tracker) CalculateCoverage() (float64, string) {
	if t == nil || t.FieldUsage == nil || len(t.FieldUsage) == 0 {
		return 0, "No keys exist in chart"
	}
	var usedReferences, unusedReferences []string
	var numReferences, numUsedReferences float64
	for key, templateTracker := range t.FieldUsage {
		for _, template := range templateTracker.Templates {
			splitKey := strings.Split(key, " : ")
			ref := fmt.Sprintf("{{ %s }} : %s", splitKey[0], strings.Join(append(splitKey[1:], template), " : "))
			numReferences++
			if templateTracker.IsCovered() {
				usedReferences = append(usedReferences, ref)
				numUsedReferences++
			} else {
				unusedReferences = append(unusedReferences, ref)
			}
		}
	}
	sort.Strings(usedReferences)
	sort.Strings(unusedReferences)

	if numUsedReferences == numReferences {
		return 1, "All keys in chart are fully covered:\n- " +
			strings.Join(usedReferences, "\n- ")
	}

	if numUsedReferences == 0 {
		return 0, "The following field references are not tested:\n- " +
			strings.Join(unusedReferences, "\n- ")
	}

	return numUsedReferences / numReferences,
		"The following field references are not tested:\n- " +
			strings.Join(unusedReferences, "\n- ") +
			"\n\nOnly the following field references are covered:\n- " +
			strings.Join(usedReferences, "\n- ")
}

type FieldTracker map[string]*TemplateTracker

func NewFieldTracker() FieldTracker {
	return map[string]*TemplateTracker{}
}

func (f FieldTracker) Track(field string, withinTemplates []string, templatePath string) {
	var key string
	if withinTemplates == nil {
		key = field
	} else {
		key = fmt.Sprintf("%s : %s", field, strings.Join(withinTemplates, " : "))
	}
	_, ok := f[key]
	if !ok {
		f[key] = NewTemplateTracker()
	}
	f[key].Track(templatePath)
}

func (f FieldTracker) Covered(fieldSeen, fieldOrNamedTemplate string) {
	for key, tt := range f {
		fieldPaths := strings.Split(key, " : ")
		currField := fieldPaths[0]
		if fieldSeen != currField {
			// only should cover when the field is actually seen
			continue
		}
		for _, fieldPath := range fieldPaths {
			// once the field has been seen, allow the fieldOrNamedTemplate to
			// match against any reference of that field. i.e. the field
			// itself or template(s) it resides in
			if fieldPath == fieldOrNamedTemplate {
				tt.Covered()
			}
		}
	}
}

type TemplateTracker struct {
	Templates []string
	covered   bool
}

func NewTemplateTracker() *TemplateTracker {
	return &TemplateTracker{}
}

func (t *TemplateTracker) Track(templatePath string) {
	if t.Templates == nil {
		t.Templates = []string{templatePath}
		return
	}
	t.Templates = append(t.Templates, templatePath)
	sort.Strings(t.Templates)
}

func (t *TemplateTracker) Covered() {
	t.covered = true
}

func (t *TemplateTracker) IsCovered() bool {
	return t.covered
}
