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

	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Getting all users and calling `PrefixedUser`", func() {
			users := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			res := users.Call("PrefixedUser", "Prefix")
			So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
		})
		env.Rollback()
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Getting one user (Jane) and checking DisplayName", func() {
			users := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(users.Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
		})
		Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
			users := env.Pool("User").OrderBy("UserName").Fetch()
			So(users.Len(), ShouldEqual, 3)
			userRecs := users.Records()
			So(userRecs[0].Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			So(userRecs[1].Get("DecoratedName"), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
			So(userRecs[2].Get("DecoratedName"), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
		})
		env.Rollback()
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Checking that user Jane is 23", func() {
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Get("Age"), ShouldEqual, 23)
		})
		Convey("Checking that user Will has no age since no profile", func() {
			userWill := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			So(userWill.Get("Age"), ShouldEqual, 0)
		})
		Convey("It's Jane's birthday, change her age, commit and check", func() {
			jane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(jane.Get("UserName"), ShouldEqual, "Jane A. Smith")
			So(jane.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 12345)
			jane.Get("Profile").(RecordCollection).Set("Age", 24)

			jane.Load()
			jane.Get("Profile").(RecordCollection).Load()
			So(jane.Get("Age"), ShouldEqual, 24)
		})
		Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
			userWill := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			userWill.Load()
			So(userWill.Get("UserName"), ShouldEqual, "Will Smith")
			willProfileData := FieldMap{
				"Age":   34,
				"Money": 5100,
			}
			willProfile := env.Pool("Profile").Call("Create", willProfileData)
			userWill.Set("Profile", willProfile)

			userWill.Load()
			So(userWill.Get("Age"), ShouldEqual, 34)
		})
		env.Commit()
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Checking that users PMoney is correct", func() {
			userJohn := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(userJohn.Len(), ShouldEqual, 1)
			So(userJohn.Get("PMoney"), ShouldEqual, 0)
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Get("PMoney"), ShouldEqual, 12345)
			userWill := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			So(userWill.Get("PMoney"), ShouldEqual, 5100)
		})
		Convey("Checking that PMoney is correct after update of Profile", func() {
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Get("PMoney"), ShouldEqual, 12345)
			userJane.Get("Profile").(RecordCollection).Set("Money", 54321)
			So(userJane.Get("PMoney"), ShouldEqual, 54321)
		})
		Convey("Checking that we can update PMoney directly", func() {
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Get("PMoney"), ShouldEqual, 12345)
			userJane.Set("PMoney", 67890)
			So(userJane.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 67890)
			So(userJane.Get("PMoney"), ShouldEqual, 67890)
			userWill := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			So(userWill.Get("PMoney"), ShouldEqual, 5100)

			userJane.Union(userWill).Set("PMoney", 100)
			So(userJane.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 100)
			So(userJane.Get("PMoney"), ShouldEqual, 100)
			So(userWill.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 100)
			So(userWill.Get("PMoney"), ShouldEqual, 100)
		})
		Convey("Checking that we can search PMoney directly", func() {
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			userWill := env.Pool("User").Filter("Email", "=", "will.smith@example.com")
			pmoneyUser := env.Pool("User").Filter("PMoney", "=", 12345)
			So(pmoneyUser.Len(), ShouldEqual, 1)
			So(pmoneyUser.Ids()[0], ShouldEqual, userJane.Ids()[0])
			pUsers := env.Pool("User").Search(NewCondition().And("PMoney", "=", 12345).Or("PMoney", "=", 5100))
			So(pUsers.Len(), ShouldEqual, 2)
			So(pUsers.Ids(), ShouldContain, userJane.Ids()[0])
			So(pUsers.Ids(), ShouldContain, userWill.Ids()[0])
		})
		env.Rollback()
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Adding a last post to Jane", func() {
			postRs := env.Pool("Post").Call("Create", FieldMap{
				"Title":   "This is my title",
				"Content": "Here we have some content",
			}).(RecordCollection)
			env.Pool("User").Filter("Email", "=", "jane.smith@example.com").Set("LastPost", postRs)
		})
		Convey("Checking that we can access jane's post directly", func() {
			userJane := env.Pool("User").Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Get("Title"), ShouldEqual, "This is my title")
			So(userJane.Get("Content"), ShouldEqual, "Here we have some content")
			So(userJane.Get("LastPost").(RecordCollection).Get("Title"), ShouldEqual, "This is my title")
			So(userJane.Get("LastPost").(RecordCollection).Get("Content"), ShouldEqual, "Here we have some content")
		})
		env.Commit()
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Checking that mixed in functions are correctly inherited", func() {
			janeProfile := env.Pool("User").Filter("Email", "=", "jane.smith@example.com").Get("Profile").(RecordCollection)
			So(janeProfile.Call("PrintAddress"), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
			So(janeProfile.Call("SayHello"), ShouldEqual, "Hello !")
		})
		env.Rollback()
	})
}
