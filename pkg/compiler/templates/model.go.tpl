package plugin

import (
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/goyang/pkg/yang"

	"github.com/onosproject/config-models/pkg/model"
)

const (
    modelName    model.Name    = {{ .Model.Name | quote }}
    modelVersion model.Version = {{ .Model.Version | quote }}
)

var modelData = []*gnmi.ModelData{
    {{- range .Model.Modules }}
	{Name: {{ .Name | quote }}, Organization: {{ .Organization | quote }}, Version: {{ .Version | quote }}},
	{Name: {{ .Name | quote }}, Organization: {{ .Organization | quote }}, Version: {{ .Version | quote }}},
	{{- end }}
}

// ConfigModel defines the config model for {{ .Model.Name }} {{ .Model.Version }}
type ConfigModel struct{}

func (m ConfigModel) Name() model.Name {
    return modelName
}

func (m ConfigModel) Version() model.Version {
    return modelVersion
}

func (m ConfigModel) Data() []*gnmi.ModelData {
    return modelData
}

func (m ConfigModel) Schema() (map[string]*yang.Entry, error) {
	return UnzipSchema()
}

func (m ConfigModel) GetStateMode() model.GetStateMode {
    return model.GetStateNone
}

func (m ConfigModel) Unmarshaller() model.Unmarshaller {
    return Unmarshaller{}
}

func (m ConfigModel) Validator() model.Validator {
    return Validator{}
}

var _ model.ConfigModel = ConfigModel{}
