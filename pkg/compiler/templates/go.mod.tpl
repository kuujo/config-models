module github.com/onosproject/config-models/{{ .Model.Name }}_{{ .Model.Version | replace "." "_" }}

go 1.14

require (
    github.com/onosproject/config-models v1.0.0
)

replace github.com/onosproject/config-models => {{ .Compiler.Root }}
