---
title: "Templating"
---

{{"{{< lead >}}"}}
Generally speaking the templating uses [Golang `text/template`](https://pkg.go.dev/text/template) template syntax. All fields with templating enabled do support the full synax from the `text/template` package.
{{"{{< /lead >}}"}}

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
< Output from the template
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
< {{ .Example.FakedOutput }}
{{- else }}
< {{ renderExample .Example }}
{{- end }}
```
{{- end }}

{{ if .Remarks -}}
{{ .Remarks }}

{{ end -}}
{{- end -}}

##  Upgrade from `v2.x` to `v3.x`

When adding [sprig](https://masterminds.github.io/sprig/) function collection some functions collided and needed replacement. You need to adapt your templates accordingly:

- Math functions (`add`, `div`, `mod`, `mul`, `multiply`, `sub`) were replaced with their sprig-equivalent and are now working with integers instead of floats. If you need them to continue to work with floats you need to use their [float-variants](https://masterminds.github.io/sprig/mathf.html).
- `now` does no longer format the current date as a string but return the current date. You need to replace this: `now "2006-01-02"` becomes `now | date "2006-01-02"`.
- `concat` is now used to concat arrays. To join strings you will need to modify your code: `concat ":" "string1" "string2"` becomes `lists "string1" "string2" | join ":"`.
- `toLower` / `toUpper` need to be replaced with their sprig equivalent `lower` and `upper`.

{{ if false }}<!-- vim: set ft=markdown: -->{{ end }}
