package parse

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/rancher/hull/pkg/tpl/utils"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	testCases := []struct {
		Name     string
		Template string
		Expect   *Result
	}{
		{
			Name: "Raw Text",
			Template: `
			hello world
			`,
			Expect: &Result{},
		},
		{
			Name: "Variables and Root",
			Template: `
			{{ $hello := .Values.Hello }}
			---
			{{ toJson $.Values.Hello.World }}
			---
			{{ toYaml .Values.Hello.World }}
			---
			{{ toToml $hello.World }}
			`,
			Expect: &Result{
				Fields: []string{".Values.Hello", ".Values.Hello.World"},
			},
		},
		{
			Name: "Use Built In Object Directly",
			Template: `
			{{ $hello := . }}
			{{ toToml $hello.World }}

			{{- with $hello }}
			{{ .World }}
			{{- end }}
			`,
			Expect: &Result{
				EmitWarning: true,
			},
		},
		{
			Name: "Has Field",
			Template: `
			{{ .Values.Hello.World }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values.Hello.World",
				},
			},
		},
		{
			Name: "Has Fields",
			Template: `
			{{ .Values.Hello.World }}
			{{ .Values.Hello.Rancher }}
			{{ .Capabilities.APIVersion.Has "hello" "world" }}
			`,
			Expect: &Result{
				Fields: []string{
					".Capabilities.APIVersion.Has",
					".Values.Hello.Rancher",
					".Values.Hello.World",
				},
			},
		},
		{
			Name: "Within Function Call",
			Template: `
			{{ printf "%s/%s" .Values.Hello.World .Values.Hello.Rancher }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values.Hello.Rancher",
					".Values.Hello.World",
				},
			},
		},
		{
			Name: "Piped",
			Template: `
			{{ .Values.Hello.World | default .Values.Hello.Hull | default "hello-world" }}
			{{ template "hello-world" (.Values.Hello | and .Values.World) }}
			{{ template "hello-world" (fromYaml .Values.Hello) }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values.Hello",
					".Values.Hello.Hull",
					".Values.Hello.World",
					".Values.World",
				},
				TemplateCalls: []string{
					"hello-world",
				},
			},
		},
		{
			Name: "Chained",
			Template: `
			{{ (.Files.Glob .Values.MyFile ).AsConfig }}
			---
			{{- with .Values }}
			{{ ($.Files.Glob .MyFileInWith ).AsConfig }}
			{{- end }}
			---
			{{- with .Values }}
			{{ ($.Files.Glob $.Values.MyFileInGlobalWith ).AsConfig }}
			{{- end }}
			---
			{{ ($.Values).MyOutsideFile }}
			{{ (.Values).MyOutsideFile2 }}
			---
			{{ include (printf "%s" .Values.MyTemplate) . }}
			`,
			Expect: &Result{
				Fields: []string{
					".Files.Glob",
					".Values",
					".Values.MyFile",
					".Values.MyFileInGlobalWith",
					".Values.MyFileInWith",
					".Values.MyOutsideFile",
					".Values.MyOutsideFile2",
					".Values.MyTemplate",
				},
			},
		},
		{
			Name: "Template Or Includes Call",
			Template: `
			{{ template "hello-world" .Values.Hello }}
			{{ include "world-hello" .Values.World }}
			{{ template "hello-rancher" (.Values.Rancher | toYaml) }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values.Hello",
					".Values.Rancher",
					".Values.World",
				},
				TemplateCalls: []string{
					"hello-rancher",
					"hello-world",
					"world-hello",
				},
			},
		},
		{
			Name: "Branch Nodes",
			Template: `
			{{- if .Values.Condition }}
			{{ template "hello-world" .Values.Hello }}
			{{- else }}
			{{ .Values.Hello.World | default .Values.Hello.Hull | default "hello-world" }}
			{{- end }}
			{{ include "world-hello" .Values.World }}

			{{- range .Values.Range }}
			- {{ .Item }}
			- {{ $.Values.RootItem }}
			{{- range $ }}
			{{ .Values.IterRoot }}
			{{- else }}
			- {{ .Values.Range.Else }}
			{{- end }}
			{{- end }}

			{{- with .Values.Hello }}
			{{ .Rancher }}
			{{- with $.Values }}
			{{ .InnerWith }}
			{{ else }}
			{{ .ElseInnerWidth }}
			{{- end }}
			{{- end }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values",
					".Values.Condition",
					".Values.ElseInnerWidth",
					".Values.Hello",
					".Values.Hello.Hull",
					".Values.Hello.Rancher",
					".Values.Hello.World",
					".Values.InnerWith",
					".Values.IterRoot",
					".Values.Range",
					".Values.Range.Else",
					".Values.Range.Item",
					".Values.RootItem",
					".Values.World",
				},
				TemplateCalls: []string{
					"hello-world",
					"world-hello",
				},
			},
		},
		{
			Name: "Edge Cases",
			Template: `
			{{- with (.Values.YAMLDoc | fromYaml) }}
			- {{ . }}
			{{- end }}
			{{- with (.) }}
			- {{ . }}
			{{- end }}
			{{- with $hi := .Values.Hi }}
			- {{ .Bye }}
			{{- end }}
			`,
			Expect: &Result{
				Fields: []string{
					".Values.Hi",
					".Values.YAMLDoc",
				},
				EmitWarning: true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tmpl, err := template.New(tc.Name).Funcs(utils.GetNoopHelmFuncMap()).Parse(string(tc.Template))
			if err != nil {
				t.Fatal(fmt.Errorf("template for %s cannot be parsed: %s", t.Name(), err))
			}
			assert.Equal(t, tc.Expect, Template(tmpl))
		})
	}
}
