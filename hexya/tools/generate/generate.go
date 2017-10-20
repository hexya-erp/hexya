// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"golang.org/x/tools/go/loader"
)

// A fieldData describes a field in a RecordSet
type fieldData struct {
	Name     string
	JSON     string
	RelModel string
	Type     string
	SanType  string
	IsRS     bool
}

// A returnType characterizes a return value of a method
type returnType struct {
	Type string
	IsRS bool
}

// A methodData describes a method in a RecordSet
type methodData struct {
	Name           string
	Doc            string
	Params         string
	ParamsWithType string
	ParamsTypes    string
	ReturnAsserts  string
	Returns        string
	ReturnString   string
	Call           string
	ToDeclare      bool
}

// an operatorDef defines an operator func
type operatorDef struct {
	Name  string
	Multi bool
}

// An fieldType holds the name and valid operators on a field type
type fieldType struct {
	Type      string
	SanType   string
	IsRS      bool
	Operators []operatorDef
}

// A modelData describes a RecordSet model
type modelData struct {
	Name           string
	ModelType      string
	IsModelMixin   bool
	Deps           []string
	Fields         []fieldData
	Methods        []methodData
	AllMethods     []methodData
	ConditionFuncs []string
	Types          []fieldType
}

// specificMethodsHandlers are functions that populate the given modelData
// for specific methods.
var specificMethodsHandlers = map[string]func(modelData *modelData, depsMap *map[string]bool){
	"Search":           searchMethodHandler,
	"SearchByName":     searchByNameMethodHandler,
	"First":            firstMethodHandler,
	"All":              allMethodHandler,
	"Create":           createMethodHandler,
	"Write":            writeMethodHandler,
	"Copy":             copyMethodHandler,
	"CartesianProduct": cartesianProductMethodHandler,
}

// searchMethodHandler returns the specific methodData for the Search method.
func searchMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Search"
	returnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("%sCondition", modelData.Name),
		ReturnString: returnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name:           name,
		Doc:            fmt.Sprintf("// Search returns a new %sSet filtering on the current one with the additional given Condition", modelData.Name),
		ToDeclare:      false,
		Params:         "condition",
		ParamsWithType: fmt.Sprintf("condition %sCondition", modelData.Name),
		ReturnAsserts:  "resTyped := res.(models.RecordSet).Collection()",
		Returns:        fmt.Sprintf("%sSet{RecordCollection: resTyped}", modelData.Name),
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// firstMethodHandler returns the specific methodData for the First method.
func firstMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         "First",
		ParamsTypes:  "",
		ReturnString: fmt.Sprintf("%sData", modelData.Name),
	})

}

// allMethodHandler returns the specific methodData for the All method.
func allMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         "All",
		ParamsTypes:  "",
		ReturnString: fmt.Sprintf("[]%sData", modelData.Name),
	})
}

// createMethodHandler returns the specific methodData for the Create method.
func createMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Create"
	returnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("*%sData", modelData.Name),
		ReturnString: returnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// Create inserts a %s record in the database from the given data.
// Returns the created %sSet.`, modelData.Name, modelData.Name),
		ToDeclare:      false,
		Params:         "data",
		ParamsWithType: fmt.Sprintf("data *%sData", modelData.Name),
		ReturnAsserts:  "resTyped := res.(models.RecordSet).Collection()",
		Returns:        fmt.Sprintf("%sSet{RecordCollection: resTyped}", modelData.Name),
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// writeMethodHandler returns the specific methodData for the Write method.
func writeMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Write"
	returnString := "bool"
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("*%sData, ...models.FieldNamer", modelData.Name),
		ReturnString: returnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// Write is the base implementation of the 'Write' method which updates
// %s records in the database with the given data.
//
// Only fields with non zero values or fields passed in the 'fieldsToReset' arg are updated`, modelData.Name),
		ToDeclare:      false,
		Params:         "data, fieldsToReset",
		ParamsWithType: fmt.Sprintf("data *%sData, fieldsToReset ...models.FieldNamer", modelData.Name),
		ReturnAsserts:  "resTyped, _ := res.(bool)",
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// copyMethodHandler returns the specific methodData for the Copy method.
func copyMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Copy"
	returnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("*%sData, ...models.FieldNamer", modelData.Name),
		ReturnString: returnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// Copy duplicates the given %s record, overridding values with overrides.
//
// Only fields with non zero values of overrides or fields passed in the 'fieldsToReset' arg are updated`, modelData.Name),
		ToDeclare:      false,
		Params:         "overrides, fieldsToReset",
		ParamsWithType: fmt.Sprintf("overrides *%sData, fieldsToReset ...models.FieldNamer", modelData.Name),
		ReturnAsserts:  "resTyped := res.(models.RecordSet).Collection()",
		Returns:        fmt.Sprintf("%sSet{RecordCollection: resTyped}", modelData.Name),
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// searchByNameMethodHandler returns the specific methodData for the Search method.
func searchByNameMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "SearchByName"
	returnString := fmt.Sprintf("%sSet", modelData.Name)
	(*depsMap)["github.com/hexya-erp/hexya/hexya/models/operator"] = true
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("string, operator.Operator, %sCondition, int", modelData.Name),
		ReturnString: returnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// SearchByName searches for %s records that have a display name matching the given
// "name" pattern when compared with the given "op" operator, while also
// matching the optional search condition ("additionalCond").
//
// This is used for example to provide suggestions based on a partial
// value for a relational field. Sometimes be seen as the inverse
// function of NameGet but it is not guaranteed to be.`, modelData.Name),
		ToDeclare:      false,
		Params:         "name, op, additionalCond, limit",
		ParamsWithType: fmt.Sprintf("name string, op operator.Operator, additionalCond %sCondition, limit int", modelData.Name),
		ReturnAsserts:  "resTyped := res.(models.RecordSet).Collection()",
		Returns:        fmt.Sprintf("%sSet{RecordCollection: resTyped}", modelData.Name),
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// cartesianProductMethodHandler returns the specific methodData for the CartesianProduct method.
func cartesianProductMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "CartesianProduct"
	returnString := fmt.Sprintf("[]%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:         name,
		ParamsTypes:  fmt.Sprintf("...%sSet", modelData.Name),
		ReturnString: returnString,
	})
}

// createTypeIdent creates a string from the given type that
// can be used inside an identifier.
func createTypeIdent(typStr string) string {
	res := strings.Replace(typStr, ".", "", -1)
	res = strings.Replace(res, "[", "Slice", -1)
	res = strings.Replace(res, "map[", "Map", -1)
	res = strings.Replace(res, "]", "", -1)
	res = strings.Title(res)
	return res
}

// CreatePool generates the pool package by parsing the source code AST
// of the given program.
// The generated package will be put in the given dir.
func CreatePool(program *loader.Program, dir string) {
	modelsASTData := GetModelsASTData(program)
	for modelName, modelASTData := range modelsASTData {
		depsMap := map[string]bool{ModelsPath: true}
		mData := modelData{
			Name:           modelName,
			ModelType:      modelASTData.ModelType,
			IsModelMixin:   modelASTData.IsModelMixin,
			ConditionFuncs: []string{"And", "AndNot", "Or", "OrNot"},
		}
		// Add fields
		addFieldsToModelData(modelASTData, &mData, &depsMap)
		// Add field types
		addFieldTypesToModelData(&mData)
		// Add methods
		addMethodsToModelData(modelsASTData, &mData, &depsMap)
		// Setting imports
		var deps []string
		for dep := range depsMap {
			if dep == "" {
				continue
			}
			deps = append(deps, dep)
		}
		mData.Deps = deps
		// Writing to file
		fileName := fmt.Sprintf("%s.go", strutils.SnakeCaseString(modelName))
		CreateFileFromTemplate(path.Join(dir, fileName), poolModelTemplate, mData)
	}
}

// addMethodsToModelData extracts data from modelsASTData to populate methods in modelData
func addMethodsToModelData(modelsASTData map[string]ModelASTData, modelData *modelData, depsMap *map[string]bool) {
	modelASTData := modelsASTData[modelData.Name]
	for methodName, methodASTData := range modelASTData.Methods {
		if handler, exists := specificMethodsHandlers[methodName]; exists {
			handler(modelData, depsMap)
			continue
		}
		var params, paramsWithType, paramsType, call, returns, returnAsserts, returnString string
		for _, astParam := range methodASTData.Params {
			paramType := astParam.Type.Type
			p := fmt.Sprintf("%s,", astParam.Name)
			if isRS, isRC := isRecordSetType(astParam.Type.Type, modelsASTData); isRS {
				if isRC {
					paramType = fmt.Sprintf("%sSet", modelData.Name)
				}
			}
			if astParam.Variadic {
				paramType = fmt.Sprintf("...%s", paramType)
			}
			params += p
			paramsWithType += fmt.Sprintf("%s %s,", astParam.Name, paramType)
			paramsType += fmt.Sprintf("%s,", paramType)
			(*depsMap)[astParam.Type.ImportPath] = true
		}
		if len(methodASTData.Returns) == 1 {
			call = "Call"
			(*depsMap)[methodASTData.Returns[0].ImportPath] = true
			typ := methodASTData.Returns[0].Type
			returnAsserts = fmt.Sprintf("resTyped, _ := res.(%s)", typ)
			returns = "resTyped"
			if isRS, isRC := isRecordSetType(typ, modelsASTData); isRS {
				if isRC {
					typ = fmt.Sprintf("%sSet", modelData.Name)
				}
				returnAsserts = "resTyped := res.(models.RecordSet).Collection()"
				returns = fmt.Sprintf("%s{RecordCollection: resTyped}", typ)
			}
			returnString = typ
		} else if len(methodASTData.Returns) > 1 {
			for i, ret := range methodASTData.Returns {
				call = "CallMulti"
				(*depsMap)[ret.ImportPath] = true
				if isRS, isRC := isRecordSetType(ret.Type, modelsASTData); isRS {
					retType := ret.Type
					if isRC {
						retType = fmt.Sprintf("%sSet", modelData.Name)
					}
					returnAsserts += fmt.Sprintf("resTyped%d := res[%d].(models.RecordSet).Collection()\n", i, i)
					returns += fmt.Sprintf("%s{RecordCollection: resTyped%d},", retType, i)
					returnString += fmt.Sprintf("%s,", retType)
				} else {
					returnAsserts += fmt.Sprintf("resTyped%d, _ := res[%d].(%s)\n", i, i, ret.Type)
					returns += fmt.Sprintf("resTyped%d,", i)
					returnString += fmt.Sprintf("%s,", ret.Type)
				}
			}
		}
		modelData.AllMethods = append(modelData.AllMethods, methodData{
			Name:         methodName,
			ParamsTypes:  strings.TrimRight(paramsType, ","),
			ReturnString: strings.TrimSuffix(returnString, ","),
		})
		modelData.Methods = append(modelData.Methods, methodData{
			Name:           methodName,
			Doc:            methodASTData.Doc,
			ToDeclare:      methodASTData.ToDeclare,
			Params:         strings.TrimRight(params, ","),
			ParamsWithType: strings.TrimRight(paramsWithType, ","),
			ReturnAsserts:  strings.TrimSuffix(returnAsserts, "\n"),
			Returns:        strings.TrimSuffix(returns, ","),
			ReturnString:   strings.TrimSuffix(returnString, ","),
			Call:           call,
		})
	}
}

// addFieldsToModelData extracts data from modelASTData to populate fields in modelData
func addFieldsToModelData(modelASTData ModelASTData, modelData *modelData, depsMap *map[string]bool) {
	for fieldName, fieldASTData := range modelASTData.Fields {
		typStr := fieldASTData.Type.Type
		if fieldASTData.RelModel != "" {
			typStr = fmt.Sprintf("%sSet", fieldASTData.RelModel)
		}
		jsonName := strutils.GetDefaultString(fieldASTData.JSON, strutils.SnakeCaseString(fieldName))
		modelData.Fields = append(modelData.Fields, fieldData{
			Name:     fieldName,
			JSON:     jsonName,
			Type:     typStr,
			IsRS:     fieldASTData.IsRS,
			RelModel: fieldASTData.RelModel,
			SanType:  createTypeIdent(typStr),
		})
		(*depsMap)[fieldASTData.Type.ImportPath] = true
	}
}

// addFieldsToModelData extracts field types from mData.Fields
// and add them to mData.Types
func addFieldTypesToModelData(mData *modelData) {
	fTypes := make(map[string]bool)
	for _, f := range mData.Fields {
		if fTypes[f.Type] {
			continue
		}
		fTypes[f.Type] = true
		mData.Types = append(mData.Types, fieldType{
			Type:    f.Type,
			SanType: f.SanType,
			IsRS:    f.IsRS,
			Operators: []operatorDef{
				{Name: "Equals"}, {Name: "NotEquals"}, {Name: "Greater"}, {Name: "GreaterOrEqual"}, {Name: "Lower"},
				{Name: "LowerOrEqual"}, {Name: "Like"}, {Name: "Contains"}, {Name: "NotContains"}, {Name: "IContains"},
				{Name: "NotIContains"}, {Name: "ILike"}, {Name: "In", Multi: true}, {Name: "NotIn", Multi: true},
				{Name: "ChildOf"},
			},
		})
	}
}

// isRecordSetType returns true if the given typ is a RecordSet according
// to the AST data stored in models.
// The second returned value is true if typ is models.RecordCollection or models.RecordSet
// and false if it is a specific RecordSet type
func isRecordSetType(typ string, models map[string]ModelASTData) (bool, bool) {
	if typ == "*RecordCollection" || typ == "*models.RecordCollection" {
		return true, true
	}
	if typ == "RecordSet" || typ == "models.RecordSet" {
		return true, true
	}
	if _, exists := models[strings.TrimSuffix(typ, "Set")]; exists {
		return true, false
	}
	return false, false
}

// CreateFileFromTemplate generates a new file from the given template and data
func CreateFileFromTemplate(fileName string, template *template.Template, data interface{}) {
	var srcBuffer bytes.Buffer
	template.Execute(&srcBuffer, data)
	srcData, err := format.Source(srcBuffer.Bytes())
	if err != nil {
		log.Panic("Error while formatting generated source file", "error", err, "fileName",
			fileName, "mData", fmt.Sprintf("%#v", data), "src", srcBuffer.String())
	}
	// Write to file
	err = ioutil.WriteFile(fileName, srcData, 0644)
	if err != nil {
		log.Panic("Error while saving generated source file", "error", err, "fileName", fileName)
	}
}

var poolModelTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package pool

import (
	"github.com/hexya-erp/hexya/hexya/tools/typesutils"
{{ range .Deps }} 	"{{ . }}"
{{ end }}
)

// ------- MODEL ---------

// {{ .Name }}Model is a strongly typed model definition that is used
// to extend the {{ .Name }} model or to get a {{ .Name }}Set through
// its NewSet() function.
//
// To get the unique instance of this type, call {{ .Name }}().
type {{ .Name }}Model struct {
	*models.Model
}

{{ if eq .ModelType "Mixin" }}
// NewSet returns a new {{ .Name }}Set instance wrapping the given model in the given Environment
func (m {{ .Name }}Model) NewSet(env models.Environment, modelName string) {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: env.Pool(modelName),
	}
}

{{ else }}
// NewSet returns a new {{ .Name }}Set instance in the given Environment
func (m {{ .Name }}Model) NewSet(env models.Environment) {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: env.Pool("{{ .Name }}"),
	}
}

// Create creates a new {{ .Name }} record and returns the newly created
// {{ .Name }}Set instance.
func (m {{ .Name }}Model) Create(env models.Environment, data *{{ .Name }}Data) {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: m.Model.Create(env, data),
	}
}

// Search searches the database and returns a new {{ .Name }}Set instance
// with the records found.
func (m {{ .Name }}Model) Search(env models.Environment, cond {{ .Name }}Condition) {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: m.Model.Search(env, cond),
	}
}

// Browse returns a new RecordSet with the records with the given ids.
// Note that this function is just a shorcut for Search on a list of ids.
func (m {{ .Name }}Model) Browse(env models.Environment, ids []int64) {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: m.Model.Browse(env, ids),
	}
}

{{ end }}
// Fields returns the Field Collection of the {{ .Name }} Model
func (m {{ .Name }}Model) Fields() {{ .Name }}FieldsCollection {
	return {{ .Name }}FieldsCollection {
		FieldsCollection: m.Model.Fields(),
	}
}

// Methods returns the Method Collection of the {{ .Name }} Model
func (m {{ .Name }}Model) Methods() {{ .Name }}MethodsCollection {
	return {{ .Name }}MethodsCollection {
		MethodsCollection: m.Model.Methods(),
	}
}

// Underlying returns the underlying models.Model instance
func (m {{ .Name }}Model) Underlying() *models.Model {
	return m.Model
}

var _ models.Modeler = {{ .Name }}Model{}

// Declare{{ .ModelType }}Model is a dummy method used for code generation
// It just returns m.
func (m {{ .Name }}Model) Declare{{ .ModelType }}Model() {{ .Name }}Model {
	return m
}

{{ range .Fields }}
{{ if .IsRS }}
// {{ .Name }}FilteredOn adds a condition with a table join on the given field and
// filters the result with the given condition
func (m {{ $.Name }}Model) {{ .Name }}FilteredOn(cond {{ .RelModel }}Condition) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: m.FilteredOn("{{ .Name }}", cond.Condition),
	}
}
{{ end }}

// {{ .Name }} adds the "{{ .Name }}" field to the Condition
func (m {{ $.Name }}Model) {{ .Name }}() {{ $.Name }}{{ .SanType }}ConditionField {
	return {{ $.Name }}{{ .SanType }}ConditionField{
		ConditionField: m.Field("{{ .Name }}"),
	}
}

{{ end }}

// {{ .Name }} returns the unique instance of the {{ .Name }}Model type
// which is used to extend the {{ .Name }} model or to get a {{ .Name }}Set through
// its NewSet() function.
func {{ .Name }}() {{ .Name }}Model {
	return {{ .Name }}Model{
		Model: models.Registry.MustGet("{{ .Name }}"),
	}
}

// ------- FIELD COLLECTION ----------

// A {{ .Name }}FieldsCollection is the collection of fields
// of the {{ .Name }} model.
type {{ .Name }}FieldsCollection struct {
	*models.FieldsCollection
}

{{ range .Fields }}
// {{ .Name }} returns a pointer to the {{ .Name }} Field.
func (c {{ $.Name }}FieldsCollection) {{ .Name }}() *models.Field {
	return c.MustGet("{{ .Name }}")
}
{{ end }}

// ------- METHOD COLLECTION ----------

// A {{ .Name }}MethodsCollection is the collection of methods
// of the {{ .Name }} model.
type {{ .Name }}MethodsCollection struct {
	*models.MethodsCollection
}

{{ range .AllMethods }}
// {{ $.Name }}_{{ .Name }} holds the metadata of the {{ $.Name }}.{{ .Name }}() method
type {{ $.Name }}_{{ .Name }} struct {
	*models.Method
}

// Extend adds the given fnct function as a new layer on this method.
func (m {{ $.Name }}_{{ .Name }}) Extend(doc string, fnct func({{ $.Name }}Set{{ if ne .ParamsTypes "" }}, {{ .ParamsTypes }}{{ end }}) ({{ .ReturnString }})) {{ $.Name }}_{{ .Name }} {
	return {{ $.Name }}_{{ .Name }} {
		Method: m.Method.Extend(doc, fnct),
	}
}

// DeclareMethod declares this method to the framework with the given function as the first layer.
func (m {{ $.Name }}_{{ .Name }}) DeclareMethod(doc string, fnct interface{}) {{ $.Name }}_{{ .Name }} {
	return {{ $.Name }}_{{ .Name }} {
		Method: m.Method.DeclareMethod(doc, fnct),
	}
}

// Underlying returns a pointer to the underlying Method data object.
func (m {{ $.Name }}_{{ .Name }}) Underlying() *models.Method {
	return m.Method
}

var _ models.Methoder = {{ $.Name }}_{{ .Name }}{}

// {{ .Name }} returns a pointer to the {{ .Name }} Method.
func (c {{ $.Name }}MethodsCollection) {{ .Name }}() {{ $.Name }}_{{ .Name }} {
	return {{ $.Name }}_{{ .Name }} {
		Method: c.MustGet("{{ .Name }}"),
	}
}
{{ end }}

// ------- CONDITION ---------

// A {{ .Name }}Condition is a type safe WHERE clause in an SQL query
type {{ .Name }}Condition struct {
	*models.Condition
}

{{ range .ConditionFuncs }}
// {{ . }} completes the current condition with a simple {{ . }} clause : c.{{ . }}().nextCond => c {{ . }} nextCond
func (c {{ $.Name }}Condition) {{ . }}() {{ $.Name }}ConditionStart {
	return {{ $.Name }}ConditionStart{
		ConditionStart: c.Condition.{{ . }}(),
	}
}

// {{ . }}Cond completes the current condition with the given cond as an {{ . }} clause
// between brackets : c.{{ . }}(cond) => c {{ . }} (cond)
func (c {{ $.Name }}Condition) {{ . }}Cond(cond {{ $.Name }}Condition) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.Condition.{{ . }}Cond(cond.Condition),
	}
}
{{ end }}

// Underlying returns the underlying models.Condition instance
func (c {{ $.Name }}Condition) Underlying() *models.Condition {
	return c.Condition
}

var _ models.Conditioner = {{ $.Name }}Condition{}

// ------- CONDITION START ---------

// A {{ .Name }}ConditionStart is an object representing a Condition when
// we just added a logical operator (AND, OR, ...) and we are
// about to add a predicate.
type {{ .Name }}ConditionStart struct {
	*models.ConditionStart
}

{{ range .Fields }}
// {{ .Name }} adds the "{{ .Name }}" field to the Condition
func (cs {{ $.Name }}ConditionStart) {{ .Name }}() {{ $.Name }}{{ .SanType }}ConditionField {
	return {{ $.Name }}{{ .SanType }}ConditionField{
		ConditionField: cs.Field("{{ .Name }}"),
	}
}

{{ if .IsRS }}
// {{ .Name }}FilteredOn adds a condition with a table join on the given field and
// filters the result with the given condition
func (cs {{ $.Name }}ConditionStart) {{ .Name }}FilteredOn(cond {{ .RelModel }}Condition) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: cs.FilteredOn("{{ .Name }}", cond.Condition),
	}
}
{{ end }}
{{ end }}

// ------- CONDITION FIELDS ----------

{{ range $typ := .Types }}
// A {{ $.Name }}{{ $typ.SanType }}ConditionField is a partial {{ $.Name }}Condition when
// we have selected a field of type {{ $typ.Type }} and expecting an operator.
type {{ $.Name }}{{ $typ.SanType }}ConditionField struct {
	*models.ConditionField
}

{{ range $typ.Operators }}
// {{ .Name }} adds a condition value to the ConditionPath
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) {{ .Name }}(arg {{ if and .Multi (not $typ.IsRS) }}[]{{ end }}{{ $typ.Type }}) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.{{ .Name }}(arg),
	}
}

// {{ .Name }}Func adds a function value to the ConditionPath.
// The function will be evaluated when the query is performed and
// it will be given the RecordSet on which the query is made as parameter
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) {{ .Name }}Func(arg func (models.RecordSet) {{ if and .Multi (not $typ.IsRS) }}[]{{ end }}{{ $typ.Type }}) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.{{ .Name }}(arg),
	}
}

// {{ .Name }}Eval adds an expression value to the ConditionPath.
// The expression value will be evaluated by the client with the
// corresponding execution context. The resulting Condition cannot
// be used server-side.
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) {{ .Name }}Eval(expression string) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.{{ .Name }}(models.ClientEvaluatedString(expression)),
	}
}

{{ end }}

// IsNull checks if the current condition field is null
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) IsNull() {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.IsNull(),
	}
}

// IsNotNull checks if the current condition field is not null
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) IsNotNull() {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.IsNotNull(),
	}
}

// AddOperator adds a condition value to the condition with the given operator and data
// If multi is true, a recordset will be converted into a slice of int64
// otherwise, it will return an int64 and panic if the recordset is not a singleton.
//
// This method is low level and should be avoided. Use operator methods such as Equals() instead.
func (c {{ $.Name }}{{ $typ.SanType }}ConditionField) AddOperator(op operator.Operator, data interface{}) {{ $.Name }}Condition {
	return {{ $.Name }}Condition{
		Condition: c.ConditionField.AddOperator(op, data),
	}
}

{{ end }}

// ------- DATA STRUCT ---------

// {{ .Name }}Data is an autogenerated struct type to handle {{ .Name }} data.
type {{ .Name }}Data struct {
{{ range .Fields }}	{{ .Name }} {{ .Type }} ` + "`json:\"{{ .JSON }}\"`" + `
{{ end }}
}

// FieldMap returns this {{ .Name }}Data as a FieldMap with JSON names as keys.
// Only {{ .Name }}Data with non zero values will be set in the FieldMap.
// To add fields with zero values to the map, give them as fields in parameters.
func (d {{ .Name }}Data) FieldMap(fields ...models.FieldNamer) models.FieldMap {
	res := make(models.FieldMap)
	fieldsMap := make(map[string]bool)
	var fJSON string
	for _, field := range fields {
		fJSON = {{ .Name }}().JSONizeFieldName(field.String())
		fieldsMap[fJSON] = true
	}
{{ range .Fields -}}
	fJSON = {{ $.Name }}().JSONizeFieldName("{{ .Name }}")
	if fieldsMap[fJSON] || !typesutils.IsZero(d.{{ .Name }}) {
		res[fJSON] = d.{{ .Name }}
	}
{{ end }}
	return res
}

// Get returns the value of this {{ .Name }}Data for the given field.
// If the field equals its go zero value, then it returns nil, except if the
// field is given in fieldsToReset, in which case the zero value is returned.
//
// The second argument is true if the field has a non-zero go value or is in the fieldsToReset.
func (d {{ .Name }}Data) Get(field models.FieldNamer, fieldsToReset ...models.FieldNamer) (interface{}, bool) {
	val, exists := d.FieldMap(fieldsToReset...).Get(field.String(), {{ .Name }}().Model)
	if exists {
		return val, true
	}
	return nil, false
}

// Fields returns the list of fields set for update, taking into account the fieldsToReset.
func (d {{ .Name }}Data) FieldsSet(fieldsToReset ...models.FieldNamer) []models.FieldNamer {
	return d.FieldMap(fieldsToReset...).FieldNames()
}

// Remove returns a new {{ .Name }}Data and []FieldNamer with the given field set to its
// zero value and removed from the returned FieldNamer slice.
func (d {{ .Name }}Data) Remove(rs {{ .Name }}Set, field models.FieldNamer, fieldsToReset ...models.FieldNamer) (*{{ .Name }}Data, []models.FieldNamer) {
	fMap := d.FieldMap(fieldsToReset...)
	fMap.Delete(field.String(), {{ .Name }}().Model)
	return rs.DataStruct(fMap)
}

var _ models.FieldMapper = {{ .Name }}Data{}
var _ models.FieldMapper = new({{ .Name }}Data)

// ------- RECORD SET ---------

// {{ .Name }}Set is an autogenerated type to handle {{ .Name }} objects.
type {{ .Name }}Set struct {
	*models.RecordCollection
}

var _ models.RecordSet = {{ .Name }}Set{}

// First returns a copy of the first Record of this RecordSet.
// It returns an empty {{ .Name }} if the RecordSet is empty.
func (s {{ .Name }}Set) First() {{ .Name }}Data {
	var res {{ .Name }}Data
	s.RecordCollection.First(&res)
	return res
}

// All Returns a copy of all records of the RecordCollection.
// It returns an empty slice if the RecordSet is empty.
func (s {{ .Name }}Set) All() []{{ .Name }}Data {
	var ptrSlice []*{{ .Name }}Data
	s.RecordCollection.All(&ptrSlice)
	res := make([]{{ .Name }}Data, len(ptrSlice))
	for i, ps := range ptrSlice {
		res[i] = *ps
	}
	return res
}

// Records returns a slice with all the records of this RecordSet, as singleton
// RecordSets
func (s {{ .Name }}Set) Records() []{{ .Name }}Set {
	recs := s.RecordCollection.Records()
	res := make([]{{ .Name }}Set, len(recs))
	for i, rec := range recs {
		res[i] = {{ .Name }}Set{
			RecordCollection: rec,
		}
	}
	return res
}

// CartesianProduct returns the cartesian product of this {{ .Name }}Set with others.
func (s {{ .Name }}Set) CartesianProduct(others ...{{ .Name }}Set) []{{ .Name }}Set {
	otherSet := make([]models.RecordSet, len(others))
	for i, o := range others {
		otherSet[i] = o
	}
	recs := s.RecordCollection.CartesianProduct(otherSet...)
	res := make([]{{ .Name }}Set, len(recs))
	for i, rec := range recs {
		res[i] = {{ .Name }}Set{
			RecordCollection: rec,
		}
	}
	return res
}

// Model returns an instance of {{ .Name }}Model
func (s {{ .Name }}Set) Model() {{ .Name }}Model {
	return {{ .Name }}Model{
		Model: s.RecordCollection.Model(),
	}
}

{{ range .Fields }}
// {{ .Name }} is a getter for the value of the "{{ .Name }}" field of the first
// record in this RecordSet. It returns the Go zero value if the RecordSet is empty.
func (s {{ $.Name }}Set) {{ .Name }}() {{ .Type }} {
{{ if .IsRS }}	return {{ .Type }}{
		RecordCollection: s.RecordCollection.Get("{{ .Name }}").(*models.RecordCollection),
	}{{ else -}}
	res, _ := s.RecordCollection.Get("{{ .Name }}").({{ .Type }})
	return res {{ end }}
}

// Set{{ .Name }} is a setter for the value of the "{{ .Name }}" field of this
// RecordSet. All Records of this RecordSet will be updated. Each call to this
// method makes an update query in the database.
//
// Set{{ .Name }} panics if the RecordSet is empty.
func (s {{ $.Name }}Set) Set{{ .Name }}(value {{ .Type }}) {
	s.RecordCollection.Set("{{ .Name }}", value)
}
{{ end }}

// Super returns a RecordSet with a modified callstack so that call to the current
// method will execute the next method layer.
//
// This method is meant to be used inside a method layer function to call its parent,
// such as:
//
//    func (rs pool.MyRecordSet) MyMethod() string {
//        res := rs.Super().MyMethod()
//        res += " ok!"
//        return res
//    }
//
// Calls to a different method than the current method will call its next layer only
// if the current method has been called from a layer of the other method. Otherwise,
// it will be the same as calling the other method directly.
func (s {{ .Name }}Set) Super() {{ .Name }}Set {
	return {{ .Name }}Set{
		RecordCollection: s.RecordCollection.Super(),
	}
}

// DataStruct returns a new {{ .Name }}Data object populated with the values
// of the given FieldMap. It returns as a second argument the list of keys of the
// given FieldMap.
func (s {{ .Name }}Set) DataStruct(fMap models.FieldMap) (*{{ .Name }}Data, []models.FieldNamer) {
	var res {{ .Name }}Data
	models.MapToStruct(s.Collection(), &res, fMap)
	return &res, fMap.FieldNames()
}

{{ range .Methods }}
{{ .Doc }}
func (s {{ $.Name }}Set) {{ .Name }}({{ .ParamsWithType }}) ({{ .ReturnString }}) {
{{- if eq .Returns "" }}
	s.Call("{{ .Name }}", {{ .Params}})
{{- else }}
	res := s.{{ .Call }}("{{ .Name }}", {{ .Params}})
	{{ .ReturnAsserts }}
	return {{ .Returns }}
{{- end }}
}

{{ end }}

func init() {
{{- if not .IsModelMixin }}
	models.New{{ .ModelType }}Model("{{ .Name }}")
{{ end }}
{{ range .Methods -}}
{{ if .ToDeclare }}	{{ $.Name }}().AddEmptyMethod("{{ .Name }}")
{{ end -}}
{{- end }}
}
`))
