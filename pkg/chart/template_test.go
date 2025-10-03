package chart

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rancher/hull/pkg/utils"
	"github.com/rancher/wrangler/v3/pkg/objectset"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	exampleChartPath            = utils.MustGetPathFromModuleRoot("testdata", "charts", "example-chart")
	withoutAnnotationsChartPath = utils.MustGetPathFromModuleRoot("testdata", "charts", "without-annotations")
	hiddenChartPath             = utils.MustGetPathFromModuleRoot("testdata", "charts", "hidden-chart")
	wrongAnnotationsChartPath   = utils.MustGetPathFromModuleRoot("testdata", "charts", "wrong-annotations")
	wrongOSAnnotationChartPath  = utils.MustGetPathFromModuleRoot("testdata", "charts", "wrong-os-annotation")
	invalidKubeConstraintPath   = utils.MustGetPathFromModuleRoot("testdata", "charts", "invalid-kube-constraint")
)

func getTemplate(t *testing.T, chartPath string, opts *TemplateOptions) Template {
	c, err := NewChart(chartPath)
	if err != nil {
		t.Errorf("unable to construct chart from chart path %s: %s", chartPath, err)
		return nil
	}
	if c == nil {
		t.Errorf("received nil chart")
		return nil
	}
	template, err := c.RenderTemplate(opts)
	if err != nil {
		t.Error(err)
		return nil
	}
	return template
}

func TestTemplate(t *testing.T) {
	// NOTE: should be changed if we add more files to the example-chart
	filesInExampleChart := []string{
		"templates/NOTES.txt",
		"templates/clusterrole.yaml",
		"templates/deployment.yaml",
		"templates/hardened.yaml",
		"templates/psp.yaml",
		"templates/rbac.yaml",
	}

	testCases := []struct {
		Name               string
		ChartPath          string
		TemplateOptions    *TemplateOptions
		HelmLintOptions    *HelmLintOptions
		NumExpectedFiles   int
		ShouldFailYamlLint bool
		ShouldFailHelmLint bool
	}{
		{
			Name:               "Default",
			ChartPath:          exampleChartPath,
			TemplateOptions:    nil,
			HelmLintOptions:    nil,
			NumExpectedFiles:   len(filesInExampleChart),
			ShouldFailYamlLint: false,
			ShouldFailHelmLint: false,
		},
		{
			Name:            "Default With Rancher HelmLint",
			ChartPath:       exampleChartPath,
			TemplateOptions: nil,
			HelmLintOptions: &HelmLintOptions{
				Rancher: RancherHelmLintOptions{
					Enabled: true,
				},
			},
			NumExpectedFiles:   len(filesInExampleChart),
			ShouldFailYamlLint: false,
			ShouldFailHelmLint: false,
		},
	}

	for _, tc := range testCases {
		template := getTemplate(t, tc.ChartPath, tc.TemplateOptions)
		if template == nil {
			t.Fatalf("could not find template %s", tc.ChartPath)
		}
		assert.NotNil(t, template.GetChart())
		assert.NotNil(t, template.GetOptions())
		assert.NotNil(t, template.GetFiles())
		assert.Equal(t, tc.NumExpectedFiles, len(template.GetFiles()), "expected %s, found %s", filesInExampleChart, template.GetFiles())
		assert.NotNil(t, template.GetObjectSets())
		assert.NotNil(t, template.GetValues())

		fakeT := &testing.T{}
		template.YamlLint(fakeT, DefaultYamllintConf)
		assert.Equal(t, tc.ShouldFailYamlLint, fakeT.Failed())

		fakeT = &testing.T{}
		template.HelmLint(fakeT, tc.HelmLintOptions)
		assert.Equal(t, tc.ShouldFailHelmLint, fakeT.Failed())

		template.Check(t, func(*testing.T, struct{}) {})
	}
}

func TestGetOptions(t *testing.T) {
	testTemplate := getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.Options = nil
	t.Run("Should pass on nil Options", func(t *testing.T) {
		assert.NotNil(t, testTemplate.GetOptions())
	})
}

func TestYamlLint(t *testing.T) {
	testTemplate := getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = nil
	t.Run("Should pass on nil ObjectSets", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "", DefaultYamllintConf)
		assert.False(t, fakeT.Failed())
	})

	testTemplate = getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = make(map[string]*objectset.ObjectSet)
	t.Run("Should pass on non-nil but empty ObjectSets", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "", DefaultYamllintConf)
		assert.False(t, fakeT.Failed())
	})

	testTemplate = getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.Files = make(map[string]string)
	t.Run("Should fail on not finding the template file associated with objects", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "", DefaultYamllintConf)
		assert.True(t, fakeT.Failed())
	})

	testTemplate = getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = map[string]*objectset.ObjectSet{
		"bad.yaml": objectset.NewObjectSet(&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "hello.cattle.io",
				"kind":       "world",
			},
		}),
	}
	testTemplate.Files = map[string]string{
		"bad.yaml": "hello:\n world: hd",
	}
	t.Run("Should fail on a bad YAML file", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "bad.yaml", DefaultYamllintConf)
		assert.True(t, fakeT.Failed())
	})
}

func TestCheck(t *testing.T) {
	testTemplate := getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = nil
	t.Run("Should pass on a bad YAML file", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.Check(fakeT, struct{}{})
		assert.False(t, fakeT.Failed())
	})
}

func TestSetKubeVersion(t *testing.T) {
	testCases := []struct {
		Name             string
		Version          string
		ShouldThrowError bool
	}{
		{
			Name:    "Valid",
			Version: "1.25.0",
		},
		{
			Name:             "Invalid",
			Version:          "1.25.",
			ShouldThrowError: true,
		},
		{
			Name:    "K3s",
			Version: "v1.25.0-k3s1",
		},
		{
			Name:    "RKE2",
			Version: "v1.25.0+rke2r1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err != nil {
					assert.True(t, tc.ShouldThrowError, "unexpected error: %s", err)
				}
				if err == nil {
					assert.False(t, tc.ShouldThrowError, "expected error to be thrown")
				}
			}()
			opts := NewTemplateOptions("example-chart", "default").SetKubeVersion(tc.Version)
			assert.NotNil(t, opts.Capabilities)
			assert.NotNil(t, opts.Capabilities.KubeVersion)
			version := tc.Version
			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}
			assert.Equal(t, opts.Capabilities.KubeVersion.Version, version)
		})
	}
}

func TestTemplateOptionsString(t *testing.T) {
	testCases := []struct {
		Name    string
		Options *TemplateOptions
		String  string
	}{
		{
			Name:    "Nil",
			Options: &TemplateOptions{},
			String:  "helm template <path-to-chart>",
		},
		{
			Name: "Custom Values Options",
			Options: &TemplateOptions{
				Values: &Values{
					ValueFiles:   []string{"testdata/values.yaml"},
					Values:       []string{"name=prod"},
					StringValues: []string{"value=1234"},
					FileValues:   []string{"myfile=testdata/values.yaml"},
					JSONValues:   []string{"myobj={\"hello\": \"world\"}"},
				},
			},
			String: "helm template -f testdata/values.yaml --set 'name=prod' --set-string 'value=1234' --set-file 'myfile=testdata/values.yaml' --set-json 'myobj={\"hello\": \"world\"}' <path-to-chart>",
		},
		{
			Name: "Custom Values Options With Multiple Values",
			Options: &TemplateOptions{
				Values: &Values{
					ValueFiles:   []string{"testdata/values.yaml", "testdata/values-2.yaml"},
					Values:       []string{"name=prod", "cluster=world"},
					StringValues: []string{"value=1234", "hello=4321"},
					FileValues:   []string{"myfile=testdata/values.yaml", "myscript=testdata/values-2.yaml"},
					JSONValues:   []string{"myobj={\"hello\": \"world\"}", "myobj2={\"hello\": \"rancher\"}"},
				},
			},
			String: "helm template -f testdata/values.yaml -f testdata/values-2.yaml --set 'name=prod' --set 'cluster=world' --set-string 'value=1234' --set-string 'hello=4321' --set-file 'myfile=testdata/values.yaml' --set-file 'myscript=testdata/values-2.yaml' --set-json 'myobj={\"hello\": \"world\"}' --set-json 'myobj2={\"hello\": \"rancher\"}' <path-to-chart>",
		},
		{
			Name:    "Default",
			Options: NewTemplateOptions("world", "hello"),
			String:  "helm template -n hello world <path-to-chart>",
		},
		{
			Name:    "Default With KubeVersion",
			Options: NewTemplateOptions("world", "hello").SetKubeVersion("1.16.0"),
			String:  "helm template -n hello --kube-version 'v1.16.0' world <path-to-chart>",
		},
		{
			Name:    "Default With Set Value",
			Options: NewTemplateOptions("world", "hello").SetValue("rancher", "hull"),
			String:  "helm template -n hello --set 'rancher=hull' world <path-to-chart>",
		},
		{
			Name:    "Default With Nil Set",
			Options: NewTemplateOptions("world", "hello").Set("rancher", nil),
			String:  "helm template -n hello --set-json 'rancher=null' world <path-to-chart>",
		},
		{
			Name:    "Default With Int Set",
			Options: NewTemplateOptions("world", "hello").Set("rancher", 5),
			String:  "helm template -n hello --set-json 'rancher=5' world <path-to-chart>",
		},
		{
			Name:    "Default With String",
			Options: NewTemplateOptions("world", "hello").Set("rancher", "world"),
			String:  "helm template -n hello --set-json 'rancher=\"world\"' world <path-to-chart>",
		},
		{
			Name:    "Default With map[string]string Set",
			Options: NewTemplateOptions("world", "hello").Set("rancher", map[string]string{"hello": "world"}),
			String:  "helm template -n hello --set-json 'rancher={\"hello\":\"world\"}' world <path-to-chart>",
		},
		{
			Name:    "Default With Upgrade",
			Options: NewTemplateOptions("world", "hello").IsUpgrade(true),
			String:  "helm template -n hello --is-upgrade world <path-to-chart>",
		},
		{
			Name:    "Default Without Upgrade",
			Options: NewTemplateOptions("world", "hello").IsUpgrade(false),
			String:  "helm template -n hello world <path-to-chart>",
		},
		{
			Name:    "Default With All",
			Options: NewTemplateOptions("world", "hello").SetKubeVersion("1.16.0").SetValue("rancher", "hull").IsUpgrade(true),
			String:  "helm template -n hello --is-upgrade --kube-version 'v1.16.0' --set 'rancher=hull' world <path-to-chart>",
		},
		{
			Name:    "Default With All",
			Options: NewTemplateOptions("world", "hello").SetKubeVersion("1.16.0").SetValue("rancher", "hull").IsUpgrade(true),
			String:  "helm template -n hello --is-upgrade --kube-version 'v1.16.0' --set 'rancher=hull' world <path-to-chart>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.String, tc.Options.String())
			if tc.Options.Values != nil {
				_, err := tc.Options.Values.ToMap()
				assert.Nil(t, err, fmt.Sprintf("could not compute values.yaml overrides for command '%s': %s", tc.String, err))
			}
		})
	}
}

func TestAdditionalLintChecks(t *testing.T) {
	testCases := []struct {
		Name                                 string
		ChartPath                            string
		ShouldFailValidateRancherAnnotations bool
	}{
		{
			Name:      "Example Chart",
			ChartPath: exampleChartPath,
		},
		{
			Name:                                 "Without Annotations",
			ChartPath:                            withoutAnnotationsChartPath,
			ShouldFailValidateRancherAnnotations: true,
		},
		{
			Name:      "Hidden Chart",
			ChartPath: hiddenChartPath,
		},
		{
			Name:                                 "Wrong Annotations",
			ChartPath:                            wrongAnnotationsChartPath,
			ShouldFailValidateRancherAnnotations: true,
		},
		{
			Name:                                 "Wrong OS Annotation",
			ChartPath:                            wrongOSAnnotationChartPath,
			ShouldFailValidateRancherAnnotations: true,
		},
		{
			Name:                                 "Invalid Kube Constraint",
			ChartPath:                            invalidKubeConstraintPath,
			ShouldFailValidateRancherAnnotations: true,
		},
	}

	for _, tc := range testCases {
		template := getTemplate(t, tc.ChartPath, nil).(*template)
		if template == nil {
			t.Fatalf("could not find template %s", tc.ChartPath)
		}

		var err error

		err = template.validateRancherAnnotations()
		if err != nil {
			assert.True(t, tc.ShouldFailValidateRancherAnnotations, "unexpected error: %s", err)
		}
		if err == nil {
			assert.False(t, tc.ShouldFailValidateRancherAnnotations, "expected error to be thrown")
		}
	}
}
