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

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		env := models.NewEnvironment(1)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohnData := pool.Test__User{
				UserName: "John Smith",
				Email:    "jsmith@example.com",
			}
			userJohn := pool.NewTest__UserSet(env).Create(&userJohnData)
			So(userJohn.Len(), ShouldEqual, 1)
			So(userJohn.ID(), ShouldBeGreaterThan, 0)
		})
		Convey("Creating user Jane with related Profile", func() {
			profileData := pool.Test__Profile{
				Age:   23,
				Money: 12345,
			}
			profile := pool.NewTest__ProfileSet(env).Create(&profileData)
			So(profile.Len(), ShouldEqual, 1)
			userJaneData := pool.Test__User{
				UserName: "Jane Smith",
				Email:    "jane.smith@example.com",
				Profile:  profile,
			}
			userJane := pool.NewTest__UserSet(env).Create(&userJaneData)
			So(userJane.Len(), ShouldEqual, 1)
			So(userJane.Profile().ID(), ShouldEqual, profile.ID())
		})
		Convey("Creating a user Will Smith", func() {
			userWillData := pool.Test__User{
				UserName: "Will Smith",
				Email:    "will.smith@example.com",
			}
			userWill := pool.NewTest__UserSet(env).Create(&userWillData)
			So(userWill.Len(), ShouldEqual, 1)
			So(userWill.ID(), ShouldBeGreaterThan, 0)
		})
		env.Cr().Commit()
	})
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID       int64
			UserName string
			Email    string
		}
		env := models.NewEnvironment(1)
		Convey("Searching User Jane", func() {
			userJane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(userJane.Len(), ShouldEqual, 1)
			Convey("Reading Jane with getters", func() {
				So(userJane.UserName(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJane.Profile().Age(), ShouldEqual, 23)
				So(userJane.Profile().Money(), ShouldEqual, 12345)
			})
			Convey("Reading Jane with ReadFirst", func() {
				var userJaneStruct UserStruct
				userJane.ReadFirst(&userJaneStruct)
				So(userJaneStruct.UserName, ShouldEqual, "Jane Smith")
				So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
				So(userJaneStruct.ID, ShouldEqual, userJane.ID())
			})
		})

		Convey("Testing search all users", func() {
			usersAll := pool.NewTest__UserSet(env).Read()
			So(usersAll.Len(), ShouldEqual, 3)
			Convey("Reading first user with getters", func() {
				So(usersAll.UserName(), ShouldEqual, "John Smith")
				So(usersAll.Email(), ShouldEqual, "jsmith@example.com")
			})
			Convey("Reading all users with Records and Get", func() {
				recs := usersAll.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Email(), ShouldEqual, "jsmith@example.com")
				So(recs[1].Email(), ShouldEqual, "jane.smith@example.com")
				So(recs[2].Email(), ShouldEqual, "will.smith@example.com")
			})
			Convey("Reading all users with ReadAll()", func() {
				var userStructs []*UserStruct
				usersAll.ReadAll(&userStructs)
				So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
				So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
				So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
			})
		})
		env.Cr().Rollback()
	})
}

//func TestUpdateRecordSet(t *testing.T) {
//	Convey("Testing updates through RecordSets", t, func() {
//		env := models.NewEnvironment(1)
//		Convey("Update on users Jane and John with Write and Set", func() {
//			jane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
//			So(jane.Len(), ShouldEqual, 1)
//			jane.Set("UserName", "Jane A. Smith")
//			jane.Read()
//			So(jane.UserName(), ShouldEqual, "Jane A. Smith")
//			So(jane.Email(), ShouldEqual, "jane.smith@example.com")
//
//			john := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
//			So(john.Len(), ShouldEqual, 1)
//			johnValues := pool.Test__User{
//				Email: "jsmith2@example.com",
//				Nums:  13,
//			}
//			john.Write(&johnValues)
//			john.Read()
//			So(john.UserName(), ShouldEqual, "John Smith")
//			So(john.Email(), ShouldEqual, "jsmith2@example.com")
//			So(john.Nums(), ShouldEqual, 13)
//		})
//		Convey("Multiple updates at once on users", func() {
//			cond := models.NewCondition().And("UserName", "=", "Jane A. Smith").Or("UserName", "=", "John Smith")
//			users := pool.NewTest__UserSet(env).Search(cond)
//			So(users.Len(), ShouldEqual, 2)
//			userRecs := users.Records()
//			So(userRecs[0].IsStaff(), ShouldBeFalse)
//			So(userRecs[1].IsStaff(), ShouldBeFalse)
//			So(userRecs[0].IsActive(), ShouldBeFalse)
//			So(userRecs[1].IsActive(), ShouldBeFalse)
//
//			users.SetIsStaff(true)
//			users.Read()
//			So(userRecs[0].IsStaff(), ShouldBeTrue)
//			So(userRecs[1].IsStaff(), ShouldBeTrue)
//
//			uData := pool.Test__User{
//				IsStaff:  false,
//				IsActive: true,
//			}
//			users.Write(&uData)
//			users.Read()
//			So(userRecs[0].IsStaff(), ShouldBeFalse)
//			So(userRecs[1].IsStaff(), ShouldBeFalse)
//			So(userRecs[0].IsActive(), ShouldBeTrue)
//			So(userRecs[1].IsActive(), ShouldBeTrue)
//		})
//		env.Cr().Commit()
//	})
//}

//func TestDeleteRecordSet(t *testing.T) {
//	env := NewEnvironment(1)
//	Convey("Delete user John Smith", t, func() {
//		users := env.Pool("User").Filter("UserName", "=", "John Smith")
//		num := users.Call("Unlink")
//		Convey("Number of deleted record should be 1", func() {
//			So(num, ShouldEqual, 1)
//		})
//	})
//	env.cr.Rollback()
//}
