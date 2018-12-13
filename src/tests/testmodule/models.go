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

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

// IsStaffHelp exported
var IsStaffHelp = "This is a var help message"

const (
	isStaffString        = "Is Staff"
	isPremiumDescription = "This is a const description string"
)

const isPremiumString = isPremiumDescription

func init() {
	user := h.User().DeclareModel()

	// Methods directly declared with AddMethod must be defined before being referenced in the field declaration

	user.AddMethod("OnChangeName", "",
		func(rs h.UserSet) *h.UserData {
			return h.NewUserData().SetDecoratedName(rs.PrefixedUser("User")[0])
		})

	user.AddMethod("ComputeDecoratedName", "",
		func(rs h.UserSet) *h.UserData {
			return h.NewUserData().SetDecoratedName(rs.PrefixedUser("User")[0])
		})

	user.AddMethod("ComputeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rs h.UserSet) *h.UserData {
			return h.NewUserData().SetAge(rs.Profile().Age())
		})

	var isPremiumHelp = "This the IsPremium Help message"

	user.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Name", Help: "The user's username", Unique: true,
			NoCopy: true, OnChange: user.Methods().OnChangeName()},
		"DecoratedName": models.CharField{Compute: user.Methods().ComputeDecoratedName()},
		"Email":         models.CharField{Help: "The user's email address", Size: 100, Index: true},
		"Password":      models.CharField{NoCopy: true},
		"Status": models.IntegerField{JSON: "status_json", GoType: new(int16),
			Default: models.DefaultValue(int16(12))},
		"IsStaff":  models.BooleanField{String: isStaffString, Help: IsStaffHelp},
		"IsActive": models.BooleanField{},
		"Profile":  models.One2OneField{RelationModel: h.Profile()},
		"Age": models.IntegerField{Compute: user.Methods().ComputeAge(),
			Inverse: user.Methods().InverseSetAge(),
			Depends: []string{"Profile", "Profile.Age"}, Stored: true, GoType: new(int16)},
		"Posts":     models.One2ManyField{RelationModel: h.Post(), ReverseFK: "User", Copy: true},
		"PMoney":    models.FloatField{Related: "Profile.Money"},
		"Resume":    models.Many2OneField{RelationModel: h.Resume(), Embed: true},
		"LastPost":  models.Many2OneField{RelationModel: h.Post()},
		"Email2":    models.CharField{},
		"IsPremium": models.BooleanField{String: isPremiumString, Help: isPremiumHelp},
		"Nums":      models.IntegerField{GoType: new(int)},
		"Size":      models.FloatField{},
		"Education": models.TextField{String: "Educational Background"},
	})
	user.Fields().Experience().SetString("Professional Experience")

	user.Methods().PrefixedUser().DeclareMethod(
		`PrefixedUser is a sample method layer for testing`,
		func(rs h.UserSet, prefix string) []string {
			var res []string
			for _, u := range rs.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.Name()))
			}
			return res
		})

	user.Methods().DecorateEmail().DeclareMethod(
		`DecorateEmail is a sample method layer for testing`,
		func(rs h.UserSet, email string) string {
			return fmt.Sprintf("<%s>", email)
		})

	user.Methods().DecorateEmail().Extend(
		`DecorateEmailExtension is a sample method layer for testing`,
		func(rs h.UserSet, email string) string {
			res := rs.Super().DecorateEmail(email)
			return fmt.Sprintf("[%s]", res)
		})

	user.Methods().RecursiveMethod().DeclareMethod(
		`RecursiveMethod is a sample method layer for testing`,
		func(rs h.UserSet, depth int, result string) string {
			if depth == 0 {
				return result
			}
			return rs.RecursiveMethod(depth-1, fmt.Sprintf("%s, recursion %d", result, depth))
		})

	user.Methods().RecursiveMethod().Extend("",
		func(rs h.UserSet, depth int, result string) string {
			result = "> " + result + " <"
			sup := rs.Super().RecursiveMethod(depth, result)
			return sup
		})

	user.Methods().SubSetSuper().DeclareMethod("",
		func(rs h.UserSet) string {
			var res string
			for _, rec := range rs.Records() {
				res += rec.Name()
			}
			return res
		})

	user.Methods().SubSetSuper().Extend("",
		func(rs h.UserSet) string {
			userJane := h.User().Search(rs.Env(), q.User().Email().Equals("jane.smith@example.com"))
			userJohn := h.User().Search(rs.Env(), q.User().Email().Equals("jsmith2@example.com"))
			users := h.User().NewSet(rs.Env())
			users = users.Union(userJane)
			users = users.Union(userJohn)
			return users.Super().SubSetSuper()
		})

	user.Methods().InverseSetAge().DeclareMethod("",
		func(rs h.UserSet, age int16) {
			rs.Profile().SetAge(age)
		})

	h.User().Methods().PrefixedUser().Extend("",
		func(rs h.UserSet, prefix string) []string {
			res := rs.Super().PrefixedUser(prefix)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	h.User().Methods().UpdateCity().DeclareMethod("",
		func(rs h.UserSet, value string) {
			rs.Profile().SetCity(value)
		})

	profile := h.Profile().DeclareModel()
	profile.AddFields(map[string]models.FieldDefinition{
		"Age":      models.IntegerField{GoType: new(int16)},
		"Gender":   models.SelectionField{Selection: types.Selection{"male": "Male", "female": "Female"}},
		"Money":    models.FloatField{},
		"User":     models.Rev2OneField{RelationModel: h.User(), ReverseFK: "Profile"},
		"BestPost": models.Many2OneField{RelationModel: h.Post()},
		"Country":  models.CharField{},
		"UserName": models.CharField{Related: "User.Name"},
	})
	profile.Fields().Zip().SetString("Zip Code")

	post := h.Post().DeclareModel()
	post.AddFields(map[string]models.FieldDefinition{
		"User":            models.Many2OneField{RelationModel: h.User()},
		"Title":           models.CharField{Required: true},
		"Content":         models.HTMLField{},
		"Tags":            models.Many2ManyField{RelationModel: h.Tag()},
		"Abstract":        models.TextField{},
		"Attachment":      models.BinaryField{},
		"LastRead":        models.DateField{},
		"Comments":        models.One2ManyField{RelationModel: h.Comment(), ReverseFK: "Post"},
		"LastCommentText": models.TextField{Related: "Comments.Text"},
		"LastTagName":     models.CharField{Related: "Tags.Name"},
	})

	h.Post().Methods().Create().Extend("",
		func(rs h.PostSet, data *h.PostData) h.PostSet {
			res := rs.Super().Create(data)
			return res
		})

	h.Post().Methods().Search().Extend("",
		func(rs h.PostSet, cond q.PostCondition) h.PostSet {
			res := rs.Super().Search(cond)
			return res
		})

	comment := h.Comment().DeclareModel()
	comment.AddFields(map[string]models.FieldDefinition{
		"Post": models.Many2OneField{RelationModel: h.Post()},
		"Text": models.TextField{},
	})

	tag := h.Tag().DeclareModel()
	tag.AddFields(map[string]models.FieldDefinition{
		"Name":        models.CharField{Constraint: tag.Methods().CheckNameDescription()},
		"BestPost":    models.Many2OneField{RelationModel: h.Post()},
		"Posts":       models.Many2ManyField{RelationModel: h.Post()},
		"Parent":      models.Many2OneField{RelationModel: h.Tag()},
		"Description": models.CharField{Constraint: tag.Methods().CheckNameDescription()},
		"Rate":        models.FloatField{Constraint: tag.Methods().CheckRate(), GoType: new(float32)},
	})

	tag.Methods().CheckNameDescription().DeclareMethod(
		`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
		func(rs h.TagSet) {
			if rs.Rate() < 0 || rs.Rate() > 10 {
				log.Panic("Tag rate must be between 0 and 10")
			}
		}).AllowGroup(security.GroupEveryone)

	tag.Methods().CheckRate().DeclareMethod(
		`CheckNameDescription checks that the description of a tag is not equal to its name`,
		func(rs h.TagSet) {
			if rs.Name() == rs.Description() {
				log.Panic("Tag name and description must be different")
			}
		})

	cv := h.Resume().DeclareModel()
	cv.AddFields(map[string]models.FieldDefinition{
		"Education":  models.TextField{},
		"Experience": models.TextField{Translate: true},
		"Leisure":    models.TextField{},
		"Other":      models.CharField{Compute: h.Resume().Methods().ComputeOther()},
	})
	cv.Methods().Create().Extend("",
		func(rs h.ResumeSet, data *h.ResumeData) h.ResumeSet {
			return rs.Super().Create(data)
		})

	cv.Methods().ComputeOther().DeclareMethod(
		`Dummy compute function`,
		func(rs h.ResumeSet) *h.ResumeData {
			return h.NewResumeData().SetOther("Other information")
		})

	addressMI := h.AddressMixIn().DeclareMixinModel()
	addressMI.AddFields(map[string]models.FieldDefinition{
		"Street": models.CharField{GoType: new(string)},
		"Zip":    models.CharField{},
		"City":   models.CharField{},
	})
	profile.InheritModel(addressMI)

	h.Profile().Methods().PrintAddress().DeclareMethod(
		`PrintAddress is a sample method layer for testing`,
		func(rs h.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("%s, %s", res, rs.Country())
		})

	h.Profile().Methods().PrintAddress().Extend("",
		func(rs h.ProfileSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("[%s]", res)
		})

	addressMI2 := h.AddressMixIn()
	addressMI2.Methods().SayHello().DeclareMethod(
		`SayHello is a sample method layer for testing`,
		func(rs h.AddressMixInSet) string {
			return "Hello !"
		})

	addressMI2.Methods().PrintAddress().DeclareMethod(
		`PrintAddressMixIn is a sample method layer for testing`,
		func(rs h.AddressMixInSet) string {
			return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
		})

	addressMI2.Methods().PrintAddress().Extend("",
		func(rs h.AddressMixInSet) string {
			res := rs.Super().PrintAddress()
			return fmt.Sprintf("<%s>", res)
		})

	activeMI := h.ActiveMixIn().DeclareMixinModel()
	activeMI.AddFields(map[string]models.FieldDefinition{
		"Active": models.BooleanField{},
	})
	h.ModelMixin().InheritModel(activeMI)

	// Chained declaration
	activeMI1 := h.ActiveMixIn()
	activeMI2 := activeMI1
	activeMI2.Methods().IsActivated().DeclareMethod(
		`IsACtivated is a sample method of ActiveMixIn"`,
		func(rs h.ActiveMixInSet) bool {
			return rs.Active()
		})

	viewModel := h.UserView().DeclareManualModel()
	viewModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{},
		"City": models.CharField{},
	})
}
