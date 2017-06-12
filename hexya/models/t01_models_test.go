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

	"github.com/hexya-erp/hexya/hexya/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestModelDeclaration(t *testing.T) {
	Convey("Creating DataBase...", t, func() {
		user := NewModel("User")
		user.AddCharField("Name", StringFieldParams{String: "Name", Help: "The user's username", Unique: true, NoCopy: true, OnChange: "computeDecoratedName"})
		user.AddCharField("DecoratedName", StringFieldParams{Compute: "computeDecoratedName"})
		user.AddCharField("Email", StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
		user.AddCharField("Password", StringFieldParams{NoCopy: true})
		user.AddIntegerField("Status", SimpleFieldParams{JSON: "status_json", GoType: new(int16), Default: DefaultValue(int16(12))})
		user.AddBooleanField("IsStaff", SimpleFieldParams{})
		user.AddBooleanField("IsActive", SimpleFieldParams{})
		user.AddMany2OneField("Profile", ForeignKeyFieldParams{RelationModel: "Profile", OnDelete: Restrict, Required: true})
		user.AddIntegerField("Age", SimpleFieldParams{Compute: "computeAge", Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)})
		user.AddOne2ManyField("Posts", ReverseFieldParams{RelationModel: "Post", ReverseFK: "User"})
		user.AddFloatField("PMoney", FloatFieldParams{Related: "Profile.Money"})
		user.AddMany2OneField("LastPost", ForeignKeyFieldParams{RelationModel: "Post", Embed: true})
		user.AddCharField("Email2", StringFieldParams{})
		user.AddBooleanField("IsPremium", SimpleFieldParams{})
		user.AddIntegerField("Nums", SimpleFieldParams{GoType: new(int)})
		user.AddFloatField("Size", FloatFieldParams{})

		profile := NewModel("Profile")
		profile.AddIntegerField("Age", SimpleFieldParams{GoType: new(int16)})
		profile.AddSelectionField("Gender", SelectionFieldParams{Selection: types.Selection{"male": "Male", "female": "Female"}})
		profile.AddFloatField("Money", FloatFieldParams{})
		profile.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: "User"})
		profile.AddOne2OneField("BestPost", ForeignKeyFieldParams{RelationModel: "Post"})
		profile.AddCharField("City", StringFieldParams{})
		profile.AddCharField("Country", StringFieldParams{})

		post := NewModel("Post")
		post.AddMany2OneField("User", ForeignKeyFieldParams{RelationModel: "User"})
		post.AddCharField("Title", StringFieldParams{})
		post.AddHTMLField("Content", StringFieldParams{})
		post.AddMany2ManyField("Tags", Many2ManyFieldParams{RelationModel: "Tag"})
		post.AddRev2OneField("BestPostProfile", ReverseFieldParams{RelationModel: "Profile", ReverseFK: "BestPost"})
		post.AddTextField("Abstract", StringFieldParams{})
		post.AddBinaryField("Attachment", SimpleFieldParams{})
		post.AddDateField("LastRead", SimpleFieldParams{})

		tag := NewModel("Tag")
		tag.AddCharField("Name", StringFieldParams{})
		tag.AddMany2OneField("BestPost", ForeignKeyFieldParams{RelationModel: "Post"})
		tag.AddMany2ManyField("Posts", Many2ManyFieldParams{RelationModel: "Post"})
		tag.AddMany2OneField("Parent", ForeignKeyFieldParams{RelationModel: "Tag"})
		tag.AddCharField("Description", StringFieldParams{})
		tag.AddFloatField("Rate", FloatFieldParams{GoType: new(float32)})

		addressMI := NewMixinModel("AddressMixIn")
		addressMI.AddCharField("Street", StringFieldParams{GoType: new(string)})
		addressMI.AddCharField("Zip", StringFieldParams{})
		addressMI.AddCharField("City", StringFieldParams{})
		profile.InheritModel(addressMI)

		activeMI := NewMixinModel("ActiveMixIn")
		activeMI.AddBooleanField("Active", SimpleFieldParams{})
		Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel := NewManualModel("UserView")
		viewModel.AddCharField("Name", StringFieldParams{})
		viewModel.AddCharField("City", StringFieldParams{})

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
				return fmt.Sprintf("<%s>", email)
			})

		user.Methods().MustGet("DecorateEmail").Extend("",
			func(rc RecordCollection, email string) string {
				res := rc.Super().Call("DecorateEmail", email).(string)
				return fmt.Sprintf("[%s]", res)
			})

		user.AddMethod("computeDecoratedName", "",
			func(rc RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["DecoratedName"] = rc.Call("PrefixedUser", "User").([]string)[0]
				return res, []FieldNamer{FieldName("DecoratedName")}
			})

		user.AddMethod("computeAge", "",
			func(rc RecordCollection) (FieldMap, []FieldNamer) {
				res := make(FieldMap)
				res["Age"] = rc.Get("Profile").(RecordCollection).Get("Age").(int16)
				return res, []FieldNamer{}
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

		addressMI.AddMethod("PrintAddress", "",
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
				userModel.AddMany2ManyField("Tags", Many2ManyFieldParams{RelationModel: "Tag", M2MOurField: "FT", M2MTheirField: "FT"})
			}, ShouldPanic)
		})
	})
}
