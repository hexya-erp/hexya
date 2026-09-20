// Copyright 2019 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
)

func TestExtBootStrap(t *testing.T) {
	t.Run("Database creation should run fine", func(t *testing.T) {
		t.Run("Modifying fields before bootstrap", func(t *testing.T) {
			models.UnBootStrap()
			visibilityField := models.Registry.MustGet("ExtPost").Fields().MustGet("Visibility")
			visibilityField.UpdateSelection(types.Selection{"logged_in": "Logged in users"})
			genderField := models.Registry.MustGet("ExtProfile").Fields().MustGet("Gender")
			genderField.SetSelection(types.Selection{"m": "Male", "f": "Female"})
		})
		t.Run("Bootstrap should not panic", func(t *testing.T) {
			models.BootStrap()
			models.SyncDatabase()
		})
		t.Run("Boostrapping twice should panic", func(t *testing.T) {
			assert.True(t, models.BootStrapped())
			assert.Panics(t, models.BootStrap)
		})
		t.Run("Creating methods after bootstrap should panic", func(t *testing.T) {
			assert.Panics(t, func() {
				models.Registry.MustGet("ExtUser").NewMethod("NewMethod", func(rc *models.RecordCollection) {})
			})
		})
		t.Run("Applying DB modifications", func(t *testing.T) {
			models.UnBootStrap()
			contentField := models.Registry.MustGet("ExtPost").Fields().MustGet("Content")
			contentField.SetRequired(false)
			profileField := models.Registry.MustGet("ExtUser").Fields().MustGet("Profile")
			profileField.SetRequired(false)
			numsField := models.Registry.MustGet("ExtUser").Fields().MustGet("Nums")
			numsField.SetDefault(nil).SetIndex(false)
			models.Registry.MustGet("ExtComment").AddFields(map[string]models.FieldDefinition{
				"Date": fields.Date{Default: func(env models.Environment) any {
					return dates.Today()
				}},
			})
			textField := models.Registry.MustGet("ExtComment").Fields().MustGet("Text")
			textField.SetFieldType(fieldtype.Text)
			assert.NotPanics(t, models.BootStrap)
			fInfos := models.Registry.MustGet("ExtPost").FieldsGet(contentField)
			assert.False(t, fInfos[contentField.JSON()].Required)
			fInfos = models.Registry.MustGet("ExtUser").FieldsGet(profileField, numsField)
			assert.False(t, fInfos[profileField.JSON()].Required)
			assert.False(t, fInfos[numsField.JSON()].Index)
			assert.NotPanics(t, models.SyncDatabase)
		})
	})

	t.Run("Post testing models modifications", func(t *testing.T) {
		visibilityField := models.Registry.MustGet("ExtPost").Fields().MustGet("Visibility")
		fInfos := models.Registry.MustGet("ExtPost").FieldsGet(visibilityField)
		assert.Len(t, fInfos[visibilityField.JSON()].Selection, 3)
		assert.Contains(t, fInfos[visibilityField.JSON()].Selection, "visible")
		assert.Contains(t, fInfos[visibilityField.JSON()].Selection, "invisible")
		assert.Contains(t, fInfos[visibilityField.JSON()].Selection, "logged_in")
		genderField := models.Registry.MustGet("ExtProfile").Fields().MustGet("Gender")
		fInfos = models.Registry.MustGet("ExtProfile").FieldsGet(genderField)
		assert.Len(t, fInfos[genderField.JSON()].Selection, 2)
		assert.Contains(t, fInfos[genderField.JSON()].Selection, "m")
		assert.Contains(t, fInfos[genderField.JSON()].Selection, "f")
	})
}
