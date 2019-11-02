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

package models

import (
	"fmt"
	"testing"

	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
	"github.com/hexya-erp/hexya/src/tools/nbutils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestModelDeclaration(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		userModel := NewModel("User")
		profileModel := NewModel("Profile")
		post := NewModel("Post")
		tag := NewModel("Tag")
		cv := NewModel("Resume")
		comment := NewModel("Comment")
		addressMI := NewMixinModel("AddressMixIn")
		activeMI := NewMixinModel("ActiveMixIn")
		viewModel := NewManualModel("UserView")
		wizard := NewTransientModel("Wizard")

		userModel.AddMethod("PrefixedUser", "",
			func(rc *RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get(u.model.FieldName("Name"))))
				}
				return res
			})

		userModel.Methods().MustGet("PrefixedUser").Extend("",
			func(rc *RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					mail := u.Get(u.model.FieldName("Email")).(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", mail))
				}
				return res
			})

		userModel.AddMethod("DecorateEmail", "",
			func(rc *RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_square_brackets") {
					return fmt.Sprintf("[%s]", email)
				}
				return fmt.Sprintf("<%s>", email)
			})

		userModel.Methods().MustGet("DecorateEmail").Extend("",
			func(rc *RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_double_square") {
					rc = rc.
						Call("WithContext", "use_square_brackets", true).(*RecordCollection).
						WithContext("fake_key", true)
				}
				res := rc.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		userModel.AddMethod("RecursiveMethod", "",
			func(rc *RecordCollection, depth int, res string) string {
				if depth == 0 {
					return res
				}
				return rc.Call("RecursiveMethod", depth-1, fmt.Sprintf("%s, recursion %d", res, depth)).(string)
			})

		userModel.Methods().MustGet("RecursiveMethod").Extend("",
			func(rc *RecordCollection, depth int, res string) string {
				res = "> " + res + " <"
				sup := rc.Super().Call("RecursiveMethod", depth, res).(string)
				return sup
			})

		userModel.AddMethod("SubSetSuper", "",
			func(rc *RecordCollection) string {
				var res string
				for _, rec := range rc.Records() {
					res += rec.Get(rec.model.FieldName("Name")).(string)
				}
				return res
			})

		userModel.Methods().MustGet("SubSetSuper").Extend("",
			func(rc *RecordCollection) string {
				users := rc.Env().Pool("User")
				emailField := users.model.FieldName("Email")
				userJane := users.Search(users.Model().Field(emailField).Equals("jane.smith@example.com"))
				userJohn := users.Search(users.Model().Field(emailField).Equals("jsmith2@example.com"))
				users = users.Call("Union", userJane).(RecordSet).Collection()
				users = users.Call("Union", userJohn).(RecordSet).Collection()
				return users.Super().Call("SubSetSuper").(string)
			})

		userModel.AddMethod("OnChangeName", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.model.FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("OnChangeNameWarning", "",
			func(rc *RecordCollection) string {
				if rc.Get(rc.model.FieldName("Name")) == "Warning User" {
					return "We have a warning here"
				}
				return ""
			})

		userModel.AddMethod("OnChangeNameFilters", "",
			func(rc *RecordCollection) map[FieldName]Conditioner {
				res := make(map[FieldName]Conditioner)
				res[rc.model.FieldName("LastPost")] = Registry.MustGet("Profile").Field(Registry.MustGet("Profile").FieldName("Street")).Equals("addr")
				return res
			})

		userModel.AddMethod("ComputeDecoratedName", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.model.FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("ComputeAge", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.model.FieldName("Age"), rc.Get(rc.model.FieldName("Profile")).(*RecordCollection).Get(Registry.MustGet("Profile").FieldName("Age")).(int16))
				return res
			})

		userModel.AddMethod("InverseSetAge", "",
			func(rc *RecordCollection, age int16) {
				rc.Get(rc.model.FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("Age"), age)
			})

		userModel.AddMethod("UpdateCity", "",
			func(rc *RecordCollection, value string) {
				rc.Get(rc.model.FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("City"), value)
			})

		userModel.AddMethod("ComputeNum", "Dummy method",
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.model)
			})

		userModel.AddMethod("EndlessRecursion", "Endless recursive method for tests",
			func(rc *RecordCollection) string {
				return rc.Call("EndlessRecursion2").(string)
			})

		userModel.AddMethod("EndlessRecursion2", "Endless recursive method for tests",
			func(rc *RecordCollection) string {
				return rc.Call("EndlessRecursion").(string)
			})

		userModel.AddMethod("TwoReturnValues", "Test method with 2 return values",
			func(rc *RecordCollection) (FieldMap, bool) {
				return FieldMap{"One": 1}, true
			})

		userModel.AddMethod("NoReturnValue", "Test method with 0 return values",
			func(rc *RecordCollection) {
				fmt.Println("NOOP")
			})

		userModel.AddMethod("WrongInverseSetAge", "",
			func(rc *RecordCollection, age int16) string {
				rc.Get(rc.model.FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("Age"), age)
				return "Ok"
			})

		userModel.AddMethod("ComputeCoolType", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.model)
				if rc.Get(rc.model.FieldName("IsCool")).(bool) {
					res.Set(rc.model.FieldName("CoolType"), "cool")
				} else {
					res.Set(rc.model.FieldName("CoolType"), "no-cool")
				}
				return res
			})

		userModel.AddMethod("OnChangeCoolType", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.model)
				if rc.Get(rc.model.FieldName("CoolType")).(string) == "cool" {
					res.Set(rc.model.FieldName("IsCool"), true)
				} else {
					res.Set(rc.model.FieldName("IsCool"), false)
				}
				return res
			})

		userModel.AddMethod("InverseCoolType", "",
			func(rc *RecordCollection, val string) {
				if val == "cool" {
					rc.Set(rc.model.FieldName("IsCool"), true)
				} else {
					rc.Set(rc.model.FieldName("IsCool"), false)
				}
			})

		userModel.AddMethod("OnChangeMana", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.model)
				post1 := rc.env.Pool("Post").SearchAll().Limit(1)
				prof := rc.env.Pool("Profile").Call("Create",
					NewModelData(Registry.MustGet("Profile")).
						Set(Registry.MustGet("Profile").FieldName("BestPost"), post1))
				prof.(RecordSet).Collection().InvalidateCache()
				res.Set(rc.model.FieldName("Profile"), prof)
				return res
			})

		userModel.Methods().MustGet("Copy").Extend("",
			func(rc *RecordCollection, overrides RecordData) *RecordCollection {
				nameField := rc.model.FieldName("Name")
				overrides.Underlying().Set(nameField, fmt.Sprintf("%s (copy)", rc.Get(nameField).(string)))
				return rc.Super().Call("Copy", overrides).(RecordSet).Collection()
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc *RecordCollection) bool {
				return rc.Get(rc.model.FieldName("Active")).(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc *RecordCollection) string {
				return "Hello !"
			})

		printAddress := addressMI.AddEmptyMethod("PrintAddress")
		printAddress.DeclareMethod("",
			func(rc *RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get(rc.model.FieldName("Street")), rc.Get(rc.model.FieldName("Zip")), rc.Get(rc.model.FieldName("City")))
			})

		profileModel.AddMethod("PrintAddress", "",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get(rc.model.FieldName("Country")))
			})

		addressMI.Methods().MustGet("PrintAddress").Extend("",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("<%s>", res)
			})

		profileModel.Methods().MustGet("PrintAddress").Extend("",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("[%s]", res)
			})

		post.AddMethod("ComputeRead", "",
			func(rc *RecordCollection) *ModelData {
				var read bool
				if !rc.Get(rc.model.FieldName("LastRead")).(dates.Date).IsZero() {
					read = true
				}
				return NewModelData(rc.model).Set(rc.model.FieldName("Read"), read)
			})

		post.Methods().MustGet("Create").Extend("",
			func(rc *RecordCollection, data RecordData) *RecordCollection {
				res := rc.Super().Call("Create", data).(RecordSet).Collection()
				return res
			})

		post.Methods().MustGet("Search").Extend("",
			func(rc *RecordCollection, cond Conditioner) *RecordCollection {
				res := rc.Super().Call("Search", cond).(RecordSet).Collection()
				return res
			})

		post.Methods().MustGet("WithContext").Extend("",
			func(rc *RecordCollection, key string, value interface{}) *RecordCollection {
				return rc.Super().Call("WithContext", key, value).(*RecordCollection)
			})

		post.AddMethod("ComputeTagsNames", "",
			func(rc *RecordCollection) *ModelData {
				var res string
				for _, rec := range rc.Records() {
					for _, tg := range rec.Get(rec.model.FieldName("Tags")).(RecordSet).Collection().Records() {
						res += tg.Get(tg.model.FieldName("Name")).(string) + " "
					}
				}
				return NewModelData(rc.model).Set(rc.model.FieldName("TagsNames"), res)
			})

		post.AddMethod("ComputeWriterAge", "",
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.model).
					Set(rc.model.FieldName("WriterAge"),
						rc.Get(rc.model.FieldName("User")).(RecordSet).Collection().Get(Registry.MustGet("User").FieldName("Age")).(int16))
			})

		tag.AddMethod("CheckRate",
			`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
			func(rc *RecordCollection) {
				if rc.Get(rc.model.FieldName("Rate")).(float32) < 0 || rc.Get(rc.model.FieldName("Rate")).(float32) > 10 {
					log.Panic("Tag rate must be between 0 and 10")
				}
			})

		tag.AddMethod("CheckNameDescription",
			`CheckNameDescription checks that the description of a tag is not equal to its name`,
			func(rc *RecordCollection) {
				if rc.Get(rc.model.FieldName("Name")).(string) == rc.Get(rc.model.FieldName("Description")).(string) {
					log.Panic("Tag name and description must be different")
				}
			})

		tag.methods.AllowAllToGroup(security.GroupEveryone)
		tag.methods.RevokeAllFromGroup(security.GroupEveryone)
		tag.methods.AllowAllToGroup(security.GroupEveryone)

		cv.AddMethod("ComputeOther",
			`Dummy compute function`,
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.model).Set(rc.model.FieldName("Other"), "Other information")
			})

		userModel.AddFields(map[string]FieldDefinition{
			"Name": CharField{String: "Name", Help: "The user's username", Unique: true,
				NoCopy: true, OnChange: userModel.Methods().MustGet("OnChangeName"),
				OnChangeFilters: userModel.Methods().MustGet("OnChangeNameFilters"),
				OnChangeWarning: userModel.Methods().MustGet("OnChangeNameWarning"),
			},
			"DecoratedName": CharField{Compute: userModel.Methods().MustGet("ComputeDecoratedName")},
			"Email":         CharField{Help: "The user's email address", Size: 100, Index: true},
			"Password":      CharField{NoCopy: true},
			"Status": IntegerField{JSON: "status_json", GoType: new(int16),
				Default: DefaultValue(int16(12)), ReadOnly: true},
			"IsStaff":  BooleanField{},
			"IsActive": BooleanField{},
			"Profile": One2OneField{RelationModel: Registry.MustGet("Profile"),
				OnDelete: SetNull, Required: true},
			"Age": IntegerField{Compute: userModel.Methods().MustGet("ComputeAge"),
				Inverse: userModel.Methods().MustGet("InverseSetAge"),
				Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
			"Posts":    One2ManyField{RelationModel: Registry.MustGet("Post"), ReverseFK: "User", Copy: true},
			"PMoney":   FloatField{Related: "Profile.Money"},
			"LastPost": Many2OneField{RelationModel: Registry.MustGet("Post")},
			"Resume":   Many2OneField{RelationModel: Registry.MustGet("Resume"), Embed: true},
			"IsCool":   BooleanField{},
			"CoolType": SelectionField{Selection: types.Selection{
				"cool":    "Yes, its a cool user",
				"no-cool": "No, forget it"},
				Compute:  userModel.Methods().MustGet("ComputeCoolType"),
				Inverse:  userModel.Methods().MustGet("InverseCoolType"),
				OnChange: userModel.Methods().MustGet("OnChangeCoolType")},
			"Email2":          CharField{},
			"IsPremium":       BooleanField{},
			"Nums":            IntegerField{GoType: new(int)},
			"Size":            FloatField{},
			"BestProfilePost": Many2OneField{RelationModel: Registry.MustGet("Post"), Related: "Profile.BestPost"},
			"Mana":            FloatField{GoType: new(float32), OnChange: userModel.Methods().MustGet("OnChangeMana")},
			"Education":       TextField{String: "Educational Background"},
		})
		userModel.AddSQLConstraint("nums_premium", "CHECK((is_premium = TRUE AND nums IS NOT NULL AND nums > 0) OR (IS_PREMIUM = false))",
			"Premium users must have positive nums")

		profileModel.AddFields(map[string]FieldDefinition{
			"Age":      IntegerField{GoType: new(int16)},
			"Gender":   SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
			"Money":    FloatField{},
			"User":     Rev2OneField{RelationModel: Registry.MustGet("User"), ReverseFK: "Profile"},
			"BestPost": Many2OneField{RelationModel: Registry.MustGet("Post")},
			"City":     CharField{},
			"Country":  CharField{},
			"UserName": CharField{Related: "User.Name"},
		})

		post.AddFields(map[string]FieldDefinition{
			"User":       Many2OneField{RelationModel: Registry.MustGet("User")},
			"Title":      CharField{Required: true},
			"Content":    HTMLField{Required: true},
			"Tags":       Many2ManyField{RelationModel: Registry.MustGet("Tag")},
			"Abstract":   TextField{},
			"Attachment": BinaryField{},
			"Read":       BooleanField{Compute: Registry.MustGet("Post").Methods().MustGet("ComputeRead")},
			"LastRead":   DateField{},
			"Visibility": SelectionField{Selection: types.Selection{
				"invisible": "Invisible",
				"visible":   "Visible",
			}},
			"Comments":        One2ManyField{RelationModel: Registry.MustGet("Comment"), ReverseFK: "Post"},
			"LastCommentText": TextField{Related: "Comments.Text"},
			"LastTagName":     CharField{Related: "Tags.Name"},
			"TagsNames":       CharField{Compute: Registry.MustGet("Post").Methods().MustGet("ComputeTagsNames")},
			"WriterAge": IntegerField{Compute: post.Methods().MustGet("ComputeWriterAge"),
				Depends: []string{"User.Age"}, Stored: true, GoType: new(int16)},
			"WriterMoney": FloatField{Related: "User.PMoney"},
		})
		post.SetDefaultOrder("Title")

		comment.AddFields(map[string]FieldDefinition{
			"Post":        Many2OneField{RelationModel: Registry.MustGet("Post")},
			"PostWriter":  Many2OneField{RelationModel: Registry.MustGet("User"), Related: "Post.User"},
			"WriterMoney": FloatField{Related: "PostWriter.PMoney"},
			"Text":        CharField{},
		})

		tag.AddFields(map[string]FieldDefinition{
			"Name":        CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"BestPost":    Many2OneField{RelationModel: Registry.MustGet("Post")},
			"Posts":       Many2ManyField{RelationModel: Registry.MustGet("Post")},
			"Parent":      Many2OneField{RelationModel: Registry.MustGet("Tag")},
			"Description": CharField{Translate: true, Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"Note":        CharField{Translate: true, Required: true, Default: DefaultValue("Default Note")},
			"Rate":        FloatField{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
		})
		tag.SetDefaultOrder("Name DESC", "ID ASC")

		cv.AddFields(map[string]FieldDefinition{
			"Education":  CharField{},
			"Experience": TextField{Translate: true},
			"Leisure":    TextField{},
			"Other":      CharField{Compute: cv.methods.MustGet("ComputeOther")},
		})

		addressMI.AddFields(map[string]FieldDefinition{
			"Street": CharField{GoType: new(string)},
			"Zip":    CharField{},
			"City":   CharField{},
		})

		profileModel.InheritModel(addressMI)

		activeMI.AddFields(map[string]FieldDefinition{
			"Active": BooleanField{Default: DefaultValue(true)},
		})

		Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel.AddFields(map[string]FieldDefinition{
			"Name": CharField{},
			"City": CharField{},
		})

		wizard.AddFields(map[string]FieldDefinition{
			"Name":  CharField{},
			"Value": IntegerField{},
		})
	})
}

func TestFieldModification(t *testing.T) {
	checkUpdates := func(f *Field, property string, value interface{}) {
		So(len(f.updates), ShouldBeGreaterThan, 0)
		So(f.updates[len(f.updates)-1], ShouldContainKey, property)
		So(f.updates[len(f.updates)-1][property], ShouldEqual, value)
	}
	Convey("Testing field modification", t, func() {
		numsField := Registry.MustGet("User").Fields().MustGet("Nums")
		numsField.SetString("Nums Reloaded")
		checkUpdates(numsField, "description", "Nums Reloaded")
		numsField.SetHelp("Num's Help")
		checkUpdates(numsField, "help", "Num's Help")
		numsField.SetCompute(Registry.MustGet("User").methods.MustGet("ComputeNum"))
		checkUpdates(numsField, "compute", "ComputeNum")
		numsField.SetCompute(nil)
		checkUpdates(numsField, "compute", "")
		numsField.SetDefault(DefaultValue("DV"))
		So(numsField.updates[len(numsField.updates)-1], ShouldContainKey, "defaultFunc")
		So(numsField.updates[len(numsField.updates)-1]["defaultFunc"].(func(Environment) interface{})(Environment{}), ShouldEqual, "DV")
		numsField.SetDepends([]string{"Dep1", "Dep2"})
		So(numsField.updates[len(numsField.updates)-1], ShouldContainKey, "depends")
		So(numsField.updates[len(numsField.updates)-1]["depends"], ShouldHaveLength, 2)
		So(numsField.updates[len(numsField.updates)-1]["depends"], ShouldContain, "Dep1")
		So(numsField.updates[len(numsField.updates)-1]["depends"], ShouldContain, "Dep2")
		numsField.SetDepends(nil)
		So(numsField.updates[len(numsField.updates)-1], ShouldContainKey, "depends")
		So(numsField.updates[len(numsField.updates)-1]["depends"], ShouldHaveLength, 0)
		numsField.SetGroupOperator("avg")
		checkUpdates(numsField, "groupOperator", "avg")
		numsField.SetGroupOperator("sum")
		checkUpdates(numsField, "groupOperator", "sum")
		numsField.SetIndex(true)
		checkUpdates(numsField, "index", true)
		numsField.SetNoCopy(true)
		checkUpdates(numsField, "noCopy", true)
		numsField.SetNoCopy(false)
		checkUpdates(numsField, "noCopy", false)
		numsField.SetRelated("Profile.Money")
		checkUpdates(numsField, "relatedPathStr", "Profile.Money")
		numsField.SetRelated("")
		checkUpdates(numsField, "relatedPathStr", "")
		numsField.SetRequired(true)
		checkUpdates(numsField, "required", true)
		numsField.SetRequired(false)
		checkUpdates(numsField, "required", false)
		numsField.SetStored(true)
		checkUpdates(numsField, "stored", true)
		numsField.SetStored(false)
		checkUpdates(numsField, "stored", false)
		numsField.SetUnique(true)
		checkUpdates(numsField, "unique", true)
		numsField.SetUnique(false)
		checkUpdates(numsField, "unique", false)
		nameField := Registry.MustGet("User").Fields().MustGet("Name")
		nameField.SetSize(127)
		checkUpdates(nameField, "size", 127)
		nameField.SetTranslate(true)
		checkUpdates(nameField, "translate", true)
		nameField.SetTranslate(false)
		checkUpdates(nameField, "translate", false)
		nameField.SetOnchange(nil)
		nameField.SetOnchange(Registry.MustGet("User").Methods().MustGet("OnChangeName"))
		nameField.SetOnchangeWarning(Registry.MustGet("User").Methods().MustGet("OnChangeNameWarning"))
		nameField.SetOnchangeFilters(Registry.MustGet("User").Methods().MustGet("OnChangeNameFilters"))
		nameField.SetConstraint(Registry.MustGet("User").Methods().MustGet("UpdateCity"))
		nameField.SetConstraint(nil)
		nameField.SetInverse(Registry.MustGet("User").Methods().MustGet("InverseSetAge"))
		nameField.SetInverse(nil)
		sizeField := Registry.MustGet("User").Fields().MustGet("Size")
		sizeField.SetDigits(nbutils.Digits{Precision: 6, Scale: 2})
		So(sizeField.updates[len(sizeField.updates)-1], ShouldContainKey, "digits")
		So(sizeField.updates[len(sizeField.updates)-1]["digits"].(nbutils.Digits).Precision, ShouldEqual, 6)
		So(sizeField.updates[len(sizeField.updates)-1]["digits"].(nbutils.Digits).Scale, ShouldEqual, 2)
		userField := Registry.MustGet("Post").Fields().MustGet("User")
		userField.SetOnDelete(Cascade)
		checkUpdates(userField, "onDelete", Cascade)
		userField.SetOnDelete(SetNull)
		checkUpdates(userField, "onDelete", SetNull)
		userField.SetEmbed(true)
		checkUpdates(userField, "embed", true)
		userField.SetEmbed(false)
		checkUpdates(userField, "embed", false)
		userField.SetFilter(Registry.MustGet("User").Field(fieldName{name: "SetActive", json: "set_active"}).Equals(true))
		userField.SetFilter(Condition{})
		visibilityField := Registry.MustGet("Post").Fields().MustGet("Visibility")
		visibilityField.UpdateSelection(types.Selection{"logged_in": "Logged in users"})
		So(visibilityField.updates[len(sizeField.updates)-1], ShouldContainKey, "selection_add")
		genderField := Registry.MustGet("Profile").Fields().MustGet("Gender")
		genderField.SetSelection(types.Selection{"m": "Male", "f": "Female"})
		So(genderField.updates[len(sizeField.updates)-1], ShouldContainKey, "selection")
		statusField := Registry.MustGet("User").Fields().MustGet("Status")
		statusField.SetReadOnly(false)
		checkUpdates(statusField, "readOnly", false)
		nFunc := func(env Environment) (b bool, conditioner Conditioner) { return }
		statusField.SetReadOnlyFunc(nFunc)
		checkUpdates(statusField, "readOnlyFunc", nFunc)
		statusField.SetInvisibleFunc(nFunc)
		checkUpdates(statusField, "invisibleFunc", nFunc)
		statusField.SetRequiredFunc(nFunc)
		checkUpdates(statusField, "requiredFunc", nFunc)
	})
}

func TestErroneousDeclarations(t *testing.T) {
	Convey("Testing wrong field declarations", t, func() {
		Convey("Ours = Theirs in M2M field def", func() {
			userModel := Registry.MustGet("User")
			So(func() {
				userModel.AddFields(map[string]FieldDefinition{
					"Tags": Many2ManyField{RelationModel: Registry.MustGet("Tag"),
						M2MOurField: "FT", M2MTheirField: "FT"},
				})
			}, ShouldPanic)
		})
	})
}

func TestMiscellaneous(t *testing.T) {
	Convey("Check that Field instances are FieldNamers", t, func() {
		So(Registry.MustGet("User").Fields().MustGet("Name").JSON(), ShouldEqual, "name")
		So(Registry.MustGet("User").Fields().MustGet("Name").Name(), ShouldEqual, "Name")
	})
}

func TestSequences(t *testing.T) {
	Convey("Testing sequences before bootstrap", t, func() {
		testSeq := CreateSequence("TestSequence", 5, 13)
		_, ok := Registry.GetSequence("TestSequence")
		So(ok, ShouldBeTrue)
		So(testSeq.Increment, ShouldEqual, 5)
		So(testSeq.Start, ShouldEqual, 13)
		testSeq.Alter(3, 14)
		So(testSeq.Increment, ShouldEqual, 3)
		So(testSeq.Start, ShouldEqual, 14)
		testSeq.Drop()
		So(func() { Registry.MustGetSequence("TestSequence") }, ShouldPanic)
		CreateSequence("TestSequence", 5, 13)
	})
}
