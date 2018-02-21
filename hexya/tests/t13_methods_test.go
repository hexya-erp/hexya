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

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Getting all users and calling `PrefixedUser`", func() {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				res := users.PrefixedUser("Prefix")
				So(res[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
			})
		})
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Getting one user (Jane) and checking DisplayName", func() {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(users.DecoratedName(), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			})
			Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
				users := h.User().NewSet(env).OrderBy("Name")
				So(users.Len(), ShouldEqual, 3)
				userRecs := users.Records()
				So(userRecs[0].DecoratedName(), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
				So(userRecs[1].DecoratedName(), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
				So(userRecs[2].DecoratedName(), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
			})
			Convey("Testing built-in DisplayName", func() {
				users := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(users.Len(), ShouldEqual, 1)
				So(users.DisplayName(), ShouldEqual, "Jane A. Smith")
			})
		})
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Checking that user Jane is 23", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Age(), ShouldEqual, 23)
			})
			Convey("Checking that user Will has no age since no profile", func() {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				So(userWill.Age(), ShouldEqual, 0)
			})
			Convey("It's Jane's birthday, change her age, commit and check", func() {
				jane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(jane.Name(), ShouldEqual, "Jane A. Smith")
				So(jane.Profile().Money(), ShouldEqual, 12345)
				jane.Profile().SetAge(24)

				jane.Load()
				jane.Profile().Load()
				So(jane.Age(), ShouldEqual, 24)
			})
			Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Name(), ShouldEqual, "Will Smith")
				willProfileData := h.ProfileData{
					Age:   36,
					Money: 5100,
				}
				willProfile := h.Profile().Create(env, &willProfileData)
				userWill.SetProfile(willProfile)

				userWill.Load()
				So(userWill.Age(), ShouldEqual, 36)
			})
			Convey("Checking inverse method by changing will's age", func() {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userWill.Load()
				So(userWill.Age(), ShouldEqual, 36)
				userWill.SetAge(34)
				So(userWill.Age(), ShouldEqual, 34)
				userWill.Load()
				So(userWill.Age(), ShouldEqual, 34)
			})
			Convey("Checking that setting a computed field with no inverse panics", func() {
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				So(func() { userWill.SetDecoratedName("FooBar") }, ShouldPanic)
			})
		})
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Checking that users PMoney is correct", func() {
				userJohn := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.PMoney(), ShouldEqual, 0)
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.PMoney(), ShouldEqual, 12345)
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				So(userWill.PMoney(), ShouldEqual, 5100)
			})
			Convey("Checking that PMoney is correct after update of Profile", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.PMoney(), ShouldEqual, 12345)
				userJane.Profile().SetMoney(54321)
				So(userJane.PMoney(), ShouldEqual, 54321)
			})
			Convey("Checking that we can update PMoney directly", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.PMoney(), ShouldEqual, 12345)
				userJane.SetPMoney(67890)
				So(userJane.Profile().Money(), ShouldEqual, 67890)
				So(userJane.PMoney(), ShouldEqual, 67890)
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				So(userWill.PMoney(), ShouldEqual, 5100)

				userJane.Union(userWill).SetPMoney(100)
				So(userJane.Profile().Money(), ShouldEqual, 100)
				So(userJane.PMoney(), ShouldEqual, 100)
				So(userWill.Profile().Money(), ShouldEqual, 100)
				So(userWill.PMoney(), ShouldEqual, 100)
			})
			Convey("Checking that we can search PMoney directly", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				pmoneyUser := h.User().Search(env, q.User().PMoney().Equals(12345))
				So(pmoneyUser.Len(), ShouldEqual, 1)
				So(pmoneyUser.Ids()[0], ShouldEqual, userJane.Ids()[0])
				pUsers := h.User().Search(env, q.User().PMoney().Equals(12345).Or().PMoney().Equals(5100))
				So(pUsers.Len(), ShouldEqual, 2)
				So(pUsers.Ids(), ShouldContain, userJane.Ids()[0])
				So(pUsers.Ids(), ShouldContain, userWill.Ids()[0])
			})
			Convey("Checking that we can order by PMoney", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userWill := h.User().Search(env, q.User().Email().Equals("will.smith@example.com"))
				userJane.SetPMoney(64)
				pUsers := h.User().NewSet(env).SearchAll().OrderBy("PMoney DESC")
				So(pUsers.Len(), ShouldEqual, 3)
				pUsersRecs := pUsers.Records()
				// pUsersRecs[0] is userJohn because its pMoney is Null.
				So(pUsersRecs[1].Equals(userWill), ShouldBeTrue)
				So(pUsersRecs[2].Equals(userJane), ShouldBeTrue)
			})
		})
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
			Convey("Checking that Jane's resume exists", func() {
				So(userJane.Resume().IsEmpty(), ShouldBeFalse)
			})
			Convey("Adding a proper resume to Jane", func() {
				userJane.Resume().SetExperience("Hexya developer for 10 years")
				userJane.Resume().SetLeisure("Music, Sports")
			})
			Convey("Checking that we can access jane's resume directly", func() {
				So(userJane.Experience(), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Leisure(), ShouldEqual, "Music, Sports")
				So(userJane.Resume().Experience(), ShouldEqual, "Hexya developer for 10 years")
				So(userJane.Resume().Leisure(), ShouldEqual, "Music, Sports")
			})
		})
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Checking that mixed in functions are correctly inherited", func() {
				janeProfile := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com")).Profile()
				So(janeProfile.PrintAddress(), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
				So(janeProfile.SayHello(), ShouldEqual, "Hello !")
			})
			Convey("Checking mixing in all models", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userJane.SetActive(true)
				So(userJane.Active(), ShouldEqual, true)
				So(userJane.IsActivated(), ShouldEqual, true)
				janeProfile := userJane.Profile()
				janeProfile.SetActive(true)
				So(janeProfile.Active(), ShouldEqual, true)
				So(janeProfile.IsActivated(), ShouldEqual, true)
			})
		})
	})
}
