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
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"golang.org/x/tools/go/loader"
)

// A PackageType describes a type of module
type PackageType int8

const (
	// Base is the PackageType for the base package of a module
	Base PackageType = iota
	// Subs is the PackageType for a sub package of a module
	Subs
	// Models is the PackageType for the hexya/models package
	Models
)

var currentFileSet *token.FileSet

// A ModuleInfo is a wrapper around loader.Package with additional data to
// describe a module.
type ModuleInfo struct {
	loader.PackageInfo
	ModType PackageType
}

// NewModuleInfo returns a pointer to a new moduleInfo instance
func NewModuleInfo(pack *loader.PackageInfo, modType PackageType) *ModuleInfo {
	return &ModuleInfo{
		PackageInfo: *pack,
		ModType:     modType,
	}
}

// GetModulePackages returns a slice of PackageInfo for packages that are hexya modules, that is:
// - A package that declares a "MODULE_NAME" constant
// - A package that is in a subdirectory of a package
// Also returns the 'hexya/models' package since all models are initialized there
func GetModulePackages(program *loader.Program) []*ModuleInfo {
	currentFileSet = program.Fset
	modules := make(map[string]*ModuleInfo)

	// We add to the modulePaths all packages which define a MODULE_NAME constant
	// and we check for 'hexya/models' package
	for _, pack := range program.AllPackages {
		obj := pack.Pkg.Scope().Lookup("MODULE_NAME")
		if obj != nil {
			modules[pack.Pkg.Path()] = NewModuleInfo(pack, Base)
			continue
		}
		if pack.Pkg.Path() == ModelsPath {
			modules[pack.Pkg.Path()] = NewModuleInfo(pack, Models)
		}
	}

	// Now we add packages that live inside another module
	for _, pack := range program.AllPackages {
		for _, module := range modules {
			if strings.HasPrefix(pack.Pkg.Path(), module.Pkg.Path()+"/") && pack.Pkg.Path() != module.Pkg.Path() {
				modules[pack.Pkg.Path()] = NewModuleInfo(pack, Subs)
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

// A TypeData holds a Type string and optional import path for this type.
type TypeData struct {
	Type       string
	ImportPath string
}

// A FieldASTData is a holder for a field's data that will be used
// for pool code generation
type FieldASTData struct {
	Name        string
	JSON        string
	Help        string
	Description string
	Selection   map[string]string
	RelModel    string
	Type        TypeData
	IsRS        bool
	MixinField  bool
	EmbedField  bool
	embed       bool
}

// A ParamData holds the name and type of a method parameter
type ParamData struct {
	Name     string
	Variadic bool
	Type     TypeData
}

// A MethodASTData is a holder for a method's data that will be used
// for pool code generation
type MethodASTData struct {
	Name      string
	Doc       string
	PkgPath   string
	Params    []ParamData
	Returns   []TypeData
	ToDeclare bool
}

// A ModelASTData holds fields and methods data of a Model
type ModelASTData struct {
	Name         string
	ModelType    string
	IsModelMixin bool
	Fields       map[string]FieldASTData
	Methods      map[string]MethodASTData
	Mixins       map[string]bool
	Embeds       map[string]bool
	Validated    bool
}

// newModelASTData returns an initialized ModelASTData instance
func newModelASTData(name string) ModelASTData {
	idField := FieldASTData{
		Name: "ID",
		JSON: "id",
		Type: TypeData{
			Type: "int64",
		},
	}
	var modelType string
	return ModelASTData{
		Name:         name,
		Fields:       map[string]FieldASTData{"ID": idField},
		Methods:      make(map[string]MethodASTData),
		Mixins:       make(map[string]bool),
		Embeds:       make(map[string]bool),
		IsModelMixin: ModelMixins[name],
		ModelType:    modelType,
	}
}

// GetModelsASTData returns the ModelASTData of all models found when parsing program.
func GetModelsASTData(program *loader.Program) map[string]ModelASTData {
	modInfos := GetModulePackages(program)
	return GetModelsASTDataForModules(modInfos, true)
}

// GetModelsASTDataForModules returns the MethodASTData for all methods in given modules.
// If validate is true, then only models that have been explicitly declared will appear in
// the result. Mixins and embeddings will be inflated too. Use this if you want validate the
// whole application.
func GetModelsASTDataForModules(modInfos []*ModuleInfo, validate bool) map[string]ModelASTData {
	modelsData := make(map[string]ModelASTData)
	for _, modInfo := range modInfos {
		for _, file := range modInfo.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CallExpr:
					fnctName, err := ExtractFunctionName(node)
					if err != nil {
						return true
					}
					switch {
					case fnctName == "DeclareMethod":
						parseDeclareMethod(node, modInfo, &modelsData)
					case fnctName == "AddMethod":
						parseAddMethod(node, modInfo, &modelsData)
					case fnctName == "InheritModel":
						parseMixInModel(node, &modelsData)
					case fnctName == "AddFields":
						parseAddFields(node, modInfo, &modelsData)
					case strutils.StartsAndEndsWith(fnctName, "Declare", "Model"):
						parseDeclareModel(node, &modelsData)
					case strutils.StartsAndEndsWith(fnctName, "New", "Model"):
						parseNewModel(node, &modelsData)
					}
				}
				return true
			})
		}
	}
	if !validate {
		// We don't want validation, so we exit early
		return modelsData
	}
	for modelName, md := range modelsData {
		// Delete models that have not been declared explicitly
		// Because it means we have a typing error
		if !md.Validated {
			delete(modelsData, modelName)
		}
		inflateMixins(modelName, &modelsData)
		inflateEmbeds(modelName, &modelsData)
	}
	return modelsData
}

// inflateEmbeds populates the given model with fields from the embedded type
func inflateEmbeds(modelName string, modelsData *map[string]ModelASTData) {
	for emb := range (*modelsData)[modelName].Embeds {
		relModel := (*modelsData)[modelName].Fields[emb].RelModel
		inflateEmbeds(relModel, modelsData)
		for fieldName, field := range (*modelsData)[relModel].Fields {
			if fieldName == "ID" {
				continue
			}
			field.EmbedField = true
			(*modelsData)[modelName].Fields[fieldName] = field
		}
	}
}

// inflateMixins populates the given model with fields
// and methods defined in its mixins
func inflateMixins(modelName string, modelsData *map[string]ModelASTData) {
	for mixin := range (*modelsData)[modelName].Mixins {
		inflateMixins(mixin, modelsData)
		for fieldName, field := range (*modelsData)[mixin].Fields {
			if fieldName == "ID" {
				continue
			}
			field.MixinField = true
			(*modelsData)[modelName].Fields[fieldName] = field
		}
		for methodName, method := range (*modelsData)[mixin].Methods {
			method.ToDeclare = false
			(*modelsData)[modelName].Methods[methodName] = method
		}
	}
}

// parseMixInModel updates the mixin tree with the given node which is a InheritModel function
func parseMixInModel(node *ast.CallExpr, modelsData *map[string]ModelASTData) {
	fNode := node.Fun.(*ast.SelectorExpr)
	modelName, err := extractModel(fNode.X)
	if err != nil {
		if _, ok := err.(generalMixinError); ok {
			return
		}
		log.Panic("Unable to extract model while visiting AST", "error", err)
	}
	mixinModel, err := extractModel(node.Args[0])
	if err != nil {
		log.Panic("Unable to extract mixin model while visiting AST", "error", err)
	}
	if _, exists := (*modelsData)[modelName]; !exists {
		(*modelsData)[modelName] = newModelASTData(modelName)
	}
	(*modelsData)[modelName].Mixins[mixinModel] = true
}

// parseDeclareModel parses the given node which is a DeclareXXXModel function
func parseDeclareModel(node *ast.CallExpr, modelsData *map[string]ModelASTData) {
	fName, _ := ExtractFunctionName(node)
	modelName, err := extractModelNameFromFunc(node.Fun.(*ast.SelectorExpr).X.(*ast.CallExpr))
	if err != nil {
		log.Panic("Unable to find model name in DeclareModel", "error", err)
	}
	modelType := strings.TrimSuffix(strings.TrimPrefix(fName, "Declare"), "Model")

	setModelData(modelsData, modelName, modelType)
}

// parseNewModel parses the given node which is a NewXXXModel function
func parseNewModel(node *ast.CallExpr, modelsData *map[string]ModelASTData) {
	fName, _ := ExtractFunctionName(node)
	modelName := strings.Trim(node.Args[0].(*ast.BasicLit).Value, "\"`")
	modelType := strings.TrimSuffix(strings.TrimPrefix(fName, "New"), "Model")

	setModelData(modelsData, modelName, modelType)
}

// setModelData adds a model with the given name and type to the given modelsData
func setModelData(modelsData *map[string]ModelASTData, modelName string, modelType string) {
	model, exists := (*modelsData)[modelName]
	if !exists {
		model = newModelASTData(modelName)
	}
	if modelName != "CommonMixin" {
		model.Mixins["CommonMixin"] = true
	}
	switch modelType {
	case "":
		model.Mixins["BaseMixin"] = true
		model.Mixins["ModelMixin"] = true
	case "Transient":
		model.Mixins["BaseMixin"] = true
	}
	model.ModelType = modelType
	model.Validated = true
	(*modelsData)[modelName] = model
}

// ExtractFunctionName returns the name of the called function
// in the given call expression.
func ExtractFunctionName(node *ast.CallExpr) (string, error) {
	var fName string
	switch nf := node.Fun.(type) {
	case *ast.SelectorExpr:
		fName = nf.Sel.Name
	case *ast.Ident:
		fName = nf.Name
	default:
		return "", errors.New("unexpected node type")
	}
	return fName, nil
}

// parseAddFields parses the given node which is an AddFields function
func parseAddFields(node *ast.CallExpr, modInfo *ModuleInfo, modelsData *map[string]ModelASTData) {
	fNode := node.Fun.(*ast.SelectorExpr)
	modelName, err := extractModel(fNode.X)
	if err != nil {
		log.Panic("Unable to extract model while visiting AST", "error", err)
	}
	if _, exists := (*modelsData)[modelName]; !exists {
		(*modelsData)[modelName] = newModelASTData(modelName)
	}
	fields := node.Args[0].(*ast.CompositeLit)
	for _, f := range fields.Elts {
		fDef := f.(*ast.KeyValueExpr)
		fieldName := strings.Trim(fDef.Key.(*ast.BasicLit).Value, "\"`")
		var typeStr string

		switch ft := fDef.Value.(*ast.CompositeLit).Type.(type) {
		case *ast.Ident:
			typeStr = strings.TrimSuffix(ft.Name, "Field")
		case *ast.SelectorExpr:
			typeStr = strings.TrimSuffix(ft.Sel.Name, "Field")
		}
		var importPath string
		if typeStr == "Date" || typeStr == "DateTime" {
			importPath = DatesPath
		}

		var fieldParams []ast.Expr
		switch fd := fDef.Value.(type) {
		case *ast.Ident:
			fieldParams = fd.Obj.Decl.(*ast.CompositeLit).Elts
		case *ast.CompositeLit:
			fieldParams = fd.Elts
		}

		fData := FieldASTData{
			Name: fieldName,
			Type: TypeData{
				Type:       fieldtype.Type(strings.ToLower(typeStr)).DefaultGoType().String(),
				ImportPath: importPath,
			},
		}
		for _, elem := range fieldParams {
			fElem := elem.(*ast.KeyValueExpr)
			fData = parseFieldAttribute(fElem, fData, modInfo)
			if fData.embed {
				(*modelsData)[modelName].Embeds[fieldName] = true
			}
		}
		(*modelsData)[modelName].Fields[fieldName] = fData
	}
}

// parseFieldAttribute parses the given KeyValueExpr of a field definition
func parseFieldAttribute(fElem *ast.KeyValueExpr, fData FieldASTData, modInfo *ModuleInfo) FieldASTData {
	switch fElem.Key.(*ast.Ident).Name {
	case "JSON":
		fData.JSON = strings.Trim(fElem.Value.(*ast.BasicLit).Value, "\"`")
	case "Help":
		fData.Help = strings.Trim(fElem.Value.(*ast.BasicLit).Value, "\"`")
	case "String":
		fData.Description = strings.Trim(fElem.Value.(*ast.BasicLit).Value, "\"`")
	case "Selection":
		fData.Selection = extractSelection(fElem.Value)
	case "RelationModel":
		modName, err := extractModel(fElem.Value)
		if err != nil {
			log.Panic("Unable to parse RelationModel", "field", fData.Name, "error", err)
		}
		fData.RelModel = modName
		fData.IsRS = true
	case "GoType":
		fData.Type = getTypeData(fElem.Value.(*ast.CallExpr).Args[0], modInfo)
	case "Embed":
		if fElem.Value.(*ast.Ident).Name == "true" {
			fData.embed = true
		}
	}
	return fData
}

// extractSelection returns a map with the keys and values of the Selection
// specified by expr.
func extractSelection(expr ast.Expr) map[string]string {
	res := make(map[string]string)
	switch e := expr.(type) {
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			elem := elt.(*ast.KeyValueExpr)
			key := elem.Key.(*ast.BasicLit).Value
			value := strings.Trim(elem.Value.(*ast.BasicLit).Value, "\"`")
			res[key] = value
		}
	}
	return res
}

// parseAddMethod parses the given node which is an AddMethod function
func parseAddMethod(node *ast.CallExpr, modInfo *ModuleInfo, modelsData *map[string]ModelASTData) {
	fNode := node.Fun.(*ast.SelectorExpr)
	modelName, err := extractModel(fNode.X)
	if err != nil {
		log.Panic("Unable to extract model while visiting AST", "error", err)
	}
	methodName := strings.Trim(node.Args[0].(*ast.BasicLit).Value, "\"`")
	docStr := strings.Trim(node.Args[1].(*ast.BasicLit).Value, "\"`")

	var funcType *ast.FuncType
	switch fd := node.Args[2].(type) {
	case *ast.Ident:
		funcType = fd.Obj.Decl.(*ast.FuncDecl).Type
	case *ast.FuncLit:
		funcType = fd.Type
	}
	if _, exists := (*modelsData)[modelName]; !exists {
		(*modelsData)[modelName] = newModelASTData(modelName)
	}
	methData := MethodASTData{
		Name:    methodName,
		Doc:     formatDocString(docStr),
		PkgPath: modInfo.Pkg.Path(),
		Params:  extractParams(funcType, modInfo),
		Returns: extractReturnType(funcType, modInfo),
	}
	(*modelsData)[modelName].Methods[methodName] = methData
}

// parseDeclareMethod parses the given node which is a DeclareMethod function
func parseDeclareMethod(node *ast.CallExpr, modInfo *ModuleInfo, modelsData *map[string]ModelASTData) {
	fNode := node.Fun.(*ast.SelectorExpr)
	modelName, err := extractModel(fNode.X)
	if err != nil {
		log.Panic("Unable to extract model while visiting AST", "error", err)
	}
	methodName := fNode.X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name
	docStr := strings.Trim(node.Args[0].(*ast.BasicLit).Value, "\"`")

	var funcType *ast.FuncType
	switch fd := node.Args[1].(type) {
	case *ast.Ident:
		funcType = fd.Obj.Decl.(*ast.FuncDecl).Type
	case *ast.FuncLit:
		funcType = fd.Type
	}
	if _, exists := (*modelsData)[modelName]; !exists {
		(*modelsData)[modelName] = newModelASTData(modelName)
	}
	methData := MethodASTData{
		Name:      methodName,
		Doc:       formatDocString(docStr),
		PkgPath:   modInfo.Pkg.Path(),
		Params:    extractParams(funcType, modInfo),
		Returns:   extractReturnType(funcType, modInfo),
		ToDeclare: true,
	}
	(*modelsData)[modelName].Methods[methodName] = methData
}

// A generalMixinError is returned if the mixin is
// a general mixin set in NewXXXXModel function.
type generalMixinError struct{}

// Error method for generalMixinError
func (gme generalMixinError) Error() string {
	return "General Mixin Error"
}

var _ error = generalMixinError{}

// extractModel returns the string name of the model of the given ident variable
// ident must point to the expr which represents a model
// Returns an error if it cannot determine the model
func extractModel(ident ast.Expr) (string, error) {
	switch idt := ident.(type) {
	case *ast.Ident:
		// Method is called on an identifier without selector such as
		// user.AddMethod. In this case, we try to find out the model from
		// the identifier declaration.
		switch decl := idt.Obj.Decl.(type) {
		case *ast.AssignStmt:
			// The declaration is also an assignment
			switch rd := decl.Rhs[0].(type) {
			case *ast.CallExpr:
				// The assignment is a call to a function
				var fnIdent *ast.Ident
				switch ft := rd.Fun.(type) {
				case *ast.Ident:
					fnIdent = ft
				case *ast.SelectorExpr:
					fnIdent = ft.Sel
				}
				switch fnIdent.Name {
				case "MustGet", "NewModel", "NewMixinModel", "NewTransientModel", "NewManualModel":
					return strings.Trim(rd.Args[0].(*ast.BasicLit).Value, "\"`"), nil
				case "createModel":
					// This is a call from inside a NewXXXXModel function
					return "", generalMixinError{}
				default:
					return extractModelNameFromFunc(rd)
				}
			case *ast.Ident:
				// The assignment is another identifier, we go to the declaration of this new ident.
				return extractModel(rd)
			default:
				return "", fmt.Errorf("unmanaged type %s (%T) for %s", rd, rd, idt.Name)
			}
		}
	case *ast.CallExpr:
		return extractModelNameFromFunc(idt)
	default:
		return "", fmt.Errorf("unmanaged call. ident: %s (%T)", idt, idt)
	}
	return "", errors.New("unmanaged situation")
}

// extractModelNameFromFunc extracts the model name from a h.ModelName()
// expression or an error if this is not a pool function.
func extractModelNameFromFunc(ce *ast.CallExpr) (string, error) {
	switch ft := ce.Fun.(type) {
	case *ast.Ident:
		// func is called without selector, then it is not from pool
		return "", errors.New("function call without selector")
	case *ast.SelectorExpr:
		switch ftt := ft.X.(type) {
		case *ast.Ident:
			if ftt.Name != PoolModelPackage && ftt.Name != "Registry" {
				return extractModel(ftt)
			}
			return ft.Sel.Name, nil
		case *ast.CallExpr:
			return extractModel(ftt)
		default:
			return "", fmt.Errorf("selector is of not managed type: %T", ftt)
		}
	}
	return "", errors.New("unparsable function call")
}

// extractParams extracts the parameters of the given FuncType
func extractParams(ft *ast.FuncType, modInfo *ModuleInfo) []ParamData {
	var params []ParamData
	for i, pl := range ft.Params.List {
		if i == 0 {
			// pass the first argument (rs)
			continue
		}
		for _, nn := range pl.Names {
			var variadic bool
			typ := pl.Type
			if el, ok := typ.(*ast.Ellipsis); ok {
				typ = el.Elt
				variadic = true
			}
			params = append(params, ParamData{
				Name:     nn.Name,
				Variadic: variadic,
				Type:     getTypeData(typ, modInfo)})
		}
	}
	return params
}

// getTypeData returns a TypeData instance representing the typ AST Expression
func getTypeData(typ ast.Expr, modInfo *ModuleInfo) TypeData {
	typStr := types.TypeString(modInfo.TypeOf(typ), (*types.Package).Name)
	if strings.HasSuffix(typStr, "invalid type") {
		// Maybe this is a pool type that is not yet defined
		byts := bytes.Buffer{}
		printer.Fprint(&byts, currentFileSet, typ)
		typStr = strings.Replace(byts.String(), PoolModelPackage+".", "", 1)
	}
	importPath := computeExportPath(modInfo.TypeOf(typ))
	if strings.Contains(importPath, PoolPath) {
		typStr = strings.Replace(typStr, PoolModelPackage+".", "", 1)
		importPath = ""
	}

	importPathTokens := strings.Split(importPath, ".")
	if len(importPathTokens) > 0 {
		importPath = strings.Join(importPathTokens[:len(importPathTokens)-1], ".")
	}
	return TypeData{
		Type:       typStr,
		ImportPath: importPath,
	}
}

// extractReturnType returns the return type of the first returned value
// of the given FuncType as a string and an import path if needed.
func extractReturnType(ft *ast.FuncType, modInfo *ModuleInfo) []TypeData {
	var res []TypeData
	if ft.Results != nil {
		for _, l := range ft.Results.List {
			res = append(res, getTypeData(l.Type, modInfo))
		}
	}
	return res
}

// computeExportPath returns the import path of the given type
func computeExportPath(typ types.Type) string {
	var res string
	switch typTyped := typ.(type) {
	case *types.Struct, *types.Named:
		res = types.TypeString(typTyped, (*types.Package).Path)
	case *types.Pointer:
		res = computeExportPath(typTyped.Elem())
	case *types.Slice:
		res = computeExportPath(typTyped.Elem())
	}
	return res
}

// formatDocString formats the given string by stripping whitespaces at the
// beginning of each line and prepend "// ". It also strips empty lines at
// the beginning.
func formatDocString(doc string) string {
	var res string
	var dataStarted bool
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if line == "" && !dataStarted {
			continue
		}
		dataStarted = true
		res += fmt.Sprintf("// %s\n", line)
	}
	return strings.TrimRight(res, "\n")
}
