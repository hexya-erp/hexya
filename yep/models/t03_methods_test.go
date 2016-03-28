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

type User_WithDisplayName struct {
	ID          int64
	UserName    string
	Email       string
	Profile     *Profile_WithID
	DisplayName string
}

func TestMethods(t *testing.T) {
	Convey("Testing simple methods", t, func() {
		env = NewEnvironment(dORM, 1)
		Convey("Getting all users and calling `PrefixedUser`", func() {
			users := env.Pool("User")
			res := users.Call("PrefixedUser", "Prefix")
			So(res.([]string)[0], ShouldEqual, "Prefix: Jane A. Smith [<jane.smith@example.com>]")
		})
	})
}

func TestComputedNonStoredFields(t *testing.T) {
	Convey("Testing non stored computed fields", t, func() {
		env = NewEnvironment(dORM, 1)
		Convey("Getting one user (Jane) and checking DisplayName", func() {
			var userJane User_WithDisplayName
			users := env.Pool("User")
			users.Filter("Email", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.DisplayName, ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
		})
		Convey("Getting all users (Jane & Will) and checking DisplayName", func() {
			var users []*User_WithDisplayName
			env.Pool("User").OrderBy("UserName").ReadAll(&users)
			So(users[0].DisplayName, ShouldEqual, "User: Jane A. Smith [<jane.smith@example.com>]")
			So(users[1].DisplayName, ShouldEqual, "User: Will Smith [<will.smith@example.com>]")
		})
	})
}

func TestComputedStoredFields(t *testing.T) {
	Convey("Testing stored computed fields", t, func() {
		env = NewEnvironment(dORM, 1)
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
		}
		Convey("Checking that user Jane is 23", func() {
			var userJane User_Simple
			env.Pool("User").Filter("Email", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.Age, ShouldEqual, 23)
		})
		Convey("Checking that user Will has no age since no profile", func() {
			var userWill User_Simple
			env.Pool("User").Filter("Email", "will.smith@example.com").ReadOne(&userWill)
			So(userWill.Age, ShouldEqual, 0)
		})
		Convey("It's Jane's birthday, change her age, commit and check", func() {
			var userJane User_Simple
			janeRs := env.Pool("User").Filter("Email", "jane.smith@example.com")
			janeRs.ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane A. Smith")
			userJane.Profile.Age = 24
			env.Sync(userJane.Profile)
			env.Pool("User").Filter("Email", "jane.smith@example.com").ReadOne(&userJane)
			So(userJane.Age, ShouldEqual, 24)
		})
		Convey("Adding a Profile to Will, writing to DB and checking Will's age", func() {
			var userWill User_Simple
			willRs := env.Pool("User").Filter("Email", "will.smith@example.com")
			willRs.ReadOne(&userWill)
			So(userWill.UserName, ShouldEqual, "Will Smith")
			userWill.Profile = &Profile_Simple{
				Age:   34,
				Money: 5100,
			}
			env.Create(userWill.Profile)
			env.Sync(&userWill)
			env.Pool("User").Filter("Email", "will.smith@example.com").ReadOne(&userWill)
			So(userWill.Age, ShouldEqual, 34)
		})
	})
}
