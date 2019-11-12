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
	"reflect"
	"testing"

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/models/types/dates"
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
					res = append(res, fmt.Sprintf("%s: %s", prefix, u.Get(u.Model().FieldName("Name"))))
				}
				return res
			})

		userModel.Methods().MustGet("PrefixedUser").Extend("",
			func(rc *RecordCollection, prefix string) []string {
				res := rc.Super().Call("PrefixedUser", prefix).([]string)
				for i, u := range rc.Records() {
					mail := u.Get(u.Model().FieldName("Email")).(string)
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
					res += rec.Get(rec.Model().FieldName("Name")).(string)
				}
				return res
			})

		userModel.Methods().MustGet("SubSetSuper").Extend("",
			func(rc *RecordCollection) string {
				users := rc.Env().Pool("User")
				emailField := users.Model().FieldName("Email")
				userJane := users.Search(users.Model().Field(emailField).Equals("jane.smith@example.com"))
				userJohn := users.Search(users.Model().Field(emailField).Equals("jsmith2@example.com"))
				users = users.Call("Union", userJane).(RecordSet).Collection()
				users = users.Call("Union", userJohn).(RecordSet).Collection()
				return users.Super().Call("SubSetSuper").(string)
			})

		userModel.AddMethod("OnChangeName", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("OnChangeNameWarning", "",
			func(rc *RecordCollection) string {
				if rc.Get(rc.Model().FieldName("Name")) == "Warning User" {
					return "We have a warning here"
				}
				return ""
			})

		userModel.AddMethod("OnChangeNameFilters", "",
			func(rc *RecordCollection) map[FieldName]Conditioner {
				res := make(map[FieldName]Conditioner)
				res[rc.Model().FieldName("LastPost")] = Registry.MustGet("Profile").Field(Registry.MustGet("Profile").FieldName("Street")).Equals("addr")
				return res
			})

		userModel.AddMethod("ComputeDecoratedName", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("DecoratedName"), rc.Call("PrefixedUser", "User").([]string)[0])
				return res
			})

		userModel.AddMethod("ComputeAge", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				res.Set(rc.Model().FieldName("Age"), rc.Get(rc.Model().FieldName("Profile")).(*RecordCollection).Get(Registry.MustGet("Profile").FieldName("Age")).(int16))
				return res
			})

		userModel.AddMethod("InverseSetAge", "",
			func(rc *RecordCollection, age int16) {
				rc.Get(rc.Model().FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("Age"), age)
			})

		userModel.AddMethod("UpdateCity", "",
			func(rc *RecordCollection, value string) {
				rc.Get(rc.Model().FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("City"), value)
			})

		userModel.AddMethod("ComputeNum", "Dummy method",
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.Model())
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
				rc.Get(rc.Model().FieldName("Profile")).(*RecordCollection).Set(Registry.MustGet("Profile").FieldName("Age"), age)
				return "Ok"
			})

		userModel.AddMethod("ComputeCoolType", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				if rc.Get(rc.Model().FieldName("IsCool")).(bool) {
					res.Set(rc.Model().FieldName("CoolType"), "cool")
				} else {
					res.Set(rc.Model().FieldName("CoolType"), "no-cool")
				}
				return res
			})

		userModel.AddMethod("OnChangeCoolType", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				if rc.Get(rc.Model().FieldName("CoolType")).(string) == "cool" {
					res.Set(rc.Model().FieldName("IsCool"), true)
				} else {
					res.Set(rc.Model().FieldName("IsCool"), false)
				}
				return res
			})

		userModel.AddMethod("InverseCoolType", "",
			func(rc *RecordCollection, val string) {
				if val == "cool" {
					rc.Set(rc.Model().FieldName("IsCool"), true)
				} else {
					rc.Set(rc.Model().FieldName("IsCool"), false)
				}
			})

		userModel.AddMethod("OnChangeMana", "",
			func(rc *RecordCollection) *ModelData {
				res := NewModelData(rc.Model())
				post1 := rc.Env().Pool("Post").SearchAll().Limit(1)
				prof := rc.Env().Pool("Profile").Call("Create",
					NewModelData(Registry.MustGet("Profile")).
						Set(Registry.MustGet("Profile").FieldName("BestPost"), post1))
				prof.(RecordSet).Collection().InvalidateCache()
				res.Set(rc.Model().FieldName("Profile"), prof)
				return res
			})

		userModel.Methods().MustGet("Copy").Extend("",
			func(rc *RecordCollection, overrides RecordData) *RecordCollection {
				nameField := rc.Model().FieldName("Name")
				overrides.Underlying().Set(nameField, fmt.Sprintf("%s (copy)", rc.Get(nameField).(string)))
				return rc.Super().Call("Copy", overrides).(RecordSet).Collection()
			})

		activeMI.AddMethod("IsActivated", "",
			func(rc *RecordCollection) bool {
				return rc.Get(rc.Model().FieldName("Active")).(bool)
			})

		addressMI.AddMethod("SayHello", "",
			func(rc *RecordCollection) string {
				return "Hello !"
			})

		addressMI.AddMethod("PrintAddress", "",
			func(rc *RecordCollection) string {
				return fmt.Sprintf("%s, %s %s", rc.Get(rc.Model().FieldName("Street")), rc.Get(rc.Model().FieldName("Zip")), rc.Get(rc.Model().FieldName("City")))
			})

		profileModel.AddMethod("PrintAddress", "",
			func(rc *RecordCollection) string {
				res := rc.Super().Call("PrintAddress").(string)
				return fmt.Sprintf("%s, %s", res, rc.Get(rc.Model().FieldName("Country")))
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
				if !rc.Get(rc.Model().FieldName("LastRead")).(dates.Date).IsZero() {
					read = true
				}
				return NewModelData(rc.Model()).Set(rc.Model().FieldName("Read"), read)
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
					for _, tg := range rec.Get(rec.Model().FieldName("Tags")).(RecordSet).Collection().Records() {
						res += tg.Get(tg.Model().FieldName("Name")).(string) + " "
					}
				}
				return NewModelData(rc.Model()).Set(rc.Model().FieldName("TagsNames"), res)
			})

		post.AddMethod("ComputeWriterAge", "",
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.Model()).
					Set(rc.Model().FieldName("WriterAge"),
						rc.Get(rc.Model().FieldName("User")).(RecordSet).Collection().Get(Registry.MustGet("User").FieldName("Age")).(int16))
			})

		tag.AddMethod("CheckRate",
			`CheckRate checks that the given RecordSet has a rate between 0 and 10`,
			func(rc *RecordCollection) {
				if rc.Get(rc.Model().FieldName("Rate")).(float32) < 0 || rc.Get(rc.Model().FieldName("Rate")).(float32) > 10 {
					log.Panic("Tag rate must be between 0 and 10")
				}
			})

		tag.AddMethod("CheckNameDescription",
			`CheckNameDescription checks that the description of a tag is not equal to its name`,
			func(rc *RecordCollection) {
				if rc.Get(rc.Model().FieldName("Name")).(string) == rc.Get(rc.Model().FieldName("Description")).(string) {
					log.Panic("Tag name and description must be different")
				}
			})

		tag.Methods().AllowAllToGroup(security.GroupEveryone)
		tag.Methods().RevokeAllFromGroup(security.GroupEveryone)
		tag.Methods().AllowAllToGroup(security.GroupEveryone)

		cv.AddMethod("ComputeOther",
			`Dummy compute function`,
			func(rc *RecordCollection) *ModelData {
				return NewModelData(rc.Model()).Set(rc.Model().FieldName("Other"), "Other information")
			})

		userModel.fields.add(&Field{
			model:           userModel,
			name:            "Name",
			json:            "name",
			description:     "Name",
			fieldType:       fieldtype.Char,
			structField:     reflect.StructField{Type: reflect.TypeOf("")},
			help:            "The user's username",
			unique:          true,
			noCopy:          true,
			onChange:        "OnChangeName",
			onChangeWarning: "OnChangeNameWarning",
			onChangeFilters: "OnChangeNameFilters",
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "DecoratedName",
			json:        "decorated_name",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			compute:     "ComputeDecoratedName",
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Email",
			json:        "email",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			help:        "The user's email address",
			size:        100,
			index:       true,
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Password",
			json:        "password",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			noCopy:      true,
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Status",
			json:        "status_json",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(int16(0))},
			defaultFunc: DefaultValue(int16(12)),
			readOnly:    true,
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "IsStaff",
			json:        "is_staff",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			defaultFunc: DefaultValue(false),
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "IsActive",
			json:        "is_active",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			defaultFunc: DefaultValue(false),
		})
		userModel.fields.add(&Field{
			model:            userModel,
			name:             "Profile",
			json:             "profile_id",
			fieldType:        fieldtype.One2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			relatedModelName: "Profile",
			onDelete:         SetNull,
			required:         true,
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Age",
			json:        "age",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(int16(0))},
			compute:     "ComputeAge",
			inverse:     "InverseSetAge",
			depends:     []string{"Profile", "Profile.Age"},
			stored:      true,
			defaultFunc: DefaultValue(0),
		})
		userModel.fields.add(&Field{
			model:            userModel,
			name:             "Posts",
			json:             "posts_ids",
			fieldType:        fieldtype.One2Many,
			structField:      reflect.StructField{Type: reflect.TypeOf([]int64{})},
			relatedModelName: "Post",
			reverseFK:        "User",
			noCopy:           false,
		})
		userModel.fields.add(&Field{
			model:          userModel,
			name:           "PMoney",
			json:           "p_money",
			fieldType:      fieldtype.Float,
			structField:    reflect.StructField{Type: reflect.TypeOf(float64(1))},
			relatedPathStr: "Profile.Money",
			defaultFunc:    DefaultValue(0),
		})
		userModel.fields.add(&Field{
			model:            userModel,
			name:             "LastPost",
			json:             "last_post_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "Post",
		})
		userModel.fields.add(&Field{
			model:            userModel,
			name:             "Resume",
			json:             "resume_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			relatedModelName: "Resume",
			onDelete:         Cascade,
			required:         false,
			noCopy:           true,
			embed:            true,
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "IsCool",
			json:        "is_cool",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			defaultFunc: DefaultValue(false),
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "CoolType",
			json:        "cool_type",
			fieldType:   fieldtype.Selection,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			selection: types.Selection{
				"cool":    "Yes, its a cool user",
				"no-cool": "No, forget it"},
			compute:  "ComputeCoolType",
			inverse:  "InverseCoolType",
			onChange: "OnChangeCoolType",
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Email2",
			json:        "email2",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "IsPremium",
			json:        "is_premium",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			defaultFunc: DefaultValue(false),
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Nums",
			json:        "nums",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(0)},
			defaultFunc: DefaultValue(0),
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Size",
			json:        "size",
			fieldType:   fieldtype.Float,
			structField: reflect.StructField{Type: reflect.TypeOf(float64(0))},
			defaultFunc: DefaultValue(0),
		})
		userModel.fields.add(&Field{
			model:            userModel,
			name:             "BestProfilePost",
			json:             "best_profile_post_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "Post",
			relatedPathStr:   "Profile.BestPost",
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Mana",
			json:        "mana",
			fieldType:   fieldtype.Float,
			structField: reflect.StructField{Type: reflect.TypeOf(float32(0))},
			onChange:    "OnChangeMana",
			defaultFunc: DefaultValue(0),
		})
		userModel.fields.add(&Field{
			model:       userModel,
			name:        "Education",
			json:        "education",
			fieldType:   fieldtype.Text,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			description: "Educational Background",
		})
		userModel.AddSQLConstraint("nums_premium", "CHECK((is_premium = TRUE AND nums IS NOT NULL AND nums > 0) OR (IS_PREMIUM = false))",
			"Premium users must have positive nums")

		profileModel.fields.add(&Field{
			model:       profileModel,
			name:        "Age",
			json:        "age",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(int16(0))},
			defaultFunc: DefaultValue(0),
		})
		profileModel.fields.add(&Field{
			model:       profileModel,
			name:        "Gender",
			json:        "gender",
			fieldType:   fieldtype.Selection,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			selection:   types.Selection{"male": "Male", "female": "Female"},
		})
		profileModel.fields.add(&Field{
			model:       profileModel,
			name:        "Money",
			json:        "money",
			fieldType:   fieldtype.Float,
			structField: reflect.StructField{Type: reflect.TypeOf(float64(0))},
			defaultFunc: DefaultValue(0),
		})
		profileModel.fields.add(&Field{
			model:            profileModel,
			name:             "User",
			json:             "user_id",
			fieldType:        fieldtype.Rev2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			relatedModelName: "User",
			reverseFK:        "Profile",
			noCopy:           true,
		})
		profileModel.fields.add(&Field{
			model:            profileModel,
			name:             "BestPost",
			json:             "best_post_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         Cascade,
			relatedModelName: "Post",
		})
		profileModel.fields.add(&Field{
			model:       profileModel,
			name:        "City",
			json:        "city",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		profileModel.fields.add(&Field{
			model:       profileModel,
			name:        "Country",
			json:        "country",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		profileModel.fields.add(&Field{
			model:          profileModel,
			name:           "UserName",
			json:           "user_name",
			fieldType:      fieldtype.Char,
			structField:    reflect.StructField{Type: reflect.TypeOf("")},
			relatedPathStr: "User.Name",
		})
		post.fields.add(&Field{
			model:            post,
			name:             "User",
			json:             "user_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "User",
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Title",
			json:        "title",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			required:    true,
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Content",
			json:        "content",
			fieldType:   fieldtype.HTML,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			required:    true,
		})
		m2mRelModel, m2mOurField, m2mTheirField := CreateM2MRelModelInfo("PostTagRel", "Post", "Tag", "Post", "Tag", false)
		post.fields.add(&Field{
			model:            post,
			name:             "Tags",
			json:             "tags_ids",
			fieldType:        fieldtype.Many2Many,
			structField:      reflect.StructField{Type: reflect.TypeOf([]int64{})},
			relatedModelName: "Tag",
			m2mRelModel:      m2mRelModel,
			m2mOurField:      m2mOurField,
			m2mTheirField:    m2mTheirField,
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Abstract",
			json:        "abstract",
			fieldType:   fieldtype.Text,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Attachment",
			json:        "attachment",
			fieldType:   fieldtype.Binary,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Read",
			json:        "read",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			compute:     "ComputeRead",
			defaultFunc: DefaultValue(false),
		})
		post.fields.add(&Field{
			model:       post,
			name:        "LastRead",
			json:        "last_read",
			fieldType:   fieldtype.Date,
			structField: reflect.StructField{Type: reflect.TypeOf(dates.Date{})},
		})
		post.fields.add(&Field{
			model:       post,
			name:        "Visibility",
			json:        "visibility",
			fieldType:   fieldtype.Selection,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			selection: types.Selection{
				"invisible": "Invisible",
				"visible":   "Visible",
			},
		})
		post.fields.add(&Field{
			model:            post,
			name:             "Comments",
			json:             "comments_ids",
			fieldType:        fieldtype.One2Many,
			structField:      reflect.StructField{Type: reflect.TypeOf([]int64{})},
			relatedModelName: "Comment",
			reverseFK:        "Post",
			noCopy:           true,
		})
		post.fields.add(&Field{
			model:          post,
			name:           "LastCommentText",
			json:           "last_comment_text",
			fieldType:      fieldtype.Text,
			structField:    reflect.StructField{Type: reflect.TypeOf("")},
			relatedPathStr: "Comments.Text",
		})
		post.fields.add(&Field{
			model:          post,
			name:           "LastTagName",
			json:           "last_tag_name",
			fieldType:      fieldtype.Char,
			structField:    reflect.StructField{Type: reflect.TypeOf("")},
			relatedPathStr: "Tags.Name",
		})
		post.fields.add(&Field{
			model:       post,
			name:        "TagsNames",
			json:        "tags_names",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			compute:     "ComputeTagsNames",
		})
		post.fields.add(&Field{
			model:       post,
			name:        "WriterAge",
			json:        "writer_age",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(int16(0))},
			compute:     "ComputeWriterAge",
			depends:     []string{"User.Age"},
			stored:      true,
			defaultFunc: DefaultValue(0),
		})
		post.fields.add(&Field{
			model:          post,
			name:           "WriterMoney",
			json:           "writer_money",
			fieldType:      fieldtype.Float,
			structField:    reflect.StructField{Type: reflect.TypeOf(float64(0))},
			relatedPathStr: "User.PMoney",
			defaultFunc:    DefaultValue(0),
		})
		post.SetDefaultOrder("Title")

		comment.fields.add(&Field{
			model:            comment,
			name:             "Post",
			json:             "post_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "Post",
		})
		comment.fields.add(&Field{
			model:            comment,
			name:             "PostWriter",
			json:             "post_writer_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "User",
			relatedPathStr:   "Post.User",
		})
		comment.fields.add(&Field{
			model:          comment,
			name:           "WriterMoney",
			json:           "writer_money",
			fieldType:      fieldtype.Float,
			structField:    reflect.StructField{Type: reflect.TypeOf(float64(0))},
			relatedPathStr: "PostWriter.PMoney",
			defaultFunc:    DefaultValue(0),
		})
		comment.fields.add(&Field{
			model:       comment,
			name:        "Text",
			json:        "text",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})

		tag.fields.add(&Field{
			model:       tag,
			name:        "Name",
			json:        "name",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			constraint:  "CheckNameDescription",
		})
		tag.fields.add(&Field{
			model:            tag,
			name:             "BestPost",
			json:             "best_post_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "Post",
		})
		tag.fields.add(&Field{
			model:            tag,
			name:             "Posts",
			json:             "posts_ids",
			fieldType:        fieldtype.Many2Many,
			structField:      reflect.StructField{Type: reflect.TypeOf([]int64{})},
			relatedModelName: "Post",
			m2mRelModel:      m2mRelModel,
			m2mOurField:      m2mTheirField,
			m2mTheirField:    m2mOurField,
		})
		tag.fields.add(&Field{
			model:            tag,
			name:             "Parent",
			json:             "parent_id",
			fieldType:        fieldtype.Many2One,
			structField:      reflect.StructField{Type: reflect.TypeOf(int64(0))},
			onDelete:         SetNull,
			relatedModelName: "Tag",
		})
		tag.fields.add(&Field{
			model:       tag,
			name:        "Description",
			json:        "description",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			contexts: FieldContexts{"lang": func(rs RecordSet) string {
				res := rs.Env().Context().GetString("lang")
				return res
			}},
			constraint: "CheckNameDescription",
		})
		tag.fields.add(&Field{
			model:       tag,
			name:        "Note",
			json:        "note",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			contexts: FieldContexts{"lang": func(rs RecordSet) string {
				res := rs.Env().Context().GetString("lang")
				return res
			}},
			required:    true,
			defaultFunc: DefaultValue("Default Note"),
		})
		tag.fields.add(&Field{
			model:       tag,
			name:        "Rate",
			json:        "rate",
			fieldType:   fieldtype.Float,
			structField: reflect.StructField{Type: reflect.TypeOf(float32(0))},
			constraint:  "CheckRate",
			defaultFunc: DefaultValue(0),
		})
		tag.SetDefaultOrder("Name DESC", "ID ASC")

		cv.fields.add(&Field{
			model:       cv,
			name:        "Education",
			json:        "education",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		cv.fields.add(&Field{
			model:       cv,
			name:        "Experience",
			json:        "experience",
			fieldType:   fieldtype.Text,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			contexts: FieldContexts{"lang": func(rs RecordSet) string {
				res := rs.Env().Context().GetString("lang")
				return res
			}},
		})
		cv.fields.add(&Field{
			model:       cv,
			name:        "Leisure",
			json:        "leisure",
			fieldType:   fieldtype.Text,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		cv.fields.add(&Field{
			model:       cv,
			name:        "Other",
			json:        "other",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
			compute:     "ComputeOther",
		})

		addressMI.fields.add(&Field{
			model:       addressMI,
			name:        "Street",
			json:        "street",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		addressMI.fields.add(&Field{
			model:       addressMI,
			name:        "Zip",
			json:        "zip",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		addressMI.fields.add(&Field{
			model:       addressMI,
			name:        "City",
			json:        "city",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		profileModel.InheritModel(addressMI)

		activeMI.fields.add(&Field{
			model:       activeMI,
			name:        "Active",
			json:        "active",
			fieldType:   fieldtype.Boolean,
			structField: reflect.StructField{Type: reflect.TypeOf(false)},
			defaultFunc: DefaultValue(true),
		})
		Registry.MustGet("ModelMixin").InheritModel(activeMI)

		viewModel.fields.add(&Field{
			model:       viewModel,
			name:        "Name",
			json:        "name",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		viewModel.fields.add(&Field{
			model:       viewModel,
			name:        "City",
			json:        "city",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})

		wizard.fields.add(&Field{
			model:       wizard,
			name:        "Name",
			json:        "name",
			fieldType:   fieldtype.Char,
			structField: reflect.StructField{Type: reflect.TypeOf("")},
		})
		wizard.fields.add(&Field{
			model:       wizard,
			name:        "Value",
			json:        "value",
			fieldType:   fieldtype.Integer,
			structField: reflect.StructField{Type: reflect.TypeOf(int64(0))},
			defaultFunc: DefaultValue(0),
		})
	})
}
