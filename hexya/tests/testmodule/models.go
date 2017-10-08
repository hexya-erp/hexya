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

	user.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().ComputeDecoratedName()},
		"DecoratedName": models.CharField{Compute: user.Methods().ComputeDecoratedName()},
		"Email":         models.CharField{Help: "The user's email address", Size: 100, Index: true},
		"Password":      models.CharField{NoCopy: true},
		"Status": models.IntegerField{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  models.BooleanField{},
		"IsActive": models.BooleanField{},
		"Profile":  models.Many2OneField{RelationModel: pool.Profile()},
		"Age": models.IntegerField{Compute: user.Methods().ComputeAge(),
			Inverse: user.Methods().InverseSetAge(),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
		"Posts":     models.One2ManyField{RelationModel: pool.Post(), ReverseFK: "User"},
		"PMoney":    models.FloatField{Related: "Profile.Money"},
		"LastPost":  models.Many2OneField{RelationModel: pool.Post(), Embed: true},
		"Email2":    models.CharField{},
		"IsPremium": models.BooleanField{},
		"Nums":      models.IntegerField{GoType: new(int)},
		"Size":      models.FloatField{},
	})

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
	profile.AddFields(map[string]models.FieldDefinition{
		"Age":      models.IntegerField{GoType: new(int16)},
		"Gender":   models.SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    models.FloatField{},
		"User":     models.Many2OneField{RelationModel: pool.User()},
		"BestPost": models.One2OneField{RelationModel: pool.Post()},
		"City":     models.CharField{},
		"Country":  models.CharField{},
	})

	post := pool.Post().DeclareModel()
	post.AddFields(map[string]models.FieldDefinition{
		"User":            models.Many2OneField{RelationModel: pool.User()},
		"Title":           models.CharField{},
		"Content":         models.HTMLField{},
		"Tags":            models.Many2ManyField{RelationModel: pool.Tag()},
		"BestPostProfile": models.Rev2OneField{RelationModel: pool.Profile(), ReverseFK: "BestPost"},
		"Abstract":        models.TextField{},
		"Attachment":      models.BinaryField{},
		"LastRead":        models.DateField{},
	})

	pool.Post().Methods().Create().Extend("",
		func(rs pool.PostSet, data *pool.PostData) pool.PostSet {
			res := rs.Super().Create(data)
			return res
		})

	pool.Post().Methods().Search().Extend("",
		func(rs pool.PostSet, cond pool.PostCondition) pool.PostSet {
			res := rs.Super().Search(cond)
			return res
		})

	tag := pool.Tag().DeclareModel()
	tag.AddFields(map[string]models.FieldDefinition{
		"Name":        models.CharField{Constraint: tag.Methods().CheckNameDescription()},
		"BestPost":    models.Many2OneField{RelationModel: pool.Post()},
		"Posts":       models.Many2ManyField{RelationModel: pool.Post()},
		"Parent":      models.Many2OneField{RelationModel: pool.Tag()},
		"Description": models.CharField{Constraint: tag.Methods().CheckNameDescription()},
		"Rate":        models.FloatField{Constraint: tag.Methods().CheckRate(), GoType: new(float32)},
	})

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
	addressMI.AddFields(map[string]models.FieldDefinition{
		"Street": models.CharField{GoType: new(string)},
		"Zip":    models.CharField{},
		"City":   models.CharField{},
	})
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
	activeMI.AddFields(map[string]models.FieldDefinition{
		"Active": models.BooleanField{},
	})
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
	viewModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{},
		"City": models.CharField{},
	})
}
