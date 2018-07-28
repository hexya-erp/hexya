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

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
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

	user.AddMethod("PrefixedUser", "",
		func(rc *models.RecordCollection, prefix string) []string {
			var res []string
			for _, u := range rc.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get("Name")))
			}
			return res
		})

	user.Methods().MustGet("PrefixedUser").Extend("",
		func(rc *models.RecordCollection, prefix string) []string {
			res := rc.Super().Call("PrefixedUser", prefix).([]string)
			for i, u := range rc.Records() {
				email := u.Get("Email").(string)
				res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
			}
			return res
		})

	user.AddMethod("DecorateEmail", "",
		func(rc *models.RecordCollection, email string) string {
			if rc.Env().Context().HasKey("use_square_brackets") {
				return fmt.Sprintf("[%s]", email)
			}
			return fmt.Sprintf("<%s>", email)
		})

	user.Methods().MustGet("DecorateEmail").Extend("",
		func(rc *models.RecordCollection, email string) string {
			if rc.Env().Context().HasKey("use_double_square") {
				rc = rc.
					Call("WithContext", "use_square_brackets", true).(*models.RecordCollection).
					WithContext("fake_key", true)
			}
			res := rc.Super().Call("DecorateEmail", email).(string)
			return fmt.Sprintf("[%s]", res)
		})

	user.AddMethod("OnChangeName", "",
		func(rc *models.RecordCollection) (models.FieldMap, []models.FieldNamer) {
			res := make(models.FieldMap)
			res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
			return res, []models.FieldNamer{models.FieldName("DecoratedName")}
		})

	user.AddMethod("ComputeDecoratedName", "",
		func(rc *models.RecordCollection) models.FieldMap {
			res := make(models.FieldMap)
			res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
			return res
		})

	user.AddMethod("ComputeAge", "",
		func(rc *models.RecordCollection) models.FieldMap {
			res := make(models.FieldMap)
			res["Age"] = rc.Get("Profile").(*models.RecordCollection).Get("Age").(int16)
			return res
		})

	user.AddMethod("InverseSetAge", "",
		func(rc *models.RecordCollection, age int16) {
			rc.Get("Profile").(*models.RecordCollection).Set("Age", age)
		})

	user.AddMethod("UpdateCity", "",
		func(rc *models.RecordCollection, value string) {
			rc.Get("Profile").(*models.RecordCollection).Set("City", value)
		})

	activeMI.AddMethod("IsActivated", "",
		func(rc *models.RecordCollection) bool {
			return rc.Get("Active").(bool)
		})

	addressMI.AddMethod("SayHello", "",
		func(rc *models.RecordCollection) string {
			return "Hello !"
		})

	printAddress := addressMI.AddEmptyMethod("PrintAddress")
	printAddress.DeclareMethod("",
		func(rc *models.RecordCollection) string {
			return fmt.Sprintf("%s, %s %s", rc.Get("Street"), rc.Get("Zip"), rc.Get("City"))
		})

	profile.AddMethod("PrintAddress", "",
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("%s, %s", res, rc.Get("Country"))
		})

	addressMI.Methods().MustGet("PrintAddress").Extend("",
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("<%s>", res)
		})

	profile.Methods().MustGet("PrintAddress").Extend("",
		func(rc *models.RecordCollection) string {
			res := rc.Super().Call("PrintAddress").(string)
			return fmt.Sprintf("[%s]", res)
		})

	post.Methods().MustGet("Create").Extend("",
		func(rc *models.RecordCollection, data models.FieldMapper) *models.RecordCollection {
			res := rc.Super().Call("Create", data).(models.RecordSet).Collection()
			return res
		})

	post.Methods().MustGet("WithContext").Extend("",
		func(rc *models.RecordCollection, key string, value interface{}) *models.RecordCollection {
			return rc.Super().Call("WithContext", key, value).(*models.RecordCollection)
		})

	tag.AddMethod("CheckRate",
		`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
		func(rc *models.RecordCollection) {
			if rc.Get("Rate").(float32) < 0 || rc.Get("Rate").(float32) > 10 {
				log.Panic("Tag rate must be between 0 and 10")
			}
		})

	tag.AddMethod("CheckNameDescription",
		`CheckNameDescription checks that the description of a tag is not equal to its name`,
		func(rc *models.RecordCollection) {
			if rc.Get("Name").(string) == rc.Get("Description").(string) {
				log.Panic("Tag name and description must be different")
			}
		})

	tag.Methods().AllowAllToGroup(security.GroupEveryone)

	user.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().MustGet("OnChangeName")},
		"DecoratedName": models.CharField{Compute: user.Methods().MustGet("ComputeDecoratedName")},
		"Email":         models.CharField{Help: "The user's email address", Size: 100, Index: true},
		"Password":      models.CharField{NoCopy: true},
		"Status": models.IntegerField{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  models.BooleanField{},
		"IsActive": models.BooleanField{},
		"Profile":  models.Many2OneField{RelationModel: models.Registry.MustGet("Profile")},
		"Age": models.IntegerField{Compute: user.Methods().MustGet("ComputeAge"),
			Inverse: user.Methods().MustGet("InverseSetAge"),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
		"Posts":     models.One2ManyField{RelationModel: models.Registry.MustGet("Post"), ReverseFK: "User"},
		"PMoney":    models.FloatField{Related: "Profile.Money"},
		"LastPost":  models.Many2OneField{RelationModel: models.Registry.MustGet("Post"), Embed: true},
		"Email2":    models.CharField{},
		"IsPremium": models.BooleanField{},
		"Nums":      models.IntegerField{GoType: new(int)},
		"Size":      models.FloatField{},
		"Education": models.TextField{String: "Educational Background"},
	})

	profile.AddFields(map[string]models.FieldDefinition{
		"Age":      models.IntegerField{GoType: new(int16)},
		"Gender":   models.SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    models.FloatField{},
		"User":     models.Many2OneField{RelationModel: models.Registry.MustGet("User")},
		"BestPost": models.One2OneField{RelationModel: models.Registry.MustGet("Post")},
		"City":     models.CharField{},
		"Country":  models.CharField{},
	})

	post.AddFields(map[string]models.FieldDefinition{
		"User":            models.Many2OneField{RelationModel: models.Registry.MustGet("User")},
		"Title":           models.CharField{},
		"Content":         models.HTMLField{},
		"Tags":            models.Many2ManyField{RelationModel: models.Registry.MustGet("Tag")},
		"BestPostProfile": models.Rev2OneField{RelationModel: models.Registry.MustGet("Profile"), ReverseFK: "BestPost"},
		"Abstract":        models.TextField{},
		"Attachment":      models.BinaryField{},
		"LastRead":        models.DateField{},
	})

	post.Methods().MustGet("Create").Extend("",
		func(rc *models.RecordCollection, data models.FieldMapper) *models.RecordCollection {
			res := rc.Super().Call("Create", data).(*models.RecordCollection)
			return res
		})

	tag.AddFields(map[string]models.FieldDefinition{
		"Name":        models.CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"BestPost":    models.Many2OneField{RelationModel: models.Registry.MustGet("Post")},
		"Posts":       models.Many2ManyField{RelationModel: models.Registry.MustGet("Post")},
		"Parent":      models.Many2OneField{RelationModel: models.Registry.MustGet("Tag")},
		"Description": models.CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
		"Rate":        models.FloatField{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
	})

	cv.AddFields(map[string]models.FieldDefinition{
		"Education":  models.TextField{},
		"Experience": models.TextField{Translate: true},
		"Leisure":    models.TextField{},
	})

	addressMI.AddFields(map[string]models.FieldDefinition{
		"Street": models.CharField{GoType: new(string)},
		"Zip":    models.CharField{},
		"City":   models.CharField{},
	})
	profile.InheritModel(addressMI)

	activeMI.AddFields(map[string]models.FieldDefinition{
		"Active": models.BooleanField{},
	})

	models.Registry.MustGet("CommonMixin").InheritModel(activeMI)

	viewModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{},
		"City": models.CharField{},
	})
}
