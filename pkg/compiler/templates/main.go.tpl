package main

import (
	"github.com/onosproject/config-models/{{ .Model.Name }}_{{ .Model.Version | replace "." "_" }}/plugin"
)

var ModelPlugin plugin.ModelPlugin
