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
		addressMI := NewMixinModel("AddressMixIn")
		activeMI := NewMixinModel("ActiveMixIn")
		viewModel := NewManualModel("UserView")

		user.AddMethod("PrefixedUser", "",
			func(rc RecordCollection, prefix string) []string {
				var res []string
				for _, u := range rc.Records() {
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get("Name")))
				}
				return res
			})

		user.Methods().MustGet("PrefixedUser").Extend("",
			func(rc RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					email := u.Get("Email").(string)
					res[i] = fmt.Sprintf("%s %s", res[i], rc.Call("DecorateEmail", email))
				}
				return res
			})

		user.AddMethod("DecorateEmail", "",
			func(rc RecordCollection, email string) string {
				if rc.Env().Context().HasKey("use_square_brackets") {
					return fmt.Sprintf("[%s]", email)
				}
				return fmt.Sprintf("<%s>", email)
			})

		user.Methods().MustGet("DecorateEmail").Extend("",
			func(rc RecordCollection, email string) string {
				rSet := rc
				if rc.Env().Context().HasKey("use_double_square") {
					rSet = rSet.
						Call("WithContext", "use_square_brackets", true).(RecordCollection).
						WithContext("fake_key", true)
				}
				res := rSet.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		user.AddMethod("ComputeDecoratedName", "",
			func(rc RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
				return res, []FieldNamer{FieldName("DecoratedName")}
			})

		user.AddMethod("ComputeAge", "",
			func(rc RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["Age"] = rc.Get("Profile").(RecordCollection).Get("Age").(int16)
				return res, []FieldNamer{}
			})

		user.AddMethod("InverseSetAge", "",
			func(rc RecordCollection, vals FieldMapper) {
				value, ok := vals.FieldMap(FieldName("Age"))["Age"]
				if !ok {
					return
				}
				rc.Get("Profile").(RecordCollection).Set("Age", value)
			})

		user.AddMethod("UpdateCity", "",
			func(rc RecordCollection, value string) {
				rc.Get("Profile").(RecordCollection).Set("City", value)
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc RecordCollection) bool {
				return rc.Get("Active").(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc RecordCollection) string {
				return "Hello !"
			})

		printAddress := addressMI.AddEmptyMethod("PrintAddress")
		printAddress.DeclareMethod("",
			func(rc RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get("Street"), rc.Get("Zip"), rc.Get("City"))
			})

		profile.AddMethod("PrintAddress", "",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get("Country"))
			})

		addressMI.Methods().MustGet("PrintAddress").Extend("",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("<%s>", res)
			})

		profile.Methods().MustGet("PrintAddress").Extend("",
			func(rc RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("[%s]", res)
			})

		post.Methods().MustGet("Create").Extend("",
			func(rc RecordCollection, data FieldMapper) RecordCollection {
				res := rc.Super().Call("Create", data).(RecordSet).Collection()
				return res
			})

		post.Methods().MustGet("WithContext").Extend("",
			func(rc RecordCollection, key string, value interface{}) RecordCollection {
				return rc.Super().Call("WithContext", key, value).(RecordCollection)
			})

		tag.AddMethod("CheckRate",
			`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
			func(rc RecordCollection) {
				if rc.Get("Rate").(float32) < 0 || rc.Get("Rate").(float32) > 10 {
					log.Panic("Tag rate must be between 0 and 10")
				}
			})

		tag.AddMethod("CheckNameDescription",
			`CheckNameDescription checks that the description of a tag is not equal to its name`,
			func(rc RecordCollection) {
				if rc.Get("Name").(string) == rc.Get("Description").(string) {
					log.Panic("Tag name and description must be different")
				}
			})

		tag.methods.AllowAllToGroup(security.GroupEveryone)
		tag.methods.RevokeAllFromGroup(security.GroupEveryone)
		tag.methods.AllowAllToGroup(security.GroupEveryone)

		user.AddCharField("Name", StringFieldParams{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().MustGet("ComputeDecoratedName")})
		user.AddCharField("DecoratedName", StringFieldParams{Compute: user.Methods().MustGet("ComputeDecoratedName")})
		user.AddCharField("Email", StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
		user.AddCharField("Password", StringFieldParams{NoCopy: true})
		user.AddIntegerField("Status", SimpleFieldParams{JSON: "status_json", GoType: new(int16),
			Default: DefaultValue(int16(12))})
		user.AddBooleanField("IsStaff", SimpleFieldParams{})
		user.AddBooleanField("IsActive", SimpleFieldParams{})
		user.AddMany2OneField("Profile", ForeignKeyFieldParams{RelationModel: Registry.MustGet("Profile"),
			OnDelete: Restrict, Required: true})
		user.AddIntegerField("Age", SimpleFieldParams{Compute: user.Methods().MustGet("ComputeAge"),
			Inverse: user.Methods().MustGet("InverseSetAge"),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true})
		user.AddOne2ManyField("Posts", ReverseFieldParams{RelationModel: Registry.MustGet("Post"),
			ReverseFK: "User"})
		user.AddFloatField("PMoney", FloatFieldParams{Related: "Profile.Money"})
		user.AddMany2OneField("LastPost", ForeignKeyFieldParams{RelationModel: Registry.MustGet("Post"),
			Embed: true})
		user.AddCharField("Email2", StringFieldParams{})
		user.AddBooleanField("IsPremium", SimpleFieldParams{})
		user.AddIntegerField("Nums", SimpleFieldParams{GoType: new(int)})
		user.AddFloatField("Size", FloatFieldParams{})

		user.AddSQLConstraint("nums_premium", "CHECK((is_premium = TRUE AND nums > 0) OR (IS_PREMIUM = false))",
			"Premium users must have positive nums")

		profile.AddIntegerField("Age", SimpleFieldParams{GoType: new(int16)})
		profile.AddSelectionField("Gender", SelectionFieldParams{Selection: types.Selection{"male": "Male", "female": "Female"}})
		profile.AddFloatField("Money", FloatFieldParams{})
		profile.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: Registry.MustGet("User")})
		profile.AddOne2OneField("BestPost", ForeignKeyFieldParams{RelationModel: Registry.MustGet("Post")})
		profile.AddCharField("City", StringFieldParams{})
		profile.AddCharField("Country", StringFieldParams{})

		post.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: Registry.MustGet("User")})
		post.AddCharField("Title", StringFieldParams{})
		post.AddHTMLField("Content", StringFieldParams{})
		post.AddMany2ManyField("Tags", Many2ManyFieldParams{RelationModel: Registry.MustGet("Tag")})
		post.AddRev2OneField("BestPostProfile", ReverseFieldParams{RelationModel: Registry.MustGet("Profile"),
			ReverseFK: "BestPost"})
		post.AddTextField("Abstract", StringFieldParams{})
		post.AddBinaryField("Attachment", SimpleFieldParams{})
		post.AddDateField("LastRead", SimpleFieldParams{})

		tag.AddCharField("Name", StringFieldParams{Constraint: tag.Methods().MustGet("CheckNameDescription")})
		tag.AddMany2OneField("BestPost", ForeignKeyFieldParams{RelationModel: Registry.MustGet("Post")})
		tag.AddMany2ManyField("Posts", Many2ManyFieldParams{RelationModel: Registry.MustGet("Post")})
		tag.AddMany2OneField("Parent", ForeignKeyFieldParams{RelationModel: Registry.MustGet("Tag")})
		tag.AddCharField("Description", StringFieldParams{Constraint: tag.Methods().MustGet("CheckNameDescription")})
		tag.AddFloatField("Rate", FloatFieldParams{Constraint: tag.Methods().MustGet("CheckRate"), GoType: new(float32)})

		addressMI.AddCharField("Street", StringFieldParams{GoType: new(string)})
		addressMI.AddCharField("Zip", StringFieldParams{})
		addressMI.AddCharField("City", StringFieldParams{})
		profile.InheritModel(addressMI)

		activeMI.AddBooleanField("Active", SimpleFieldParams{})
		Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel.AddCharField("Name", StringFieldParams{})
		viewModel.AddCharField("City", StringFieldParams{})

	})
}

func TestFieldModification(t *testing.T) {
	Convey("Testing field modification", t, func() {
		numsField := Registry.MustGet("User").Fields().MustGet("Nums")
		So(numsField.SetString("Nums Reloaded").description, ShouldEqual, "Nums Reloaded")
		So(numsField.SetHelp("Num's Help").help, ShouldEqual, "Num's Help")
		So(numsField.SetCompute("ComputeNum").compute, ShouldEqual, "ComputeNum")
		So(numsField.SetCompute("").compute, ShouldEqual, "")
		So(numsField.SetDefault(DefaultValue("DV")).defaultFunc(Environment{}, FieldMap{}).(string), ShouldEqual, "DV")
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
				userModel.AddMany2ManyField("Tags", Many2ManyFieldParams{RelationModel: Registry.MustGet("Tag"),
					M2MOurField: "FT", M2MTheirField: "FT"})
			}, ShouldPanic)
		})
	})
}
