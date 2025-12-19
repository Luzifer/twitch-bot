---
title: Templating
---

> [!TIP]
> Generally speaking the templating uses [Golang `text/template`](https://pkg.go.dev/text/template) template syntax. All fields with templating enabled do support the full synax from the `text/template` package.

## Variables

There are certain variables available in the strings with templating enabled:

- `channel` - Channel the message was sent to, only available for regular messages not events
- `msg` - The message object, used in functions, should not be sent to chat
- `permitTimeout` - Value of `permit_timeout` in seconds
- `username` - The username of the message author


## Functions

Within templates following functions can be used:

- built-in functions in `text/template` engine
- functions from [sprig](https://masterminds.github.io/sprig/) function collection
- functions mentioned below

Examples below are using this syntax in the code block:

```
! Message matcher used for the input message
> Input message if used in the example
# Template used in the fields
< Output from the template (Rendered during docs generation)
* Output from the template (Static output, template not rendered)
```

{{ range .Funcs -}}
### `{{ .Name }}`

{{ .Description }}

Syntax: `{{ .Syntax }}`

{{- if .Example }}

Example:

```
{{- if .Example.MatchMessage }}
! {{ .Example.MatchMessage }}
{{- end }}
{{- if .Example.MessageContent }}
> {{ .Example.MessageContent }}
{{- end }}
# {{ .Example.Template }}
{{- if .Example.FakedOutput }}
* {{ .Example.FakedOutput }}
{{- else }}
< {{ renderExample .Example }}
{{- end }}
```
{{- end }}

{{ if .Remarks -}}
{{ .Remarks }}

{{ end -}}
{{- end -}}

{{ if false }}<!-- vim: set ft=markdown: -->{{ end }}
