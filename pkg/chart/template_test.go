package chart

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/stretchr/testify/assert"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	exampleChartPath = filepath.Join("..", "..", "testdata", "charts", "example-chart")
)

func getTemplate(t *testing.T, chartPath string, opts *TemplateOptions) Template {
	c, err := NewChart(chartPath)
	if err != nil {
		t.Errorf("unable to construct chart from chart path: %s", err)
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
		"configmap.yaml",
		"NOTES.txt",
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
			return
		}
		assert.NotNil(t, template.GetChart())
		assert.NotNil(t, template.GetOptions())
		assert.NotNil(t, template.GetFiles())
		assert.Equal(t, tc.NumExpectedFiles, len(template.GetFiles()))
		assert.NotNil(t, template.GetObjectSets())
		assert.NotNil(t, template.GetValues())

		fakeT := &testing.T{}
		template.YamlLint(fakeT)
		assert.Equal(t, tc.ShouldFailYamlLint, fakeT.Failed())

		fakeT = &testing.T{}
		template.HelmLint(fakeT, tc.HelmLintOptions)
		assert.Equal(t, tc.ShouldFailHelmLint, fakeT.Failed())

		template.Check(t, nil, func(*testing.T, struct{}) {})
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
		testTemplate.yamlLint(fakeT, "")
		assert.False(t, fakeT.Failed())
	})

	testTemplate = getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = make(map[string]*objectset.ObjectSet)
	t.Run("Should pass on non-nil but empty ObjectSets", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "")
		assert.False(t, fakeT.Failed())
	})

	testTemplate = getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.Files = make(map[string]string)
	t.Run("Should fail on not finding the template file associated with objects", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.yamlLint(fakeT, "")
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
		testTemplate.yamlLint(fakeT, "bad.yaml")
		assert.True(t, fakeT.Failed())
	})
}

func TestCheck(t *testing.T) {
	testTemplate := getTemplate(t, exampleChartPath, nil).(*template)
	testTemplate.ObjectSets = nil
	t.Run("Should pass on a bad YAML file", func(t *testing.T) {
		fakeT := &testing.T{}
		testTemplate.Check(fakeT, nil, struct{}{})
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
				ValuesOptions: &helmValues.Options{
					ValueFiles:   []string{"values.yaml"},
					Values:       []string{"name=prod"},
					StringValues: []string{"value=1234"},
					FileValues:   []string{"myfile=hello"},
				},
			},
			String: "helm template -f values.yaml --set name=prod --set-string value=1234 --set-file myfile=hello <path-to-chart>",
		},
		{
			Name: "Custom Values Options With Multiple Values",
			Options: &TemplateOptions{
				ValuesOptions: &helmValues.Options{
					ValueFiles:   []string{"values.yaml", "values-2.yaml"},
					Values:       []string{"name=prod", "cluster=world"},
					StringValues: []string{"value=1234", "hello=4321"},
					FileValues:   []string{"myfile=hello", "myscript=world"},
				},
			},
			String: "helm template -f values.yaml -f values-2.yaml --set name=prod --set cluster=world --set-string value=1234 --set-string hello=4321 --set-file myfile=hello --set-file myscript=world <path-to-chart>",
		},
		{
			Name:    "Default",
			Options: NewTemplateOptions("world", "hello"),
			String:  "helm template -n hello world <path-to-chart>",
		},
		{
			Name:    "Default With KubeVersion",
			Options: NewTemplateOptions("world", "hello").SetKubeVersion("1.16.0"),
			String:  "helm template -n hello --kube-version v1.16.0 world <path-to-chart>",
		},
		{
			Name:    "Default With Set Value",
			Options: NewTemplateOptions("world", "hello").SetValue("rancher", "hull"),
			String:  "helm template -n hello --set rancher=hull world <path-to-chart>",
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
			String:  "helm template -n hello --is-upgrade --kube-version v1.16.0 --set rancher=hull world <path-to-chart>",
		},
		{
			Name:    "Default With All",
			Options: NewTemplateOptions("world", "hello").SetKubeVersion("1.16.0").SetValue("rancher", "hull").IsUpgrade(true),
			String:  "helm template -n hello --is-upgrade --kube-version v1.16.0 --set rancher=hull world <path-to-chart>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.String, tc.Options.String())
		})
	}
}
