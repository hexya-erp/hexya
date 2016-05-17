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
	"github.com/npiganeau/yep/yep/orm"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var env Environment

type Profile_WithID struct {
	ID    int64
	Age   int16
	Money float64
}

type User_WithID struct {
	ID       int64
	UserName string
	Email    string
	Profile  *Profile_WithID
}

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		env = NewEnvironment(dORM, 1)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohn := User_WithID{
				UserName: "John Smith",
				Email:    "jsmith@example.com",
			}
			users := env.Create(&userJohn)
			So(users.Ids(), ShouldContain, 1)
			So(userJohn.ID, ShouldEqual, 1)
		})
		Convey("Creating user Jane with related Profile using Call('Create')", func() {
			userJane := User_WithID{
				UserName: "Jane Smith",
				Email:    "jane.smith@example.com",
				Profile: &Profile_WithID{
					Age:   23,
					Money: 12345,
				},
			}
			rsProfile := NewRecordSet(env, "Profile")
			profile := rsProfile.Call("Create", userJane.Profile).(RecordSet)
			So(profile.Ids(), ShouldContain, 1)
			So(userJane.Profile.ID, ShouldEqual, 1)
			rsUsers := NewRecordSet(env, "User")
			users2 := rsUsers.Call("Create", &userJane).(RecordSet)
			So(users2.Ids(), ShouldContain, 2)
			So(userJane.Profile.ID, ShouldEqual, 1)
		})
	})
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecorSets", t, func() {
		Convey("Searching User Jane and getting struct through ReadOne", func() {
			users := env.Pool(new(User)).Filter("UserName", "Jane Smith").Search()
			So(len(users.Ids()), ShouldEqual, 1)
			var userJane User_WithID
			users.RelatedSel(1).ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane Smith")
			So(userJane.Email, ShouldEqual, "jane.smith@example.com")
		})

		Convey("Testing search all users and getting struct slice", func() {
			usersAll := env.Pool(new(User)).Search()
			var userStructs []*User_PartialWithPosts
			num := usersAll.ReadAll(&userStructs)
			So(num, ShouldEqual, 2)
			So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
			So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
		})
	})

}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		Convey("Simple update with params to user Jane", func() {
			rsJane := env.Pool(new(User)).Filter("UserName", "Jane Smith").Search()
			So(rsJane.Ids(), ShouldContain, 2)
			So(len(rsJane.Ids()), ShouldEqual, 1)
			num := rsJane.Call("Write", orm.Params{"UserName": "Jane A. Smith"})
			So(num, ShouldEqual, 1)
			var userJane User_WithID
			rsJane.ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane A. Smith")
			So(userJane.Email, ShouldEqual, "jane.smith@example.com")
		})
		Convey("Simple update with struct", func() {
			rsJane := env.Pool(new(User)).Filter("UserName", "Jane A. Smith").Search()
			var userJohn User_WithID
			rsJohn := env.Pool("User").Filter("UserName", "John Smith")
			rsJohn.ReadOne(&userJohn)
			userJohn.Email = "jsmith2@example.com"
			env.Sync(&userJohn)
			var userJane2 User_WithID
			rsJane.ReadOne(&userJane2)
			So(userJane2.UserName, ShouldEqual, "Jane A. Smith")
			So(userJane2.Email, ShouldEqual, "jane.smith@example.com")
			var userJohn2 User_WithID
			env.Pool("User").Filter("UserName", "John Smith").ReadOne(&userJohn2)
			So(userJohn2.UserName, ShouldEqual, "John Smith")
			So(userJohn2.Email, ShouldEqual, "jsmith2@example.com")
		})
	})
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		users := env.Pool("User").Filter("UserName", "John Smith")
		num := users.Unlink()
		Convey("Number of deleted record should be 1", func() {
			So(num, ShouldEqual, 1)
		})
	})
	Convey("Creating a user Will Smith instead", t, func() {
		userWill := User_WithID{
			UserName: "Will Smith",
			Email:    "will.smith@example.com",
		}
		users := env.Create(&userWill)
		Convey("Created user ids should be [3] ", func() {
			So(users.Ids(), ShouldContain, 3)
		})
	})
}
