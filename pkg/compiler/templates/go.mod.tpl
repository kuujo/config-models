module github.com/onosproject/config-models/{{ .Model.Name }}_{{ .Model.Version | replace "." "_" }}

go 1.14

require (
    github.com/onosproject/config-models {{ .Compiler.Version }}
)

{{- if eq .Compiler.Mode "module" }}
replace github.com/onosproject/config-models => {{ .Compiler.RelativePath }}
{{- end }}
