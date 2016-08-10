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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type User_WithDecoratedName struct {
	ID            int64
	UserName      string
	Email         string
	Profile       *Profile_WithID
	DecoratedName string
	DisplayName   string
}

type Profile_Simple struct {
	ID    int64
	Age   int16
	Money float64
}

type User_Simple struct {
	ID       int64
	UserName string
	Profile  *Profile_Simple
	Age      int16
	PMoney   float64
	Title    string
	Content  string
}

type User_WithLastPost struct {
	ID       int64
	Title    string
	Content  string
	LastPost *Post
}

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		env := NewEnvironment(1)
		Convey("Getting all users and calling `PrefixedUser`", func() {
			users := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			res := users.Call("PrefixedUser", "Prefix")
			So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
		})
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		env := NewEnvironment(1)
		Convey("Getting one user (Jane) and checking DisplayName", func() {
			var userJane User_WithDecoratedName
			users := env.Pool("User")
			users.Filter("Email", "=", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.DecoratedName, ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
		})
		Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
			var users []*User_WithDecoratedName
			env.Pool("User").OrderBy("UserName").ReadAll(&users)
			So(len(users), ShouldEqual, 3)
			So(users[0].DecoratedName, ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			So(users[1].DecoratedName, ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
			So(users[2].DecoratedName, ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
		})
		Convey("Getting all users (Jane & Will) by values and checking DecoratedName", func() {
			var params []FieldMap
			env.Pool("User").OrderBy("UserName").ReadValues(&params)
			So(params[0]["decorated_name"], ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			So(params[1]["decorated_name"], ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
			So(params[2]["decorated_name"], ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
		})
		env.cr.Rollback()
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		env := NewEnvironment(1)
		Convey("Checking that user Jane is 23", func() {
			var userJane User_Simple
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.Age, ShouldEqual, 23)
		})
		Convey("Checking that user Will has no age since no profile", func() {
			var userWill User_Simple
			env.Pool("User").Filter("Email", "=", "will.smith@example.com").ReadOne(&userWill)
			So(userWill.Age, ShouldEqual, 0)
		})
		Convey("It's Jane's birthday, change her age, commit and check", func() {
			var userJane User_Simple
			janeRs := env.Pool("User").RelatedDepth(1).Filter("Email", "=", "jane.smith@example.com")
			janeRs.ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane A. Smith")
			So(userJane.Profile.Money, ShouldEqual, 12345)
			userJane.Profile.Age = 24
			env.Sync(userJane.Profile)
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.Age, ShouldEqual, 24)
		})
		Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
			var userWill User_Simple
			willRs := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			willRs.ReadOne(&userWill)
			So(userWill.UserName, ShouldEqual, "Will Smith")
			userWill.Profile = &Profile_Simple{
				Age:   34,
				Money: 5100,
			}
			env.Create(userWill.Profile)
			env.Sync(&userWill)
			env.Pool("User").Filter("Email", "=", "will.smith@example.com").ReadOne(&userWill)
			So(userWill.Age, ShouldEqual, 34)
		})
		env.cr.Commit()
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		env := NewEnvironment(1)
		Convey("Checking that user Jane money equals is 12345", func() {
			var userJane User_Simple
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.PMoney, ShouldEqual, 12345)
		})
		env.cr.Rollback()
	})
}

func TestInheritedModels(t *testing.T) {
	Convey("Testing inherits-ed models", t, func() {
		env := NewEnvironment(1)
		Convey("Adding a last post to Jane", func() {
			postRs := env.Pool("Post").Create(FieldMap{
				"Title":   "This is my title",
				"Content": "Here we have some content",
			})
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").Write(FieldMap{
				"LastPost": postRs.ID(),
			})
		})
		Convey("Checking that we can access jane's post directly", func() {
			var userJane User_Simple
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.Title, ShouldEqual, "This is my title")
			So(userJane.Content, ShouldEqual, "Here we have some content")
			var userJane2 User_WithLastPost
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").RelatedDepth(1).ReadOne(&userJane2)
			So(userJane2.Title, ShouldEqual, "This is my title")
			So(userJane2.Content, ShouldEqual, "Here we have some content")
			So(userJane2.LastPost.Title, ShouldEqual, "This is my title")
			So(userJane2.LastPost.Content, ShouldEqual, "Here we have some content")
		})
		env.cr.Commit()
	})
}
