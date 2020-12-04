package main

import (
	"github.com/onosproject/config-models/{{ .Model.Name }}_{{ .Model.Version | replace "." "_" }}/model"
)

var ConfigPlugin model.ConfigPlugin
