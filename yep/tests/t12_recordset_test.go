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
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := pool.UserData{
					UserName: "John Smith",
					Email:    "jsmith@example.com",
				}
				userJohn := pool.User().NewSet(env).Create(&userJohnData)
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Creating user Jane with related Profile and Posts and Tags", func() {
				profileData := pool.ProfileData{
					Age:     23,
					Money:   12345,
					Street:  "165 5th Avenue",
					City:    "New York",
					Zip:     "0305",
					Country: "USA",
				}
				profile := pool.Profile().NewSet(env).Create(&profileData)
				So(profile.Len(), ShouldEqual, 1)
				userJaneData := pool.UserData{
					UserName: "Jane Smith",
					Email:    "jane.smith@example.com",
					Profile:  profile,
				}
				userJane := pool.User().NewSet(env).Create(&userJaneData)
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Profile().ID(), ShouldEqual, profile.ID())
				post1Data := pool.PostData{
					User:    userJane,
					Title:   "1st Post",
					Content: "Content of first post",
				}
				post1 := pool.Post().NewSet(env).Create(&post1Data)
				So(post1.Len(), ShouldEqual, 1)
				post2Data := pool.PostData{
					User:    userJane,
					Title:   "2nd Post",
					Content: "Content of second post",
				}
				post2 := pool.Post().NewSet(env).Create(&post2Data)
				So(post2.Len(), ShouldEqual, 1)
				So(userJane.Posts().Len(), ShouldEqual, 2)

				tag1 := pool.Tag().NewSet(env).Create(&pool.TagData{
					Name: "Trending",
				})
				tag2 := pool.Tag().NewSet(env).Create(&pool.TagData{
					Name: "Books",
				})
				tag3 := pool.Tag().NewSet(env).Create(&pool.TagData{
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
				userWillData := pool.UserData{
					UserName: "Will Smith",
					Email:    "will.smith@example.com",
				}
				userWill := pool.User().NewSet(env).Create(&userWillData)
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.ID(), ShouldBeGreaterThan, 0)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on creation (create only)", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot create records", func() {
				userTomData := pool.UserData{
					UserName: "Tom Smith",
					Email:    "tsmith@example.com",
				}
				So(func() { pool.User().NewSet(env).Create(&userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check creation", func() {
				pool.User().AllowModelAccess(group1, security.Create)
				userTomData := pool.UserData{
					UserName: "Tom Smith",
					Email:    "tsmith@example.com",
				}
				userTom := pool.User().NewSet(env).Create(&userTomData)
				So(func() { userTom.UserName() }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				pool.User().AllowModelAccess(group1, security.Create|security.Read)
				userTomData := pool.UserData{
					UserName: "Tom Smith",
					Email:    "tsmith@example.com",
				}
				userTom := pool.User().NewSet(env).Create(&userTomData)
				So(userTom.UserName(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldEqual, "tsmith@example.com")
			})
			Convey("Removing Create right on Email field", func() {
				pool.User().AllowModelAccess(group1, security.Create|security.Read)
				pool.User().DenyFieldAccess(pool.User().Email(), group1, security.Create)
				userTomData := pool.UserData{
					UserName: "Tom Smith",
					Email:    "tsmith@example.com",
				}
				userTom := pool.User().NewSet(env).Create(&userTomData)
				So(userTom.UserName(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldBeBlank)
				pool.User().DenyModelAccess(group1, security.Create|security.Read)
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID       int64
			UserName string
			Email    string
		}
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Searching User Jane", func() {
				userJane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
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
				usersAll := pool.User().NewSet(env).OrderBy("UserName").Load()
				So(usersAll.Len(), ShouldEqual, 3)
				Convey("Reading first user with getters", func() {
					So(usersAll.UserName(), ShouldEqual, "Jane Smith")
					So(usersAll.Email(), ShouldEqual, "jane.smith@example.com")
				})
				Convey("Reading all users with Records and Get", func() {
					recs := usersAll.Records()
					So(len(recs), ShouldEqual, 3)
					So(recs[0].Email(), ShouldEqual, "jane.smith@example.com")
					So(recs[1].Email(), ShouldEqual, "jsmith@example.com")
					So(recs[2].Email(), ShouldEqual, "will.smith@example.com")
				})
				Convey("Reading all users with ReadAll()", func() {
					userStructs := usersAll.All()
					So(userStructs[0].Email, ShouldEqual, "jane.smith@example.com")
					So(userStructs[1].Email, ShouldEqual, "jsmith@example.com")
					So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
				})
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list while searching", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot access records", func() {
				userJane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and checking access", func() {
				pool.User().AllowModelAccess(group1, security.Read)

				userJane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.UserName(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJane.Age(), ShouldEqual, 23)
				So(func() { userJane.Profile().Age() }, ShouldPanic)
			})
			Convey("Adding field access rights to user 2 and checking access", func() {
				pool.User().AllowModelAccess(group1, security.Read)
				pool.User().DenyFieldAccess(pool.User().Email(), group1, security.Read)
				pool.User().DenyFieldAccess(pool.User().Age(), group1, security.Read)

				userJane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.UserName(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldBeBlank)
				So(userJane.Age(), ShouldEqual, 0)
			})
			Convey("Checking record rules", func() {
				pool.User().AllowModelAccess(group1, security.Read)
				users := pool.User().NewSet(env).Load()
				So(users.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().UserName().ILike("j").Condition,
					Perms:     security.Read,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: pool.User().UserName().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&notUsedRule)

				users = pool.User().NewSet(env).Load()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].UserName(), ShouldBeIn, []string{"Jane Smith", "John Smith"})
				pool.User().DenyModelAccess(group1, security.Read)
				pool.User().RemoveRecordRule("jOnly")
				pool.User().RemoveRecordRule("writeRule")
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	Convey("Testing advanced queries on M2O relations", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields", func() {
				users := pool.User().NewSet(env).Search(pool.User().Profile().Equals(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
			Convey("Condition on m2o relation fields with IN operator", func() {
				users := pool.User().NewSet(env).Search(pool.User().Profile().In(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
		})
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Conditions on o2m relation", func() {
				users := pool.User().NewSet(env).Search(pool.User().Posts().Equals(jane.Posts().Records()[0]))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with IN operator", func() {
				users := pool.User().NewSet(env).Search(pool.User().Posts().In(jane.Posts()))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
		})
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Update on users Jane and John with Write and Set", func() {
				jane := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set("UserName", "Jane A. Smith")
				jane.Load()
				So(jane.UserName(), ShouldEqual, "Jane A. Smith")
				So(jane.Email(), ShouldEqual, "jane.smith@example.com")

				john := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
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
				cond := pool.User().UserName().Equals("Jane A. Smith").Or().UserName().Equals("John Smith")
				users := pool.User().NewSet(env).Search(cond)
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

				uData := pool.UserData{
					IsStaff:  false,
					IsActive: true,
				}
				users.Write(&uData, pool.User().IsActive(), pool.User().IsStaff())
				users.Load()
				So(userRecs[0].IsStaff(), ShouldBeFalse)
				So(userRecs[1].IsStaff(), ShouldBeFalse)
				So(userRecs[0].IsActive(), ShouldBeTrue)
				So(userRecs[1].IsActive(), ShouldBeTrue)
			})
			Convey("Updating many2many fields", func() {
				post1 := pool.Post().NewSet(env).Search(pool.Post().Title().Equals("1st Post"))
				tagBooks := pool.Tag().NewSet(env).Search(pool.Tag().Name().Equals("Books"))
				post1.SetTags(tagBooks)

				post1Tags := post1.Tags()
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Name(), ShouldEqual, "Books")
				post2Tags := pool.Post().NewSet(env).Search(pool.Post().Title().Equals("2nd Post")).Tags()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on update (write only)", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot update records", func() {
				pool.User().AllowModelAccess(group1, security.Read)
				john := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				So(func() { john.Write(&johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				pool.User().AllowModelAccess(group1, security.Read|security.Write)
				john := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
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
				pool.User().AllowModelAccess(group1, security.Write|security.Read)
				pool.User().DenyFieldAccess(pool.User().Email(), group1, security.Write)
				john := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				john.Write(&johnValues)
				john.Load()
				So(john.UserName(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith2@example.com")
				So(john.Nums(), ShouldEqual, 13)
			})
			Convey("Checking record rules", func() {
				pool.User().AllowModelAccess(group1, security.Read|security.Write)
				userJane := pool.User().NewSet(env).Load()
				So(userJane.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().UserName().ILike("j").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "unlinkRule",
					Group:     group1,
					Condition: pool.User().UserName().Equals("Nobody").Condition,
					Perms:     security.Unlink,
				}
				pool.User().AddRecordRule(&notUsedRule)

				userJane = pool.User().NewSet(env).Search(pool.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.UserName(), ShouldEqual, "Jane A. Smith")
				userJane.SetUserName("Jane B. Smith")
				So(userJane.UserName(), ShouldEqual, "Jane B. Smith")

				userWill := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Will Smith"))
				So(func() { userWill.SetUserName("Will Jr. Smith") }, ShouldPanic)

				pool.User().DenyModelAccess(group1, security.Read|security.Write)
				pool.User().RemoveRecordRule("jOnly")
				pool.User().RemoveRecordRule("unlinkRule")
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			users := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
			num := users.Unlink()
			Convey("Number of deleted record should be 1", func() {
				So(num, ShouldEqual, 1)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Checking unlink access permissions", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot delete records", func() {
				pool.User().AllowModelAccess(group1, security.Read)
				users := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				So(func() { users.Unlink() }, ShouldPanic)
			})
			Convey("Adding unlink permission to user2", func() {
				pool.User().AllowModelAccess(group1, security.Read|security.Unlink)
				users := pool.User().NewSet(env).Search(pool.User().UserName().Equals("John Smith"))
				num := users.Unlink()
				So(num, ShouldEqual, 1)
			})
			Convey("Checking record rules", func() {
				pool.User().AllowModelAccess(group1, security.Read|security.Unlink)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().UserName().ILike("j").Condition,
					Perms:     security.Unlink,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: pool.User().UserName().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&notUsedRule)

				userJane := pool.User().NewSet(env).Search(pool.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Unlink(), ShouldEqual, 1)

				userWill := pool.User().NewSet(env).Search(pool.User().UserName().Equals("Will Smith"))
				So(userWill.Unlink(), ShouldEqual, 0)

				pool.User().DenyModelAccess(group1, security.Read|security.Unlink)
				pool.User().RemoveRecordRule("jOnly")
				pool.User().RemoveRecordRule("writeRule")
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}
