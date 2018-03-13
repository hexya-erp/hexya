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

	"github.com/hexya-erp/hexya/hexya/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Getting all users and calling `PrefixedUser`", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				res := users.Call("PrefixedUser", "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Calling `PrefixedUser` with context", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				res := users.WithContext("use_double_square", true).Call("PrefixedUser", "Prefix")
				So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [[jane.smith@example.com]]")
			})
			Convey("Calling super on subset", func() {
				users := env.Pool("User").SearchAll()
				So(users.Call("SubSetSuper").(string), ShouldEqual, "Jane A. SmithJohn Smith")
			})
			Convey("Calling recursive method", func() {
				users := env.Pool("User")
				So(users.Call("RecursiveMethod", 3, "Start"), ShouldEqual, "> > > > Start <, recursion 3 <, recursion 2 <, recursion 1 <")
			})
		}), ShouldBeNil)
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Getting one user (Jane) and checking DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(users.Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
				users := env.Pool("User").OrderBy("Name").Call("Fetch").(RecordSet).Collection()
				So(users.Len(), ShouldEqual, 3)
				userRecs := users.Records()
				So(userRecs[0].Get("DecoratedName"), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
				So(userRecs[1].Get("DecoratedName"), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
				So(userRecs[2].Get("DecoratedName"), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
			})
			Convey("Testing built-in DisplayName", func() {
				users := env.Pool("User")
				users = users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(users.Get("DisplayName").(string), ShouldEqual, "Jane A. Smith")
			})
		}), ShouldBeNil)
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that user Jane is 23", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("Age"), ShouldEqual, 23)
			})
			Convey("Checking that user Will has no age since no profile", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("Age"), ShouldEqual, 0)
			})
			Convey("It's Jane's birthday, change her age, commit and check", func() {
				jane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(jane.Get("Name"), ShouldEqual, "Jane A. Smith")
				So(jane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 12345)
				jane.Get("Profile").(RecordSet).Collection().Set("Age", 24)

				jane.Load()
				jane.Get("Profile").(RecordSet).Collection().Load()
				So(jane.Get("Age"), ShouldEqual, 24)
			})
			Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get("Name"), ShouldEqual, "Will Smith")
				willProfileData := FieldMap{
					"Age":   36,
					"Money": 5100,
				}
				willProfile := env.Pool("Profile").Call("Create", willProfileData)
				userWill.Set("Profile", willProfile)

				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 36)
			})
			Convey("Checking inverse method by changing will's age", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 36)
				userWill.Set("Age", int16(34))
				So(userWill.Get("Age"), ShouldEqual, 34)
				userWill.Load()
				So(userWill.Get("Age"), ShouldEqual, 34)
			})
			Convey("Checking that setting a computed field with no inverse panics", func() {
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(func() { userWill.Set("DecoratedName", "FooBar") }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that users PMoney is correct", func() {
				userJohn := users.Search(users.Model().Field("Name").Equals("John Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.Get("PMoney"), ShouldEqual, 0)
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("PMoney"), ShouldEqual, 5100)
			})
			Convey("Checking that PMoney is correct after update of Profile", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userJane.Get("Profile").(RecordSet).Collection().Set("Money", 54321)
				So(userJane.Get("PMoney"), ShouldEqual, 54321)
			})
			Convey("Checking that we can update PMoney directly", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Get("PMoney"), ShouldEqual, 12345)
				userJane.Set("PMoney", 67890)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 67890)
				So(userJane.Get("PMoney"), ShouldEqual, 67890)
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				So(userWill.Get("PMoney"), ShouldEqual, 5100)

				userJane.Union(userWill).Set("PMoney", 100)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 100)
				So(userJane.Get("PMoney"), ShouldEqual, 100)
				So(userWill.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 100)
				So(userWill.Get("PMoney"), ShouldEqual, 100)
			})
			Convey("Checking that we can search PMoney directly", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				pmoneyUser := users.Search(users.Model().Field("PMoney").Equals(12345))
				So(pmoneyUser.Len(), ShouldEqual, 1)
				So(pmoneyUser.Ids()[0], ShouldEqual, userJane.Ids()[0])
				pUsers := users.Search(users.Model().Field("PMoney").Equals(12345).Or().Field("PMoney").Equals(5100))
				So(pUsers.Len(), ShouldEqual, 2)
				So(pUsers.Ids(), ShouldContain, userJane.Ids()[0])
				So(pUsers.Ids(), ShouldContain, userWill.Ids()[0])
			})
			Convey("Checking that we can order by PMoney", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userWill := users.Search(users.Model().Field("Email").Equals("will.smith@example.com"))
				userJane.Set("PMoney", 64)
				pUsers := users.SearchAll().OrderBy("PMoney DESC")
				So(pUsers.Len(), ShouldEqual, 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				So(pUsersRecs[1].Equals(userWill), ShouldBeTrue)
				So(pUsersRecs[2].Equals(userJane), ShouldBeTrue)
			})
		}), ShouldBeNil)
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
			Convey("Checking that Jane's resume exists", func() {
				So(userJane.Get("Resume").(RecordSet).IsEmpty(), ShouldBeFalse)
			})
			Convey("Adding a proper resume to Jane", func() {
				userJane.Get("Resume").(RecordSet).Collection().Set("Experience", "Hexya developer for 10 years")
				userJane.Get("Resume").(RecordSet).Collection().Set("Leisure", "Music, Sports")
			})
			Convey("Checking that we can access jane's resume directly", func() {
				So(userJane.Get("Experience"), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get("Leisure"), ShouldEqual, "Music, Sports")
				So(userJane.Get("Resume").(RecordSet).Collection().Get("Experience"), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Get("Resume").(RecordSet).Collection().Get("Leisure"), ShouldEqual, "Music, Sports")
			})
		}), ShouldBeNil)
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			Convey("Checking that mixed in functions are correctly inherited", func() {
				janeProfile := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com")).Get("Profile").(RecordSet).Collection()
				So(janeProfile.Call("PrintAddress"), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
				So(janeProfile.Call("SayHello"), ShouldEqual, "Hello !")
			})
			Convey("Checking mixing in all models", func() {
				userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
				userJane.Set("Active", true)
				So(userJane.Get("Active").(bool), ShouldEqual, true)
				So(userJane.Call("IsActivated").(bool), ShouldEqual, true)
				janeProfile := userJane.Get("Profile").(RecordSet).Collection()
				janeProfile.Set("Active", true)
				So(janeProfile.Get("Active").(bool), ShouldEqual, true)
				So(janeProfile.Call("IsActivated").(bool), ShouldEqual, true)
			})
		}), ShouldBeNil)
	})
}
