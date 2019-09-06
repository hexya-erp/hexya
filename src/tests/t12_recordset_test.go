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

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := h.User().NewData().
					SetName("John Smith").
					SetEmail("jsmith@example.com").
					SetIsStaff(true)
				userJohn := h.User().Create(env, userJohnData)
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Creating user Jane with related Profile and Posts and Comments and Tags", func() {
				userJaneData := h.User().NewData().
					SetName("Jane Smith").
					SetEmail("jane.smith@example.com").
					SetNums(2).
					CreateProfile(h.Profile().NewData().
						SetAge(23).
						SetMoney(12345).
						SetStreet("165 5th Avenue").
						SetCity("New York").
						SetZip("0305").
						SetCountry("USA")).
					CreatePosts(h.Post().NewData().
						SetTitle("1st Post").
						SetContent("Content of first post")).
					CreatePosts(h.Post().NewData().
						SetTitle("2nd Post").
						SetContent("Content of second post"))
				userJane := h.User().Create(env, userJaneData)
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Profile().ID(), ShouldNotEqual, 0)
				So(userJane.Profile().UserName(), ShouldEqual, "Jane Smith")

				post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
				post2 := h.Post().Search(env, q.Post().Title().Equals("2nd Post"))
				So(post1.Len(), ShouldEqual, 1)
				So(post2.Len(), ShouldEqual, 1)

				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, userJane.ID())
				So(userJane.Posts().Len(), ShouldEqual, 2)

				userJane.Profile().SetBestPost(post1)

				tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Trending"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Books"))
				tag3 := h.Tag().Create(env, h.Tag().NewData().SetName("Jane's"))
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
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("First Comment"))
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("Another Comment"))
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("Third Comment"))
				So(post1.LastCommentText(), ShouldEqual, "Third Comment")
				So(post1.Comments().Len(), ShouldEqual, 3)
			})
			Convey("Creating a user Will Smith", func() {
				userWillData := h.User().NewData().
					SetName("Will Smith").
					SetEmail("will.smith@example.com")
				userWill := h.User().Create(env, userWillData)
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.ID(), ShouldBeGreaterThan, 0)
			})
			Convey("Checking constraint methods enforcement", func() {
				tag1Data := h.Tag().NewData().SetName("Tag1").SetDescription("Tag1")
				So(func() { h.Tag().Create(env, tag1Data) }, ShouldPanic)
				tag2Data := h.Tag().NewData().SetName("Tag2").SetRate(12)
				So(func() { h.Tag().Create(env, tag2Data) }, ShouldPanic)
				tag3Data := h.Tag().NewData().SetName("Tag2").SetDescription("Tag2").SetRate(-3)
				So(func() { h.Tag().Create(env, tag3Data) }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on creation (create only)", t, func() {
		So(models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			Convey("Checking that user 2 cannot create records", func() {
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				So(func() { h.User().Create(env, userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				h.User().Methods().Create().AllowGroup(group1)
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Write())
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				So(func() { h.User().Create(env, userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for posts and it works", func() {
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Create())
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				userTom := h.User().Create(env, userTomData)
				So(func() { userTom.Name() }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				h.User().Methods().Load().AllowGroup(group1)
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				userTom := h.User().Create(env, userTomData)
				So(userTom.Name(), ShouldEqual, "Tom Smith")
				So(userTom.Email(), ShouldEqual, "tsmith@example.com")
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
				Convey("Reading Jane with First", func() {
					ujData := userJane.First()
					So(ujData.Name(), ShouldEqual, "Jane Smith")
					So(ujData.HasName(), ShouldBeTrue)
					So(ujData.Email(), ShouldEqual, "jane.smith@example.com")
					So(ujData.HasEmail(), ShouldBeTrue)
					So(ujData.ID(), ShouldEqual, userJane.ID())
					So(ujData.HasID(), ShouldBeTrue)
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
					usersData := usersAll.All()
					So(usersData[0].Email(), ShouldEqual, "jane.smith@example.com")
					So(usersData[0].HasEmail(), ShouldBeTrue)
					So(usersData[1].Email(), ShouldEqual, "jsmith@example.com")
					So(usersData[1].HasEmail(), ShouldBeTrue)
					So(usersData[2].Email(), ShouldEqual, "will.smith@example.com")
					So(usersData[2].HasEmail(), ShouldBeTrue)
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
			Convey("Checking ModelData methods", func() {
				johnValues := h.User().NewData().
					SetEmail("jsmith2@example.com").
					SetNums(13).
					SetIsStaff(false)
				So(johnValues.Nums(), ShouldEqual, 13)
				So(johnValues.HasNums(), ShouldBeTrue)
				jv2 := johnValues.Copy()
				johnValues.UnsetNums()
				So(johnValues.Nums(), ShouldEqual, 0)
				So(johnValues.HasNums(), ShouldBeFalse)
				So(jv2.Nums(), ShouldEqual, 13)
				So(jv2.HasNums(), ShouldBeTrue)
			})
			Convey("Checking FieldMap conversion to ModelData", func() {
				fm := models.FieldMap{
					"Email": "jsmith2@example.com",
					"Nums":  13,
				}
				ud := h.User().NewData(fm)
				So(ud.Email(), ShouldEqual, "jsmith2@example.com")
				So(ud.HasEmail(), ShouldBeTrue)
				So(ud.HasIsStaff(), ShouldBeFalse)
			})
			Convey("Update on users Jane and John with Write and Set", func() {
				jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set("Name", "Jane A. Smith")
				jane.Load()
				So(jane.Name(), ShouldEqual, "Jane A. Smith")
				So(jane.Email(), ShouldEqual, "jane.smith@example.com")

				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.User().NewData().
					SetEmail("jsmith2@example.com").
					SetNums(13).
					SetIsStaff(false)
				john.Write(johnValues)
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

				uData := h.User().NewData().
					SetIsStaff(false).
					SetIsActive(true)
				users.Write(uData)
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

				post1 := profile.BestPost()
				profile.Write(h.Profile().NewData().
					CreateBestPost(h.Post().NewData().
						SetTitle("Post created on the Fly")))
				So(profile.BestPost().Title(), ShouldEqual, "Post created on the Fly")
				profile.SetBestPost(post1)
			})
			Convey("Updating many2many fields", func() {
				post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))

				post1.Write(h.Post().NewData().
					CreateTags(h.Tag().NewData().
						SetName("Tag created on the fly")).
					CreateTags(h.Tag().NewData().
						SetName("Second Tag on the fly")))
				post1Tags := post1.Tags()
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Name(), ShouldBeIn, []string{"Tag created on the fly", "Second Tag on the fly"})
				So(post1Tags.Records()[1].Name(), ShouldBeIn, []string{"Tag created on the fly", "Second Tag on the fly"})

				tagBooks := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				post1.SetTags(tagBooks)
				post1Tags = post1.Tags()
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
				post3 := posts.Create(h.Post().NewData().
					SetTitle("3rd Post").
					SetContent("Content of third post"))
				So(post3.Title(), ShouldEqual, "3rd Post")
				So(post3.Content(), ShouldEqual, "Content of third post")
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userJane.SetPosts(post1.Union(post3))
				So(post1.User().ID(), ShouldEqual, userJane.ID())
				So(post3.User().ID(), ShouldEqual, userJane.ID())
				So(post2.User().ID(), ShouldEqual, 0)

				userJane.SetPosts(nil)
				userJane.Write(h.User().NewData().
					CreatePosts(h.Post().NewData().
						SetTitle("Another post created on the fly")).
					CreatePosts(h.Post().NewData().
						SetTitle("One more post created on the fly")))
				So(userJane.Posts().Len(), ShouldEqual, 2)
				So(userJane.Posts().Records()[0].Title(), ShouldBeIn, []string{"Another post created on the fly", "One more post created on the fly"})
				So(userJane.Posts().Records()[1].Title(), ShouldBeIn, []string{"Another post created on the fly", "One more post created on the fly"})

				userJane.SetPosts(post1.Union(post3))
				So(userJane.Posts().Len(), ShouldEqual, 2)
			})
			Convey("Checking constraint methods enforcement", func() {
				tag1 := h.Tag().Search(env, q.Tag().Name().Equals("Trending"))
				So(func() { tag1.SetDescription("Trending") }, ShouldPanic)
				tag2 := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				So(func() { tag2.SetRate(12) }, ShouldPanic)
				So(func() {
					tag2.Write(h.Tag().NewData().
						SetDescription("Books").
						SetRate(-3))
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
				johnValues := h.User().NewData().
					SetEmail("jsmith3@example.com").
					SetNums(13)
				So(func() { john.Write(johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				h.User().Methods().Write().AllowGroup(group1)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := h.User().NewData().
					SetEmail("jsmith3@example.com").
					SetNums(13)
				john.Write(johnValues)
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
				h.Profile().Methods().Write().AllowGroup(group1, h.User().Methods().UpdateCity())
				jane := h.User().Search(env, q.User().Name().Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.UpdateCity("London") }, ShouldNotPanic)
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
			Convey("Number of deleted record should be 1", func() {
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				num := users.Unlink()
				So(num, ShouldEqual, 1)
			})
			Convey("Deleted RecordSet should update themselves when reloading", func() {
				userJohn := h.User().Search(env, q.User().Name().Equals("John Smith"))
				userJohn2 := h.User().Search(env, q.User().Name().Equals("John Smith"))
				users := h.User().Search(env, q.User().Name().Equals("John Smith").Or().Name().Equals("Jane A. Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn2.Len(), ShouldEqual, 1)
				So(users.Len(), ShouldEqual, 2)
				userJohn.Unlink()
				userJohn.ForceLoad()
				So(userJohn.Len(), ShouldEqual, 0)
				userJohn2.ForceLoad()
				So(userJohn2.Len(), ShouldEqual, 0)
				users.ForceLoad()
				So(users.Len(), ShouldEqual, 1)
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
				So(func() { users.Unlink() }, ShouldPanic)
			})
			Convey("Adding permissions to user2 on Profile and Post", func() {
				h.Profile().Methods().Load().AllowGroup(group1)
				h.Post().Methods().Load().AllowGroup(group1)
				h.Post().Methods().Write().AllowGroup(group1)
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
