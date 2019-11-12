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

package models_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExtModelDeclaration(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		userModel := models.NewModel("ExtUser")
		profileModel := models.NewModel("ExtProfile")
		post := models.NewModel("ExtPost")
		tag := models.NewModel("ExtTag")
		cv := models.NewModel("ExtResume")
		comment := models.NewModel("ExtComment")
		addressMI := models.NewMixinModel("ExtAddressMixIn")
		activeMI := models.NewMixinModel("ExtActiveMixIn")
		viewModel := models.NewManualModel("ExtUserView")
		wizard := models.NewTransientModel("ExtWizard")

		userModel.AddMethod("PrefixedUser", "",
			func(rc *models.RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get(u.Model().FieldName("Name"))))
				}
				return res
			})

		userModel.Methods().MustGet("PrefixedUser").Extend("",
			func(rc *models.RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					mail := u.Get(u.Model().FieldName("Email")).(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", mail))
				}
				return res
			})

		userModel.AddMethod("DecorateEmail", "",
			func(rc *models.RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_square_brackets") {
					return fmt.Sprintf("[%s]", email)
				}
				return fmt.Sprintf("<%s>", email)
			})

		userModel.Methods().MustGet("DecorateEmail").Extend("",
			func(rc *models.RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_double_square") {
					rc = rc.
						Call("WithContext", "use_square_brackets", true).(*models.RecordCollection).
						WithContext("fake_key", true)
				}
				res := rc.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		userModel.AddMethod("RecursiveMethod", "",
			func(rc *models.RecordCollection, depth int, res string) string {
				if depth == 0 {
					return res
				}
				return rc.Call("RecursiveMethod", depth-1, fmt.Sprintf("%s, recursion %d", res, depth)).(string)
			})

		userModel.Methods().MustGet("RecursiveMethod").Extend("",
			func(rc *models.RecordCollection, depth int, res string) string {
				res = "> " + res + " <"
				sup := rc.Super().Call("RecursiveMethod", depth, res).(string)
				return sup
			})

		userModel.AddMethod("SubSetSuper", "",
			func(rc *models.RecordCollection) string {
				var res string
				for _, rec := range rc.Records() {
					res += rec.Get(rec.Model().FieldName("Name")).(string)
				}
				return res
			})

		userModel.Methods().MustGet("SubSetSuper").Extend("",
			func(rc *models.RecordCollection) string {
				users := rc.Env().Pool("User")
				emailField := users.Model().FieldName("Email")
				userJane := users.Search(users.Model().Field(emailField).Equals("jane.smith@example.com"))
				userJohn := users.Search(users.Model().Field(emailField).Equals("jsmith2@example.com"))
				users = users.Call("Union", userJane).(models.RecordSet).Collection()
				users = users.Call("Union", userJohn).(models.RecordSet).Collection()
				return users.Super().Call("SubSetSuper").(string)
			})

		userModel.AddMethod("OnChangeName", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("OnChangeNameWarning", "",
			func(rc *models.RecordCollection) string {
				if rc.Get(rc.Model().FieldName("Name")) == "Warning User" {
					return "We have a warning here"
				}
				return ""
			})

		userModel.AddMethod("OnChangeNameFilters", "",
			func(rc *models.RecordCollection) map[models.FieldName]models.Conditioner {
				res := make(map[models.FieldName]models.Conditioner)
				res[rc.Model().FieldName("LastPost")] = models.Registry.MustGet("ExtProfile").Field(models.Registry.MustGet("ExtProfile").FieldName("Street")).Equals("addr")
				return res
			})

		userModel.AddMethod("ComputeDecoratedName", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("ComputeAge", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("Age"), rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Get(models.Registry.MustGet("ExtProfile").FieldName("Age")).(int16))
				return res
			})

		userModel.AddMethod("InverseSetAge", "",
			func(rc *models.RecordCollection, age int16) {
				rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Set(models.Registry.MustGet("ExtProfile").FieldName("Age"), age)
			})

		userModel.AddMethod("UpdateCity", "",
			func(rc *models.RecordCollection, value string) {
				rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Set(models.Registry.MustGet("ExtProfile").FieldName("City"), value)
			})

		userModel.AddMethod("ComputeNum", "Dummy method",
			func(rc *models.RecordCollection) *models.ModelData {
				return models.NewModelData(rc.Model())
			})

		userModel.AddMethod("EndlessRecursion", "Endless recursive method for tests",
			func(rc *models.RecordCollection) string {
				return rc.Call("EndlessRecursion2").(string)
			})

		userModel.AddMethod("EndlessRecursion2", "Endless recursive method for tests",
			func(rc *models.RecordCollection) string {
				return rc.Call("EndlessRecursion").(string)
			})

		userModel.AddMethod("TwoReturnValues", "Test method with 2 return values",
			func(rc *models.RecordCollection) (models.FieldMap, bool) {
				return models.FieldMap{"One": 1}, true
			})

		userModel.AddMethod("NoReturnValue", "Test method with 0 return values",
			func(rc *models.RecordCollection) {
				fmt.Println("NOOP")
			})

		userModel.AddMethod("WrongInverseSetAge", "",
			func(rc *models.RecordCollection, age int16) string {
				rc.Get(rc.Model().FieldName("Profile")).(*models.RecordCollection).Set(models.Registry.MustGet("ExtProfile").FieldName("Age"), age)
				return "Ok"
			})

		userModel.AddMethod("ComputeCoolType", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				if rc.Get(rc.Model().FieldName("IsCool")).(bool) {
					res.Set(rc.Model().FieldName("CoolType"), "cool")
				} else {
					res.Set(rc.Model().FieldName("CoolType"), "no-cool")
				}
				return res
			})

		userModel.AddMethod("OnChangeCoolType", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				if rc.Get(rc.Model().FieldName("CoolType")).(string) == "cool" {
					res.Set(rc.Model().FieldName("IsCool"), true)
				} else {
					res.Set(rc.Model().FieldName("IsCool"), false)
				}
				return res
			})

		userModel.AddMethod("InverseCoolType", "",
			func(rc *models.RecordCollection, val string) {
				if val == "cool" {
					rc.Set(rc.Model().FieldName("IsCool"), true)
				} else {
					rc.Set(rc.Model().FieldName("IsCool"), false)
				}
			})

		userModel.AddMethod("OnChangeMana", "",
			func(rc *models.RecordCollection) *models.ModelData {
				res := models.NewModelData(rc.Model())
				post1 := rc.Env().Pool("Post").SearchAll().Limit(1)
				prof := rc.Env().Pool("Profile").Call("Create",
					models.NewModelData(models.Registry.MustGet("Profile")).
						Set(models.Registry.MustGet("ExtProfile").FieldName("BestPost"), post1))
				prof.(models.RecordSet).Collection().InvalidateCache()
				res.Set(rc.Model().FieldName("Profile"), prof)
				return res
			})

		userModel.Methods().MustGet("Copy").Extend("",
			func(rc *models.RecordCollection, overrides models.RecordData) *models.RecordCollection {
				nameField := rc.Model().FieldName("Name")
				overrides.Underlying().Set(nameField, fmt.Sprintf("%s (copy)", rc.Get(nameField).(string)))
				return rc.Super().Call("Copy", overrides).(models.RecordSet).Collection()
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc *models.RecordCollection) bool {
				return rc.Get(rc.Model().FieldName("Active")).(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc *models.RecordCollection) string {
				return "Hello !"
			})

		addressMI.AddMethod("PrintAddress", "",
			func(rc *models.RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get(rc.Model().FieldName("Street")), rc.Get(rc.Model().FieldName("Zip")), rc.Get(rc.Model().FieldName("City")))
			})

		profileModel.AddMethod("PrintAddress", "",
			func(rc *models.RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get(rc.Model().FieldName("Country")))
			})

		addressMI.Methods().MustGet("PrintAddress").Extend("",
			func(rc *models.RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("<%s>", res)
			})

		profileModel.Methods().MustGet("PrintAddress").Extend("",
			func(rc *models.RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("[%s]", res)
			})

		post.AddMethod("ComputeRead", "",
			func(rc *models.RecordCollection) *models.ModelData {
				var read bool
				if !rc.Get(rc.Model().FieldName("LastRead")).(dates.Date).IsZero() {
					read = true
				}
				return models.NewModelData(rc.Model()).Set(rc.Model().FieldName("Read"), read)
			})

		post.Methods().MustGet("Create").Extend("",
			func(rc *models.RecordCollection, data models.RecordData) *models.RecordCollection {
				res := rc.Super().Call("Create", data).(models.RecordSet).Collection()
				return res
			})

		post.Methods().MustGet("Search").Extend("",
			func(rc *models.RecordCollection, cond models.Conditioner) *models.RecordCollection {
				res := rc.Super().Call("Search", cond).(models.RecordSet).Collection()
				return res
			})

		post.Methods().MustGet("WithContext").Extend("",
			func(rc *models.RecordCollection, key string, value interface{}) *models.RecordCollection {
				return rc.Super().Call("WithContext", key, value).(*models.RecordCollection)
			})

		post.AddMethod("ComputeTagsNames", "",
			func(rc *models.RecordCollection) *models.ModelData {
				var res string
				for _, rec := range rc.Records() {
					for _, tg := range rec.Get(rec.Model().FieldName("Tags")).(models.RecordSet).Collection().Records() {
						res += tg.Get(tg.Model().FieldName("Name")).(string) + " "
					}
				}
				return models.NewModelData(rc.Model()).Set(rc.Model().FieldName("TagsNames"), res)
			})

		post.AddMethod("ComputeWriterAge", "",
			func(rc *models.RecordCollection) *models.ModelData {
				return models.NewModelData(rc.Model()).
					Set(rc.Model().FieldName("WriterAge"),
						rc.Get(rc.Model().FieldName("User")).(models.RecordSet).Collection().Get(models.Registry.MustGet("ExtUser").FieldName("Age")).(int16))
			})

		tag.AddMethod("CheckRate",
			`CheckRate checks that the given models.RecordSet has a rate between 0 and 10`,
			func(rc *models.RecordCollection) {
				if rc.Get(rc.Model().FieldName("Rate")).(float32) < 0 || rc.Get(rc.Model().FieldName("Rate")).(float32) > 10 {
					log.Panic("Tag rate must be between 0 and 10")
				}
			})

		tag.AddMethod("CheckNameDescription",
			`CheckNameDescription checks that the description of a tag is not equal to its name`,
			func(rc *models.RecordCollection) {
				if rc.Get(rc.Model().FieldName("Name")).(string) == rc.Get(rc.Model().FieldName("Description")).(string) {
					log.Panic("Tag name and description must be different")
				}
			})

		tag.Methods().AllowAllToGroup(security.GroupEveryone)
		tag.Methods().RevokeAllFromGroup(security.GroupEveryone)
		tag.Methods().AllowAllToGroup(security.GroupEveryone)

		cv.AddMethod("ComputeOther",
			`Dummy compute function`,
			func(rc *models.RecordCollection) *models.ModelData {
				return models.NewModelData(rc.Model()).Set(rc.Model().FieldName("Other"), "Other information")
			})

		userModel.AddFields(map[string]models.FieldDefinition{
			"Name": fields.Char{String: "Name", Help: "The user's username", Unique: true,
				NoCopy: true, OnChange: userModel.Methods().MustGet("OnChangeName"),
				OnChangeFilters: userModel.Methods().MustGet("OnChangeNameFilters"),
				OnChangeWarning: userModel.Methods().MustGet("OnChangeNameWarning"),
			},
			"DecoratedName": fields.Char{Compute: userModel.Methods().MustGet("ComputeDecoratedName")},
			"Email":         fields.Char{Help: "The user's email address", Size: 100, Index: true},
			"Password":      fields.Char{NoCopy: true},
			"Status": fields.Integer{JSON: "status_json", GoType: new(int16),
				Default: models.DefaultValue(int16(12)), ReadOnly: true},
			"IsStaff":  fields.Boolean{},
			"IsActive": fields.Boolean{},
			"Profile": fields.One2One{RelationModel: models.Registry.MustGet("ExtProfile"),
				OnDelete: models.SetNull, Required: true},
			"Age": fields.Integer{Compute: userModel.Methods().MustGet("ComputeAge"),
				Inverse: userModel.Methods().MustGet("InverseSetAge"),
				Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
			"Posts":    fields.One2Many{RelationModel: models.Registry.MustGet("ExtPost"), ReverseFK: "User", Copy: true},
			"PMoney":   fields.Float{Related: "Profile.Money"},
			"LastPost": fields.Many2One{RelationModel: models.Registry.MustGet("ExtPost")},
			"Resume":   fields.Many2One{RelationModel: models.Registry.MustGet("ExtResume"), Embed: true},
			"IsCool":   fields.Boolean{},
			"CoolType": fields.Selection{Selection: types.Selection{
				"cool":    "Yes, its a cool user",
				"no-cool": "No, forget it"},
				Compute:  userModel.Methods().MustGet("ComputeCoolType"),
				Inverse:  userModel.Methods().MustGet("InverseCoolType"),
				OnChange: userModel.Methods().MustGet("OnChangeCoolType")},
			"Email2":          fields.Char{},
			"IsPremium":       fields.Boolean{},
			"Nums":            fields.Integer{GoType: new(int)},
			"Size":            fields.Float{},
			"BestProfilePost": fields.Many2One{RelationModel: models.Registry.MustGet("ExtPost"), Related: "Profile.BestPost"},
			"Mana":            fields.Float{GoType: new(float32), OnChange: userModel.Methods().MustGet("OnChangeMana")},
			"Education":       fields.Text{String: "Educational Background"},
		})
		userModel.AddSQLConstraint("nums_premium", "CHECK((is_premium = TRUE AND nums IS NOT NULL AND nums > 0) OR (IS_PREMIUM = false))",
			"Premium users must have positive nums")

		profileModel.AddFields(map[string]models.FieldDefinition{
			"Age":      fields.Integer{GoType: new(int16)},
			"Gender":   fields.Selection{Selection: types.Selection{"male": "Male", "female": "Female"}},
			"Money":    fields.Float{},
			"User":     fields.Rev2One{RelationModel: models.Registry.MustGet("ExtUser"), ReverseFK: "Profile"},
			"BestPost": fields.Many2One{RelationModel: models.Registry.MustGet("ExtPost")},
			"City":     fields.Char{},
			"Country":  fields.Char{},
			"UserName": fields.Char{Related: "User.Name"},
		})

		post.AddFields(map[string]models.FieldDefinition{
			"User":       fields.Many2One{RelationModel: models.Registry.MustGet("ExtUser")},
			"Title":      fields.Char{Required: true},
			"Content":    fields.HTML{Required: true},
			"Tags":       fields.Many2Many{RelationModel: models.Registry.MustGet("ExtTag")},
			"Abstract":   fields.Text{},
			"Attachment": fields.Binary{},
			"Read":       fields.Boolean{Compute: models.Registry.MustGet("ExtPost").Methods().MustGet("ComputeRead")},
			"LastRead":   fields.Date{},
			"Visibility": fields.Selection{Selection: types.Selection{
				"invisible": "Invisible",
				"visible":   "Visible",
			}},
			"Comments":        fields.One2Many{RelationModel: models.Registry.MustGet("ExtComment"), ReverseFK: "Post"},
			"LastCommentText": fields.Text{Related: "Comments.Text"},
			"LastTagName":     fields.Char{Related: "Tags.Name"},
			"TagsNames":       fields.Char{Compute: models.Registry.MustGet("ExtPost").Methods().MustGet("ComputeTagsNames")},
			"WriterAge": fields.Integer{Compute: post.Methods().MustGet("ComputeWriterAge"),
				Depends: []string{"User.Age"}, Stored: true, GoType: new(int16)},
			"WriterMoney": fields.Float{Related: "User.PMoney"},
		})
		post.SetDefaultOrder("Title")

		comment.AddFields(map[string]models.FieldDefinition{
			"Post":        fields.Many2One{RelationModel: models.Registry.MustGet("ExtPost")},
			"PostWriter":  fields.Many2One{RelationModel: models.Registry.MustGet("ExtUser"), Related: "Post.User"},
			"WriterMoney": fields.Float{Related: "PostWriter.PMoney"},
			"Text":        fields.Char{},
		})

		tag.AddFields(map[string]models.FieldDefinition{
			"Name":        fields.Char{Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"BestPost":    fields.Many2One{RelationModel: models.Registry.MustGet("ExtPost")},
			"Posts":       fields.Many2Many{RelationModel: models.Registry.MustGet("ExtPost")},
			"Parent":      fields.Many2One{RelationModel: models.Registry.MustGet("ExtTag")},
			"Description": fields.Char{Translate: true, Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"Note":        fields.Char{Translate: true, Required: true, Default: models.DefaultValue("Default Note")},
			"Rate":        fields.Float{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
		})
		tag.SetDefaultOrder("Name DESC", "ID ASC")

		cv.AddFields(map[string]models.FieldDefinition{
			"Education":  fields.Char{},
			"Experience": fields.Text{Translate: true},
			"Leisure":    fields.Text{},
			"Other":      fields.Char{Compute: cv.Methods().MustGet("ComputeOther")},
		})

		addressMI.AddFields(map[string]models.FieldDefinition{
			"Street": fields.Char{GoType: new(string)},
			"Zip":    fields.Char{},
			"City":   fields.Char{},
		})

		profileModel.InheritModel(addressMI)

		activeMI.AddFields(map[string]models.FieldDefinition{
			"Active": fields.Boolean{Default: models.DefaultValue(true)},
		})

		models.Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel.AddFields(map[string]models.FieldDefinition{
			"Name": fields.Char{},
			"City": fields.Char{},
		})

		wizard.AddFields(map[string]models.FieldDefinition{
			"Name":  fields.Char{},
			"Value": fields.Integer{},
		})
	})
}

func TestExtErroneousDeclarations(t *testing.T) {
	Convey("Testing wrong field declarations", t, func() {
		Convey("Ours = Theirs in M2M field def", func() {
			userModel := models.Registry.MustGet("User")
			So(func() {
				userModel.AddFields(map[string]models.FieldDefinition{
					"Tags": fields.Many2Many{RelationModel: models.Registry.MustGet("ExtTag"),
						M2MOurField: "FT", M2MTheirField: "FT"},
				})
			}, ShouldPanic)
		})
	})
}
