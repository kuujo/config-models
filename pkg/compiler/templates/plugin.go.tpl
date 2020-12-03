package plugin

import (
	_ "github.com/golang/protobuf/proto"
	_ "github.com/openconfig/gnmi/proto/gnmi"
	_ "github.com/openconfig/goyang/pkg/yang"
	_ "github.com/openconfig/ygot/genutil"
	_ "github.com/openconfig/ygot/ygen"
	_ "github.com/openconfig/ygot/ygot"
	_ "github.com/openconfig/ygot/ytypes"

	"github.com/onosproject/config-models/pkg/model"
)

// ModelPlugin defines the model plugin for {{ .Model.Name }} {{ .Model.Version }}
type ModelPlugin struct{}

func (p ModelPlugin) Model() model.ConfigModel {
    return ConfigModel{}
}

var _ model.ModelPlugin = ModelPlugin{}
