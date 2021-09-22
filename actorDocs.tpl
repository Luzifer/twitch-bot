# Available Actions

{{ range .Actors }}
## {{ .Name }}

{{ .Description }}

```yaml
- type: {{ .Type }}
{{- if gt (len .Fields) 0 }}
  attributes:
{{- range .Fields }}
    # {{ .Description }}
    # Optional: {{ .Optional }}
{{- if eq .Type "bool" }}
    # Type:     {{ .Type }}{{ if .SupportTemplate }} (Supports Templating) {{ end }}
    {{ .Key }}: {{ eq .Default "true" }}
{{- end }}
{{- if eq .Type "duration" }}
    # Type:     {{ .Type }}{{ if .SupportTemplate }} (Supports Templating) {{ end }}
    {{ .Key }}: {{ if eq .Default "" }}0s{{ else }}{{ .Default }}{{ end }}
{{- end }}
{{- if eq .Type "int64" }}
    # Type:     {{ .Type }}{{ if .SupportTemplate }} (Supports Templating) {{ end }}
    {{ .Key }}: {{ if eq .Default "" }}0{{ else }}{{ .Default }}{{ end }}
{{- end }}
{{- if eq .Type "string" }}
    # Type:     {{ .Type }}{{ if .SupportTemplate }} (Supports Templating) {{ end }}
    {{ .Key }}: "{{ .Default }}"
{{- end }}
{{- if eq .Type "stringslice" }}
    # Type:     array of strings{{ if .SupportTemplate }} (Supports Templating in each string) {{ end }}
    {{ .Key }}: []
{{- end }}
{{- end }}
{{- else }}
  # Does not have configuration attributes
{{- end }}
```
{{ end }}

{{ if false }}<!-- vim: set ft=markdown: -->{{ end }}
