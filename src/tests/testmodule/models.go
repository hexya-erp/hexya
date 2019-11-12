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

	"github.com/hexya-erp/hexya/src/actions"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/fields"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/m"
	"github.com/hexya-erp/pool/q"
)

// IsStaffHelp exported
var IsStaffHelp = "This is a var help message"

const (
	isStaffString        = "Is Staff"
	isPremiumDescription = "This is a const description string"
)

const isPremiumString = isPremiumDescription

// UserDecorateEmail decorates the email of the given user
func UserDecorateEmail(rs m.UserSet, email string) string {
	return fmt.Sprintf("<%s>", email)
}

func init() {
	models.NewModel("User")

	// Methods directly declared with AddMethod must be defined before being referenced in the field declaration

	h.User().AddMethod("OnChangeName", "",
		func(rs m.UserSet) m.UserData {
			return h.User().NewData().SetDecoratedName(rs.PrefixedUser("User")[0])
		})

	h.User().AddMethod("ComputeDecoratedName", "",
		func(rs m.UserSet) m.UserData {
			return h.User().NewData().SetDecoratedName(rs.PrefixedUser("User")[0])
		})

	h.User().AddMethod("ComputeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rs m.UserSet) m.UserData {
			return h.User().NewData().SetAge(rs.Profile().Age())
		})

	var isPremiumHelp = "This the IsPremium Help message"

	h.User().Fields().Experience().SetString("Professional Experience")

	h.User().AddMethod("PrefixedUser",
		`PrefixedUser is a sample method layer for testing`,
		func(rs m.UserSet, prefix string) []string {
			var res []string
			for _, u := range rs.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.Name()))
			}
			return res
		})

	h.User().AddMethod("DecorateEmail",
		`DecorateEmail is a sample method layer for testing`,
		UserDecorateEmail)

	h.User().Methods().DecorateEmail().Extend(
		`DecorateEmailExtension is a sample method layer for testing`,
		func(rs m.UserSet, email string) string {
			res := rs.Super().DecorateEmail(email)
			return fmt.Sprintf("[%s]", res)
		})

	h.User().AddMethod("RecursiveMethod",
		`RecursiveMethod is a sample method layer for testing`,
		func(rs m.UserSet, depth int, result string) string {
			if depth == 0 {
				return result
			}
			return rs.RecursiveMethod(depth-1, fmt.Sprintf("%s, recursion %d", result, depth))
		})

	h.User().Methods().RecursiveMethod().Extend("",
		func(rs m.UserSet, depth int, result string) string {
			result = "> " + result + " <"
			sup := rs.Super().RecursiveMethod(depth, result)
			return sup
		})

	h.User().AddMethod("SubSetSuper", "",
		func(rs m.UserSet) string {
			var res string
			for _, rec := range rs.Records() {
				res += rec.Name()
			}
			return res
		})

	h.User().Methods().SubSetSuper().Extend("",
		func(rs m.UserSet) string {
			userJane := h.User().Search(rs.Env(), q.User().Email().Equals("jane.smith@example.com"))
			userJohn := h.User().Search(rs.Env(), q.User().Email().Equals("jsmith2@example.com"))
			users := h.User().NewSet(rs.Env())
			users = users.Union(userJane)
			users = users.Union(userJohn)
			return users.Super().SubSetSuper()
		})

	h.User().AddMethod("InverseSetAge", "",
		func(rs m.UserSet, age int16) {
			rs.Profile().SetAge(age)
		})

	h.User().Methods().PrefixedUser().Extend("",
		func(rs m.UserSet, prefix string) []string {
			res := rs.Super().PrefixedUser(prefix)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	h.User().AddMethod("UpdateCity", "",
		func(rs m.UserSet, value string) {
			rs.Profile().SetCity(value)
		})

	h.User().AddFields(map[string]models.FieldDefinition{
		"Name": fields.Char{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: h.User().Methods().OnChangeName()},
		"DecoratedName": fields.Char{Compute: h.User().Methods().ComputeDecoratedName()},
		"Email":         fields.Char{Help: "The user's email address", Size: 100, Index: true},
		"Password":      fields.Char{NoCopy: true},
		"Status": fields.Integer{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  fields.Boolean{String: isStaffString, Help: IsStaffHelp},
		"IsActive": fields.Boolean{},
		"Profile":  fields.One2One{OnDelete: models.SetNull, RelationModel: h.Profile()},
		"Age": fields.Integer{Compute: h.User().Methods().ComputeAge(),
			Inverse: h.User().Methods().InverseSetAge(),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
		"Posts":     fields.One2Many{RelationModel: h.Post(), ReverseFK: "User", Copy: true},
		"PMoney":    fields.Float{Related: "Profile.Money"},
		"Resume":    fields.Many2One{RelationModel: h.Resume(), Embed: true},
		"LastPost":  fields.Many2One{RelationModel: h.Post()},
		"Email2":    fields.Char{},
		"IsPremium": fields.Boolean{String: isPremiumString, Help: isPremiumHelp},
		"Nums":      fields.Integer{GoType: new(int)},
		"Size":      fields.Float{},
		"Education": fields.Text{String: "Educational Background"},
	})

	models.NewModel("Profile")
	h.Profile().InheritModel(h.AddressMixIn())

	h.Profile().AddMethod("PrintAddress",
		`PrintAddress is a sample method layer for testing`,
		func(rs m.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("%s, %s", res, rs.Country())
		})

	h.Profile().Methods().PrintAddress().Extend("",
		func(rs m.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("[%s]", res)
		})

	h.Profile().AddFields(map[string]models.FieldDefinition{
		"Age":      fields.Integer{GoType: new(int16)},
		"Gender":   fields.Selection{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    fields.Float{},
		"User":     fields.Rev2One{RelationModel: h.User(), ReverseFK: "Profile"},
		"BestPost": fields.Many2One{RelationModel: h.Post()},
		"Country":  fields.Char{},
		"UserName": fields.Char{Related: "User.Name"},
		"Action":   fields.Char{GoType: new(actions.ActionRef)},
	})
	h.Profile().Fields().Zip().SetString("Zip Code")

	models.NewModel("Post")
	h.Post().AddFields(map[string]models.FieldDefinition{
		"User":             fields.Many2One{RelationModel: h.User()},
		"Title":            fields.Char{Required: true},
		"Content":          fields.HTML{},
		"Tags":             fields.Many2Many{RelationModel: h.Tag()},
		"Abstract":         fields.Text{},
		"Attachment":       fields.Binary{},
		"LastRead":         fields.Date{},
		"Comments":         fields.One2Many{RelationModel: h.Comment(), ReverseFK: "Post"},
		"FirstCommentText": fields.Text{Related: "Comments.Text"},
		"FirstTagName":     fields.Char{Related: "Tags.Name"},
		"WriterMoney":      fields.Float{Related: "User.PMoney"},
	})

	h.Post().Methods().Create().Extend("",
		func(rs m.PostSet, data m.PostData) m.PostSet {
			res := rs.Super().Create(data)
			return res
		})

	h.Post().Methods().Search().Extend("",
		func(rs m.PostSet, cond q.PostCondition) m.PostSet {
			res := rs.Super().Search(cond)
			return res
		})

	models.NewModel("Comment")
	h.Comment().AddFields(map[string]models.FieldDefinition{
		"Post":        fields.Many2One{RelationModel: h.Post()},
		"PostWriter":  fields.Many2One{RelationModel: h.User(), Related: "Post.User"},
		"WriterMoney": fields.Float{Related: "PostWriter.PMoney"},
		"Text":        fields.Text{},
	})

	models.NewModel("Tag")
	h.Tag().SetDefaultOrder("Name DESC", "ID ASC")

	h.Tag().AddMethod("CheckNameDescription",
		`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
		func(rs m.TagSet) {
			if rs.Rate() < 0 || rs.Rate() > 10 {
				log.Panic("Tag rate must be between 0 and 10")
			}
		}).AllowGroup(security.GroupEveryone)

	h.Tag().AddMethod("CheckRate",
		`CheckNameDescription checks that the description of a tag is not equal to its name`,
		func(rs m.TagSet) {
			if rs.Name() == rs.Description() {
				log.Panic("Tag name and description must be different")
			}
		})

	h.Tag().AddFields(map[string]models.FieldDefinition{
		"Name":        fields.Char{Constraint: h.Tag().Methods().CheckNameDescription()},
		"BestPost":    fields.Many2One{RelationModel: h.Post()},
		"Posts":       fields.Many2Many{RelationModel: h.Post()},
		"Parent":      fields.Many2One{RelationModel: h.Tag()},
		"Description": fields.Char{Constraint: h.Tag().Methods().CheckNameDescription()},
		"Rate":        fields.Float{Constraint: h.Tag().Methods().CheckRate(), GoType: new(float32)},
	})

	models.NewModel("Resume")

	h.Resume().Methods().Create().Extend("",
		func(rs m.ResumeSet, data m.ResumeData) m.ResumeSet {
			return rs.Super().Create(data)
		})

	h.Resume().AddMethod("ComputeOther",
		`Dummy compute function`,
		func(rs m.ResumeSet) m.ResumeData {
			return h.Resume().NewData().SetOther("Other information")
		})

	h.Resume().AddFields(map[string]models.FieldDefinition{
		"Education":  fields.Text{},
		"Experience": fields.Text{Translate: true},
		"Leisure":    fields.Text{},
		"Other":      fields.Char{Compute: h.Resume().Methods().ComputeOther()},
	})

	addressMI2 := h.AddressMixIn()
	addressMI2.AddMethod("SayHello",
		`SayHello is a sample method layer for testing`,
		func(rs m.AddressMixInSet) string {
			return "Hello !"
		})

	addressMI2.AddMethod("PrintAddress",
		`PrintAddressMixIn is a sample method layer for testing`,
		func(rs m.AddressMixInSet) string {
			return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
		})

	addressMI2.Methods().PrintAddress().Extend("",
		func(rs m.AddressMixInSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("<%s>", res)
		})

	models.NewMixinModel("AddressMixIn")
	h.AddressMixIn().AddFields(map[string]models.FieldDefinition{
		"Street": fields.Char{GoType: new(string)},
		"Zip":    fields.Char{},
		"City":   fields.Char{},
	})

	models.NewMixinModel("ActiveMixIn")
	h.ActiveMixIn().AddFields(map[string]models.FieldDefinition{
		"Active": fields.Boolean{},
	})
	h.ModelMixin().InheritModel(h.ActiveMixIn())

	// Chained declaration
	activeMI1 := h.ActiveMixIn()
	activeMI2 := activeMI1
	activeMI2.AddMethod("IsActivated",
		`IsACtivated is a sample method of ActiveMixIn"`,
		func(rs m.ActiveMixInSet) bool {
			return rs.Active()
		})

	models.NewManualModel("UserView")
	h.UserView().AddFields(map[string]models.FieldDefinition{
		"Name": fields.Char{},
		"City": fields.Char{},
	})
}
