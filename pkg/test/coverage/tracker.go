package coverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aiyengar2/hull/pkg/chart"
	"github.com/aiyengar2/hull/pkg/test/coverage/internal"
	"github.com/aiyengar2/hull/pkg/tpl"
	"github.com/aiyengar2/hull/pkg/tpl/parse"
	"github.com/gobwas/glob"
)

type Tracker struct {
	FieldUsage map[string]TemplateTracker
}

func NewTracker(usage *tpl.TemplateUsage, includeSubcharts bool) *Tracker {
	if usage == nil {
		return nil
	}

	var fieldUsage map[string]TemplateTracker
	var trackFieldsFromResult func(result *parse.Result, templatePath string)

	trackFieldsFromResult = func(result *parse.Result, templatePath string) {
		for _, field := range result.Fields {
			if !strings.HasPrefix(field, ".Values") {
				continue
			}
			if fieldUsage == nil {
				fieldUsage = map[string]TemplateTracker{}
			}
			if _, ok := fieldUsage[field]; !ok {
				fieldUsage[field] = NewTemplateTracker()
			}
			fieldUsage[field].Track(templatePath)
		}
		for _, templateCall := range result.TemplateCalls {
			trackFieldsFromResult(usage.NamedTemplates[templateCall], fmt.Sprintf("%s : %s", templateCall, templatePath))
		}
	}

	for templatePath, result := range usage.Files {
		if !includeSubcharts && strings.HasPrefix(templatePath, "charts/") {
			continue
		}
		trackFieldsFromResult(result, templatePath)
	}
	return &Tracker{
		FieldUsage: fieldUsage,
	}
}

func (t *Tracker) Record(templateOptions *chart.TemplateOptions, templateGlobPaths []string) error {
	if templateOptions == nil {
		// nothing to track, nothing is modified
		return nil
	}
	if len(templateGlobPaths) == 0 {
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
	for _, templateGlobPath := range templateGlobPaths {
		for _, field := range setFields {
			templateTracker, ok := t.FieldUsage[field]
			if !ok {
				continue
			}
			if err := templateTracker.SeenIn(templateGlobPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Tracker) CalculateCoverage() (float64, string) {
	if t == nil || t.FieldUsage == nil {
		return 0, "No keys exist in chart"
	}
	var usedReferences, unusedReferences []string
	var numReferences, numUsedReferences float64
	for field, templateTracker := range t.FieldUsage {
		for template, tracked := range templateTracker {
			ref := fmt.Sprintf("{{ %s }} : %s", field, template)
			numReferences++
			if tracked {
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

type TemplateTracker map[string]bool

func NewTemplateTracker() TemplateTracker {
	return map[string]bool{}
}

func (t TemplateTracker) Track(templatePath string) {
	t[templatePath] = false
}

func (t TemplateTracker) SeenIn(templateGlobPath string) error {
	glob, err := glob.Compile(templateGlobPath)
	if err != nil {
		return err
	}
	for templatePaths := range t {
		for _, templatePath := range strings.Split(templatePaths, " : ") {
			if glob.Match(templatePath) {
				t[templatePaths] = true
			}
		}
	}
	return nil
}
