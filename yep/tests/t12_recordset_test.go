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
	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		env := models.NewEnvironment(security.SuperUserID)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohnData := pool.Test__User{
				UserName: "John Smith",
				Email:    "jsmith@example.com",
			}
			userJohn := pool.NewTest__UserSet(env).Create(&userJohnData)
			So(userJohn.Len(), ShouldEqual, 1)
			So(userJohn.ID(), ShouldBeGreaterThan, 0)
		})
		Convey("Creating user Jane with related Profile and Posts and Tags", func() {
			profileData := pool.Test__Profile{
				Age:     23,
				Money:   12345,
				Street:  "165 5th Avenue",
				City:    "New York",
				Zip:     "0305",
				Country: "USA",
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
			post1Data := pool.Test__Post{
				User:    userJane,
				Title:   "1st Post",
				Content: "Content of first post",
			}
			post1 := pool.NewTest__PostSet(env).Create(&post1Data)
			So(post1.Len(), ShouldEqual, 1)
			post2Data := pool.Test__Post{
				User:    userJane,
				Title:   "2nd Post",
				Content: "Content of second post",
			}
			post2 := pool.NewTest__PostSet(env).Create(&post2Data)
			So(post2.Len(), ShouldEqual, 1)
			So(userJane.Posts().Len(), ShouldEqual, 2)

			tag1 := pool.NewTest__TagSet(env).Create(&pool.Test__Tag{
				Name: "Trending",
			})
			tag2 := pool.NewTest__TagSet(env).Create(&pool.Test__Tag{
				Name: "Books",
			})
			tag3 := pool.NewTest__TagSet(env).Create(&pool.Test__Tag{
				Name: "Jane's",
			})
			post1.SetTags(tag1.Union(tag3))
			post2.SetTags(tag2.Union(tag3))
			post1Tags := post1.Tags()
			So(post1Tags.Len(), ShouldEqual, 2)
			So(post1Tags.Records()[0].Name(), ShouldBeIn, "Trending", "Jane's")
			So(post1Tags.Records()[1].Name(), ShouldBeIn, "Trending", "Jane's")
			post2Tags := post2.Tags()
			So(post2Tags.Len(), ShouldEqual, 2)
			So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
			So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")
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
		env.Commit()
	})
	Convey("Testing access control list on creation (create only)", t, func() {
		env := models.NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot create records", func() {
			userTomData := pool.Test__User{
				UserName: "Tom Smith",
				Email:    "tsmith@example.com",
			}
			So(func() { pool.NewTest__UserSet(env).Create(&userTomData) }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and check creation", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Create)
			userTomData := pool.Test__User{
				UserName: "Tom Smith",
				Email:    "tsmith@example.com",
			}
			userTom := pool.NewTest__UserSet(env).Create(&userTomData)
			So(func() { userTom.UserName() }, ShouldPanic)
		})
		Convey("Checking creation again with read rights too", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Create|security.Read)
			userTomData := pool.Test__User{
				UserName: "Tom Smith",
				Email:    "tsmith@example.com",
			}
			userTom := pool.NewTest__UserSet(env).Create(&userTomData)
			So(userTom.UserName(), ShouldEqual, "Tom Smith")
			So(userTom.Email(), ShouldEqual, "tsmith@example.com")
		})
		Convey("Removing Create right on Email field", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Create|security.Read)
			models.DenyFieldAccess(pool.ModelTest__User, pool.Test__User_Email, group1, security.Create)
			userTomData := pool.Test__User{
				UserName: "Tom Smith",
				Email:    "tsmith@example.com",
			}
			userTom := pool.NewTest__UserSet(env).Create(&userTomData)
			So(userTom.UserName(), ShouldEqual, "Tom Smith")
			So(userTom.Email(), ShouldBeBlank)
			models.DenyModelAccess(pool.ModelTest__User, group1, security.Create|security.Read)
		})
		env.Rollback()
	})
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID       int64
			UserName string
			Email    string
		}
		env := models.NewEnvironment(security.SuperUserID)
		Convey("Searching User Jane", func() {
			userJane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(userJane.Len(), ShouldEqual, 1)
			Convey("Reading Jane with getters", func() {
				So(userJane.UserName(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJane.Profile().Age(), ShouldEqual, 23)
				So(userJane.Profile().Money(), ShouldEqual, 12345)
				So(userJane.Profile().Country(), ShouldEqual, "USA")
				So(userJane.Profile().Zip(), ShouldEqual, "0305")
				recs := userJane.Posts().Records()
				So(recs[0].Title(), ShouldEqual, "1st Post")
				So(recs[1].Title(), ShouldEqual, "2nd Post")
			})
			Convey("Reading Jane with ReadFirst", func() {
				userJaneStruct := userJane.First()
				So(userJaneStruct.UserName, ShouldEqual, "Jane Smith")
				So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
				So(userJaneStruct.ID, ShouldEqual, userJane.ID())
			})
		})

		Convey("Testing search all users", func() {
			usersAll := pool.NewTest__UserSet(env).Load()
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
				userStructs := usersAll.All()
				So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
				So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
				So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
			})
		})
		env.Rollback()
	})
	Convey("Testing access control list while searching", t, func() {
		env := models.NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot access records", func() {
			userJane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and checking access", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read)

			userJane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldNotPanic)
			So(userJane.UserName(), ShouldEqual, "Jane Smith")
			So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
			So(userJane.Age(), ShouldEqual, 23)
			So(func() { userJane.Profile().Age() }, ShouldPanic)
		})
		Convey("Adding field access rights to user 2 and checking access", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read)
			models.DenyFieldAccess(pool.ModelTest__User, pool.Test__User_Email, group1, security.Read)
			models.DenyFieldAccess(pool.ModelTest__User, pool.Test__User_Age, group1, security.Read)

			userJane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldNotPanic)
			So(userJane.UserName(), ShouldEqual, "Jane Smith")
			So(userJane.Email(), ShouldBeBlank)
			So(userJane.Age(), ShouldEqual, 0)
		})
		env.Rollback()
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		env := models.NewEnvironment(security.SuperUserID)
		Convey("Update on users Jane and John with Write and Set", func() {
			jane := pool.NewTest__UserSet(env).Filter("UserName", "=", "Jane Smith")
			So(jane.Len(), ShouldEqual, 1)
			jane.Set("UserName", "Jane A. Smith")
			jane.Load()
			So(jane.UserName(), ShouldEqual, "Jane A. Smith")
			So(jane.Email(), ShouldEqual, "jane.smith@example.com")

			john := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := pool.Test__User{
				Email: "jsmith2@example.com",
				Nums:  13,
			}
			john.Write(&johnValues)
			john.Load()
			So(john.UserName(), ShouldEqual, "John Smith")
			So(john.Email(), ShouldEqual, "jsmith2@example.com")
			So(john.Nums(), ShouldEqual, 13)
		})
		Convey("Multiple updates at once on users", func() {
			cond := models.NewCondition().And("UserName", "=", "Jane A. Smith").Or("UserName", "=", "John Smith")
			users := pool.NewTest__UserSet(env).Search(cond)
			So(users.Len(), ShouldEqual, 2)
			userRecs := users.Records()
			So(userRecs[0].IsStaff(), ShouldBeFalse)
			So(userRecs[1].IsStaff(), ShouldBeFalse)
			So(userRecs[0].IsActive(), ShouldBeFalse)
			So(userRecs[1].IsActive(), ShouldBeFalse)

			users.SetIsStaff(true)
			users.Load()
			So(userRecs[0].IsStaff(), ShouldBeTrue)
			So(userRecs[1].IsStaff(), ShouldBeTrue)

			uData := pool.Test__User{
				IsStaff:  false,
				IsActive: true,
			}
			users.Write(&uData, pool.Test__User_IsActive, pool.Test__User_IsStaff)
			users.Load()
			So(userRecs[0].IsStaff(), ShouldBeFalse)
			So(userRecs[1].IsStaff(), ShouldBeFalse)
			So(userRecs[0].IsActive(), ShouldBeTrue)
			So(userRecs[1].IsActive(), ShouldBeTrue)
		})
		Convey("Updating many2many fields", func() {
			post1 := pool.NewTest__PostSet(env).Filter("title", "=", "1st Post")
			tagBooks := pool.NewTest__TagSet(env).Filter("name", "=", "Books")
			post1.SetTags(tagBooks)

			post1Tags := post1.Tags()
			So(post1Tags.Len(), ShouldEqual, 1)
			So(post1Tags.Name(), ShouldEqual, "Books")
			post2Tags := pool.NewTest__PostSet(env).Filter("title", "=", "2nd Post").Tags()
			So(post2Tags.Len(), ShouldEqual, 2)
			So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
			So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")
		})
		env.Commit()
	})
	Convey("Testing access control list on update (write only)", t, func() {
		env := models.NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot update records", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read)
			john := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := pool.Test__User{
				Email: "jsmith3@example.com",
				Nums:  13,
			}
			So(func() { john.Write(&johnValues) }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and check update", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read|security.Write)
			john := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := pool.Test__User{
				Email: "jsmith3@example.com",
				Nums:  13,
			}
			john.Write(&johnValues)
			john.Load()
			So(john.UserName(), ShouldEqual, "John Smith")
			So(john.Email(), ShouldEqual, "jsmith3@example.com")
			So(john.Nums(), ShouldEqual, 13)
		})
		Convey("Removing Update right on Email field", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Write|security.Read)
			models.DenyFieldAccess(pool.ModelTest__User, pool.Test__User_Email, group1, security.Write)
			john := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := pool.Test__User{
				Email: "jsmith3@example.com",
				Nums:  13,
			}
			john.Write(&johnValues)
			john.Load()
			So(john.UserName(), ShouldEqual, "John Smith")
			So(john.Email(), ShouldEqual, "jsmith2@example.com")
			So(john.Nums(), ShouldEqual, 13)
		})
		env.Rollback()
	})
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		env := models.NewEnvironment(security.SuperUserID)
		users := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
		num := users.Unlink()
		Convey("Number of deleted record should be 1", func() {
			So(num, ShouldEqual, 1)
		})
		env.Rollback()
	})
	Convey("Checking unlink access permissions", t, func() {
		env := models.NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot delete records", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read)
			users := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			So(func() { users.Unlink() }, ShouldPanic)
		})
		Convey("Adding unlink permission to user2", func() {
			models.AllowModelAccess(pool.ModelTest__User, group1, security.Read|security.Unlink)
			users := pool.NewTest__UserSet(env).Filter("UserName", "=", "John Smith")
			num := users.Unlink()
			So(num, ShouldEqual, 1)
		})
		env.Rollback()
	})
}
