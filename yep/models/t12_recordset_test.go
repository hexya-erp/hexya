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

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		env := NewEnvironment(1)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohnData := FieldMap{
				"UserName": "John Smith",
				"Email":    "jsmith@example.com",
			}
			users := env.Pool("User").Call("Create", userJohnData).(RecordCollection)
			So(users.Len(), ShouldEqual, 1)
			So(users.Get("ID"), ShouldBeGreaterThan, 0)
		})
		Convey("Creating user Jane with related Profile", func() {
			userJaneProfileData := FieldMap{
				"Age":   23,
				"Money": 12345,
			}
			profile := env.Pool("Profile").Call("Create", userJaneProfileData).(RecordCollection)
			So(profile.Len(), ShouldEqual, 1)
			userJaneData := FieldMap{
				"UserName": "Jane Smith",
				"Email":    "jane.smith@example.com",
				"Profile":  profile,
			}
			userJane := env.Pool("User").Call("Create", userJaneData).(RecordCollection)
			So(userJane.Len(), ShouldEqual, 1)
			So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, profile.Get("ID"))
		})
		Convey("Creating a user Will Smith", func() {
			userWillData := FieldMap{
				"UserName": "Will Smith",
				"Email":    "will.smith@example.com",
			}
			userWill := env.Pool("User").Call("Create", userWillData).(RecordCollection)
			So(userWill.Len(), ShouldEqual, 1)
			So(userWill.Get("ID"), ShouldBeGreaterThan, 0)
		})
		env.cr.Commit()
	})
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID       int64
			UserName string
			Email    string
		}
		env := NewEnvironment(1)
		Convey("Searching User Jane", func() {
			userJane := env.Pool("User").Filter("UserName", "=", "Jane Smith").Load()
			So(userJane.Len(), ShouldEqual, 1)
			Convey("Reading Jane with Get", func() {
				So(userJane.Get("UserName").(string), ShouldEqual, "Jane Smith")
				So(userJane.Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(userJane.Get("Profile").(RecordCollection).Get("Age"), ShouldEqual, 23)
				So(userJane.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 12345)
			})
			Convey("Reading Jane with ReadFirst", func() {
				var userJaneStruct UserStruct
				userJane.ReadFirst(&userJaneStruct)
				So(userJaneStruct.UserName, ShouldEqual, "Jane Smith")
				So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
				So(userJaneStruct.ID, ShouldEqual, userJane.Get("ID").(int64))
			})
		})

		Convey("Testing search all users", func() {
			usersAll := env.Pool("User").Load()
			So(usersAll.Len(), ShouldEqual, 3)
			Convey("Reading first user with Get", func() {
				So(usersAll.Get("UserName"), ShouldEqual, "John Smith")
				So(usersAll.Get("Email"), ShouldEqual, "jsmith@example.com")
			})
			Convey("Reading all users with Records and Get", func() {
				recs := usersAll.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Get("Email"), ShouldEqual, "jsmith@example.com")
				So(recs[1].Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(recs[2].Get("Email"), ShouldEqual, "will.smith@example.com")
			})
			Convey("Reading all users with ReadAll()", func() {
				var userStructs []*UserStruct
				usersAll.ReadAll(&userStructs)
				So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
				So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
				So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
			})
		})
		env.cr.Rollback()
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		env := NewEnvironment(1)
		Convey("Update on users Jane and John with Write and Set", func() {
			jane := env.Pool("User").Filter("UserName", "=", "Jane Smith").Load()
			So(jane.Len(), ShouldEqual, 1)
			jane.Set("UserName", "Jane A. Smith")
			jane.Read()
			So(jane.Get("UserName"), ShouldEqual, "Jane A. Smith")
			So(jane.Get("Email"), ShouldEqual, "jane.smith@example.com")

			john := env.Pool("User").Filter("UserName", "=", "John Smith").Load()
			So(john.Len(), ShouldEqual, 1)
			johnValues := FieldMap{
				"Email": "jsmith2@example.com",
				"Nums":  13,
			}
			john.Call("Write", johnValues)
			john.Read()
			So(john.Get("UserName"), ShouldEqual, "John Smith")
			So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
			So(john.Get("Nums"), ShouldEqual, 13)
		})
		Convey("Multiple updates at once on users", func() {
			cond := NewCondition().And("UserName", "=", "Jane A. Smith").Or("UserName", "=", "John Smith")
			users := env.Pool("User").Search(cond).Read()
			So(users.Len(), ShouldEqual, 2)
			userRecs := users.Records()
			So(userRecs[0].Get("IsStaff").(bool), ShouldBeFalse)
			So(userRecs[1].Get("IsStaff").(bool), ShouldBeFalse)
			So(userRecs[0].Get("IsActive").(bool), ShouldBeFalse)
			So(userRecs[1].Get("IsActive").(bool), ShouldBeFalse)

			users.Set("IsStaff", true)
			users.Read()
			So(userRecs[0].Get("IsStaff").(bool), ShouldBeTrue)
			So(userRecs[1].Get("IsStaff").(bool), ShouldBeTrue)

			fMap := FieldMap{
				"IsStaff":  false,
				"IsActive": true,
			}
			users.Call("Write", fMap)
			users.Read()
			So(userRecs[0].Get("IsStaff").(bool), ShouldBeFalse)
			So(userRecs[1].Get("IsStaff").(bool), ShouldBeFalse)
			So(userRecs[0].Get("IsActive").(bool), ShouldBeTrue)
			So(userRecs[1].Get("IsActive").(bool), ShouldBeTrue)
		})
		env.cr.Commit()
	})
}

func TestDeleteRecordSet(t *testing.T) {
	env := NewEnvironment(1)
	Convey("Delete user John Smith", t, func() {
		users := env.Pool("User").Filter("UserName", "=", "John Smith")
		num := users.Call("Unlink")
		Convey("Number of deleted record should be 1", func() {
			So(num, ShouldEqual, 1)
		})
	})
	env.cr.Rollback()
}
