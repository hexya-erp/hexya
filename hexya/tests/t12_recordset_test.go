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

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := h.UserData{
					Name:    "John Smith",
					Email:   "jsmith@example.com",
					IsStaff: true,
				}
				userJohn := h.User().Create(env, &userJohnData)
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Creating user Jane with related Profile and Posts and Comments and Tags", func() {
				profileData := h.ProfileData{
					Age:     int16(23),
					Money:   12345,
					Street:  "165 5th Avenue",
					City:    "New York",
					Zip:     "0305",
					Country: "USA",
				}
				profile := h.Profile().Create(env, &profileData)
				So(profile.Len(), ShouldEqual, 1)
				So(profile.UserName(), ShouldBeBlank)

				post1Data := h.PostData{
					Title:   "1st Post",
					Content: "Content of first post",
				}
				post1 := h.Post().Create(env, &post1Data)
				So(post1.Len(), ShouldEqual, 1)
				post2Data := h.PostData{
					Title:   "2nd Post",
					Content: "Content of second post",
				}
				post2 := h.Post().Create(env, &post2Data)
				So(post2.Len(), ShouldEqual, 1)
				posts := post1.Union(post2)
				profile.SetBestPost(post1)
				userJaneData := h.UserData{
					Name:    "Jane Smith",
					Email:   "jane.smith@example.com",
					Profile: profile,
					Posts:   posts,
					Nums:    2,
				}
				userJane := h.User().Create(env, &userJaneData)
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Profile().ID(), ShouldEqual, profile.ID())
				So(profile.UserName(), ShouldEqual, "Jane Smith")

				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, userJane.ID())
				So(userJane.Posts().Len(), ShouldEqual, 2)

				tag1 := h.Tag().Create(env, &h.TagData{
					Name: "Trending",
				})
				tag2 := h.Tag().Create(env, &h.TagData{
					Name: "Books",
				})
				tag3 := h.Tag().Create(env, &h.TagData{
					Name: "Jane's",
				})
				So(post1.LastTagName(), ShouldBeBlank)
				post1.SetTags(tag1.Union(tag3))
				post2.SetTags(tag2.Union(tag3))
				So(post1.LastTagName(), ShouldEqual, "Jane's")
				post1Tags := post1.Tags()
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Name(), ShouldBeIn, "Trending", "Jane's")
				So(post1Tags.Records()[1].Name(), ShouldBeIn, "Trending", "Jane's")
				post2Tags := post2.Tags()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")

				So(post1.LastCommentText(), ShouldBeBlank)
				h.Comment().Create(env, &h.CommentData{
					Post: post1,
					Text: "First Comment",
				})
				h.Comment().Create(env, &h.CommentData{
					Post: post1,
					Text: "Another Comment",
				})
				h.Comment().Create(env, &h.CommentData{
					Post: post1,
					Text: "Third Comment",
				})
				So(post1.LastCommentText(), ShouldEqual, "Third Comment")
				So(post1.Comments().Len(), ShouldEqual, 3)
			})
			Convey("Creating a user Will Smith", func() {
				userWillData := h.UserData{
					Name:  "Will Smith",
					Email: "will.smith@example.com",
				}
				userWill := h.User().Create(env, &userWillData)
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Checking constraint methods enforcement", func() {
				tag1Data := &h.TagData{
					Name:        "Tag1",
					Description: "Tag1",
				}
				So(func() { h.Tag().Create(env, tag1Data) }, ShouldPanic)
				tag2Data := &h.TagData{
					Name: "Tag2",
					Rate: 12,
				}
				So(func() { h.Tag().Create(env, tag2Data) }, ShouldPanic)
				tag3Data := &h.TagData{
					Name:        "Tag2",
					Description: "Tag2",
					Rate:        -3,
				}
				So(func() { h.Tag().Create(env, tag3Data) }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on creation (create only)", t, func() {
		So(models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot create records", func() {
				userTomData := h.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				So(func() { h.User().Create(env, &userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				h.User().Methods().Create().AllowGroup(group1)
				h.User().Methods().Load().AllowGroup(group1, h.User().Methods().Create())
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Write())
				userTomData := h.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				So(func() { h.User().Create(env, &userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for posts and it works", func() {
				h.Resume().Methods().Load().AllowGroup(group1)
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Create())
				userTomData := h.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := h.User().Create(env, &userTomData)
				So(func() { userTom.Name() }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				h.User().Methods().Load().AllowGroup(group1)
				userTomData := h.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := h.User().Create(env, &userTomData)
				So(userTom.Name(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldEqual, "tsmith@example.com")
			})
			Convey("Removing Create right on Email field", func() {
				h.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Write)
				userTomData := h.UserData{
					Name:  "Tom Smith",
					Email: "tsmith@example.com",
				}
				userTom := h.User().Create(env, &userTomData)
				So(userTom.Name(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldBeBlank)
				h.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Write)
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Searching User Jane", func() {
				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(userJane.Len(), ShouldEqual, 1)
				Convey("Reading Jane with getters", func() {
					So(userJane.Name(), ShouldEqual, "Jane Smith")
					So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
					So(userJane.Profile().Age(), ShouldEqual, 23)
					So(userJane.Profile().Money(), ShouldEqual, 12345)
					So(userJane.Profile().Country(), ShouldEqual, "USA")
					So(userJane.Profile().Zip(), ShouldEqual, "0305")
					recs := userJane.Posts().Records()
					So(len(recs), ShouldEqual, 2)
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
				usersAll := h.User().NewSet(env).OrderBy("Name").Load()
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
				userViews := h.UserView().NewSet(env).SearchAll()
				So(userViews.Len(), ShouldEqual, 3)
				userViews = h.UserView().NewSet(env).OrderBy("Name")
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
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list while searching", t, func() {
		So(models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot access records", func() {
				h.User().Methods().Search().AllowGroup(group1)
				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and checking access", func() {
				h.User().Methods().Load().AllowGroup(group1)

				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Name(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldEqual, "jane.smith@example.com")
				So(userJane.Age(), ShouldEqual, 23)
				So(func() { userJane.Profile().Age() }, ShouldPanic)
			})
			Convey("Adding field access rights to user 2 and checking access", func() {
				h.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Read)
				h.User().Fields().Age().RevokeAccess(security.GroupEveryone, security.Read)

				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Name(), ShouldEqual, "Jane Smith")
				So(userJane.Email(), ShouldBeBlank)
				So(userJane.Age(), ShouldEqual, 0)

				h.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Read)
				h.User().Fields().Age().GrantAccess(security.GroupEveryone, security.Read)
			})
			Convey("Checking record rules", func() {
				users := h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: q.User().Name().IContains("j").Condition,
					Perms:     security.Read,
				}
				h.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: q.User().Name().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				h.User().AddRecordRule(&notUsedRule)

				users = h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Name(), ShouldBeIn, []string{"Jane Smith", "John Smith"})
				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("writeRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	Convey("Testing advanced queries on M2O relations", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields", func() {
				users := h.User().Search(env, q.User().Profile().Equals(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
			Convey("Empty RecordSet on m2o relation fields", func() {
				users := h.User().Search(env, q.User().Profile().Equals(h.Profile().NewSet(env)))
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Empty RecordSet on m2o relation fields with IsNull", func() {
				users := h.User().Search(env, q.User().Profile().IsNull())
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2o relation fields with IN operator", func() {
				users := h.User().Search(env, q.User().Profile().In(jane.Profile()))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
			Convey("Empty RecordSet on m2o relation fields with IN operator", func() {
				users := h.User().Search(env, q.User().Profile().In(h.Profile().NewSet(env)))
				So(users.Len(), ShouldEqual, 0)
			})
			Convey("M2O chain", func() {
				users := h.User().Search(env, q.User().ProfileFilteredOn(q.Profile().BestPostFilteredOn(q.Post().Title().Equals("1st Post"))))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Conditions on o2m relation", func() {
				users := h.User().Search(env, q.User().Posts().Equals(jane.Posts().Records()[0]))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with IN operator", func() {
				users := h.User().Search(env, q.User().Posts().In(jane.Posts()))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with null", func() {
				users := h.User().Search(env, q.User().Posts().IsNull())
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].Name(), ShouldEqual, "John Smith")
				So(userRecs[1].Name(), ShouldEqual, "Will Smith")
			})
			Convey("O2M Chain", func() {
				users := h.User().Search(env, q.User().PostsFilteredOn(q.Post().Title().Equals("1st Post")))
				So(users.Len(), ShouldEqual, 1)
				So(users.ID(), ShouldEqual, jane.ID())
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on M2M relations", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
			So(post1.Len(), ShouldEqual, 1)
			post2 := h.Post().Search(env, q.Post().Title().Equals("2nd Post"))
			So(post2.Len(), ShouldEqual, 1)
			tag1 := h.Tag().Search(env, q.Tag().Name().Equals("Trending"))
			tag2 := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
			So(tag1.Len(), ShouldEqual, 1)
			Convey("Condition on m2m relation", func() {
				posts := h.Post().Search(env, q.Post().Tags().Equals(tag1))
				So(posts.Len(), ShouldEqual, 1)
				So(posts.ID(), ShouldEqual, post1.ID())
			})
			Convey("Condition on m2m relation with null", func() {
				posts := h.Post().Search(env, q.Post().Tags().IsNull())
				So(posts.Len(), ShouldEqual, 0)
			})
			Convey("Condition on m2m relation with IN operator", func() {
				tags := tag1.Union(tag2)
				posts := h.Post().Search(env, q.Post().Tags().In(tags))
				So(posts.Len(), ShouldEqual, 2)
			})
			Convey("M2M Chain", func() {
				posts := h.Post().Search(env, q.Post().TagsFilteredOn(q.Tag().Name().Equals("Trending")))
				So(posts.Len(), ShouldEqual, 1)
				So(posts.ID(), ShouldEqual, post1.ID())
			})
		}), ShouldBeNil)
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Update on users Jane and John with Write and Set", func() {
				jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set("Name", "Jane A. Smith")
				jane.Load()
				So(jane.Name(), ShouldEqual, "Jane A. Smith")
				So(jane.Email(), ShouldEqual, "jane.smith@example.com")

				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.UserData{
					Email:   "jsmith2@example.com",
					Nums:    13,
					IsStaff: false,
				}
				john.Write(&johnValues, h.User().Email(), h.User().IsStaff())
				john.Load()
				So(john.Name(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith2@example.com")
				So(john.Nums(), ShouldEqual, 13)
				So(john.IsStaff(), ShouldBeFalse)
				john.SetIsStaff(true)
				So(john.IsStaff(), ShouldBeTrue)
				john.SetIsStaff(false)
				So(john.IsStaff(), ShouldBeFalse)
				john.SetIsStaff(true)
				So(john.IsStaff(), ShouldBeTrue)
			})
			Convey("Multiple updates at once on users", func() {
				cond := q.User().Name().Equals("Jane A. Smith").Or().Name().Equals("John Smith")
				users := h.User().Search(env, cond)
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].IsStaff(), ShouldBeTrue)
				So(userRecs[1].IsStaff(), ShouldBeFalse)
				So(userRecs[0].IsActive(), ShouldBeFalse)
				So(userRecs[1].IsActive(), ShouldBeFalse)

				users.SetIsStaff(true)
				users.Load()
				So(userRecs[0].IsStaff(), ShouldBeTrue)
				So(userRecs[1].IsStaff(), ShouldBeTrue)

				uData := h.UserData{
					IsStaff:  false,
					IsActive: true,
				}
				users.Write(&uData, h.User().IsActive(), h.User().IsStaff())
				users.Load()
				So(userRecs[0].IsStaff(), ShouldBeFalse)
				So(userRecs[1].IsStaff(), ShouldBeFalse)
				So(userRecs[0].IsActive(), ShouldBeTrue)
				So(userRecs[1].IsActive(), ShouldBeTrue)
			})
			Convey("Updating many2one fields", func() {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				profile := userJane.Profile()
				userJane.SetProfile(h.Profile().NewSet(env))
				So(userJane.Profile().ID(), ShouldEqual, 0)
				userJane.SetProfile(profile)
				So(userJane.Profile().ID(), ShouldEqual, profile.Ids()[0])
			})
			Convey("Updating many2many fields", func() {
				post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
				tagBooks := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				post1.SetTags(tagBooks)

				post1Tags := post1.Tags()
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Name(), ShouldEqual, "Books")
				post2Tags := h.Post().Search(env, q.Post().Title().Equals("2nd Post")).Tags()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Name(), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Name(), ShouldBeIn, "Books", "Jane's")
			})
			Convey("Updating One2many fields", func() {
				posts := h.Post().NewSet(env)
				post1 := posts.Search(q.Post().Title().Equals("1st Post"))
				post2 := posts.Search(q.Post().Title().Equals("2nd Post"))
				post3 := posts.Create(&h.PostData{
					Title:   "3rd Post",
					Content: "Content of third post",
				})
				So(post3.Title(), ShouldEqual, "3rd Post")
				So(post3.Content(), ShouldEqual, "Content of third post")
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userJane.SetPosts(post1.Union(post3))
				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post3.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, 0)
			})
			Convey("Checking constraint methods enforcement", func() {
				tag1 := h.Tag().Search(env, q.Tag().Name().Equals("Trending"))
				So(func() { tag1.SetDescription("Trending") }, ShouldPanic)
				tag2 := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				So(func() { tag2.SetRate(12) }, ShouldPanic)
				So(func() {
					tag2.Write(&h.TagData{
						Description: "Books",
						Rate:        -3,
					}, h.Tag().Description(), h.Tag().Rate())
				}, ShouldPanic)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on update (write only)", t, func() {
		So(models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot update records", func() {
				h.User().Methods().Load().AllowGroup(group1)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				So(func() { john.Write(&johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				h.User().Methods().Write().AllowGroup(group1)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.UserData{
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
				h.User().Methods().Load().AllowGroup(group1)
				h.User().Methods().UpdateCity().AllowGroup(group1)
				jane := h.User().Search(env, q.User().Name().Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.UpdateCity("London") }, ShouldPanic)
			})
			Convey("Checking that user 2 can run UpdateCity after giving permission for caller", func() {
				h.User().Methods().Load().AllowGroup(group1)
				h.Profile().Methods().Load().AllowGroup(group1, h.User().Methods().UpdateCity())
				h.Profile().Methods().Write().AllowGroup(group1, h.User().Methods().UpdateCity())
				jane := h.User().Search(env, q.User().Name().Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.UpdateCity("London") }, ShouldNotPanic)
			})
			Convey("Removing Update right on Email field", func() {
				h.User().Fields().Email().RevokeAccess(security.GroupEveryone, security.Write)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.UserData{
					Email: "jsmith3@example.com",
					Nums:  13,
				}
				john.Write(&johnValues)
				john.Load()
				So(john.Name(), ShouldEqual, "John Smith")
				So(john.Email(), ShouldEqual, "jsmith2@example.com")
				So(john.Nums(), ShouldEqual, 13)
				h.User().Fields().Email().GrantAccess(security.GroupEveryone, security.Write)
			})
			Convey("Checking record rules", func() {
				userJane := h.User().NewSet(env).SearchAll()
				So(userJane.Len(), ShouldEqual, 3)

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: q.User().Name().IContains("j").Condition,
					Perms:     security.Write,
				}
				h.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "unlinkRule",
					Group:     group1,
					Condition: q.User().Name().Equals("Nobody").Condition,
					Perms:     security.Unlink,
				}
				h.User().AddRecordRule(&notUsedRule)

				userJane = h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Name(), ShouldEqual, "Jane A. Smith")
				userJane.SetName("Jane B. Smith")
				So(userJane.Name(), ShouldEqual, "Jane B. Smith")

				userWill := h.User().Search(env, q.User().Name().Equals("Will Smith"))
				So(func() { userWill.SetName("Will Jr. Smith") }, ShouldPanic)

				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("unlinkRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			users := h.User().Search(env, q.User().Name().Equals("John Smith"))
			num := users.Unlink()
			Convey("Number of deleted record should be 1", func() {
				So(num, ShouldEqual, 1)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Checking unlink access permissions", t, func() {
		So(models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot unlink records", func() {
				h.User().Methods().Load().AllowGroup(group1)
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(func() { users.Unlink() }, ShouldPanic)
			})
			Convey("Adding unlink permission to user2", func() {
				h.User().Methods().Unlink().AllowGroup(group1)
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				num := users.Unlink()
				So(num, ShouldEqual, 1)
			})
			Convey("Checking record rules", func() {

				rule := models.RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: q.User().Name().IContains("j").Condition,
					Perms:     security.Unlink,
				}
				h.User().AddRecordRule(&rule)

				notUsedRule := models.RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: q.User().Name().Equals("Nobody").Condition,
					Perms:     security.Write,
				}
				h.User().AddRecordRule(&notUsedRule)

				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Unlink(), ShouldEqual, 1)

				userWill := h.User().Search(env, q.User().Name().Equals("Will Smith"))
				So(userWill.Unlink(), ShouldEqual, 0)

				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("writeRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}
