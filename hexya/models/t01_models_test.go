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

	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestModelDeclaration(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		user := NewModel("User")
		profile := NewModel("Profile")
		post := NewModel("Post")
		tag := NewModel("Tag")
		cv := NewModel("Resume")
		addressMI := NewMixinModel("AddressMixIn")
		activeMI := NewMixinModel("ActiveMixIn")
		viewModel := NewManualModel("UserView")

		user.AddMethod("PrefixedUser", "",
			func(rc *RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get("Name")))
				}
				return res
			})

		user.Methods().MustGet("PrefixedUser").Extend("",
			func(rc *RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					email := u.Get("Email").(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
				}
				return res
			})

		user.AddMethod("DecorateEmail", "",
			func(rc *RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_square_brackets") {
					return fmt.Sprintf("[%s]", email)
				}
				return fmt.Sprintf("<%s>", email)
			})

		user.Methods().MustGet("DecorateEmail").Extend("",
			func(rc *RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_double_square") {
					rc = rc.
						Call("WithContext", "use_square_brackets", true).(*RecordCollection).
						WithContext("fake_key", true)
				}
				res := rc.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		user.AddMethod("ComputeDecoratedName", "",
			func(rc *RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
				return res, []FieldNamer{FieldName("DecoratedName")}
			})

		user.AddMethod("ComputeAge", "",
			func(rc *RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["Age"] = rc.Get("Profile").(*RecordCollection).Get("Age").(int16)
				return res, []FieldNamer{}
			})

		user.AddMethod("InverseSetAge", "",
			func(rc *RecordCollection, age int16) {
				rc.Get("Profile").(*RecordCollection).Set("Age", age)
			})

		user.AddMethod("UpdateCity", "",
			func(rc *RecordCollection, value string) {
				rc.Get("Profile").(*RecordCollection).Set("City", value)
			})

		user.AddMethod("ComputeNum", "Dummy method",
			func(rc *RecordCollection) (FieldMap, []FieldNamer) {
				return FieldMap{}, []FieldNamer{}
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc *RecordCollection) bool {
				return rc.Get("Active").(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc *RecordCollection) string {
				return "Hello !"
			})

		printAddress := addressMI.AddEmptyMethod("PrintAddress")
		printAddress.DeclareMethod("",
			func(rc *RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get("Street"), rc.Get("Zip"), rc.Get("City"))
			})

		profile.AddMethod("PrintAddress", "",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get("Country"))
			})

		addressMI.Methods().MustGet("PrintAddress").Extend("",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("<%s>", res)
			})

		profile.Methods().MustGet("PrintAddress").Extend("",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("[%s]", res)
			})

		post.Methods().MustGet("Create").Extend("",
			func(rc *RecordCollection, data FieldMapper) *RecordCollection {
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

		tag.AddMethod("CheckRate",
			`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
			func(rc *RecordCollection) {
				if rc.Get("Rate").(float32) < 0 || rc.Get("Rate").(float32) > 10 {
					log.Panic("Tag rate must be between 0 and 10")
				}
			})

		tag.AddMethod("CheckNameDescription",
			`CheckNameDescription checks that the description of a tag is not equal to its name`,
			func(rc *RecordCollection) {
				if rc.Get("Name").(string) == rc.Get("Description").(string) {
					log.Panic("Tag name and description must be different")
				}
			})

		tag.methods.AllowAllToGroup(security.GroupEveryone)
		tag.methods.RevokeAllFromGroup(security.GroupEveryone)
		tag.methods.AllowAllToGroup(security.GroupEveryone)

		user.AddFields(map[string]FieldDefinition{
			"Name": CharField{String: "Name", Help: "The user's username", Unique: true,
				NoCopy: true, OnChange: user.Methods().MustGet("ComputeDecoratedName")},
			"DecoratedName": CharField{Compute: user.Methods().MustGet("ComputeDecoratedName")},
			"Email":         CharField{Help: "The user's email address", Size: 100, Index: true},
			"Password":      CharField{NoCopy: true},
			"Status": IntegerField{JSON: "status_json", GoType: new(int16),
				Default: DefaultValue(int16(12))},
			"IsStaff":  BooleanField{},
			"IsActive": BooleanField{},
			"Profile": Many2OneField{RelationModel: Registry.MustGet("Profile"),
				OnDelete: Restrict, Required: true},
			"Age": IntegerField{Compute: user.Methods().MustGet("ComputeAge"),
				Inverse: user.Methods().MustGet("InverseSetAge"),
				Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
			"Posts":     One2ManyField{RelationModel: Registry.MustGet("Post"), ReverseFK: "User"},
			"PMoney":    FloatField{Related: "Profile.Money"},
			"LastPost":  Many2OneField{RelationModel: Registry.MustGet("Post")},
			"Resume":    Many2OneField{RelationModel: Registry.MustGet("Resume"), Embed: true},
			"Email2":    CharField{},
			"IsPremium": BooleanField{},
			"Nums":      IntegerField{GoType: new(int)},
			"Size":      FloatField{},
		})
		user.AddSQLConstraint("nums_premium", "CHECK((is_premium = TRUE AND nums > 0) OR (IS_PREMIUM = false))",
			"Premium users must have positive nums")

		profile.AddFields(map[string]FieldDefinition{
			"Age":      IntegerField{GoType: new(int16)},
			"Gender":   SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
			"Money":    FloatField{},
			"User":     Many2OneField{RelationModel: Registry.MustGet("User")},
			"BestPost": One2OneField{RelationModel: Registry.MustGet("Post")},
			"City":     CharField{},
			"Country":  CharField{},
		})

		post.AddFields(map[string]FieldDefinition{
			"User":            Many2OneField{RelationModel: Registry.MustGet("User")},
			"Title":           CharField{Required: true},
			"Content":         HTMLField{},
			"Tags":            Many2ManyField{RelationModel: Registry.MustGet("Tag")},
			"BestPostProfile": Rev2OneField{RelationModel: Registry.MustGet("Profile"), ReverseFK: "BestPost"},
			"Abstract":        TextField{},
			"Attachment":      BinaryField{},
			"LastRead":        DateField{},
		})

		tag.AddFields(map[string]FieldDefinition{
			"Name":        CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"BestPost":    Many2OneField{RelationModel: Registry.MustGet("Post")},
			"Posts":       Many2ManyField{RelationModel: Registry.MustGet("Post")},
			"Parent":      Many2OneField{RelationModel: Registry.MustGet("Tag")},
			"Description": CharField{Constraint: tag.Methods().MustGet("CheckNameDescription")},
			"Rate":        FloatField{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)},
		})

		cv.AddFields(map[string]FieldDefinition{
			"Education":  TextField{},
			"Experience": TextField{},
			"Leisure":    TextField{},
		})

		addressMI.AddFields(map[string]FieldDefinition{
			"Street": CharField{GoType: new(string)},
			"Zip":    CharField{},
			"City":   CharField{},
		})

		profile.InheritModel(addressMI)

		activeMI.AddFields(map[string]FieldDefinition{
			"Active": BooleanField{},
		})

		Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel.AddFields(map[string]FieldDefinition{
			"Name": CharField{},
			"City": CharField{},
		})
	})
}

func TestFieldModification(t *testing.T) {
	Convey("Testing field modification", t, func() {
		numsField := Registry.MustGet("User").Fields().MustGet("Nums")
		So(numsField.SetString("Nums Reloaded").description, ShouldEqual, "Nums Reloaded")
		So(numsField.SetHelp("Num's Help").help, ShouldEqual, "Num's Help")
		So(numsField.SetCompute(Registry.MustGet("User").methods.MustGet("ComputeNum")).compute, ShouldEqual, "ComputeNum")
		So(numsField.SetCompute(nil).compute, ShouldEqual, "")
		So(numsField.SetDefault(DefaultValue("DV")).defaultFunc(Environment{}).(string), ShouldEqual, "DV")
		numsField.SetDepends([]string{"Dep1", "Dep2"})
		So(numsField.depends, ShouldHaveLength, 2)
		So(numsField.depends, ShouldContain, "Dep1")
		So(numsField.depends, ShouldContain, "Dep2")
		numsField.SetDepends(nil)
		So(numsField.depends, ShouldBeEmpty)
		So(numsField.SetGroupOperator("avg").groupOperator, ShouldEqual, "avg")
		So(numsField.SetGroupOperator("sum").groupOperator, ShouldEqual, "sum")
		So(numsField.SetIndex(true).index, ShouldBeTrue)
		So(numsField.SetNoCopy(true).noCopy, ShouldBeTrue)
		So(numsField.SetNoCopy(false).noCopy, ShouldBeFalse)
		So(numsField.SetRelated("Profile.Money").relatedPath, ShouldEqual, "Profile.Money")
		So(numsField.SetRelated("").relatedPath, ShouldEqual, "")
		So(numsField.SetRequired(true).required, ShouldBeTrue)
		So(numsField.SetRequired(false).required, ShouldBeFalse)
		So(numsField.SetStored(true).stored, ShouldBeTrue)
		So(numsField.SetStored(false).stored, ShouldBeFalse)
		So(numsField.SetTranslate(true).translate, ShouldBeTrue)
		So(numsField.SetUnique(true).unique, ShouldBeTrue)
		So(numsField.SetUnique(false).unique, ShouldBeFalse)
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
