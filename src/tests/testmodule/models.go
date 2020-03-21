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

var (
	// IsStaffHelp exported
	IsStaffHelp   = "This is a var help message"
	isPremiumHelp = "This the IsPremium Help message"
)

const (
	isStaffString        = "Is Staff"
	isPremiumDescription = "This is a const description string"
)

const isPremiumString = isPremiumDescription

var fields_User = map[string]models.FieldDefinition{
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
}

// DecorateEmail decorates the email of the given user
func user_DecorateEmail(_ m.UserSet, email string) string {
	return fmt.Sprintf("<%s>", email)
}

func user_OnChangeName(rs m.UserSet) m.UserData {
	return h.User().NewData().SetDecoratedName(rs.PrefixedUser("User")[0])
}

func user_ComputeDecoratedName(rs m.UserSet) m.UserData {
	return h.User().NewData().SetDecoratedName(rs.PrefixedUser("User")[0])
}

func user_ComputeAge(rs m.UserSet) m.UserData {
	return h.User().NewData().SetAge(rs.Profile().Age())
}

func user_PrefixedUser(rs m.UserSet, prefix string) []string {
	var res []string
	for _, u := range rs.Records() {
		res = append(res, fmt.Sprintf("%s: %s", prefix, u.Name()))
	}
	return res
}

func user_ext_DecorateEmail(rs m.UserSet, email string) string {
	res := rs.Super().DecorateEmail(email)
	return fmt.Sprintf("[%s]", res)
}

func user_RecursiveMethod(rs m.UserSet, depth int, result string) string {
	if depth == 0 {
		return result
	}
	return rs.RecursiveMethod(depth-1, fmt.Sprintf("%s, recursion %d", result, depth))
}

func user_ext_RecursiveMethod(rs m.UserSet, depth int, result string) string {
	result = "> " + result + " <"
	sup := rs.Super().RecursiveMethod(depth, result)
	return sup
}

func user_SubSetSuper(rs m.UserSet) string {
	var res string
	for _, rec := range rs.Records() {
		res += rec.Name()
	}
	return res
}

func user_ext_SubSetSuper(rs m.UserSet) string {
	userJane := h.User().Search(rs.Env(), q.User().Email().Equals("jane.smith@example.com"))
	userJohn := h.User().Search(rs.Env(), q.User().Email().Equals("jsmith2@example.com"))
	users := h.User().NewSet(rs.Env())
	users = users.Union(userJane)
	users = users.Union(userJohn)
	return users.Super().SubSetSuper()
}

func user_InverseSetAge(rs m.UserSet, age int16) {
	rs.Profile().SetAge(age)
}

func user_ext_PrefixedUser(rs m.UserSet, prefix string) []string {
	res := rs.Super().PrefixedUser(prefix)
	for i, u := range rs.Records() {
		res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
	}
	return res
}

func user_UpdateCity(rs m.UserSet, value string) {
	rs.Profile().SetCity(value)
}

func user_Aggregates(rs m.UserSet, fieldNames ...models.FieldName) []m.UserGroupAggregateRow {
	return rs.Super().Aggregates(fieldNames...)
}

var fields_Profile = map[string]models.FieldDefinition{
	"Age":      fields.Integer{GoType: new(int16)},
	"Gender":   fields.Selection{Selection: types.Selection{"male": "Male", "female": "Female"}},
	"Money":    fields.Float{},
	"User":     fields.Rev2One{RelationModel: h.User(), ReverseFK: "Profile"},
	"BestPost": fields.Many2One{RelationModel: h.Post()},
	"Country":  fields.Char{},
	"UserName": fields.Char{Related: "User.Name"},
	"Action":   fields.Char{GoType: new(actions.ActionRef)},
}

func profile_PrintAddress(rs m.ProfileSet) string {
	res := rs.Super().PrintAddress()
	return fmt.Sprintf("%s, %s", res, rs.Country())
}

func profile_ext_PrintAddress(rs m.ProfileSet) string {
	res := rs.Super().PrintAddress()
	return fmt.Sprintf("[%s]", res)
}

var fields_Post = map[string]models.FieldDefinition{
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
}

func post_Create(rs m.PostSet, data m.PostData) m.PostSet {
	res := rs.Super().Create(data)
	return res
}

func post_Search(rs m.PostSet, cond q.PostCondition) m.PostSet {
	res := rs.Super().Search(cond)
	return res
}

var fields_Comment = map[string]models.FieldDefinition{
	"Post":        fields.Many2One{RelationModel: h.Post()},
	"PostWriter":  fields.Many2One{RelationModel: h.User(), Related: "Post.User"},
	"WriterMoney": fields.Float{Related: "PostWriter.PMoney"},
	"Text":        fields.Text{},
}

var fields_Tag = map[string]models.FieldDefinition{
	"Name":        fields.Char{Constraint: h.Tag().Methods().CheckNameDescription()},
	"BestPost":    fields.Many2One{RelationModel: h.Post()},
	"Posts":       fields.Many2Many{RelationModel: h.Post()},
	"Parent":      fields.Many2One{RelationModel: h.Tag()},
	"Description": fields.Char{Constraint: h.Tag().Methods().CheckNameDescription()},
	"Rate":        fields.Float{Constraint: h.Tag().Methods().CheckRate(), GoType: new(float32)},
}

func tag_CheckNameDescription(rs m.TagSet) {
	if rs.Rate() < 0 || rs.Rate() > 10 {
		log.Panic("Tag rate must be between 0 and 10")
	}
}

func tag_CheckRate(rs m.TagSet) {
	if rs.Name() == rs.Description() {
		log.Panic("Tag name and description must be different")
	}
}

var fields_Resume = map[string]models.FieldDefinition{
	"Education":  fields.Text{},
	"Experience": fields.Text{Translate: true},
	"Leisure":    fields.Text{},
	"Other":      fields.Char{Compute: h.Resume().Methods().ComputeOther()},
}

func resume_Create(rs m.ResumeSet, data m.ResumeData) m.ResumeSet {
	return rs.Super().Create(data)
}

func resume_ComputeOther(_ m.ResumeSet) m.ResumeData {
	return h.Resume().NewData().SetOther("Other information")
}

var fields_AddressMixIn = map[string]models.FieldDefinition{
	"Street": fields.Char{GoType: new(string)},
	"Zip":    fields.Char{},
	"City":   fields.Char{},
}

func addressMixIn_SayHello(_ m.AddressMixInSet) string {
	return "Hello !"
}

func addressMixIn_PrintAddress(rs m.AddressMixInSet) string {
	return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
}

func addressMixIn_ext_PrintAddress(rs m.AddressMixInSet) string {
	res := rs.Super().PrintAddress()
	return fmt.Sprintf("<%s>", res)
}

var fields_ActiveMixin = map[string]models.FieldDefinition{
	"Active": fields.Boolean{},
}

func activeMixIn_IsActivated(rs m.ActiveMixInSet) bool {
	return rs.Active()
}

var fields_UserView = map[string]models.FieldDefinition{
	"Name": fields.Char{},
	"City": fields.Char{},
}

func init() {
	models.NewModel("User")

	h.User().AddFields(fields_User)
	h.User().Fields().Experience().SetString("Professional Experience")

	h.User().NewMethod("OnChangeName", user_OnChangeName)
	h.User().NewMethod("ComputeDecoratedName", user_ComputeDecoratedName)
	h.User().NewMethod("ComputeAge", user_ComputeAge)
	h.User().NewMethod("PrefixedUser", user_PrefixedUser)
	h.User().NewMethod("DecorateEmail", user_DecorateEmail)
	h.User().NewMethod("RecursiveMethod", user_RecursiveMethod)
	h.User().NewMethod("SubSetSuper", user_SubSetSuper)
	h.User().NewMethod("InverseSetAge", user_InverseSetAge)
	h.User().NewMethod("UpdateCity", user_UpdateCity)
	h.User().Methods().DecorateEmail().Extend(user_ext_DecorateEmail)
	h.User().Methods().RecursiveMethod().Extend(user_ext_RecursiveMethod)
	h.User().Methods().SubSetSuper().Extend(user_ext_SubSetSuper)
	h.User().Methods().PrefixedUser().Extend(user_ext_PrefixedUser)
	h.User().Methods().Aggregates().Extend(user_Aggregates)

	models.NewModel("Profile")
	h.Profile().InheritModel(h.AddressMixIn())

	h.Profile().AddFields(fields_Profile)
	h.Profile().Fields().Zip().SetString("Zip Code")

	h.Profile().Methods().PrintAddress().Extend(profile_PrintAddress)
	h.Profile().Methods().PrintAddress().Extend(profile_ext_PrintAddress)

	models.NewModel("Post")

	h.Post().AddFields(fields_Post)

	h.Post().Methods().Create().Extend(post_Create)
	h.Post().Methods().Search().Extend(post_Search)

	models.NewModel("Comment")

	h.Comment().AddFields(fields_Comment)

	models.NewModel("Tag")
	h.Tag().SetDefaultOrder("Name DESC", "ID ASC")

	h.Tag().AddFields(fields_Tag)

	h.Tag().NewMethod("CheckNameDescription", tag_CheckNameDescription).AllowGroup(security.GroupEveryone)
	h.Tag().NewMethod("CheckRate", tag_CheckRate)

	models.NewModel("Resume")

	h.Resume().AddFields(fields_Resume)

	h.Resume().Methods().Create().Extend(resume_Create)
	h.Resume().NewMethod("ComputeOther", resume_ComputeOther)

	addressMI2 := h.AddressMixIn()
	addressMI2.NewMethod("SayHello", addressMixIn_SayHello)
	addressMI2.NewMethod("PrintAddress", addressMixIn_PrintAddress)
	addressMI2.Methods().PrintAddress().Extend(addressMixIn_ext_PrintAddress)

	models.NewMixinModel("AddressMixIn")
	h.AddressMixIn().AddFields(fields_AddressMixIn)

	models.NewMixinModel("ActiveMixIn")
	h.ActiveMixIn().AddFields(fields_ActiveMixin)
	h.ModelMixin().InheritModel(h.ActiveMixIn())

	// Chained declaration
	activeMI1 := h.ActiveMixIn()
	activeMI2 := activeMI1
	activeMI2.NewMethod("IsActivated", activeMixIn_IsActivated)

	models.NewManualModel("UserView")
	h.UserView().AddFields(fields_UserView)
}
