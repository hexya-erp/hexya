// Copyright 2016 Nicolas Piganeau. All Rights Reserved.
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

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func TestCreateRecordSet(t *testing.T) {
	t.Run("Test record creation", func(t *testing.T) {
		assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Creating simple user John with no relations and checking ID", func(t *testing.T) {
				userJohnData := h.User().NewData().
					SetName("John Smith").
					SetEmail("jsmith@example.com").
					SetIsStaff(true)
				userJohn := h.User().Create(env, userJohnData)
				assert.EqualValues(t, userJohn.Len(), 1)
				assert.Greater(t, userJohn.ID(), int64(0))
			})
			t.Run("Creating user Jane with related Profile and Posts and Comments and Tags", func(t *testing.T) {
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
				assert.EqualValues(t, userJane.Len(), 1)
				assert.NotEqualValues(t, userJane.Profile().ID(), int64(0))
				assert.EqualValues(t, userJane.Profile().UserName(), "Jane Smith")

				post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
				post2 := h.Post().Search(env, q.Post().Title().Equals("2nd Post"))
				assert.EqualValues(t, post1.Len(), 1)
				assert.EqualValues(t, post2.Len(), 1)

				assert.EqualValues(t, post1.User().ID(), userJane.ID())
				assert.EqualValues(t, post2.User().ID(), userJane.ID())
				assert.EqualValues(t, userJane.Posts().Len(), 2)

				userJane.Profile().SetBestPost(post1)

				tag1 := h.Tag().Create(env, h.Tag().NewData().SetName("Trending"))
				tag2 := h.Tag().Create(env, h.Tag().NewData().SetName("Books"))
				tag3 := h.Tag().Create(env, h.Tag().NewData().SetName("Jane's"))
				assert.Empty(t, post1.FirstTagName())
				post1.SetTags(tag1.Union(tag3))
				post2.SetTags(tag2.Union(tag3))
				assert.EqualValues(t, post1.FirstTagName(), "Trending")
				post1Tags := post1.Tags()
				assert.EqualValues(t, post1Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Trending", "Jane's"}, post1Tags.Records()[0].Name())
				assert.Contains(t, []interface{}{"Trending", "Jane's"}, post1Tags.Records()[1].Name())
				post2Tags := post2.Tags()
				assert.EqualValues(t, post2Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[0].Name())
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[1].Name())

				assert.Empty(t, post1.FirstCommentText())
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("First Comment"))
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("Another Comment"))
				h.Comment().Create(env, h.Comment().NewData().SetPost(post1).SetText("Third Comment"))
				assert.EqualValues(t, post1.FirstCommentText(), "First Comment")
				assert.EqualValues(t, post1.Comments().Len(), 3)
			})
			t.Run("Creating a user Will Smith", func(t *testing.T) {
				userWillData := h.User().NewData().
					SetName("Will Smith").
					SetEmail("will.smith@example.com")
				userWill := h.User().Create(env, userWillData)
				assert.EqualValues(t, userWill.Len(), 1)
				assert.Greater(t, userWill.ID(), int64(0))
			})
			t.Run("Checking constraint methods enforcement", func(t *testing.T) {
				tag1Data := h.Tag().NewData().SetName("Tag1").SetDescription("Tag1")
				assert.Panics(t, func() { h.Tag().Create(env, tag1Data) })
				tag2Data := h.Tag().NewData().SetName("Tag2").SetRate(12)
				assert.Panics(t, func() { h.Tag().Create(env, tag2Data) })
				tag3Data := h.Tag().NewData().SetName("Tag2").SetDescription("Tag2").SetRate(-3)
				assert.Panics(t, func() { h.Tag().Create(env, tag3Data) })
			})
		}))
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	t.Run("Testing access control list on creation (create only)", func(t *testing.T) {
		t.Run("Checking that user 2 cannot create records", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				assert.Panics(t, func() { h.User().Create(env, userTomData) })
			}))
		})
		t.Run("Adding model access rights to user 2 and check failure again", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.User().Methods().Create().AllowGroup(group1)
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Write())
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				assert.Panics(t, func() { h.User().Create(env, userTomData) })
			}))
		})
		t.Run("Adding model access rights to user 2 for posts and it works", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.Resume().Methods().Create().AllowGroup(group1, h.User().Methods().Create())
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				userTom := h.User().Create(env, userTomData)
				assert.Panics(t, func() { userTom.Name() })
			}))
		})
		t.Run("Checking creation again with read rights too", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.User().Methods().Load().AllowGroup(group1)
				userTomData := h.User().NewData().
					SetName("Tom Smith").
					SetEmail("tsmith@example.com")
				userTom := h.User().Create(env, userTomData)
				assert.EqualValues(t, userTom.Name(), "Tom Smith")
				assert.EqualValues(t, userTom.Email(), "tsmith@example.com")
			}))
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	t.Run("Testing search through RecordSets", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Searching User Jane", func(t *testing.T) {
				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				assert.EqualValues(t, userJane.Len(), 1)
				t.Run("Reading Jane with getters", func(t *testing.T) {
					assert.EqualValues(t, userJane.Name(), "Jane Smith")
					assert.EqualValues(t, userJane.Email(), "jane.smith@example.com")
					assert.EqualValues(t, userJane.Profile().Age(), 23)
					assert.EqualValues(t, userJane.Profile().Money(), 12345)
					assert.EqualValues(t, userJane.Profile().Country(), "USA")
					assert.EqualValues(t, userJane.Profile().Zip(), "0305")
					recs := userJane.Posts().Records()
					assert.EqualValues(t, len(recs), 2)
					assert.EqualValues(t, recs[0].Title(), "1st Post")
					assert.EqualValues(t, recs[1].Title(), "2nd Post")
				})
				t.Run("Reading Jane with low level getters", func(t *testing.T) {
					assert.EqualValues(t, userJane.Get(h.User().Fields().Name()), "Jane Smith")
					assert.EqualValues(t, userJane.Get(q.User().Name()), "Jane Smith")
					assert.EqualValues(t, userJane.Get(h.User().Fields().Profile()).(models.RecordSet).Get(h.Profile().Fields().Age()), 23)
					assert.EqualValues(t, userJane.Get(q.User().Profile()).(models.RecordSet).Get(q.Profile().Age()), 23)
				})
				t.Run("Reading Jane with First", func(t *testing.T) {
					ujData := userJane.First()
					assert.EqualValues(t, ujData.Name(), "Jane Smith")
					assert.True(t, ujData.HasName())
					assert.EqualValues(t, ujData.Email(), "jane.smith@example.com")
					assert.True(t, ujData.HasEmail())
					assert.EqualValues(t, ujData.ID(), userJane.ID())
					assert.True(t, ujData.HasID())
				})
			})

			t.Run("Testing search all users", func(t *testing.T) {
				usersAll := h.User().NewSet(env).OrderBy("Name").Load()
				assert.EqualValues(t, usersAll.Len(), 3)
				t.Run("Reading first user with getters", func(t *testing.T) {
					assert.EqualValues(t, usersAll.Name(), "Jane Smith")
					assert.EqualValues(t, usersAll.Email(), "jane.smith@example.com")
				})
				t.Run("Reading all users with Records and Get", func(t *testing.T) {
					recs := usersAll.Records()
					assert.EqualValues(t, len(recs), 3)
					assert.EqualValues(t, recs[0].Email(), "jane.smith@example.com")
					assert.EqualValues(t, recs[1].Email(), "jsmith@example.com")
					assert.EqualValues(t, recs[2].Email(), "will.smith@example.com")
				})
				t.Run("Reading all users with ReadAll()", func(t *testing.T) {
					usersData := usersAll.All()
					assert.EqualValues(t, usersData[0].Email(), "jane.smith@example.com")
					assert.True(t, usersData[0].HasEmail())
					assert.EqualValues(t, usersData[1].Email(), "jsmith@example.com")
					assert.True(t, usersData[1].HasEmail())
					assert.EqualValues(t, usersData[2].Email(), "will.smith@example.com")
					assert.True(t, usersData[2].HasEmail())
				})
			})

			t.Run("Testing search on manual model", func(t *testing.T) {
				userViews := h.UserView().NewSet(env).SearchAll()
				assert.EqualValues(t, userViews.Len(), 3)
				userViews = h.UserView().NewSet(env).OrderBy("Name")
				assert.EqualValues(t, userViews.Len(), 3)
				recs := userViews.Records()
				assert.EqualValues(t, len(recs), 3)
				assert.EqualValues(t, recs[0].Name(), "Jane Smith")
				assert.EqualValues(t, recs[1].Name(), "John Smith")
				assert.EqualValues(t, recs[2].Name(), "Will Smith")
				assert.EqualValues(t, recs[0].City(), "New York")
				assert.EqualValues(t, recs[1].City(), "")
				assert.EqualValues(t, recs[2].City(), "")
			})
		}))
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	t.Run("Testing access control list while searching", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			t.Run("Checking that user 2 cannot access records", func(t *testing.T) {
				h.User().Methods().Search().AllowGroup(group1)
				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				assert.Panics(t, func() { userJane.Load() })
			})
			t.Run("Adding model access rights to user 2 and checking access", func(t *testing.T) {
				h.User().Methods().Load().AllowGroup(group1)

				userJane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				assert.NotPanics(t, func() { userJane.Load() })
				assert.EqualValues(t, userJane.Name(), "Jane Smith")
				assert.EqualValues(t, userJane.Email(), "jane.smith@example.com")
				assert.EqualValues(t, userJane.Age(), 23)
				assert.Panics(t, func() { userJane.Profile().Age() })
			})
			t.Run("Checking record rules", func(t *testing.T) {
				users := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, users.Len(), 3)

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
				assert.EqualValues(t, users.Len(), 2)
				assert.Contains(t, []string{"Jane Smith", "John Smith"}, users.Records()[0].Name())
				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("writeRule")
			})
		}))
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	t.Run("Testing advanced queries on M2O relations", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
			assert.EqualValues(t, jane.Len(), 1)
			t.Run("Condition on m2o relation fields", func(t *testing.T) {
				users := h.User().Search(env, q.User().Profile().Equals(jane.Profile()))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
			t.Run("Empty RecordSet on m2o relation fields", func(t *testing.T) {
				users := h.User().Search(env, q.User().Profile().Equals(h.Profile().NewSet(env)))
				assert.EqualValues(t, users.Len(), 2)
			})
			t.Run("Empty RecordSet on m2o relation fields with IsNull", func(t *testing.T) {
				users := h.User().Search(env, q.User().Profile().IsNull())
				assert.EqualValues(t, users.Len(), 2)
			})
			t.Run("Condition on m2o relation fields with IN operator", func(t *testing.T) {
				users := h.User().Search(env, q.User().Profile().In(jane.Profile()))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
			t.Run("Empty RecordSet on m2o relation fields with IN operator", func(t *testing.T) {
				users := h.User().Search(env, q.User().Profile().In(h.Profile().NewSet(env)))
				assert.EqualValues(t, users.Len(), 0)
			})
			t.Run("M2O chain", func(t *testing.T) {
				users := h.User().Search(env, q.User().ProfileFilteredOn(q.Profile().BestPostFilteredOn(q.Post().Title().Equals("1st Post"))))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
		}))
	})
	t.Run("Testing advanced queries on O2M relations", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
			assert.EqualValues(t, jane.Len(), 1)
			t.Run("Conditions on o2m relation", func(t *testing.T) {
				users := h.User().Search(env, q.User().Posts().Equals(jane.Posts().Records()[0]))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
			t.Run("Conditions on o2m relation with IN operator", func(t *testing.T) {
				users := h.User().Search(env, q.User().Posts().In(jane.Posts()))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
			t.Run("Conditions on o2m relation with null", func(t *testing.T) {
				users := h.User().Search(env, q.User().Posts().IsNull())
				assert.EqualValues(t, users.Len(), 2)
				userRecs := users.Records()
				assert.EqualValues(t, userRecs[0].Name(), "John Smith")
				assert.EqualValues(t, userRecs[1].Name(), "Will Smith")
			})
			t.Run("O2M Chain", func(t *testing.T) {
				users := h.User().Search(env, q.User().PostsFilteredOn(q.Post().Title().Equals("1st Post")))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.ID(), jane.ID())
			})
		}))
	})
	t.Run("Testing advanced queries on M2M relations", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))
			assert.EqualValues(t, post1.Len(), 1)
			post2 := h.Post().Search(env, q.Post().Title().Equals("2nd Post"))
			assert.EqualValues(t, post2.Len(), 1)
			tag1 := h.Tag().Search(env, q.Tag().Name().Equals("Trending"))
			tag2 := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
			assert.EqualValues(t, tag1.Len(), 1)
			t.Run("Condition on m2m relation", func(t *testing.T) {
				posts := h.Post().Search(env, q.Post().Tags().Equals(tag1))
				assert.EqualValues(t, posts.Len(), 1)
				assert.EqualValues(t, posts.ID(), post1.ID())
			})
			t.Run("Condition on m2m relation with null", func(t *testing.T) {
				posts := h.Post().Search(env, q.Post().Tags().IsNull())
				assert.EqualValues(t, posts.Len(), 0)
			})
			t.Run("Condition on m2m relation with IN operator", func(t *testing.T) {
				tags := tag1.Union(tag2)
				posts := h.Post().Search(env, q.Post().Tags().In(tags))
				assert.EqualValues(t, posts.Len(), 2)
			})
			t.Run("M2M Chain", func(t *testing.T) {
				posts := h.Post().Search(env, q.Post().TagsFilteredOn(q.Tag().Name().Equals("Trending")))
				assert.EqualValues(t, posts.Len(), 1)
				assert.EqualValues(t, posts.ID(), post1.ID())
			})
		}))
	})
}

func TestUpdateRecordSet(t *testing.T) {
	t.Run("Testing updates through RecordSets", func(t *testing.T) {
		assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			t.Run("Checking ModelData methods", func(t *testing.T) {
				johnValues := h.User().NewData().
					SetEmail("jsmith2@example.com").
					SetNums(13).
					SetIsStaff(false)
				assert.EqualValues(t, johnValues.Nums(), 13)
				assert.True(t, johnValues.HasNums())
				jv2 := johnValues.Copy()
				johnValues.UnsetNums()
				assert.EqualValues(t, johnValues.Nums(), 0)
				assert.False(t, johnValues.HasNums())
				assert.EqualValues(t, jv2.Nums(), 13)
				assert.True(t, jv2.HasNums())
			})
			t.Run("Checking FieldMap conversion to ModelData", func(t *testing.T) {
				fm := models.FieldMap{
					"Email": "jsmith2@example.com",
					"Nums":  13,
				}
				ud := h.User().NewData(fm)
				assert.EqualValues(t, ud.Email(), "jsmith2@example.com")
				assert.True(t, ud.HasEmail())
				assert.False(t, ud.HasIsStaff())
			})
			t.Run("Update on users Jane and John with Write and Set", func(t *testing.T) {
				jane := h.User().Search(env, q.User().Name().Equals("Jane Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				jane.SetName("Jane A. Smith")
				jane.Load()
				assert.EqualValues(t, jane.Name(), "Jane A. Smith")
				assert.EqualValues(t, jane.Email(), "jane.smith@example.com")

				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := h.User().NewData().
					SetEmail("jsmith2@example.com").
					SetNums(13).
					SetIsStaff(false)
				john.Write(johnValues)
				john.Load()
				assert.EqualValues(t, john.Name(), "John Smith")
				assert.EqualValues(t, john.Email(), "jsmith2@example.com")
				assert.EqualValues(t, john.Nums(), 13)
				assert.False(t, john.IsStaff())
				john.SetIsStaff(true)
				assert.True(t, john.IsStaff())
				john.SetIsStaff(false)
				assert.False(t, john.IsStaff())
				john.SetIsStaff(true)
				assert.True(t, john.IsStaff())
			})
			t.Run("Multiple updates at once on users", func(t *testing.T) {
				cond := q.User().Name().Equals("Jane A. Smith").Or().Name().Equals("John Smith")
				users := h.User().Search(env, cond)
				assert.EqualValues(t, users.Len(), 2)
				userRecs := users.Records()
				assert.True(t, userRecs[0].IsStaff())
				assert.False(t, userRecs[1].IsStaff())
				assert.False(t, userRecs[0].IsActive())
				assert.False(t, userRecs[1].IsActive())

				users.SetIsStaff(true)
				users.Load()
				assert.True(t, userRecs[0].IsStaff())
				assert.True(t, userRecs[1].IsStaff())

				uData := h.User().NewData().
					SetIsStaff(false).
					SetIsActive(true)
				users.Write(uData)
				users.Load()
				assert.False(t, userRecs[0].IsStaff())
				assert.False(t, userRecs[1].IsStaff())
				assert.True(t, userRecs[0].IsActive())
				assert.True(t, userRecs[1].IsActive())
			})
			t.Run("Updating many2one fields", func(t *testing.T) {
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				profile := userJane.Profile()
				userJane.SetProfile(h.Profile().NewSet(env))
				assert.EqualValues(t, userJane.Profile().ID(), int64(0))
				userJane.SetProfile(profile)
				assert.EqualValues(t, userJane.Profile().ID(), profile.Ids()[0])

				post1 := profile.BestPost()
				profile.Write(h.Profile().NewData().
					CreateBestPost(h.Post().NewData().
						SetTitle("Post created on the Fly")))
				assert.EqualValues(t, profile.BestPost().Title(), "Post created on the Fly")
				profile.SetBestPost(post1)
			})
			t.Run("Updating many2many fields", func(t *testing.T) {
				post1 := h.Post().Search(env, q.Post().Title().Equals("1st Post"))

				post1.Write(h.Post().NewData().
					CreateTags(h.Tag().NewData().
						SetName("Tag created on the fly")).
					CreateTags(h.Tag().NewData().
						SetName("Second Tag on the fly")))
				post1Tags := post1.Tags()
				assert.EqualValues(t, post1Tags.Len(), 2)
				assert.Contains(t, []string{"Tag created on the fly", "Second Tag on the fly"}, post1Tags.Records()[0].Name())
				assert.Contains(t, []string{"Tag created on the fly", "Second Tag on the fly"}, post1Tags.Records()[1].Name())

				tagBooks := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				post1.SetTags(tagBooks)
				post1Tags = post1.Tags()
				assert.EqualValues(t, post1Tags.Len(), 1)
				assert.EqualValues(t, post1Tags.Name(), "Books")

				post2Tags := h.Post().Search(env, q.Post().Title().Equals("2nd Post")).Tags()
				assert.EqualValues(t, post2Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[0].Name())
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[1].Name())
			})
			t.Run("Updating One2many fields", func(t *testing.T) {
				posts := h.Post().NewSet(env)
				post1 := posts.Search(q.Post().Title().Equals("1st Post"))
				post2 := posts.Search(q.Post().Title().Equals("2nd Post"))
				post3 := posts.Create(h.Post().NewData().
					SetTitle("3rd Post").
					SetContent("Content of third post"))
				assert.EqualValues(t, post3.Title(), "3rd Post")
				assert.EqualValues(t, post3.Content(), "Content of third post")
				userJane := h.User().Search(env, q.User().Email().Equals("jane.smith@example.com"))
				userJane.SetPosts(post1.Union(post3))
				assert.EqualValues(t, post1.User().ID(), userJane.ID())
				assert.EqualValues(t, post3.User().ID(), userJane.ID())
				assert.EqualValues(t, post2.User().ID(), int64(0))

				userJane.SetPosts(nil)
				userJane.Write(h.User().NewData().
					CreatePosts(h.Post().NewData().
						SetTitle("Another post created on the fly")).
					CreatePosts(h.Post().NewData().
						SetTitle("One more post created on the fly")))
				assert.EqualValues(t, userJane.Posts().Len(), 2)
				assert.Contains(t, []string{"Another post created on the fly", "One more post created on the fly"}, userJane.Posts().Records()[0].Title())
				assert.Contains(t, []string{"Another post created on the fly", "One more post created on the fly"}, userJane.Posts().Records()[1].Title())

				userJane.SetPosts(post1.Union(post3))
				assert.EqualValues(t, userJane.Posts().Len(), 2)
			})
			t.Run("Checking constraint methods enforcement", func(t *testing.T) {
				tag1 := h.Tag().Search(env, q.Tag().Name().Equals("Trending"))
				assert.Panics(t, func() { tag1.SetDescription("Trending") })
				tag2 := h.Tag().Search(env, q.Tag().Name().Equals("Books"))
				assert.Panics(t, func() { tag2.SetRate(12) })
				assert.Panics(t, func() {
					tag2.Write(h.Tag().NewData().
						SetDescription("Books").
						SetRate(-3))
				})
			})
		}))
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	t.Run("Testing access control list on update (write only)", func(t *testing.T) {
		assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
			security.Registry.AddMembership(2, group1)
			t.Run("Checking that user 2 cannot update records", func(t *testing.T) {
				h.User().Methods().Load().AllowGroup(group1)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := h.User().NewData().
					SetEmail("jsmith3@example.com").
					SetNums(13)
				assert.Panics(t, func() { john.Write(johnValues) })
			})
			t.Run("Adding model access rights to user 2 and check update", func(t *testing.T) {
				h.User().Methods().Write().AllowGroup(group1)
				john := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := h.User().NewData().
					SetEmail("jsmith3@example.com").
					SetNums(13)
				john.Write(johnValues)
				john.Load()
				assert.EqualValues(t, john.Name(), "John Smith")
				assert.EqualValues(t, john.Email(), "jsmith3@example.com")
				assert.EqualValues(t, john.Nums(), 13)
			})
			t.Run("Checking that user 2 cannot update profile through UpdateCity method", func(t *testing.T) {
				h.User().Methods().Load().AllowGroup(group1)
				h.User().Methods().UpdateCity().AllowGroup(group1)
				jane := h.User().Search(env, q.User().Name().Equals("Jane A. Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				assert.Panics(t, func() { jane.UpdateCity("London") })
			})
			t.Run("Checking that user 2 can run UpdateCity after giving permission for caller", func(t *testing.T) {
				h.User().Methods().Load().AllowGroup(group1)
				h.Profile().Methods().Write().AllowGroup(group1, h.User().Methods().UpdateCity())
				jane := h.User().Search(env, q.User().Name().Equals("Jane A. Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				assert.NotPanics(t, func() { jane.UpdateCity("London") })
			})
			t.Run("Checking record rules", func(t *testing.T) {
				userJane := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, userJane.Len(), 3)

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
				assert.EqualValues(t, userJane.Len(), 1)
				assert.EqualValues(t, userJane.Name(), "Jane A. Smith")
				userJane.SetName("Jane B. Smith")
				assert.EqualValues(t, userJane.Name(), "Jane B. Smith")

				userWill := h.User().Search(env, q.User().Name().Equals("Will Smith"))
				assert.Panics(t, func() { userWill.SetName("Will Jr. Smith") })

				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("unlinkRule")
			})
		}))
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	t.Run("Delete user John Smith", func(t *testing.T) {
		t.Run("Number of deleted record should be 1", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				num := users.Unlink()
				assert.EqualValues(t, num, 1)
			}))
		})
		t.Run("Deleted RecordSet should update themselves when reloading", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userJohn := h.User().Search(env, q.User().Name().Equals("John Smith"))
				userJohn2 := h.User().Search(env, q.User().Name().Equals("John Smith"))
				users := h.User().Search(env, q.User().Name().Equals("John Smith").Or().Name().Equals("Jane A. Smith"))
				assert.EqualValues(t, userJohn.Len(), 1)
				assert.EqualValues(t, userJohn2.Len(), 1)
				assert.EqualValues(t, users.Len(), 2)
				userJohn.Unlink()
				userJohn.ForceLoad()
				assert.EqualValues(t, userJohn.Len(), 0)
				userJohn2.ForceLoad()
				assert.EqualValues(t, userJohn2.Len(), 0)
				users.ForceLoad()
				assert.EqualValues(t, users.Len(), 1)
			}))
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	t.Run("Checking unlink access permissions", func(t *testing.T) {
		t.Run("Checking that user 2 cannot unlink records", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.User().Methods().Load().AllowGroup(group1)
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.Panics(t, func() { users.Unlink() })
			}))
		})
		t.Run("Adding unlink permission to user2", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.User().Methods().Unlink().AllowGroup(group1)
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				assert.Panics(t, func() { users.Unlink() })
			}))
		})
		t.Run("Adding permissions to user2 on Profile and Post", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)
				h.Profile().Methods().Load().AllowGroup(group1)
				h.Post().Methods().Load().AllowGroup(group1)
				h.Post().Methods().Write().AllowGroup(group1)
				users := h.User().Search(env, q.User().Name().Equals("John Smith"))
				num := users.Unlink()
				assert.EqualValues(t, num, 1)
			}))
		})
		t.Run("Checking record rules", func(t *testing.T) {
			assert.Nil(t, models.SimulateInNewEnvironment(2, func(env models.Environment) {
				security.Registry.AddMembership(2, group1)

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
				assert.EqualValues(t, userJane.Len(), 1)
				assert.EqualValues(t, userJane.Unlink(), 1)

				userWill := h.User().Search(env, q.User().Name().Equals("Will Smith"))
				assert.EqualValues(t, userWill.Unlink(), 0)

				h.User().RemoveRecordRule("jOnly")
				h.User().RemoveRecordRule("writeRule")
			}))
		})
	})
	security.Registry.UnregisterGroup(group1)
}
