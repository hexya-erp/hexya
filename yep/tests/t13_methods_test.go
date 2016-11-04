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

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		env := models.NewEnvironment(1)
		Convey("Getting all users and calling `PrefixedUser`", func() {
			users := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			res := users.PrefixedUser("Prefix")
			So(res[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
		})
		env.Rollback()
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		env := models.NewEnvironment(1)
		Convey("Getting one user (Jane) and checking DisplayName", func() {
			users := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			So(users.DecoratedName(), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
		})
		Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
			users := pool.NewTest__UserSet(env).OrderBy("UserName")
			So(users.Len(), ShouldEqual, 3)
			userRecs := users.Records()
			So(userRecs[0].DecoratedName(), ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			So(userRecs[1].DecoratedName(), ShouldEqual, "User: John Smith [<jsmith2@example.com>]")
			So(userRecs[2].DecoratedName(), ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
		})
		env.Rollback()
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		env := models.NewEnvironment(1)
		Convey("Checking that user Jane is 23", func() {
			userJane := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Age(), ShouldEqual, 23)
		})
		Convey("Checking that user Will has no age since no profile", func() {
			userWill := pool.NewTest__UserSet(env).Filter("Email", "=", "will.smith@example.com")
			So(userWill.Age(), ShouldEqual, 0)
		})
		Convey("It's Jane's birthday, change her age, commit and check", func() {
			jane := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			So(jane.UserName(), ShouldEqual, "Jane A. Smith")
			So(jane.Profile().Money(), ShouldEqual, 12345)
			jane.Profile().SetAge(24)

			jane.Load()
			jane.Profile().Load()
			So(jane.Age(), ShouldEqual, 24)
		})
		Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
			userWill := pool.NewTest__UserSet(env).Filter("Email", "=", "will.smith@example.com")
			userWill.Load()
			So(userWill.UserName(), ShouldEqual, "Will Smith")
			willProfileData := pool.Test__Profile{
				Age:   34,
				Money: 5100,
			}
			willProfile := pool.NewTest__ProfileSet(env).Create(&willProfileData)
			userWill.SetProfile(willProfile)

			userWill.Load()
			So(userWill.Age(), ShouldEqual, 34)
		})
		env.Commit()
	})
}

func TestRelatedNonStoredFields(t *testing.T) {
	Convey("Testing non stored related fields", t, func() {
		env := models.NewEnvironment(1)
		Convey("Checking that user Jane PMoney equals is 12345", func() {
			userJane := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			So(userJane.PMoney(), ShouldEqual, 12345)
		})
		env.Rollback()
	})
}

func TestEmbeddedModels(t *testing.T) {
	Convey("Testing embedded models", t, func() {
		env := models.NewEnvironment(1)
		Convey("Adding a last post to Jane", func() {
			postRs := pool.NewTest__PostSet(env).Create(&pool.Test__Post{
				Title:   "This is my title",
				Content: "Here we have some content",
			})
			pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com").SetLastPost(postRs)
		})
		Convey("Checking that we can access jane's post directly", func() {
			userJane := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com")
			So(userJane.Title(), ShouldEqual, "This is my title")
			So(userJane.Content(), ShouldEqual, "Here we have some content")
			So(userJane.LastPost().Title(), ShouldEqual, "This is my title")
			So(userJane.LastPost().Content(), ShouldEqual, "Here we have some content")
		})
		env.Commit()
	})
}

func TestMixedInModels(t *testing.T) {
	Convey("Testing mixed in models", t, func() {
		env := models.NewEnvironment(1)
		Convey("Checking that mixed in functions are correctly inherited", func() {
			janeProfile := pool.NewTest__UserSet(env).Filter("Email", "=", "jane.smith@example.com").Profile()
			So(janeProfile.PrintAddress(), ShouldEqual, "[<165 5th Avenue, 0305 New York>, USA]")
			So(janeProfile.SayHello(), ShouldEqual, "Hello !")
		})
		env.Rollback()
	})
}
