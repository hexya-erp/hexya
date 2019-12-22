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

package testllmodule

import (
	"fmt"
	"log"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
)

func declareModels() {
	user := models.NewModel("User")
	profile := models.NewModel("Profile")
	post := models.NewModel("Post")
	tag := models.NewModel("Tag")
	cv := models.NewModel("Resume")
	addressMI := models.NewMixinModel("AddressMixIn")
	activeMI := models.NewMixinModel("ActiveMixIn")
	viewModel := models.NewManualModel("UserView")

	user.AddMethod("PrefixedUser",
		func(rc *models.RecordCollection, prefix string) []string {
			var res []string
			for _, u := range rc.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get(models.Name)))
			}
			return res
		})

	user.Methods().MustGet("PrefixedUser").Extend(
		func(rc *models.RecordCollection, prefix string) []string {
			res := rc.Super().Call("PrefixedUser", prefix).([]string)
			for i, u := range rc.Records() {
				email := u.Get(rc.Model().FieldName("Email")).(string)
				res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
			}
			return res
		})

	user.AddMethod("DecorateEmail",
		func(rc *models.RecordCollection, email string) string {
			if rc.Env().Context().HasKey("use_square_brackets") {
				return fmt.Sprintf("[%s]", email)
			}
			return fmt.Sprintf("<%s>", email)
		})

	user.Methods().MustGet("DecorateEmail").Extend(
		func(rc *models.RecordCollection, email string) string {
			if rc.Env().Context().HasKey("use_double_square") {
				rc = rc.
					Call("WithContext", "use_square_brackets", true).(*models.RecordCollection).
					WithContext("fake_key", true)
			}
			res := rc.Super().Call("DecorateEmail", email).(string)
			return fmt.Sprintf("[%s]", res)
		})

	user.AddMethod("OnChangeName",
		func(rc *models.RecordCollection) *models.ModelData {
			res := make(models.FieldMap)
			res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
			return models.NewModelDataFromRS(rc, res)
		})

	user.AddMethod("ComputeDecoratedName",
		func(rc *models.RecordCollection) *models.ModelData {
			res := make(models.FieldMap)
			res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
			return models.NewModelDataFromRS(rc, res)
		})

	user.AddMethod("ComputeAge",
		func(rc *models.RecordCollection) *models.ModelData {
			res := make(models.FieldMap)
			res["Age"] = rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Get(rc.Model().FieldName("Age")).(int16)
			return models.NewModelDataFromRS(rc, res)
		})

	user.AddMethod("InverseSetAge",
		func(rc *models.RecordCollection, age int16) {
			rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Set(rc.Model().FieldName("Age"), age)
		})

	user.AddMethod("UpdateCity",
		func(rc *models.RecordCollection, value string) {
			rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Set(rc.Model().FieldName("City"), value)
		})

	activeMI.AddMethod("IsActivated",
		func(rc *models.RecordCollection) bool {
			return rc.Get(rc.Model().FieldName("Active")).(bool)
		})

	addressMI.AddMethod("SayHello",
		func(rc *models.RecordCollection) string {
			return "Hello !"
		})

	addressMI.AddMethod("PrintAddress",
		func(rc *models.RecordCollection) string {
			return fmt.Sprintf("%s, %s %s", rc.Get(rc.Model().FieldName("Street")), rc.Get(rc.Model().FieldName("Zip")), rc.Get(rc.Model().FieldName("City")))
		})

	profile.AddMethod("PrintAddress",
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("%s, %s", res, rc.Get(rc.Model().FieldName("Country")))
		})

	addressMI.Methods().MustGet("PrintAddress").Extend(
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("<%s>", res)
		})

	profile.Methods().MustGet("PrintAddress").Extend(
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("[%s]", res)
		})

	post.Methods().MustGet("Create").Extend(
		func(rc *models.RecordCollection, data models.RecordData) *models.RecordCollection {
			res := rc.Super().Call("Create", data).(models.RecordSet).Collection()
			return res
		})

	post.Methods().MustGet("WithContext").Extend(
		func(rc *models.RecordCollection, key string, value interface{}) *models.RecordCollection {
			return rc.Super().Call("WithContext", key, value).(*models.RecordCollection)
		})

	tag.AddMethod("CheckRate",
		func(rc *models.RecordCollection) {
			if rc.Get(rc.Model().FieldName("Rate")).(float32) < 0 || rc.Get(rc.Model().FieldName("Rate")).(float32) > 10 {
				log.Panic("Tag rate must be between 0 and 10")
			}
		})

	tag.AddMethod("CheckNameDescription",
		func(rc *models.RecordCollection) {
			if rc.Get(rc.Model().FieldName("Name")).(string) == rc.Get(rc.Model().FieldName("Description")).(string) {
				log.Panic("Tag name and description must be different")
			}
		})

	tag.Methods().AllowAllToGroup(security.GroupEveryone)

	user.AddFields(map[string]models.FieldDefinition{
		"Name": fields.Char{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().MustGet("OnChangeName")},
		"DecoratedName": fields.Char{Compute: user.Methods().MustGet("ComputeDecoratedName")},
		"Email":         fields.Char{Help: "The user's email address", Size: 100, Index: true},
		"Password":      fields.Char{NoCopy: true},
		"Status": fields.Integer{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  fields.Boolean{},
		"IsActive": fields.Boolean{},
		"Profile":  fields.Many2One{RelationModel: models.Registry.MustGet("Profile")},
		"Age": fields.Integer{Compute: user.Methods().MustGet("ComputeAge"),
			Inverse: user.Methods().MustGet("InverseSetAge"),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
		"Posts":     fields.One2Many{RelationModel: models.Registry.MustGet("Post"), ReverseFK: "User"},
		"PMoney":    fields.Float{Related: "Profile.Money"},
		"LastPost":  fields.Many2One{RelationModel: models.Registry.MustGet("Post"), Embed: true},
		"Email2":    fields.Char{},
		"IsPremium": fields.Boolean{},
		"Nums":      fields.Integer{GoType: new(int)},
		"Size":      fields.Float{},
		"Education": fields.Text{String: "Educational Background"},
	})

	profile.AddFields(map[string]models.FieldDefinition{
		"Age":      fields.Integer{GoType: new(int16)},
		"Gender":   fields.Selection{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    fields.Float{},
		"User":     fields.Many2One{RelationModel: models.Registry.MustGet("User")},
		"BestPost": fields.One2One{RelationModel: models.Registry.MustGet("Post")},
		"City":     fields.Char{},
		"Country":  fields.Char{},
	})

	post.AddFields(map[string]models.FieldDefinition{
		"User":            fields.Many2One{RelationModel: models.Registry.MustGet("User")},
		"Title":           fields.Char{},
		"Content":         fields.HTML{},
		"Tags":            fields.Many2Many{RelationModel: models.Registry.MustGet("Tag")},
		"BestPostProfile": fields.Rev2One{RelationModel: models.Registry.MustGet("Profile"), ReverseFK: "BestPost"},
		"Abstract":        fields.Text{},
		"Attachment":      fields.Binary{},
		"LastRead":        fields.Date{},
	})

	post.Methods().MustGet("Create").Extend(
		func(rc *models.RecordCollection, data models.RecordData) *models.RecordCollection {
			res := rc.Super().Call("Create", data).(*models.RecordCollection)
			return res
		})

	tag.AddFields(map[string]models.FieldDefinition{
		"Name":        fields.Char{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"BestPost":    fields.Many2One{RelationModel: models.Registry.MustGet("Post")},
		"Posts":       fields.Many2Many{RelationModel: models.Registry.MustGet("Post")},
		"Parent":      fields.Many2One{RelationModel: models.Registry.MustGet("Tag")},
		"Description": fields.Char{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"Rate":        fields.Float{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
	})

	cv.AddFields(map[string]models.FieldDefinition{
		"Education":  fields.Text{},
		"Experience": fields.Text{Translate: true},
		"Leisure":    fields.Text{},
	})

	addressMI.AddFields(map[string]models.FieldDefinition{
		"Street": fields.Char{GoType: new(string)},
		"Zip":    fields.Char{},
		"City":   fields.Char{},
	})
	profile.InheritModel(addressMI)

	activeMI.AddFields(map[string]models.FieldDefinition{
		"Active": fields.Boolean{},
	})

	models.Registry.MustGet("CommonMixin").InheritModel(activeMI)

	viewModel.AddFields(map[string]models.FieldDefinition{
		"Name": fields.Char{},
		"City": fields.Char{},
	})
}
