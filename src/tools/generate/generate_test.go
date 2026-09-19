// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package generate

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/stretchr/testify/assert"
)

func TestCreateTypeIdent(t *testing.T) {
	assert.Equal(t, "String", createTypeIdent("string"))
	assert.Equal(t, "DatesDate", createTypeIdent("dates.Date"))
	assert.Equal(t, "Sliceint64", createTypeIdent("[]int64"))
	assert.Equal(t, "Mapstringinterface{}", createTypeIdent("map[string]interface{}"))
}

func TestTrimInterfacePackagePrefix(t *testing.T) {
	assert.Equal(t, "UserSet", trimInterfacePackagePrefix("m.UserSet"))
	assert.Equal(t, "string", trimInterfacePackagePrefix("string"))
	assert.Equal(t, "[]UserSet", trimInterfacePackagePrefix("[]m.UserSet"))
	assert.Equal(t, "map[string]UserSet", trimInterfacePackagePrefix("map[string]m.UserSet"))
}

func TestIsRecordSetType(t *testing.T) {
	models := map[string]ModelASTData{
		"User": newModelASTData("User"),
	}
	t.Run("Generic RecordSet types", func(t *testing.T) {
		for _, typ := range []string{"*RecordCollection", "*models.RecordCollection", "RecordSet", "models.RecordSet"} {
			isRS, generic := isRecordSetType(typ, models)
			assert.True(t, isRS, typ)
			assert.True(t, generic, typ)
		}
	})
	t.Run("Specific RecordSet types", func(t *testing.T) {
		isRS, generic := isRecordSetType("UserSet", models)
		assert.True(t, isRS)
		assert.False(t, generic)
	})
	t.Run("Other types", func(t *testing.T) {
		isRS, generic := isRecordSetType("string", models)
		assert.False(t, isRS)
		assert.False(t, generic)
		isRS, _ = isRecordSetType("PartnerSet", models)
		assert.False(t, isRS)
	})
}

func TestModelDataSort(t *testing.T) {
	mData := modelData{
		Deps:       []string{"z", "a"},
		RelModels:  []string{"Partner", "Currency"},
		Fields:     []fieldData{{Name: "Name"}, {Name: "Age"}},
		Methods:    []methodData{{Name: "Write"}, {Name: "Create"}},
		AllMethods: []methodData{{Name: "Write"}, {Name: "Create"}},
		Types:      []fieldType{{Type: "string"}, {Type: "int64"}},
	}
	mData.sort()
	assert.Equal(t, []string{"a", "z"}, mData.Deps)
	assert.Equal(t, []string{"Currency", "Partner"}, mData.RelModels)
	assert.Equal(t, "Age", mData.Fields[0].Name)
	assert.Equal(t, "Create", mData.Methods[0].Name)
	assert.Equal(t, "Create", mData.AllMethods[0].Name)
	assert.Equal(t, "int64", mData.Types[0].Type)
}

func TestAddFieldsToModelData(t *testing.T) {
	mASTData := ModelASTData{
		Name: "User",
		Fields: map[string]FieldASTData{
			"Name": {
				Name:  "Name",
				Type:  TypeData{Type: "string"},
				FType: fieldtype.Char,
			},
			"Profile": {
				Name:     "Profile",
				Type:     TypeData{Type: "m.ProfileSet"},
				RelModel: "Profile",
				IsRS:     true,
				FType:    fieldtype.Many2One,
			},
		},
	}
	mData := modelData{Name: "User"}
	depsMap := map[string]bool{}
	addFieldsToModelData(mASTData, &mData, &depsMap)
	mData.sort()

	assert.Len(t, mData.Fields, 2)
	assert.Equal(t, "Name", mData.Fields[0].Name)
	assert.Equal(t, "name", mData.Fields[0].JSON)
	assert.Equal(t, "string", mData.Fields[0].Type)
	assert.Equal(t, "String", mData.Fields[0].SanType)

	assert.Equal(t, "Profile", mData.Fields[1].Name)
	assert.Equal(t, "profile_id", mData.Fields[1].JSON)
	assert.Equal(t, "m.ProfileSet", mData.Fields[1].Type)
	assert.Equal(t, "ProfileSet", mData.Fields[1].IType)
	assert.Equal(t, []string{"Profile"}, mData.RelModels)
}

func TestAddFieldTypesToModelData(t *testing.T) {
	mData := modelData{
		Fields: []fieldData{
			{Name: "Name", IType: "string", SanType: "String"},
			{Name: "Other", IType: "string", SanType: "String"},
			{Name: "Profile", IType: "ProfileSet", SanType: "MProfileSet", IsRS: true, ImportPath: DatesPath},
		},
	}
	addFieldTypesToModelData(&mData)
	assert.Len(t, mData.Types, 2)
	mData.sort()
	assert.Equal(t, "ProfileSet", mData.Types[0].Type)
	assert.True(t, mData.Types[0].IsRS)
	assert.Equal(t, "string", mData.Types[1].Type)
	assert.NotEmpty(t, mData.Types[0].Operators)
	// Empty import paths are not added to the deps
	assert.Equal(t, []string{DatesPath}, mData.TypesDeps)
}

func TestAddMethodsToModelData(t *testing.T) {
	modelsASTData := map[string]ModelASTData{
		"User": {
			Name: "User",
			Methods: map[string]MethodASTData{
				"NoReturn": {
					Name:   "NoReturn",
					Doc:    "// NoReturn does nothing",
					Params: []ParamData{{Name: "value", Type: TypeData{Type: "string"}}},
				},
				"SingleReturn": {
					Name:    "SingleReturn",
					Params:  []ParamData{{Name: "values", Variadic: true, Type: TypeData{Type: "int64"}}},
					Returns: []TypeData{{Type: "string"}},
				},
				"MultiReturn": {
					Name:    "MultiReturn",
					Returns: []TypeData{{Type: "string"}, {Type: "int64"}},
				},
				"RecordSetReturn": {
					Name:    "RecordSetReturn",
					Returns: []TypeData{{Type: "UserSet"}},
				},
			},
		},
	}
	mData := modelData{Name: "User"}
	depsMap := map[string]bool{}
	addMethodsToModelData(modelsASTData, &mData, &depsMap)
	mData.sort()
	assert.Len(t, mData.Methods, 4)

	methods := make(map[string]methodData)
	for _, m := range mData.Methods {
		methods[m.Name] = m
	}

	assert.Equal(t, "value string", methods["NoReturn"].ParamsWithType)
	assert.Equal(t, "value", methods["NoReturn"].Params)
	assert.Equal(t, "", methods["NoReturn"].Call)

	assert.Equal(t, "values ...int64", methods["SingleReturn"].ParamsWithType)
	assert.Equal(t, "Call", methods["SingleReturn"].Call)
	assert.Equal(t, "string", methods["SingleReturn"].ReturnString)
	assert.Equal(t, "resTyped", methods["SingleReturn"].Returns)

	assert.Equal(t, "CallMulti", methods["MultiReturn"].Call)
	assert.Equal(t, "string,int64", methods["MultiReturn"].ReturnString)
	assert.Equal(t, "resTyped0,resTyped1", methods["MultiReturn"].Returns)

	assert.Equal(t, "m.UserSet", methods["RecordSetReturn"].ReturnString)
	assert.Contains(t, methods["RecordSetReturn"].ReturnAsserts, `Wrap("User")`)
}

func TestAddMethodsToModelDataWithSpecificHandler(t *testing.T) {
	modelsASTData := map[string]ModelASTData{
		"User": {
			Name: "User",
			Methods: map[string]MethodASTData{
				"Search": {Name: "Search"},
			},
		},
	}
	mData := modelData{Name: "User"}
	depsMap := map[string]bool{}
	addMethodsToModelData(modelsASTData, &mData, &depsMap)
	assert.Len(t, mData.Methods, 1)
	assert.Equal(t, "Search", mData.Methods[0].Name)
	assert.Equal(t, "m.UserSet", mData.Methods[0].ReturnString)
	assert.Equal(t, "UserSet", mData.AllMethods[0].IReturnString)
}

func TestSpecificMethodsHandlers(t *testing.T) {
	for methodName := range specificMethodsHandlers {
		t.Run(methodName, func(t *testing.T) {
			modelsASTData := map[string]ModelASTData{
				"User": {
					Name:    "User",
					Methods: map[string]MethodASTData{methodName: {Name: methodName}},
				},
			}
			mData := modelData{Name: "User"}
			depsMap := map[string]bool{}
			addMethodsToModelData(modelsASTData, &mData, &depsMap)
			// All handlers declare the method in the model interface. Some of them
			// (e.g. CartesianProduct) don't generate an implementation.
			assert.Len(t, mData.AllMethods, 1)
			assert.Equal(t, methodName, mData.AllMethods[0].Name)
			assert.NotEmpty(t, mData.AllMethods[0].ReturnString)
			assert.NotEmpty(t, mData.AllMethods[0].IReturnString)
			for _, m := range mData.Methods {
				assert.Equal(t, methodName, m.Name)
				assert.NotEmpty(t, m.Call)
				assert.NotEmpty(t, m.ReturnString)
			}
		})
	}
}

func TestCreatePoolFiles(t *testing.T) {
	dir := t.TempDir()
	for _, pack := range []string{PoolInterfacesPackage, PoolModelPackage, PoolQueryPackage} {
		assert.Nil(t, os.MkdirAll(filepath.Join(dir, pack), 0755))
	}
	mData := modelData{
		Name:                  "User",
		SnakeName:             "user",
		ModelsPackageName:     PoolModelPackage,
		QueryPackageName:      PoolQueryPackage,
		InterfacesPackageName: PoolInterfacesPackage,
		ModelType:             "",
		ConditionFuncs:        []string{"And", "AndNot", "Or", "OrNot"},
		Deps:                  []string{ModelsPath},
		Fields: []fieldData{
			{Name: "Name", JSON: "name", Type: "string", IType: "string", SanType: "String"},
		},
	}
	addFieldTypesToModelData(&mData)
	assert.NotPanics(t, func() { createPoolFiles(dir, &mData) })

	for _, fileName := range []string{
		filepath.Join(PoolInterfacesPackage, "user.go"),
		filepath.Join(PoolModelPackage, "user.go"),
		filepath.Join(PoolModelPackage, "user", "user.go"),
		filepath.Join(PoolQueryPackage, "user.go"),
		filepath.Join(PoolQueryPackage, "user", "user.go"),
	} {
		content, err := os.ReadFile(filepath.Join(dir, fileName))
		assert.Nil(t, err, fileName)
		assert.Contains(t, string(content), "User", fileName)
	}
}

func TestCreateFileFromTemplate(t *testing.T) {
	tmpl := template.Must(template.New("test").Parse("package {{ .Package }}\n\nvar  {{ .Var }}    =   1\n"))
	t.Run("Valid template", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "test.go")
		CreateFileFromTemplate(fileName, tmpl, struct {
			Package string
			Var     string
		}{Package: "mypack", Var: "myVar"})
		content, err := os.ReadFile(fileName)
		assert.Nil(t, err)
		// The generated source must have been formatted with gofmt
		assert.Equal(t, "package mypack\n\nvar myVar = 1\n", string(content))
	})
	t.Run("Invalid generated source should panic", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "test.go")
		assert.Panics(t, func() {
			CreateFileFromTemplate(fileName, tmpl, struct {
				Package string
				Var     string
			}{Package: "mypack", Var: "not a valid ident"})
		})
	})
}
