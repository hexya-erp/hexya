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
		env := NewEnvironment(1)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohn := User_WithID{
				UserName: "John Smith",
				Email:    "jsmith@example.com",
			}
			users := env.Create(&userJohn)
			So(len(users.Ids()), ShouldEqual, 1)
			So(userJohn.ID, ShouldEqual, users.Ids()[0])
		})
		Convey("Creating user Jane with related Profile using rs.Create / rs.Call('Create')", func() {
			userJane := User_WithID{
				UserName: "Jane Smith",
				Email:    "jane.smith@example.com",
				Profile: &Profile_WithID{
					Age:   23,
					Money: 12345,
				},
			}
			rsProfile := env.Pool("Profile")
			profile := rsProfile.Create(userJane.Profile)
			So(len(profile.Ids()), ShouldEqual, 1)
			So(userJane.Profile.ID, ShouldEqual, profile.Ids()[0])
			rsUsers := env.Pool("User")
			users2 := rsUsers.Call("Create", &userJane).(*RecordSet)
			So(len(users2.Ids()), ShouldEqual, 1)
			So(userJane.ID, ShouldEqual, users2.Ids()[0])
		})
		Convey("Creating a user Will Smith", func() {
			userWill := User_WithID{
				UserName: "Will Smith",
				Email:    "will.smith@example.com",
			}
			users := env.Create(&userWill)
			Convey("Created user ids should match struct's ID ", func() {
				So(len(users.Ids()), ShouldEqual, 1)
				So(users.Ids()[0], ShouldEqual, userWill.ID)
			})
		})
		env.cr.Commit()
	})
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		env := NewEnvironment(1)
		Convey("Searching User Jane and getting struct through ReadOne", func() {
			users := env.Pool("User").Filter("UserName", "=", "Jane Smith").Search()
			So(len(users.Ids()), ShouldEqual, 1)
			var userJane User_WithID
			users.RelatedDepth(1).ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane Smith")
			So(userJane.Email, ShouldEqual, "jane.smith@example.com")
			So(userJane.Profile.Age, ShouldEqual, 23)
			So(userJane.Profile.Money, ShouldEqual, 12345)
		})

		Convey("Testing search all users and getting struct slice", func() {
			usersAll := env.Pool("User").Search()
			So(len(usersAll.Ids()), ShouldEqual, 3)
			var userStructs []*User_PartialWithPosts
			num := usersAll.ReadAll(&userStructs)
			So(num, ShouldEqual, 3)
			So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
			So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
			So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
		})
		env.cr.Rollback()
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		env := NewEnvironment(1)
		Convey("Simple update with params to user Jane", func() {
			rsJane := env.Pool("User").Filter("UserName", "=", "Jane Smith").Search()
			So(len(rsJane.Ids()), ShouldEqual, 1)
			res := rsJane.Call("Write", FieldMap{"UserName": "Jane A. Smith"})
			So(res, ShouldEqual, true)
			var userJane User_WithID
			rsJane.ReadOne(&userJane)
			So(userJane.UserName, ShouldEqual, "Jane A. Smith")
			So(userJane.Email, ShouldEqual, "jane.smith@example.com")
		})
		Convey("Simple update with struct", func() {
			rsJane := env.Pool("User").Filter("UserName", "=", "Jane A. Smith").Search()
			var userJohn User_WithID
			rsJohn := env.Pool("User").Filter("UserName", "=", "John Smith")
			rsJohn.ReadOne(&userJohn)
			userJohn.Email = "jsmith2@example.com"
			env.Sync(&userJohn)
			var userJane2 User_WithID
			rsJane.ReadOne(&userJane2)
			So(userJane2.UserName, ShouldEqual, "Jane A. Smith")
			So(userJane2.Email, ShouldEqual, "jane.smith@example.com")
			var userJohn2 User_WithID
			env.Pool("User").Filter("UserName", "=", "John Smith").ReadOne(&userJohn2)
			So(userJohn2.UserName, ShouldEqual, "John Smith")
			So(userJohn2.Email, ShouldEqual, "jsmith2@example.com")
		})
		env.cr.Commit()
	})
}

func TestDeleteRecordSet(t *testing.T) {
	env := NewEnvironment(1)
	Convey("Delete user John Smith", t, func() {
		users := env.Pool("User").Filter("UserName", "=", "John Smith")
		num := users.Unlink()
		Convey("Number of deleted record should be 1", func() {
			So(num, ShouldEqual, 1)
		})
	})
	env.cr.Rollback()
}
