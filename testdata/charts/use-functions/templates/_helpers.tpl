{{- define "json-content" -}}
{
    "hello": "world"
}
{{- end -}}

{{- define "multi-json-content" -}}
[
    {
      "hello": "world"
    },
    {
      "hello": "rancher"
    }
]
{{- end -}}

{{- define "yaml-content" -}}
hello: world
{{- end -}}
