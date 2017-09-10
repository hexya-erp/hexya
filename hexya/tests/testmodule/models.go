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

package testmodule

import (
	"fmt"

	"log"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	user := pool.User().DeclareModel()

	// Methods directly declared with AddMethod must be defined before being referenced in the field declaration

	user.AddMethod("ComputeDecoratedName", "",
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			res := pool.UserData{
				DecoratedName: rs.PrefixedUser("User")[0],
			}
			return &res, []models.FieldNamer{pool.User().DecoratedName()}
		})

	user.AddMethod("ComputeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			res := pool.UserData{
				Age: rs.Profile().Age(),
			}
			return &res, []models.FieldNamer{pool.User().Age()}
		})

	user.AddCharField("Name", models.StringFieldParams{String: "Name", Help: "The user's username", Unique: true})
	user.AddCharField("DecoratedName", models.StringFieldParams{Compute: pool.User().Methods().ComputeDecoratedName()})
	user.AddCharField("Email", models.StringFieldParams{Help: "The user's email address", Size: 100, Index: true})
	user.AddCharField("Password", models.StringFieldParams{NoCopy: true})
	user.AddIntegerField("Status", models.SimpleFieldParams{JSON: "status_json", GoType: new(int16)})
	user.AddBooleanField("IsStaff", models.SimpleFieldParams{})
	user.AddBooleanField("IsActive", models.SimpleFieldParams{})
	user.AddMany2OneField("Profile", models.ForeignKeyFieldParams{RelationModel: pool.Profile()})
	user.AddIntegerField("Age", models.SimpleFieldParams{Compute: pool.User().Methods().ComputeAge(),
		Inverse: pool.User().Methods().InverseSetAge(),
		Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)})
	user.AddOne2ManyField("Posts", models.ReverseFieldParams{RelationModel: pool.Post(), ReverseFK: "User"})
	user.AddFloatField("PMoney", models.FloatFieldParams{Related: "Profile.Money"})
	user.AddMany2OneField("LastPost", models.ForeignKeyFieldParams{RelationModel: pool.Post(), Embed: true})
	user.AddCharField("Email2", models.StringFieldParams{})
	user.AddBooleanField("IsPremium", models.SimpleFieldParams{})
	user.AddIntegerField("Nums", models.SimpleFieldParams{GoType: new(int)})

	user.Methods().PrefixedUser().DeclareMethod(
		`PrefixedUser is a sample method layer for testing`,
		func(rs pool.UserSet, prefix string) []string {
			var res []string
			for _, u := range rs.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.Name()))
			}
			return res
		})

	user.Methods().DecorateEmail().DeclareMethod(
		`DecorateEmail is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			return fmt.Sprintf("<%s>", email)
		})

	user.Methods().DecorateEmail().Extend(
		`DecorateEmailExtension is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			res := rs.Super().DecorateEmail(email)
			return fmt.Sprintf("[%s]", res)
		})

	user.Methods().InverseSetAge().DeclareMethod("",
		func(rs pool.UserSet, vals models.FieldMapper) {
			values, _ := rs.DataStruct(vals.FieldMap())
			rs.Profile().SetAge(values.Age)
		})

	pool.User().Methods().PrefixedUser().Extend("",
		func(rs pool.UserSet, prefix string) []string {
			res := rs.Super().PrefixedUser(prefix)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	pool.User().Methods().UpdateCity().DeclareMethod("",
		func(rs pool.UserSet, value string) {
			rs.Profile().SetCity(value)
		})

	profile := pool.Profile().DeclareModel()
	profile.AddIntegerField("Age", models.SimpleFieldParams{GoType: new(int16)})
	profile.AddSelectionField("Gender", models.SelectionFieldParams{Selection: types.Selection{"male": "Male", "female": "Female"}})
	profile.AddFloatField("Money", models.FloatFieldParams{})
	profile.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User()})
	profile.AddOne2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: pool.Post()})
	profile.AddCharField("City", models.StringFieldParams{})
	profile.AddCharField("Country", models.StringFieldParams{})

	post := pool.Post().DeclareModel()
	post.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User()})
	post.AddCharField("Title", models.StringFieldParams{})
	post.AddTextField("Content", models.StringFieldParams{})
	post.AddMany2ManyField("Tags", models.Many2ManyFieldParams{RelationModel: pool.Tag()})

	pool.Post().Methods().Create().Extend("",
		func(rs pool.PostSet, data models.FieldMapper) pool.PostSet {
			res := rs.Super().Create(data)
			return res
		})

	tag := pool.Tag().DeclareModel()
	tag.AddCharField("Name", models.StringFieldParams{Constraint: pool.Tag().Methods().CheckNameDescription()})
	tag.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.Tag()})
	tag.AddMany2OneField("BestPost", models.ForeignKeyFieldParams{RelationModel: pool.Post()})
	tag.AddMany2ManyField("Posts", models.Many2ManyFieldParams{RelationModel: pool.Post()})
	tag.AddCharField("Description", models.StringFieldParams{Constraint: pool.Tag().Methods().CheckNameDescription()})
	tag.AddFloatField("Rate", models.FloatFieldParams{Constraint: pool.Tag().Methods().CheckRate(), GoType: new(float32)})

	tag.Methods().CheckNameDescription().DeclareMethod(
		`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
		func(rs pool.TagSet) {
			if rs.Rate() < 0 || rs.Rate() > 10 {
				log.Panic("Tag rate must be between 0 and 10")
			}
		}).AllowGroup(security.GroupEveryone)

	tag.Methods().CheckRate().DeclareMethod(
		`CheckNameDescription checks that the description of a tag is not equal to its name`,
		func(rs pool.TagSet) {
			if rs.Name() == rs.Description() {
				log.Panic("Tag name and description must be different")
			}
		})

	addressMI := pool.AddressMixIn().DeclareMixinModel()
	addressMI.AddCharField("Street", models.StringFieldParams{})
	addressMI.AddCharField("Zip", models.StringFieldParams{})
	addressMI.AddCharField("City", models.StringFieldParams{})
	profile.InheritModel(addressMI)

	pool.Profile().Methods().PrintAddress().DeclareMethod(
		`PrintAddress is a sample method layer for testing`,
		func(rs pool.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("%s, %s", res, rs.Country())
		})

	pool.Profile().Methods().PrintAddress().Extend("",
		func(rs pool.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("[%s]", res)
		})

	addressMI2 := pool.AddressMixIn()
	addressMI2.Methods().SayHello().DeclareMethod(
		`SayHello is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return "Hello !"
		})

	addressMI2.Methods().PrintAddress().DeclareMethod(
		`PrintAddressMixIn is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
		})

	addressMI2.Methods().PrintAddress().Extend("",
		func(rs pool.AddressMixInSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("<%s>", res)
		})

	activeMI := pool.ActiveMixIn().DeclareMixinModel()
	activeMI.AddBooleanField("Active", models.SimpleFieldParams{})
	pool.ModelMixin().InheritModel(activeMI)

	// Chained declaration
	activeMI1 := pool.ActiveMixIn()
	activeMI2 := activeMI1
	activeMI2.Methods().IsActivated().DeclareMethod(
		`IsACtivated is a sample method of ActiveMixIn"`,
		func(rs pool.ActiveMixInSet) bool {
			return rs.Active()
		})

	viewModel := pool.UserView().DeclareManualModel()
	viewModel.AddCharField("Name", models.StringFieldParams{})
	viewModel.AddCharField("City", models.StringFieldParams{})
}
