// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package generate

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestNewModuleInfo(t *testing.T) {
	fSet := token.NewFileSet()
	pack := packages.Package{
		PkgPath: "github.com/hexya-erp/hexya-addons/base",
		Name:    "base",
	}
	modInfo := NewModuleInfo(&pack, Base, fSet)
	assert.Equal(t, "github.com/hexya-erp/hexya-addons/base", modInfo.PkgPath)
	assert.Equal(t, Base, modInfo.ModType)
	assert.Equal(t, fSet, modInfo.FSet)
}

func TestDefaultFields(t *testing.T) {
	t.Run("Any model has an ID field", func(t *testing.T) {
		fields := defaultFields("User")
		assert.Len(t, fields, 1)
		assert.Equal(t, "ID", fields["ID"].Name)
		assert.Equal(t, "id", fields["ID"].JSON)
		assert.Equal(t, "int64", fields["ID"].Type.Type)
		assert.Equal(t, fieldtype.Integer, fields["ID"].FType)
	})
	t.Run("BaseMixin has the audit fields", func(t *testing.T) {
		fields := defaultFields("BaseMixin")
		for _, fName := range []string{"ID", "CreateDate", "CreateUID", "WriteDate", "WriteUID", "LastUpdate", "DisplayName"} {
			assert.Contains(t, fields, fName)
		}
		assert.Equal(t, DatesPath, fields["CreateDate"].Type.ImportPath)
		assert.Equal(t, "__last_update", fields["LastUpdate"].JSON)
	})
	t.Run("ModelMixin has the external id fields", func(t *testing.T) {
		fields := defaultFields("ModelMixin")
		assert.Len(t, fields, 3)
		assert.Equal(t, "hexya_external_id", fields["HexyaExternalID"].JSON)
		assert.Equal(t, fieldtype.Integer, fields["HexyaVersion"].FType)
	})
}

func TestNewModelASTData(t *testing.T) {
	t.Run("Standard model", func(t *testing.T) {
		mData := newModelASTData("User")
		assert.Equal(t, "User", mData.Name)
		assert.False(t, mData.IsModelMixin)
		assert.Contains(t, mData.Fields, "ID")
		assert.NotNil(t, mData.Methods)
		assert.NotNil(t, mData.Mixins)
		assert.NotNil(t, mData.Embeds)
		assert.Empty(t, mData.ModelType)
	})
	t.Run("Mixin model", func(t *testing.T) {
		assert.True(t, newModelASTData("BaseMixin").IsModelMixin)
		assert.True(t, newModelASTData("ModelMixin").IsModelMixin)
	})
}

func TestFormatDocString(t *testing.T) {
	assert.Equal(t, "", formatDocString(""))
	assert.Equal(t, "// MyMethod does something", formatDocString("MyMethod does something"))
	assert.Equal(t, "// MyMethod does something\n// on several lines",
		formatDocString("\n\n  MyMethod does something\n\t on several lines  \n"))
	assert.Equal(t, "// First line\n// \n// Third line", formatDocString("First line\n\nThird line\n"))
}

func TestComputeExportPath(t *testing.T) {
	pkg := types.NewPackage("github.com/hexya-erp/hexya/src/models/types/dates", "dates")
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Date", nil), types.Typ[types.String], nil)
	expected := "github.com/hexya-erp/hexya/src/models/types/dates.Date"

	assert.Equal(t, expected, computeExportPath(named))
	assert.Equal(t, expected, computeExportPath(types.NewPointer(named)))
	assert.Equal(t, expected, computeExportPath(types.NewSlice(named)))
	assert.Equal(t, "", computeExportPath(types.Typ[types.String]))
}

func TestGeneralMixinError(t *testing.T) {
	assert.Equal(t, "General Mixin Error", generalMixinError{}.Error())
}
