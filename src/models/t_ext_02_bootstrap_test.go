// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models_test

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExtBootStrap(t *testing.T) {
	Convey("Database creation should run fine", t, func() {
		Convey("Modifying fields before bootstrap", func() {
			models.UnBootStrap()
			visibilityField := models.Registry.MustGet("ExtPost").Fields().MustGet("Visibility")
			visibilityField.UpdateSelection(types.Selection{"logged_in": "Logged in users"})
			genderField := models.Registry.MustGet("ExtProfile").Fields().MustGet("Gender")
			genderField.SetSelection(types.Selection{"m": "Male", "f": "Female"})
		})
		Convey("Bootstrap should not panic", func() {
			models.BootStrap()
			models.SyncDatabase()
		})
		Convey("Boostrapping twice should panic", func() {
			So(models.BootStrapped(), ShouldBeTrue)
			So(models.BootStrap, ShouldPanic)
		})
		Convey("Creating methods after bootstrap should panic", func() {
			So(func() {
				models.Registry.MustGet("ExtUser").AddMethod("NewMethod", func(rc *models.RecordCollection) {})
			}, ShouldPanic)
		})
		Convey("Applying DB modifications", func() {
			models.UnBootStrap()
			contentField := models.Registry.MustGet("ExtPost").Fields().MustGet("Content")
			contentField.SetRequired(false)
			profileField := models.Registry.MustGet("ExtUser").Fields().MustGet("Profile")
			profileField.SetRequired(false)
			numsField := models.Registry.MustGet("ExtUser").Fields().MustGet("Nums")
			numsField.SetDefault(nil).SetIndex(false)
			models.Registry.MustGet("ExtComment").AddFields(map[string]models.FieldDefinition{
				"Date": fields.Date{Default: func(env models.Environment) interface{} {
					return dates.Today()
				}},
			})
			textField := models.Registry.MustGet("ExtComment").Fields().MustGet("Text")
			textField.SetFieldType(fieldtype.Text)
			So(models.BootStrap, ShouldNotPanic)
			fInfos := models.Registry.MustGet("ExtPost").FieldsGet(contentField)
			So(fInfos[contentField.JSON()].Required, ShouldBeFalse)
			fInfos = models.Registry.MustGet("ExtUser").FieldsGet(profileField, numsField)
			So(fInfos[profileField.JSON()].Required, ShouldBeFalse)
			So(fInfos[numsField.JSON()].Index, ShouldBeFalse)
			So(models.SyncDatabase, ShouldNotPanic)
		})
	})

	Convey("Post testing models modifications", t, func() {
		visibilityField := models.Registry.MustGet("ExtPost").Fields().MustGet("Visibility")
		fInfos := models.Registry.MustGet("ExtPost").FieldsGet(visibilityField)
		So(fInfos[visibilityField.JSON()].Selection, ShouldHaveLength, 3)
		So(fInfos[visibilityField.JSON()].Selection, ShouldContainKey, "visible")
		So(fInfos[visibilityField.JSON()].Selection, ShouldContainKey, "invisible")
		So(fInfos[visibilityField.JSON()].Selection, ShouldContainKey, "logged_in")
		genderField := models.Registry.MustGet("ExtProfile").Fields().MustGet("Gender")
		fInfos = models.Registry.MustGet("ExtProfile").FieldsGet(genderField)
		So(fInfos[genderField.JSON()].Selection, ShouldHaveLength, 2)
		So(fInfos[genderField.JSON()].Selection, ShouldContainKey, "m")
		So(fInfos[genderField.JSON()].Selection, ShouldContainKey, "f")
	})
}
