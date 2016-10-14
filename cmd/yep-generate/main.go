// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/loader"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/tools/generate"
	"strings"
)

const (
	POOL_DIR     string = "pool"
	TEMP_STRUCTS string = "temp_structs.go"
	TEMP_METHODS string = "temp_methods.go"
	STRUCT_GEN   string = "yep-temp.go"
)

func main() {
	// TODO Set this through flags
	models.Testing = true

	cleanPoolDir(POOL_DIR)
	conf := loader.Config{
		AllowErrors: true,
	}
	fmt.Print(`
YEP Generate
------------
Loading program...
Warnings may appear here, just ignore them if yep-generate doesn't crash
`)
	conf.Import(generate.CONFIG_PATH)
	conf.Import(generate.TEST_MODULE_PATH)
	program, _ := conf.Load()
	fmt.Println("Ok")
	fmt.Print("Identifying modules...")
	modules := generate.GetModulePackages(program)
	fmt.Println("Ok")

	fmt.Print("Stage 1: Generating temporary structs...")
	missingDecls := getMissingDeclarations(modules)
	generateTempStructs(path.Join(POOL_DIR, TEMP_STRUCTS), missingDecls)
	fmt.Println("Ok")

	fmt.Print("Stage 2: Generating final structs...")
	defsModules := filterDefsModules(modules)
	generateFromModelRegistry(POOL_DIR, defsModules)
	os.Remove(path.Join(POOL_DIR, TEMP_STRUCTS))
	fmt.Println("Ok")

	fmt.Print("Stage 3: Generating temporary methods...")
	generateTempMethods(path.Join(POOL_DIR, TEMP_METHODS))
	fmt.Println("Ok")

	fmt.Print("Stage 4: Generating final methods...")
	generateFromModelRegistry(POOL_DIR, []string{generate.CONFIG_PATH, generate.TEST_MODULE_PATH})
	os.Remove(path.Join(POOL_DIR, TEMP_METHODS))
	fmt.Println("Ok")

	fmt.Println("Pool successfully generated")
}

// cleanPoolDir removes all files in the given directory and leaves only
// one empty file declaring package 'pool'.
func cleanPoolDir(dirName string) {
	os.RemoveAll(dirName)
	os.MkdirAll(dirName, 0755)
	generate.CreateFileFromTemplate(path.Join(dirName, "temp.go"), emptyPoolTemplate, nil)
}

// getMissingDeclarations parses the errors from the program for
// identifiers not declared in package pool, and returns a slice
// with all these names.
func getMissingDeclarations(packages []*generate.ModuleInfo) []string {
	// We scan all packages and populate a map to have distinct values
	missing := make(map[string]bool)
	for _, pack := range packages {
		for _, err := range pack.Errors {
			typeErr, ok := err.(types.Error)
			if !ok {
				continue
			}
			var identName string
			n, e := fmt.Sscanf(typeErr.Msg, "%s not declared by package pool", &identName)
			if n == 0 || e != nil {
				continue
			}
			missing[identName] = true
		}
	}

	// We create our result slice from the missing map
	res := make([]string, len(missing))
	var i int
	for m := range missing {
		res[i] = m
		i++
	}
	return res
}

// generateTempStructs creates a temporary file with empty struct
// definitions with the given names.
//
// This is typically done so that yep can compile to have access to
// reflection and generate the final structs.
func generateTempStructs(fileName string, names []string) {
	generate.CreateFileFromTemplate(fileName, tempStructsTemplate, names)
}

// generateMethodsStructs creates a temporary file with empty methods
// definitions of all models
func generateTempMethods(fileName string) {
	type methData struct {
		Model      string
		Name       string
		Params     string
		ReturnType string
	}
	type templData struct {
		Methods []methData
		Imports []string
	}

	astData := generate.GetMethodsASTData()
	var data templData
	for ref, mData := range astData {
		params := strings.Join(mData.Params, " interface{}, ") + " interface{}"
		if ref.Model == "" {
			// RecordCollection methods have already been generated in previous step
			continue
		} else {
			data.Methods = append(data.Methods, methData{
				Model:      fmt.Sprintf("%sSet", ref.Model),
				Name:       ref.Method,
				Params:     params,
				ReturnType: strings.Replace(mData.ReturnType.Type, "pool.", "", 1),
			})
			if mData.ReturnType.ImportPath != "" && mData.ReturnType.ImportPath != generate.POOL_PATH {
				data.Imports = append(data.Imports, mData.ReturnType.ImportPath)
			}
		}
	}
	generate.CreateFileFromTemplate(fileName, tempMethodsTemplate, data)
}

// filterDefsModules returns the names of modules of type DEFS from the given
// modules list.
func filterDefsModules(modules []*generate.ModuleInfo) []string {
	var modulesList []string
	for _, modInfo := range modules {
		if modInfo.ModType == generate.DEFS {
			modulesList = append(modulesList, modInfo.String())
		}
	}
	return modulesList
}

// generateFromModelRegistry will generate the structs in the pool from the data
// in the model registry that will be created by importing the given modules.
func generateFromModelRegistry(dirName string, modules []string) {
	generatorFileName := path.Join(os.TempDir(), STRUCT_GEN)
	//defer os.Remove(generatorFileName)

	data := struct {
		Imports    []string
		DirName    string
		ModelsPath string
	}{
		Imports:    modules,
		DirName:    dirName,
		ModelsPath: generate.MODELS_PATH,
	}
	generate.CreateFileFromTemplate(generatorFileName, buildTemplate, data)

	cmd := exec.Command("go", "run", generatorFileName)
	if output, err := cmd.CombinedOutput(); err != nil {
		panic(string(output))
	}
}

var emptyPoolTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package pool
`))

var tempStructsTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package pool

{{ range . }}
type {{ . }} struct {}
{{ end }}
`))

var tempMethodsTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package pool

import (
{{ range .Imports }} 	"{{ . }}"
{{ end }}
)

{{ range .Methods }}
func (s {{ .Model }}) {{ .Name }}({{ .Params }}) {{ .ReturnType }} {
	return *new({{ .ReturnType }})
}
{{ end }}
`))

var buildTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package main

import (
	"{{ .ModelsPath }}"
{{ range .Imports }} 	_ "{{ . }}"
{{ end }}
)

func main() {
	models.GeneratePool("{{ .DirName }}")
}
`))
