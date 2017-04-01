// Copyright 2017 NDP Systèmes. All Rights Reserved.
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

package cmd

import (
	"fmt"
	"go/build"
	"go/types"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	// We need to import models because of generated code
	_ "github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/tools/generate"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/loader"
)

const (
	// PoolDirRel is the name of the generated pool directory (relative to the yep root)
	PoolDirRel string = "pool"
	// TempEmpty is the name of the temporary go file in the pool directory for startup
	TempEmpty string = "temp.go"
	// TempStructs is the name of the temporary go file in the pool directory used in stage 1
	TempStructs string = "temp_structs.go"
	// TempMethods is the name of the temporary go file in the pool directory used in stage 3
	TempMethods string = "temp_methods.go"
	// StructGen is the name of the generated script of stage 4
	StructGen string = "yep-temp.go"
)

var generateCmd = &cobra.Command{
	Use:   "generate [projectDir]",
	Short: "Generate the source code of the model pool",
	Long: `Generate the source code of the pool package which includes the definition of all the models.
Additionally, this command creates the startup file of the project.
This command must be rerun after each source code modification, including module import.

  projectDir: the directory in which to find the go package that imports all the modules we want.
              If not set, projectDir defaults to the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := "."
		if len(args) > 0 {
			projectDir = args[0]
		}
		runGenerate(projectDir)
	},
}

var (
	generateEmptyPool bool
	testedModule      string
	importedPaths     []string
)

func initGenerate() {
	YEPCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&testedModule, "test", "t", "", "Generate pool for testing the module in the given source directory. When set projectDir is ignored.")
	generateCmd.Flags().BoolVar(&generateEmptyPool, "empty", false, "Generate an empty pool package. When set projectDir is ignored.")
}

func runGenerate(projectDir string) {
	poolDir := path.Join(generate.YEPDir, PoolDirRel)
	cleanPoolDir(poolDir)
	if generateEmptyPool {
		return
	}

	conf := loader.Config{
		AllowErrors: true,
	}

	fmt.Println(`YEP Generate
------------`)
	fmt.Printf("Detected YEP root directory at %s.\n", generate.YEPDir)

	targetDir := path.Join(projectDir, "config")
	if testedModule != "" {
		targetDir, _ = filepath.Abs(testedModule)
	}
	fmt.Println("target dir", targetDir)
	importPack, err := build.ImportDir(targetDir, 0)
	if err != nil {
		panic(fmt.Errorf("Error while importing project: %s", err))
	}
	fmt.Printf("Project package found: %s.\n", importPack.Name)

	importedPaths = importPack.Imports
	if testedModule != "" {
		importedPaths = []string{importPack.ImportPath}
	}
	for _, ip := range importedPaths {
		conf.Import(ip)
	}

	fmt.Println(`Loading program...
Warnings may appear here, just ignore them if yep-generate doesn't crash.`)

	program, _ := conf.Load()
	fmt.Println("Ok")
	fmt.Print("Identifying modules...")
	modules := generate.GetModulePackages(program)
	fmt.Println("Ok")

	fmt.Print("Stage 1: Generating temporary structs (1/2)...")
	missingDecls := getMissingDeclarations(modules)
	generateTempStructs(path.Join(poolDir, TempStructs), missingDecls)
	fmt.Println("Ok")

	fmt.Print("Stage 1: Generating temporary structs (2/2)...")
	program, _ = conf.Load()
	modules = generate.GetModulePackages(program)
	newMissingDecls := getMissingFields(modules)
	for k, v := range newMissingDecls {
		missingDecls[k] = v
	}
	generateTempStructs(path.Join(poolDir, TempStructs), missingDecls)
	fmt.Println("Ok")

	fmt.Print("Stage 2: Generating final structs...")
	defsModules := filterDefsModules(modules)
	generateFromModelRegistry(poolDir, defsModules)
	os.Remove(path.Join(poolDir, TempStructs))
	fmt.Println("Ok")

	fmt.Print("Stage 3: Generating temporary methods...")
	generateTempMethods(path.Join(poolDir, TempMethods))
	os.Remove(path.Join(poolDir, TempEmpty))
	fmt.Println("Ok")

	fmt.Print("Stage 4: Generating final methods...")
	generateFromModelRegistry(poolDir, importedPaths)
	os.Remove(path.Join(poolDir, TempMethods))
	fmt.Println("Ok")

	fmt.Println("Pool generated successfully")
}

// cleanPoolDir removes all files in the given directory and leaves only
// one empty file declaring package 'pool'.
func cleanPoolDir(dirName string) {
	os.RemoveAll(dirName)
	os.MkdirAll(dirName, 0755)
	generate.CreateFileFromTemplate(path.Join(dirName, TempEmpty), emptyPoolTemplate, nil)
}

// getMissingDeclarations parses the errors from the program for
// identifiers not declared in package pool, and returns a slice
// with all these names.
func getMissingDeclarations(packages []*generate.ModuleInfo) map[string][]string {
	res, _ := scanPackagesErrors(packages, "%s not declared by package pool", 1, 1)
	result := make(map[string][]string)
	for r := range res {
		result[r] = []string{}
	}
	return result
}

// getMissingFields returns a map with models as keys and fields as values
// for missing fields.
func getMissingFields(packages []*generate.ModuleInfo) map[string][]string {
	res := make(map[string][]string)
	done := make(map[[2]string]bool)
	_, raws := scanPackagesErrors(packages, "invalid operation: pool.%s (value of type github.com/npiganeau/yep/pool.%s has no field or method %s", 3, 3)
	for _, lst := range raws {
		firstItem := *lst[0].(*string)
		model := firstItem[:len(firstItem)-2]
		field := *lst[2].(*string)
		if !done[[2]string{model, field}] {
			res[model] = append(res[model], field)
			done[[2]string{model, field}] = true
		}
	}
	return res
}

// scanPackagesErrors parses the errors from the program for
// identifiers matched by the given error format string at given position.
// It returns the list of identifiers as first result, and the raw scanned values as second output.
func scanPackagesErrors(packages []*generate.ModuleInfo, format string, num, pos int) (map[string]bool, [][]interface{}) {
	// We scan all packages and populate a map to have distinct values
	missing := make(map[string]bool)
	var raws [][]interface{}
	for _, pack := range packages {
		for _, err := range pack.Errors {
			typeErr, ok := err.(types.Error)
			if !ok {
				continue
			}
			identNames := make([]interface{}, num)
			for i := range identNames {
				identNames[i] = new(string)
			}
			n, e := fmt.Sscanf(typeErr.Msg, format, identNames...)
			if n == 0 || e != nil {
				continue
			}
			identName := *identNames[pos-1].(*string)
			missing[identName] = true
			raws = append(raws, identNames)
		}
	}
	return missing, raws
}

// generateTempStructs creates a temporary file with empty struct
// definitions with the given names.
//
// This is typically done so that yep can compile to have access to
// reflection and generate the final structs.
func generateTempStructs(fileName string, names map[string][]string) {
	var tempStructData struct {
		Import  string
		Funcs   map[string][]string
		Structs []string
	}
	tempStructData.Funcs = make(map[string][]string)
	for name, fields := range names {
		switch {
		case len(name) < 4:
			tempStructData.Funcs[name] = fields
		case name[len(name)-3:] == "Set":
			tempStructData.Structs = append(tempStructData.Structs, name)
		case name[len(name)-4:] == "Data":
			tempStructData.Structs = append(tempStructData.Structs, name)
		default:
			tempStructData.Funcs[name] = fields
		}
	}
	tempStructData.Import = generate.ModelsPath
	generate.CreateFileFromTemplate(fileName, tempStructsTemplate, tempStructData)
}

// generateMethodsStructs creates a temporary file with empty methods
// definitions of all models
func generateTempMethods(fileName string) {
	type methData struct {
		Model        string
		Name         string
		Params       string
		ReturnTypes  string
		ReturnString string
	}
	type templData struct {
		Methods []methData
		Imports []string
	}

	astData := generate.GetMethodsASTData(importedPaths)
	var data templData
	for ref, mData := range astData {
		var params string
		if len(mData.Params) > 0 {
			params = strings.Join(mData.Params, " interface{}, ") + " interface{}"
		}
		if mData.PkgPath == generate.ModelsPath {
			// BaseMixin methods defined in models have already been generated in previous step
			continue
		}
		retNews := make([]string, len(mData.ReturnType.Types))
		for i, rt := range mData.ReturnType.Types {
			retNews[i] = fmt.Sprintf("*new(%s)", rt)
		}
		data.Methods = append(data.Methods, methData{
			Model:        fmt.Sprintf("%sSet", ref.Model),
			Name:         ref.Method,
			Params:       params,
			ReturnTypes:  strings.Join(retNews, ","),
			ReturnString: strings.Join(mData.ReturnType.Types, ","),
		})
		for _, impPath := range mData.ReturnType.ImportPaths {
			if impPath != "" && impPath != generate.PoolPath {
				data.Imports = append(data.Imports, impPath)
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
		if modInfo.ModType == generate.Defs {
			modulesList = append(modulesList, modInfo.String())
		}
	}
	return modulesList
}

// generateFromModelRegistry will generate the structs in the pool from the data
// in the model registry that will be created by importing the given modules.
func generateFromModelRegistry(dirName string, modules []string) {
	generatorFileName := path.Join(os.TempDir(), StructGen)
	defer os.Remove(generatorFileName)

	data := struct {
		Imports      []string
		DirName      string
		ModelsPath   string
		GeneratePath string
	}{
		Imports:      modules,
		DirName:      dirName,
		ModelsPath:   generate.ModelsPath,
		GeneratePath: generate.GeneratePath,
	}
	generate.CreateFileFromTemplate(generatorFileName, buildTemplate, data)

	output, err := exec.Command("go", "run", generatorFileName).CombinedOutput()
	if err != nil {
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

import "{{ .Import }}"

{{ range.Structs }}
type {{ . }} struct {}
{{ end }}

{{ range  $model, $fields := .Funcs }}
type {{ $model }}Model struct {
	*models.Model
	__Name  string
}

func {{ $model }}() {{ $model }}Model {
	return {{ $model }}Model{
		__Name:  "{{ $model }}",
	}
}
{{ range $field := $fields }}
func (m {{ $model }}Model) {{ $field }}() models.FieldNamer {
	return models.FieldName("{{ $field }}")
}
{{ end }}
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
func (s {{ .Model }}) {{ .Name }}({{ .Params }}) ({{ .ReturnString }}) {
	{{ if .ReturnTypes }}
	return {{ .ReturnTypes }}
	{{ end }}
}
{{ end }}
`))

var buildTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package main

import (
	"{{ .ModelsPath }}"
	"{{ .GeneratePath }}"
{{ range .Imports }} 	_ "{{ . }}"
{{ end }}
)

func main() {
	astData := generate.GetMethodsASTData({{ printf "%#v" .Imports }})
	models.GeneratePool("{{ .DirName }}", astData)
}
`))
