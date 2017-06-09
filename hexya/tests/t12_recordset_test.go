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
	"github.com/hexya-erp/hexya/pool"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := pool.UserData{
					Name:  "John Smith",
					Email: "jsmith@example.com",
				}
				userJohn := pool.User().Create(env, &userJohnData)
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Creating user Jane with related Profile and Posts and Tags", func() {
				profileData := pool.ProfileData{
					Age:     int16(23),
					Money:   12345,
					Street:  "165 5th Avenue",
					City:    "New York",
					Zip:     "0305",
					Country: "USA",
				}
				profile := pool.Profile().Create(env, &profileData)
				So(profile.Len(), ShouldEqual, 1)
				post1Data := pool.PostData{
					Title:   "1st Post",
					Content: "Content of first post",
				}
				post1 := pool.Post().Create(env, &post1Data)
				So(post1.Len(), ShouldEqual, 1)
				post2Data := pool.PostData{
					Title:   "2nd Post",
					Content: "Content of second post",
				}
				post2 := pool.Post().Create(env, &post2Data)
				So(post2.Len(), ShouldEqual, 1)
				posts := post1.Union(post2)
				userJaneData := pool.UserData{
					Name:    "Jane Smith",
					Email:   "jane.smith@example.com",
					Profile: profile,
					Posts:   posts,
				}
				userJane := pool.User().Create(env, &userJaneData)
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Profile().ID(), ShouldEqual, profile.ID())
				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, userJane.ID())
				So(userJane.Posts().Len(), ShouldEqual, 2)

				tag1 := pool.Tag().Create(env, &pool.TagData{
					Name: "Trending",
				})
				tag2 := pool.Tag().Create(env, &pool.TagData{
					Name: "Books",
				})
				tag3 := pool.Tag().Create(env, &pool.TagData{
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
					Name:  "Will Smith",
					Email: "will.smith@example.com",
				}
				userWill := pool.User().Create(env, &userWillData)
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
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				So(func() { pool.User().Create(env, &userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				pool.User().Methods().Create().AllowGroup(group1)
				pool.Post().Methods().Create().AllowGroup(group1, pool.User().Methods().Write())
				userTomData := pool.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				So(func() { pool.User().Create(env, &userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for posts and it works", func() {
				pool.Post().Methods().Create().AllowGroup(group1, pool.User().Methods().Create())
				userTomData := pool.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := pool.User().Create(env, &userTomData)
				So(func() { userTom.Name() }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				pool.User().Methods().Load().AllowGroup(group1)
				userTomData := pool.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := pool.User().Create(env, &userTomData)
				So(userTom.Name(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldEqual, "tsmith@example.com")
			})
			Convey("Removing Create right on Email field", func() {
				pool.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Write)
				userTomData := pool.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := pool.User().Create(env, &userTomData)
				So(userTom.Name(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldBeBlank)
				pool.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Write)
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID    int64
			Name  string
			Email string
		}
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Searching User Jane", func() {
				userJane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
				So(userJane.Len(), ShouldEqual, 1)
				Convey("Reading Jane with getters", func() {
					So(userJane.Name(), ShouldEqual, "Jane Smith")
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
					So(userJaneStruct.Name, ShouldEqual, "Jane Smith")
					So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
					So(userJaneStruct.ID, ShouldEqual, userJane.ID())
				})
			})

			Convey("Testing search all users", func() {
				usersAll := pool.User().NewSet(env).OrderBy("Name").Load()
				So(usersAll.Len(), ShouldEqual, 3)
				Convey("Reading first user with getters", func() {
					So(usersAll.Name(), ShouldEqual, "Jane Smith")
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

			Convey("Testing search on manual model", func() {
				userViews := pool.UserView().NewSet(env).FetchAll()
				So(userViews.Len(), ShouldEqual, 3)
				userViews = pool.UserView().NewSet(env).OrderBy("Name")
				So(userViews.Len(), ShouldEqual, 3)
				recs := userViews.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Name(), ShouldEqual, "Jane Smith")
				So(recs[1].Name(), ShouldEqual, "John Smith")
				So(recs[2].Name(), ShouldEqual, "Will Smith")
				So(recs[0].City(), ShouldEqual, "New York")
				So(recs[1].City(), ShouldEqual, "")
				So(recs[2].City(), ShouldEqual, "")
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list while searching", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot access records", func() {
				pool.User().Methods().Search().AllowGroup(group1)
				userJane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and checking access", func() {
				pool.User().Methods().Load().AllowGroup(group1)

				userJane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Name(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJane.Age(), ShouldEqual, 23)
				So(func() { userJane.Profile().Age() }, ShouldPanic)
			})
			Convey("Adding field access rights to user 2 and checking access", func() {
				pool.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Read)
				pool.User().Fields().Age().RevokeAccess(security.GroupEveryone, security.Read)

				userJane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Name(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldBeBlank)
				So(userJane.Age(), ShouldEqual, 0)

				pool.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Read)
				pool.User().Fields().Age().GrantAccess(security.GroupEveryone, security.Read)
			})
			Convey("Checking record rules", func() {
				users := pool.User().NewSet(env).FetchAll()
				So(users.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().Name().ILike("j").Condition,
					Perms:     security.Read,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: pool.User().Name().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&notUsedRule)

				users = pool.User().NewSet(env).FetchAll()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Name(), ShouldBeIn, []string{"Jane Smith", "John Smith"})
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
			jane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields", func() {
				users := pool.User().Search(env, pool.User().Profile().Equals(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
			Convey("Empty RecordSet on m2o relation fields", func() {
				users := pool.User().Search(env, pool.User().Profile().Equals(pool.Profile().NewSet(env)))
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2o relation fields with IN operator", func() {
				users := pool.User().Search(env, pool.User().Profile().In(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
			Convey("Empty RecordSet on m2o relation fields with IN operator", func() {
				users := pool.User().Search(env, pool.User().Profile().In(pool.Profile().NewSet(env)))
				So(users.Len(), ShouldEqual, 0)
			})
		})
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Conditions on o2m relation", func() {
				users := pool.User().Search(env, pool.User().Posts().Equals(jane.Posts().Records()[0]))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with IN operator", func() {
				users := pool.User().Search(env, pool.User().Posts().In(jane.Posts()))
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
				jane := pool.User().Search(env, pool.User().Name().Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set("Name", "Jane A. Smith")
				jane.Load()
				So(jane.Name(), ShouldEqual, "Jane A. Smith")
				So(jane.Email(), ShouldEqual, "jane.smith@example.com")

				john := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith2@example.com",
					Nums:  13,
				}
				john.Write(&johnValues)
				john.Load()
				So(john.Name(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith2@example.com")
				So(john.Nums(), ShouldEqual, 13)
			})
			Convey("Multiple updates at once on users", func() {
				cond := pool.User().Name().Equals("Jane A. Smith").Or().Name().Equals("John Smith")
				users := pool.User().Search(env, cond)
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
			Convey("Updating many2one fields", func() {
				userJane := pool.User().Search(env, pool.User().Email().Equals("jane.smith@example.com"))
				profile := userJane.Profile()
				userJane.SetProfile(pool.Profile().NewSet(env))
				So(userJane.Profile().ID(), ShouldEqual, 0)
				userJane.SetProfile(profile)
				So(userJane.Profile().ID(), ShouldEqual, profile.Ids()[0])
			})
			Convey("Updating many2many fields", func() {
				post1 := pool.Post().Search(env, pool.Post().Title().Equals("1st Post"))
				tagBooks := pool.Tag().Search(env, pool.Tag().Name().Equals("Books"))
				post1.SetTags(tagBooks)

				post1Tags := post1.Tags()
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Name(), ShouldEqual, "Books")
				post2Tags := pool.Post().Search(env, pool.Post().Title().Equals("2nd Post")).Tags()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")
			})
			Convey("Updating One2many fields", func() {
				posts := pool.Post().NewSet(env)
				post1 := posts.Search(pool.Post().Title().Equals("1st Post"))
				post2 := posts.Search(pool.Post().Title().Equals("2nd Post"))
				post3 := posts.Create(pool.PostData{
					Title:   "3rd Post",
					Content: "Content of third post",
				})
				userJane := pool.User().Search(env, pool.User().Email().Equals("jane.smith@example.com"))
				userJane.SetPosts(post1.Union(post3))
				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post3.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, 0)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on update (write only)", t, func() {
		models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot update records", func() {
				pool.User().Methods().Load().AllowGroup(group1)
				john := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				So(func() { john.Write(&johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				pool.User().Methods().Write().AllowGroup(group1)
				john := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				john.Write(&johnValues)
				john.Load()
				So(john.Name(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith3@example.com")
				So(john.Nums(), ShouldEqual, 13)
			})
			Convey("Checking that user 2 cannot update profile through UpdateCity method", func() {
				pool.User().Methods().Load().AllowGroup(group1)
				pool.User().Methods().UpdateCity().AllowGroup(group1)
				jane := pool.User().Search(env, pool.User().Name().Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.UpdateCity("London") }, ShouldPanic)
			})
			Convey("Checking that user 2 can run UpdateCity after giving permission for caller", func() {
				pool.User().Methods().Load().AllowGroup(group1)
				pool.Profile().Methods().Load().AllowGroup(group1, pool.User().Methods().UpdateCity())
				pool.Profile().Methods().Write().AllowGroup(group1, pool.User().Methods().UpdateCity())
				jane := pool.User().Search(env, pool.User().Name().Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.UpdateCity("London") }, ShouldNotPanic)
			})
			Convey("Removing Update right on Email field", func() {
				pool.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Write)
				john := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := pool.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				john.Write(&johnValues)
				john.Load()
				So(john.Name(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith2@example.com")
				So(john.Nums(), ShouldEqual, 13)
				pool.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Write)
			})
			Convey("Checking record rules", func() {
				userJane := pool.User().NewSet(env).FetchAll()
				So(userJane.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().Name().ILike("j").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "unlinkRule",
					Group:     group1,
					Condition: pool.User().Name().Equals("Nobody").Condition,
					Perms:     security.Unlink,
				}
				pool.User().AddRecordRule(&notUsedRule)

				userJane = pool.User().Search(env, pool.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Name(), ShouldEqual, "Jane A. Smith")
				userJane.SetName("Jane B. Smith")
				So(userJane.Name(), ShouldEqual, "Jane B. Smith")

				userWill := pool.User().Search(env, pool.User().Name().Equals("Will Smith"))
				So(func() { userWill.SetName("Will Jr. Smith") }, ShouldPanic)

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
			users := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
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
			Convey("Checking that user 2 cannot unlink records", func() {
				pool.User().Methods().Load().AllowGroup(group1)
				users := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				So(func() { users.Unlink() }, ShouldPanic)
			})
			Convey("Adding unlink permission to user2", func() {
				pool.User().Methods().Unlink().AllowGroup(group1)
				users := pool.User().Search(env, pool.User().Name().Equals("John Smith"))
				num := users.Unlink()
				So(num, ShouldEqual, 1)
			})
			Convey("Checking record rules", func() {

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: pool.User().Name().ILike("j").Condition,
					Perms:     security.Unlink,
				}
				pool.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: pool.User().Name().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				pool.User().AddRecordRule(&notUsedRule)

				userJane := pool.User().Search(env, pool.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Unlink(), ShouldEqual, 1)

				userWill := pool.User().Search(env, pool.User().Name().Equals("Will Smith"))
				So(userWill.Unlink(), ShouldEqual, 0)

				pool.User().RemoveRecordRule("jOnly")
				pool.User().RemoveRecordRule("writeRule")
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}
