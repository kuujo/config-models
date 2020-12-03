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

package main

import (
	"errors"
	"github.com/onosproject/config-models/pkg/compiler"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "config-model",
	}

	compileCmd := &cobra.Command{
		Use: "compile-plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			version, _ := cmd.Flags().GetString("version")
			modules, _ := cmd.Flags().GetStringToString("module")

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				dir, err := os.Getwd()
				if err != nil {
					return err
				}
				output = dir
			}

			templates, _ := cmd.Flags().GetString("templates")
			if templates == "" {
				templates = "pkg/compiler/templates"
			}

			model := compiler.ModelInfo{
				Name:    name,
				Version: version,
				Modules: []compiler.ModuleInfo{},
			}
			for nameVersion, module := range modules {
				names := strings.Split(nameVersion, "@")
				if len(names) != 2 {
					return errors.New("module name must be in the format $name@$version")
				}
				name, version := names[0], names[1]
				data, err := ioutil.ReadFile(module)
				if err != nil {
					return err
				}
				model.Modules = append(model.Modules, compiler.ModuleInfo{
					Name:    name,
					Version: version,
					Data:    data,
				})
			}

			config := compiler.PluginCompilerConfig{
				TemplatePath: templates,
				OutputPath:   output,
			}
			_, err := compiler.CompilePlugin(model, config)
			return err
		},
		SilenceUsage: true,
	}
	compileCmd.Flags().StringP("name", "n", "", "the model name")
	compileCmd.Flags().StringP("version", "v", "", "the model version")
	compileCmd.Flags().StringToStringP("module", "m", map[string]string{}, "model files")
	compileCmd.Flags().StringP("output", "o", "", "the output path")
	compileCmd.Flags().StringP("templates", "t", "", "the templates path")
	rootCmd.AddCommand(compileCmd)

	if err := rootCmd.Execute(); err != nil {
		println(err)
		os.Exit(1)
	}
}
