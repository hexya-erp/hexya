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

package generate

import (
	"bytes"
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools"
	"go/ast"
	"go/format"
	"go/parser"
	"go/types"
	"golang.org/x/tools/go/loader"
	"io/ioutil"
	"strings"
	"text/template"
)

const (
	CONFIG_PATH string = "github.com/npiganeau/yep/config"
	MODELS_PATH string = "github.com/npiganeau/yep/yep/models"
	POOL_PATH   string = "github.com/npiganeau/yep/pool"
)

var log log15.Logger

// CreateFileFromTemplate generates a new file from the given template and data
func CreateFileFromTemplate(fileName string, template *template.Template, data interface{}) {
	var srcBuffer bytes.Buffer
	template.Execute(&srcBuffer, data)
	srcData, err := format.Source(srcBuffer.Bytes())
	if err != nil {
		tools.LogAndPanic(log, "Error while formatting generated source file", "error", err, "fileName", fileName, "mData", fmt.Sprintf("%#v", data), "src", srcBuffer.String())
	}
	// Write to file
	err = ioutil.WriteFile(fileName, srcData, 0644)
	if err != nil {
		tools.LogAndPanic(log, "Error while saving generated source file", "error", err, "fileName", fileName)
	}
}

// moduleType describes a type of module
type PackageType int8

const (
	// The base package of a module
	BASE PackageType = iota
	// The defs package of a module
	DEFS
	// A sub package of a module (that is not defs)
	SUB
	// The yep/models package
	MODELS
)

// moduleInfo is a wrapper around loader.Package with additional data to
// describe a module.
type ModuleInfo struct {
	loader.PackageInfo
	ModType PackageType
}

// newModuleInfo returns a pointer to a new moduleInfo instance
func NewModuleInfo(pack *loader.PackageInfo, modType PackageType) *ModuleInfo {
	return &ModuleInfo{
		PackageInfo: *pack,
		ModType:     modType,
	}
}

// GetModulePackages returns a slice of PackageInfo for packages that are yep modules, that is:
// - A package that declares a "MODULE_NAME" constant
// - A package that is in a subdirectory of a package
// Also returns the 'yep/models' package since all models are initialized there
func GetModulePackages(program *loader.Program) []*ModuleInfo {
	modules := make(map[string]*ModuleInfo)

	// We add to the modulePaths all packages which define a MODULE_NAME constant
	// and we check for 'yep/models' package
	for _, pack := range program.AllPackages {
		obj := pack.Pkg.Scope().Lookup("MODULE_NAME")
		_, ok := obj.(*types.Const)
		if ok {
			modules[pack.Pkg.Path()] = NewModuleInfo(pack, BASE)
			continue
		}
		if pack.Pkg.Path() == MODELS_PATH {
			modules[pack.Pkg.Path()] = NewModuleInfo(pack, MODELS)
		}
	}

	// Now we add packages that live inside another module
	for _, pack := range program.AllPackages {
		for _, module := range modules {
			if strings.HasPrefix(pack.Pkg.Path(), module.Pkg.Path()) {
				typ := SUB
				if strings.HasSuffix(pack.String(), "defs") {
					typ = DEFS
				}
				modules[pack.Pkg.Path()] = NewModuleInfo(pack, typ)
			}
		}
	}

	// Finally, we build up our result slice from modules map
	modSlice := make([]*ModuleInfo, len(modules))
	var i int
	for _, mod := range modules {
		modSlice[i] = mod
		i++
	}
	return modSlice
}

// A MethodRef is a map key for a method in a model
type MethodRef struct {
	Model  string
	Method string
}

// DocAndParams is a holder for a function's doc string and parameters names
type DocAndParams struct {
	Doc    string
	Params []string
}

// GetMethodsDocAndParamsNames returns the doc string and parameters name of all
// methods of  all YEP modules.
func GetMethodsDocAndParamsNames() map[MethodRef]DocAndParams {
	res := make(map[MethodRef]DocAndParams)
	// Parse source code
	conf := loader.Config{
		AllowErrors: true,
		ParserMode:  parser.ParseComments,
	}
	conf.Import(CONFIG_PATH)
	program, _ := conf.Load()
	modInfos := GetModulePackages(program)

	// Parse all modules for comments and params names
	// In the same loop, we both :
	// - Get doc and params for all functions
	// - Get the list of methods by parsing 'CreateMethod'
	meths := make(map[MethodRef]*ast.FuncDecl)
	funcs := make(map[*ast.FuncDecl]DocAndParams)
	for _, modInfo := range modInfos {
		for _, file := range modInfo.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.FuncDecl:
					// Extract doc
					var docString string
					if node.Doc != nil {
						for _, d := range node.Doc.List {
							docString = fmt.Sprintf("%s\n%s", docString, d.Text)
						}
					}
					// Extract params
					var params []string
					for _, pl := range node.Type.Params.List {
						for _, nn := range pl.Names {
							params = append(params, nn.Name)
						}
					}
					funcs[node] = DocAndParams{
						Doc:    docString,
						Params: params,
					}
				case *ast.CallExpr:
					var fNameNode *ast.Ident
					switch nf := node.Fun.(type) {
					case *ast.SelectorExpr:
						fNameNode = nf.Sel
					case *ast.Ident:
						fNameNode = nf
					default:
						return true
					}
					if fNameNode.Name != "CreateMethod" {
						return true
					}
					modelName := ""
					if mn, ok := node.Args[0].(*ast.BasicLit); ok {
						modelName = strings.Trim(mn.Value, `"`)
					}
					methodName := ""
					if mn, ok := node.Args[1].(*ast.BasicLit); ok {
						methodName = strings.Trim(mn.Value, `"`)
					}

					funcDecl := node.Args[2].(*ast.Ident).Obj.Decl.(*ast.FuncDecl)
					meths[MethodRef{Model: modelName, Method: methodName}] = funcDecl
				}
				return true
			})
		}
	}
	// Now we extract the doc and params from funcs only for methods
	for ref, meth := range meths {
		res[ref] = funcs[meth]
	}
	return res
}

func init() {
	log = tools.GetLogger("tools/generate")
}
