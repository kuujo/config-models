// Copyright 2020-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

const pluginExt = ".so."
const pluginSymbol = "ModelPlugin"

// RepositoryConfig is a model plugin repository config
type RepositoryConfig struct {
	Path string `yaml:"path" json:"path"`
}

// NewRepository creates a new config model repository
func NewRepository(config RepositoryConfig) *Repository {
	return &Repository{
		config: config,
	}
}

// Repository is a repository of config models
type Repository struct {
	config RepositoryConfig
}

// GetModel gets a model by name and version
func (r *Repository) GetModel(name Name, version Version) (ConfigModel, error) {
	path := filepath.Join(r.config.Path, fmt.Sprintf("%s.so.%s", name, version))
	return loadModel(path)
}

// ListModels lists models in the repository
func (r *Repository) ListModels() ([]ConfigModel, error) {
	return loadModels(r.config.Path)
}

func loadModels(path string) ([]ConfigModel, error) {
	var files []string
	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, pluginExt) {
			files = append(files, filepath.Join(path, file))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var models []ConfigModel
	for _, file := range files {
		model, err := loadModel(file)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

func loadModel(path string) (ConfigModel, error) {
	module, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symbol, err := module.Lookup(pluginSymbol)
	if err != nil {
		return nil, err
	}
	plugin, ok := symbol.(ModelPlugin)
	if !ok {
		return nil, errors.NewInvalid("symbol loaded from module %s is not a ModelPlugin", filepath.Base(path))
	}
	return plugin.Model(), nil
}
