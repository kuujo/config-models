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

package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	_ "github.com/golang/protobuf/proto"
	_ "github.com/openconfig/gnmi/proto/gnmi"
	_ "github.com/openconfig/goyang/pkg/yang"
	_ "github.com/openconfig/ygot/genutil"
	_ "github.com/openconfig/ygot/ygen"
	_ "github.com/openconfig/ygot/ygot"
	_ "github.com/openconfig/ygot/ytypes"
)

const (
	pluginDir       = "plugin"
	yangModelDir    = "yang"
	compilerVersion = "v0.6.10"
)

const (
	modTemplate          = "go.mod.tpl"
	mainTemplate         = "main.go.tpl"
	pluginTemplate       = "plugin.go.tpl"
	modelTemplate        = "model.go.tpl"
	unmarshallerTemplate = "unmarshaller.go.tpl"
	validatorTemplate    = "validator.go.tpl"
)

const (
	modFile          = "go.mod"
	mainFile         = "main.go"
	pluginFile       = "plugin.go"
	modelFile        = "model.go"
	unmarshallerFile = "unmarshaller.go"
	validatorFile    = "validator.go"
)

// Mode is a compiler mode
type Mode string

const (
	// ModeBinary runs the compiler binary
	ModeBinary Mode = "binary"
	// ModeModule runs the compiler as a module
	ModeModule Mode = "module"
)

// ModelInfo is config model info
type ModelInfo struct {
	Name    string
	Version string
	Modules []ModuleInfo
}

// ModuleInfo is a config module info
type ModuleInfo struct {
	Name         string
	Organization string
	Version      string
	Data         []byte
}

// PluginInfo is config model plugin info
type PluginInfo struct {
	Name    string
	Version string
	File    string
}

// CompilerInfo is the compiler info
type CompilerInfo struct {
	Mode         Mode
	Version      string
	RelativePath string
}

// TemplateInfo provides all the variables for templates
type TemplateInfo struct {
	Model    ModelInfo
	Compiler CompilerInfo
}

// CompilePlugin compiles a model plugin to the given path
func CompilePlugin(model ModelInfo, config PluginCompilerConfig) (PluginInfo, error) {
	return NewPluginCompiler(config).CompilePlugin(model)
}

// PluginCompilerConfig is a plugin compiler configuration
type PluginCompilerConfig struct {
	TemplatePath string
	OutputPath   string
	Mode         Mode
}

// NewPluginCompiler creates a new model plugin compiler
func NewPluginCompiler(config PluginCompilerConfig) *PluginCompiler {
	return &PluginCompiler{config}
}

// PluginCompiler is a model plugin compiler
type PluginCompiler struct {
	config PluginCompilerConfig
}

// CompilePlugin compiles a model plugin to the given path
func (c *PluginCompiler) CompilePlugin(model ModelInfo) (PluginInfo, error) {
	// Ensure the output directory exists
	c.createDir(c.config.OutputPath)

	// Create the module files
	c.createDir(c.getModuleDir(model))
	if err := c.generateMod(model); err != nil {
		return PluginInfo{}, err
	}
	if err := c.generateMain(model); err != nil {
		return PluginInfo{}, err
	}

	// Create the model plugin
	c.createDir(c.getPackageDir(model))
	if err := c.generateConfigModel(model); err != nil {
		return PluginInfo{}, err
	}
	if err := c.generateUnmarshaller(model); err != nil {
		return PluginInfo{}, err
	}
	if err := c.generateValidator(model); err != nil {
		return PluginInfo{}, err
	}
	if err := c.generateModelPlugin(model); err != nil {
		return PluginInfo{}, err
	}

	// Generate the YANG bindings
	c.createDir(c.getYangDir(model))
	if err := c.copyModules(model); err != nil {
		return PluginInfo{}, err
	}
	if err := c.generateYangBindings(model); err != nil {
		return PluginInfo{}, err
	}

	// Compile the plugin
	if err := c.compilePlugin(model); err != nil {
		return PluginInfo{}, err
	}
	return PluginInfo{
		Name:    model.Name,
		Version: model.Version,
		File:    c.getPluginFile(model),
	}, nil
}

func (c *PluginCompiler) getTemplateInfo(model ModelInfo) (TemplateInfo, error) {
	wd, err := os.Getwd()
	if err != nil {
		return TemplateInfo{}, err
	}

	md, err := filepath.Abs(c.getModuleDir(model))
	if err != nil {
		return TemplateInfo{}, err
	}

	path, err := filepath.Rel(md, wd)
	if err != nil {
		return TemplateInfo{}, err
	}

	return TemplateInfo{
		Model: model,
		Compiler: CompilerInfo{
			Mode:         c.config.Mode,
			Version:      compilerVersion,
			RelativePath: path,
		},
	}, nil
}

func (c *PluginCompiler) compilePlugin(model ModelInfo) error {
	cmd := exec.Command("go", "build", "-o", c.getPluginFile(model), "-buildmode=plugin", "github.com/onosproject/config-models/"+c.getSafeQualifiedName(model))
	cmd.Dir = c.getModuleDir(model)
	cmd.Env = append(os.Environ(), "GO111MODULE=on", "CGO_ENABLED=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *PluginCompiler) copyModules(model ModelInfo) error {
	for _, module := range model.Modules {
		if err := c.copyModule(model, module); err != nil {
			return err
		}
	}
	return nil
}

func (c *PluginCompiler) copyModule(model ModelInfo, module ModuleInfo) error {
	path := c.getYangPath(model, module)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte(module.Data), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *PluginCompiler) generateYangBindings(model ModelInfo) error {
	args := []string{
		"run",
		"github.com/openconfig/ygot/generator",
		"-path=yang",
		"-output_file=plugin/generated.go",
		"-package_name=plugin",
		"-generate_fakeroot",
	}

	for _, module := range model.Modules {
		args = append(args, c.getYangFile(module))
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = c.getModuleDir(model)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *PluginCompiler) getTemplatePath(name string) string {
	return filepath.Join(c.config.TemplatePath, name)
}

func (c *PluginCompiler) generateMain(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(mainTemplate, c.getTemplatePath(mainTemplate), c.getModulePath(model, mainFile), info)
}

func (c *PluginCompiler) generateMod(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(modTemplate, c.getTemplatePath(modTemplate), c.getModulePath(model, modFile), info)
}

func (c *PluginCompiler) generateModelPlugin(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(pluginTemplate, c.getTemplatePath(pluginTemplate), c.getPackagePath(model, pluginFile), info)
}

func (c *PluginCompiler) generateConfigModel(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(modelTemplate, c.getTemplatePath(modelTemplate), c.getPackagePath(model, modelFile), info)
}

func (c *PluginCompiler) generateUnmarshaller(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(unmarshallerTemplate, c.getTemplatePath(unmarshallerTemplate), c.getPackagePath(model, unmarshallerFile), info)
}

func (c *PluginCompiler) generateValidator(model ModelInfo) error {
	info, err := c.getTemplateInfo(model)
	if err != nil {
		return err
	}
	return applyTemplate(validatorTemplate, c.getTemplatePath(validatorTemplate), c.getPackagePath(model, validatorFile), info)
}

func (c *PluginCompiler) getPluginFile(model ModelInfo) string {
	return fmt.Sprintf("%s.so.%s", model.Name, model.Version)
}

func (c *PluginCompiler) getModuleDir(model ModelInfo) string {
	return filepath.Join(c.config.OutputPath, c.getSafeQualifiedName(model))
}

func (c *PluginCompiler) getModulePath(model ModelInfo, name string) string {
	return filepath.Join(c.getModuleDir(model), name)
}

func (c *PluginCompiler) getPackageDir(model ModelInfo) string {
	return filepath.Join(c.getModuleDir(model), pluginDir)
}

func (c *PluginCompiler) getPackagePath(model ModelInfo, name string) string {
	return filepath.Join(c.getPackageDir(model), name)
}

func (c *PluginCompiler) getYangDir(model ModelInfo) string {
	return filepath.Join(c.getModuleDir(model), yangModelDir)
}

func (c *PluginCompiler) getYangPath(model ModelInfo, module ModuleInfo) string {
	return filepath.Join(c.getYangDir(model), c.getYangFile(module))
}

func (c *PluginCompiler) getYangFile(module ModuleInfo) string {
	return fmt.Sprintf("%s@%s.yang", module.Name, module.Version)
}

func (c *PluginCompiler) getSafeQualifiedName(model ModelInfo) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", model.Name, model.Version), ".", "_")
}

func (c *PluginCompiler) createDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

func (c *PluginCompiler) removeDir(dir string) {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		os.RemoveAll(dir)
	}
}

func applyTemplate(name, tplPath, outPath string, data TemplateInfo) error {
	var funcs template.FuncMap = map[string]interface{}{
		"quote": func(value string) string {
			return "\"" + value + "\""
		},
		"replace": func(search, replace, value string) string {
			return strings.ReplaceAll(value, search, replace)
		},
	}

	tpl, err := template.New(name).
		Funcs(funcs).
		ParseFiles(tplPath)
	if err != nil {
		return err
	}

	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tpl.Execute(file, data)
}
