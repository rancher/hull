package utils

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

var testTemplate = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-map
  namespace: {{ .Release.Namespace }}
data:
  lookup.yaml: |-
{{ lookup "" "" "" "" | toYaml | indent 4 }}
  required.txt: |-
{{ required "should never fail" "dummy-value" | toYaml | indent 4 }}
  include.json: |-
{{ include "json-content" . | toYaml | indent 4 }}
  tpl.txt: |-
{{ tpl "{{ .Values.data }}" . | toYaml | indent 4 }}
  data.json: |- 
{{ fromJson (include "json-content" .) | toJson | indent 4 }}
  multi.json: |-
{{ fromJsonArray (include "multi-json-content" .) | toJson | indent 4 }}
  data.yaml: |- 
{{ fromYaml (include "json-content" .) | toYaml | indent 4 }}
  multi.yaml: |-
{{ fromYamlArray (include "multi-json-content" .) | toYaml | indent 4 }}
  data.toml: |-
{{ fromYaml (include "json-content" .) | toToml | indent 4 }}
`

func TestGetNoopHelmFuncMap(t *testing.T) {
	tpl, err := template.New(t.Name()).Funcs(GetNoopHelmFuncMap()).Parse(testTemplate)
	if err != nil {
		assert.Fail(t, "unexpected error while parsing template", err)
	}
	var renderedTplBytes bytes.Buffer
	if err := tpl.Execute(&renderedTplBytes, map[string]interface{}{}); err != nil {
		assert.Fail(t, "unexpected error while parsing template", err)
	}
	if t.Failed() {
		// useful to see where execution failed
		t.Log(renderedTplBytes.String())
	}
}
