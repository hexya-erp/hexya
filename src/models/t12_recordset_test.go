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

package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
)

func TestCreateRecordSet(t *testing.T) {
	t.Run("Test record creation", func(t *testing.T) {
		assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")
			tagModel := Registry.MustGet("Tag")
			postModel := Registry.MustGet("Post")
			commentModel := Registry.MustGet("Comment")
			t.Run("Creating simple user John with no relations and checking ID", func(t *testing.T) {
				userJohnData := NewModelData(userModel).
					Set(Name, "John Smith").
					Set(email, "jsmith@example.com").
					Set(isStaff, true).
					Set(nums, 1)
				users := env.Pool("User").Call("Create", userJohnData).(RecordSet).Collection()
				assert.EqualValues(t, users.Len(), 1)
				assert.Greater(t, users.Get(ID).(int64), int64(0))
				assert.False(t, users.Get(resume).(RecordSet).IsEmpty())
			})
			t.Run("Creating user Jane with related Profile and Posts and Tags and Comments", func(t *testing.T) {
				tag1 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Trending",
				})).(RecordSet).Collection()
				tag2 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Books",
				})).(RecordSet).Collection()
				tag3 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Jane's",
				})).(RecordSet).Collection()
				assert.EqualValues(t, tag1.Len(), 1)
				assert.EqualValues(t, tag2.Len(), 1)
				assert.EqualValues(t, tag3.Len(), 1)

				userJaneData := NewModelData(userModel).
					Set(Name, "Jane Smith").
					Set(email, "jane.smith@example.com").
					Set(nums, 2).
					Create(profile, NewModelData(profileModel).
						Set(age, 23).
						Set(money, 12345).
						Set(street, "165 5th Avenue").
						Set(city, "New York").
						Set(zip, "0305").
						Set(country, "USA")).
					Create(posts, NewModelData(postModel).
						Set(title, "1st Post").
						Set(content, "Content of first post").
						Set(tags, tag1.Union(tag3))).
					Create(posts, NewModelData(postModel).
						Set(title, "2nd Post").
						Set(content, "Content of second post"))
				userJane := env.Pool("User").Call("Create", userJaneData).(RecordSet).Collection()
				assert.EqualValues(t, userJane.Len(), 1)
				assert.NotEqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(ID), 0)
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(userName), "Jane Smith")

				post1 := env.Pool("Post").Search(postModel.Field(title).Equals("1st Post"))
				post2 := env.Pool("Post").Search(postModel.Field(title).Equals("2nd Post"))
				assert.EqualValues(t, post1.Len(), 1)
				assert.EqualValues(t, post2.Len(), 1)
				assert.EqualValues(t, post1.Get(user).(RecordSet).Collection().Get(ID), userJane.Get(ID))
				assert.EqualValues(t, post2.Get(user).(RecordSet).Collection().Get(ID), userJane.Get(ID))
				janePosts := userJane.Get(posts).(RecordSet).Collection()
				assert.EqualValues(t, janePosts.Len(), 2)

				userJane.Get(profile).(RecordSet).Collection().Set(bestPost, post1)

				assert.Empty(t, post2.Get(lastTagName))
				post2.Set(tags, tag2.Union(tag3))
				assert.EqualValues(t, post1.Get(lastTagName), "Jane's")
				post1Tags := post1.Get(tags).(RecordSet).Collection()
				assert.EqualValues(t, post1Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Trending", "Jane's"}, post1Tags.Records()[0].Get(Name))
				assert.Contains(t, []interface{}{"Trending", "Jane's"}, post1Tags.Records()[1].Get(Name))
				post2Tags := post2.Get(tags).(RecordSet).Collection()
				assert.EqualValues(t, post2Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[0].Get(Name))
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[1].Get(Name))

				assert.Empty(t, post1.Get(lastCommentText).(string))
				env.Pool("Comment").Call("Create", NewModelData(commentModel, FieldMap{
					"Post": post1,
					"Text": "First Comment",
				}))
				env.Pool("Comment").Call("Create", NewModelData(commentModel, FieldMap{
					"Post": post1,
					"Text": "Another Comment",
				}))
				env.Pool("Comment").Call("Create", NewModelData(commentModel, FieldMap{
					"Post": post1,
					"Text": "Third Comment",
				}))
				assert.EqualValues(t, post1.Get(lastCommentText).(string), "First Comment")
				assert.EqualValues(t, post1.Get(comments).(RecordSet).Len(), 3)
			})
			t.Run("Creating a user Will Smith", func(t *testing.T) {
				userWillData := NewModelData(userModel, FieldMap{
					"Name":    "Will Smith",
					"Email":   "will.smith@example.com",
					"IsStaff": true,
					"Nums":    3,
				})
				userWill := env.Pool("User").Call("Create", userWillData).(RecordSet).Collection()
				assert.EqualValues(t, userWill.Len(), 1)
				assert.Greater(t, userWill.Get(ID).(int64), int64(0))
			})
		}))
		t.Run("Checking constraint methods enforcement", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				tagModel := Registry.MustGet("Tag")
				tag1Data := NewModelData(tagModel, FieldMap{
					"Name":        "Tag1",
					"Description": "Tag1",
				})
				assert.Panics(t, func() { env.Pool("Tag").Call("Create", tag1Data) })
				tag2Data := NewModelData(tagModel, FieldMap{
					"Name": "Tag2",
					"Rate": 12,
				})
				assert.Panics(t, func() { env.Pool("Tag").Call("Create", tag2Data) })
				tag3Data := NewModelData(tagModel, FieldMap{
					"Name":        "Tag2",
					"Description": "Tag2",
					"Rate":        -3,
				})
				assert.Panics(t, func() { env.Pool("Tag").Call("Create", tag3Data) })
			}))
		})
		t.Run("Checking that we can't create two users with the same name", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				user1Data := NewModelData(userModel, FieldMap{
					"Name": "User1",
				})
				assert.NotPanics(t, func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() })
				assert.Panics(t, func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() })
			}))
		})
		t.Run("Checking that we can't create two users with a empty string name", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				user1Data := NewModelData(userModel, FieldMap{
					"Name": "",
				})
				assert.NotPanics(t, func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() })
				assert.Panics(t, func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() })
			}))
		})
		t.Run("Checking that we can create as many users with a NULL name", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userModel := Registry.MustGet("User")
				user2Data := NewModelData(userModel, FieldMap{
					"Email": "user2@example.com",
				})
				assert.NotPanics(t, func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() })
				assert.NotPanics(t, func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() })
				assert.NotPanics(t, func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() })
			}))
		})
	})
	t.Run("Checking SQL Constraint enforcement", func(t *testing.T) {
		err := SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userRobData := NewModelData(userModel, FieldMap{
				"Name":      "Rob Smith",
				"IsPremium": true,
			})
			env.Pool("User").Call("Create", userRobData)
		})
		assert.NotNil(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "pq: Premium users must have positive nums"))
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	t.Run("Testing access control list on creation (create only)", func(t *testing.T) {
		t.Run("Checking that user 2 cannot create records", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				assert.Panics(t, func() { env.Pool("User").Call("Create", userTomData) })
			}))
		})
		t.Run("Adding model access rights to user 2 and check failure again", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				resumeModel := Registry.MustGet("Resume")
				userModel.methods.MustGet("Create").AllowGroup(group1)
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Write"))
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				assert.Panics(t, func() { env.Pool("User").Call("Create", userTomData) })
			}))
		})
		t.Run("Adding model access rights to user 2 for resume and it works", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				resumeModel := Registry.MustGet("Resume")
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				resumeModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("Create"))
				updateContextModelsSecurity()
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				assert.Panics(t, func() { userTom.Get(Name) })
			}))
		})
		t.Run("Revoking model access rights to user 2 for resume and it doesn't works", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				resumeModel := Registry.MustGet("Resume")
				resumeModel.methods.MustGet("Create").RevokeGroup(group1)
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				assert.Panics(t, func() { env.Pool("User").Call("Create", userTomData) })
			}))
		})
		t.Run("Regranting model access rights to user 2 for posts and it works", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				resumeModel := Registry.MustGet("Resume")
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				assert.Panics(t, func() { userTom.Get(Name) })
			}))
		})
		t.Run("Checking creation again with read rights too", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				userModel := Registry.MustGet("User")
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				assert.EqualValues(t, userTom.Get(Name), "Tom Smith")
				assert.EqualValues(t, userTom.Get(email), "tsmith@example.com")
			}))
		})
		t.Run("Checking that we can create tags", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
				security.Registry.AddMembership(2, group1)
				tagModel := Registry.MustGet("Tag")
				tagData := NewModelData(tagModel, FieldMap{
					"Name": "My Tag",
				})
				env.Pool("Tag").Call("Create", tagData)
				assert.NotPanics(t, func() {})
			}))
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	t.Run("Testing search through RecordSets", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Searching User Jane", func(t *testing.T) {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				assert.EqualValues(t, userJane.Len(), 1)
				t.Run("Reading Jane with Get", func(t *testing.T) {
					assert.EqualValues(t, userJane.Get(Name).(string), "Jane Smith")
					assert.EqualValues(t, userJane.Get(email), "jane.smith@example.com")
					assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(age), 23)
					assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(money), 12345)
					assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(country), "USA")
					assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(zip), "0305")
					recs := userJane.Get(posts).(RecordSet).Collection().Records()
					assert.Len(t, recs, 2)
					assert.EqualValues(t, recs[0].Get(title), "1st Post")
					assert.EqualValues(t, recs[1].Get(title), "2nd Post")
				})
				t.Run("Reading Jane with ReadFirst", func(t *testing.T) {
					ujData := userJane.First()
					assert.EqualValues(t, ujData.Get(Name), "Jane Smith")
					assert.True(t, ujData.Has(Name))
					assert.EqualValues(t, ujData.Get(email), "jane.smith@example.com")
					assert.True(t, ujData.Has(email))
					assert.EqualValues(t, ujData.Get(ID), userJane.Get(ID).(int64))
					assert.True(t, ujData.Has(ID))
					assert.EqualValues(t, ujData.Get(profile).(RecordSet).Collection().Get(ID), userJane.Get(profile).(RecordSet).Collection().Get(ID))
					assert.True(t, ujData.Has(profile))
				})
				t.Run("Reading an empty RecordSet should return zero value", func(t *testing.T) {
					empty := env.Pool("User")
					assert.EqualValues(t, empty.Get(Name), "")
				})
				t.Run("Reading an invalid RecordSet should return zero value", func(t *testing.T) {
					empty := &RecordCollection{model: Registry.MustGet("User")}
					assert.EqualValues(t, empty.Get(Name), "")
				})
			})

			t.Run("Testing search all users", func(t *testing.T) {
				usersAll := env.Pool("User").Call("SearchAll").(RecordSet).Collection()
				assert.EqualValues(t, usersAll.Len(), 3)
				usersAll = env.Pool("User").OrderBy("Name")
				assert.EqualValues(t, usersAll.Len(), 3)
				t.Run("Reading first user with Get", func(t *testing.T) {
					assert.EqualValues(t, usersAll.Get(Name), "Jane Smith")
					assert.EqualValues(t, usersAll.Get(email), "jane.smith@example.com")
				})
				t.Run("Reading all users with Records and Get", func(t *testing.T) {
					recs := usersAll.Records()
					assert.EqualValues(t, len(recs), 3)
					assert.EqualValues(t, recs[0].Get(email), "jane.smith@example.com")
					assert.EqualValues(t, recs[1].Get(email), "jsmith@example.com")
					assert.EqualValues(t, recs[2].Get(email), "will.smith@example.com")
				})
				t.Run("Reading all users with ReadAll()", func(t *testing.T) {
					usersData := usersAll.All()
					assert.EqualValues(t, usersData[0].Get(email), "jane.smith@example.com")
					assert.True(t, usersData[0].Has(email))
					assert.EqualValues(t, usersData[1].Get(email), "jsmith@example.com")
					assert.True(t, usersData[1].Has(email))
					assert.EqualValues(t, usersData[2].Get(email), "will.smith@example.com")
					assert.True(t, usersData[2].Has(email))
				})
			})
			t.Run("Testing search on manual model", func(t *testing.T) {
				userViews := env.Pool("UserView").SearchAll()
				assert.EqualValues(t, userViews.Len(), 3)
				userViews = env.Pool("UserView").OrderBy("Name")
				assert.EqualValues(t, userViews.Len(), 3)
				recs := userViews.Records()
				assert.EqualValues(t, len(recs), 3)
				assert.EqualValues(t, recs[0].Get(Name), "Jane Smith")
				assert.EqualValues(t, recs[1].Get(Name), "John Smith")
				assert.EqualValues(t, recs[2].Get(Name), "Will Smith")
				assert.EqualValues(t, recs[0].Get(city), "New York")
				assert.EqualValues(t, recs[1].Get(city), "")
				assert.EqualValues(t, recs[2].Get(city), "")
			})
			t.Run("Testing browse with empty ids", func(t *testing.T) {
				var ids []int64
				users := env.Pool("User").Model().Browse(env, ids)
				assert.EqualValues(t, users.Len(), 0)
			})
		}))
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	t.Run("Testing access control list while searching", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			t.Run("Checking that user 2 cannot access records", func(t *testing.T) {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				assert.Panics(t, func() { userJane.Load() })
			})
			t.Run("Adding model access rights to user 2 and checking access", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				assert.NotPanics(t, func() { userJane.Load() })
				assert.EqualValues(t, userJane.Get(Name).(string), "Jane Smith")
				assert.EqualValues(t, userJane.Get(email).(string), "jane.smith@example.com")
				assert.EqualValues(t, userJane.Get(age), 23)
				assert.Panics(t, func() { userJane.Get(profile).(RecordSet).Collection().Get(age) })
			})
			t.Run("Revoking model access rights to user 2 and checking access", func(t *testing.T) {
				userModel.methods.MustGet("Load").RevokeGroup(group1)
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.Panics(t, func() { userJohn.Load() })
			})
			t.Run("Regranting model access rights to user 2 and checking access", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				assert.NotPanics(t, func() { userJane.Load() })
				assert.Panics(t, func() { userJane.Get(profile).(RecordSet).Collection().Get(age) })
			})
			t.Run("Checking record rules", func(t *testing.T) {
				users := env.Pool("User").SearchAll()
				assert.EqualValues(t, users.Len(), 3)

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: users.Model().Field(Name).IContains("j"),
					Perms:     security.Read,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: users.Model().Field(Name).Equals("Nobody"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&notUsedRule)

				users = env.Pool("User").SearchAll()
				assert.EqualValues(t, users.Len(), 2)
				assert.Contains(t, []string{"Jane Smith", "John Smith"}, users.Records()[0].Get(Name))
				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		}))
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	t.Run("Testing advanced queries on M2O relations", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
			assert.EqualValues(t, jane.Len(), 1)
			t.Run("Condition on m2o relation fields with ids", func(t *testing.T) {
				profileID := jane.Get(profile).(RecordSet).Collection().Get(ID).(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(profileID))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Condition on m2o relation fields with recordset", func(t *testing.T) {
				janeProfile := jane.Get(profile).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(janeProfile))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Empty recordset", func(t *testing.T) {
				emptyProfile := env.Pool("Profile")
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(emptyProfile))
				assert.EqualValues(t, users.Len(), 2)
			})
			t.Run("Empty recordset with IsNull", func(t *testing.T) {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).IsNull())
				assert.EqualValues(t, users.Len(), 2)
			})
			t.Run("Condition on m2o relation fields with IN operator and ids", func(t *testing.T) {
				profileID := jane.Get(profile).(RecordSet).Collection().Get(ID).(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).In(profileID))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Condition on m2o relation fields with IN operator and recordset", func(t *testing.T) {
				janeProfile := jane.Get(profile).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).In(janeProfile))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Empty recordset with IN operator", func(t *testing.T) {
				emptyProfile := env.Pool("Profile")
				users := env.Pool("User").Search(
					env.Pool("User").Model().Field(profile).In(emptyProfile))
				assert.EqualValues(t, users.Len(), 0)
				users = env.Pool("User").Search(
					env.Pool("User").Model().Field(profile).In(emptyProfile).
						And().Field(isStaff).Equals(false))
				assert.EqualValues(t, users.Len(), 0)
			})
			t.Run("M2O chain", func(t *testing.T) {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profileBestPostTitle).Equals("1st Post"))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
		}))
	})
	t.Run("Testing advanced queries on O2M relations", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
			assert.EqualValues(t, jane.Len(), 1)
			t.Run("Condition on o2m relation with slice of ids", func(t *testing.T) {
				postID := jane.Get(posts).(RecordSet).Collection().Ids()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).Equals(postID))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Conditions on o2m relation with recordset", func(t *testing.T) {
				post := jane.Get(posts).(RecordSet).Collection().Records()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).Equals(post))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Conditions on o2m relation with null", func(t *testing.T) {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).IsNull())
				assert.EqualValues(t, users.Len(), 2)
				userRecs := users.Records()
				assert.EqualValues(t, userRecs[0].Get(Name), "John Smith")
				assert.EqualValues(t, userRecs[1].Get(Name), "Will Smith")
			})
			t.Run("Condition on o2m relation with IN operator and slice of ids", func(t *testing.T) {
				postIds := jane.Get(posts).(RecordSet).Collection().Ids()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).In(postIds))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("Conditions on o2m relation with IN operator and recordset", func(t *testing.T) {
				janePosts := jane.Get(posts).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).In(janePosts))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
			t.Run("O2M Chain", func(t *testing.T) {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(postsTitle).Equals("1st Post"))
				assert.EqualValues(t, users.Len(), 1)
				assert.EqualValues(t, users.Get(ID).(int64), jane.Get(ID).(int64))
			})
		}))
	})
	t.Run("Testing advanced queries on M2M relations", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			post1 := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Equals("1st Post"))
			assert.EqualValues(t, post1.Len(), 1)
			post2 := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Equals("2nd Post"))
			assert.EqualValues(t, post2.Len(), 1)
			tag1 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Trending"))
			tag2 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Books"))
			assert.EqualValues(t, tag1.Len(), 1)
			t.Run("Condition on m2m relation with slice of ids", func(t *testing.T) {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).Equals(tag1.Get(ID)))
				assert.EqualValues(t, rPosts.Len(), 1)
				assert.EqualValues(t, rPosts.Get(ID).(int64), post1.Get(ID).(int64))
			})
			t.Run("Condition on m2m relation with recordset", func(t *testing.T) {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).Equals(tag1))
				assert.EqualValues(t, rPosts.Len(), 1)
				assert.EqualValues(t, rPosts.Get(ID).(int64), post1.Get(ID).(int64))
			})
			t.Run("Condition on m2m relation with null", func(t *testing.T) {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).IsNull())
				assert.EqualValues(t, rPosts.Len(), 0)
			})
			t.Run("Condition on m2m relation with IN operator and ids", func(t *testing.T) {
				tags12 := tag1.Union(tag2)
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tags12.Ids()))
				assert.EqualValues(t, rPosts.Len(), 2)
			})
			t.Run("Condition on m2m relation with IN operator and empty ids", func(t *testing.T) {
				var tagIds []int64
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tagIds))
				assert.EqualValues(t, rPosts.Len(), 0)
			})
			t.Run("Condition on m2m relation with IN operator and recordset", func(t *testing.T) {
				tags12 := tag1.Union(tag2)
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tags12))
				assert.EqualValues(t, rPosts.Len(), 2)
			})
			t.Run("Condition on m2m relation with IN operator and empty recordset", func(t *testing.T) {
				emptyTags := env.Pool("Tag")
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(emptyTags))
				assert.EqualValues(t, rPosts.Len(), 0)
			})
			t.Run("M2M Chain", func(t *testing.T) {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tagsName).Equals("Trending"))
				assert.EqualValues(t, rPosts.Len(), 1)
				assert.EqualValues(t, rPosts.Get(ID).(int64), post1.Get(ID).(int64))
			})
		}))
	})
	t.Run("Testing advanced queries with multiple joins", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			jane := users.Search(users.Model().Field(Name).Equals("Jane Smith"))
			john := users.Search(users.Model().Field(Name).Equals("John Smith"))
			assert.EqualValues(t, jane.Len(), 1)
			assert.EqualValues(t, john.Len(), 1)
			t.Run("Testing M2O-M2O-M2O", func(t *testing.T) {
				assert.True(t, jane.Get(profileBestPostUser).(RecordSet).Collection().Equals(jane))
				assert.True(t, john.Get(profileBestPostUser).(RecordSet).Collection().IsEmpty())
				johnProfile := env.Pool("Profile").Call("Create", NewModelData(Registry.MustGet("Profile")))
				john.Set(profile, johnProfile)
				assert.Empty(t, john.Get(profileBestPostUser).(RecordSet).Collection().Get(email))
			})
		}))
	})
}

func TestGroupedQueries(t *testing.T) {
	t.Run("Testing grouped queries", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Simple grouped query on the whole table", func(t *testing.T) {
				groupedUsers := env.Pool("User").SearchAll().Call("GroupBy", []FieldName{isStaff}).(RecordSet).Collection().Aggregates(isStaff, nums)
				assert.EqualValues(t, len(groupedUsers), 2)
				assert.True(t, groupedUsers[0].Values.Has(isStaff))
				assert.Equal(t, false, groupedUsers[0].Values.Get(isStaff))
				assert.True(t, groupedUsers[0].Values.Has(nums))
				assert.EqualValues(t, groupedUsers[0].Values.Get(nums), 2)
				assert.EqualValues(t, groupedUsers[0].Count, 1)
				assert.True(t, groupedUsers[1].Values.Has(isStaff))
				assert.Equal(t, true, groupedUsers[1].Values.Get(isStaff))
				assert.True(t, groupedUsers[1].Values.Has(nums))
				assert.EqualValues(t, groupedUsers[1].Values.Get(nums), 4)
				assert.EqualValues(t, groupedUsers[1].Count, 2)
			})
		}))
	})
}

func TestUpdateRecordSet(t *testing.T) {
	t.Run("Testing updates through RecordSets", func(t *testing.T) {
		assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			postModel := Registry.MustGet("Post")
			tagModel := Registry.MustGet("Tag")
			t.Run("Update on users Jane and John with Write and Set", func(t *testing.T) {
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				jane.Set(Name, "Jane A. Smith")
				jane.Load()
				assert.EqualValues(t, jane.Get(Name), "Jane A. Smith")
				assert.EqualValues(t, jane.Get(email), "jane.smith@example.com")

				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := NewModelData(userModel).
					Set(email, "jsmith2@example.com").
					Set(nums, 13).
					Set(isStaff, false)
				john.Call("Write", johnValues)
				john.Load()
				assert.EqualValues(t, john.Get(Name), "John Smith")
				assert.EqualValues(t, john.Get(email), "jsmith2@example.com")
				assert.EqualValues(t, john.Get(nums), 13)
				assert.Equal(t, false, john.Get(isStaff))
				john.Set(isStaff, true)
				assert.Equal(t, true, john.Get(isStaff))
			})
			t.Run("Updating an empty RecordSet should do nothing", func(t *testing.T) {
				empty := env.Pool("User")
				assert.NotPanics(t, func() { empty.Set(Name, "Foo") })
				assert.NotPanics(t, func() {
					empty.Call("Write", NewModelData(userModel).
						Set(Name, "Bar"))
				})
			})
			t.Run("Multiple updates at once on users", func(t *testing.T) {
				cond := env.Pool("User").Model().Field(Name).Equals("Jane A. Smith").Or().Field(Name).Equals("John Smith")
				users := env.Pool("User").Search(cond).Load()
				assert.EqualValues(t, users.Len(), 2)
				userRecs := users.Records()
				assert.True(t, userRecs[0].Get(isStaff).(bool))
				assert.False(t, userRecs[1].Get(isStaff).(bool))
				assert.False(t, userRecs[0].Get(isActive).(bool))
				assert.False(t, userRecs[1].Get(isActive).(bool))

				users.Set(isStaff, true)
				users.Load()
				assert.True(t, userRecs[0].Get(isStaff).(bool))
				assert.True(t, userRecs[1].Get(isStaff).(bool))

				data := NewModelData(userModel).
					Set(isStaff, false).
					Set(isActive, true)
				users.Call("Write", data)
				users.Load()
				assert.False(t, userRecs[0].Get(isStaff).(bool))
				assert.False(t, userRecs[1].Get(isStaff).(bool))
				assert.True(t, userRecs[0].Get(isActive).(bool))
				assert.True(t, userRecs[1].Get(isActive).(bool))
			})
			t.Run("Updating many2one fields", func(t *testing.T) {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				userJane.Set(profile, nil)
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(ID), 0)
				userJane.Set(profile, janeProfile.Get(ID))
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(ID), janeProfile.ids[0])
				userJane.Set(profile, env.Pool("Profile"))
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(ID), 0)
				userJane.Set(profile, janeProfile)
				assert.EqualValues(t, userJane.Get(profile).(RecordSet).Collection().Get(ID), janeProfile.ids[0])

				post1 := janeProfile.Get(bestPost)
				janeProfile.Call("Write", NewModelData(janeProfile.model).
					Create(bestPost, NewModelData(postModel).
						Set(title, "Post created on the Fly")))
				assert.EqualValues(t, janeProfile.Get(bestPost).(RecordSet).Collection().Get(title), "Post created on the Fly")
				janeProfile.Set(bestPost, post1)
			})
			t.Run("Updating many2many fields", func(t *testing.T) {
				emptyPosts := env.Pool("Post")
				post1 := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("1st Post"))
				post1.Call("Write", NewModelData(postModel).
					Create(tags, NewModelData(tagModel).
						Set(Name, "Tag created on the fly")).
					Create(tags, NewModelData(tagModel).
						Set(Name, "Second Tag on the fly")))
				post1Tags := post1.Get(tags).(RecordSet).Collection()
				assert.EqualValues(t, post1Tags.Len(), 2)
				assert.Contains(t, []string{"Tag created on the fly", "Second Tag on the fly"}, post1Tags.Records()[0].Get(Name))
				assert.Contains(t, []string{"Tag created on the fly", "Second Tag on the fly"}, post1Tags.Records()[1].Get(Name))

				tagBooks := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Books"))
				post1.Set(tags, tagBooks)
				post1Tags = post1.Get(tags).(RecordSet).Collection()
				assert.EqualValues(t, post1Tags.Len(), 1)
				assert.EqualValues(t, post1Tags.Get(Name), "Books")

				post2Tags := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("2nd Post")).Get(tags).(RecordSet).Collection()
				assert.EqualValues(t, post2Tags.Len(), 2)
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[0].Get(Name))
				assert.Contains(t, []interface{}{"Books", "Jane's"}, post2Tags.Records()[1].Get(Name))
			})
			t.Run("Updating One2many fields", func(t *testing.T) {
				mPosts := env.Pool("Post")
				post1 := mPosts.Search(mPosts.Model().Field(title).Equals("1st Post"))
				post2 := mPosts.Search(mPosts.Model().Field(title).Equals("2nd Post"))
				post3 := mPosts.Call("Create", NewModelData(postModel, FieldMap{
					"Title":   "3rd Post",
					"Content": "Content of third post",
				})).(RecordSet).Collection()
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Set(posts, post1.Call("Union", post3).(RecordSet).Collection())
				assert.EqualValues(t, post1.Get(user).(RecordSet).Collection().Get(ID), userJane.Get(ID))
				assert.EqualValues(t, post3.Get(user).(RecordSet).Collection().Get(ID), userJane.Get(ID))
				assert.EqualValues(t, post2.Get(user).(RecordSet).Collection().Get(ID), 0)

				userJane.Set(posts, nil)
				userJane.Call("Write", NewModelData(userModel).
					Create(posts, NewModelData(postModel).
						Set(title, "Another post created on the fly")).
					Create(posts, NewModelData(postModel).
						Set(title, "One more post created on the fly")))
				assert.EqualValues(t, userJane.Get(posts).(RecordSet).Len(), 2)
				assert.Contains(t, []string{"Another post created on the fly", "One more post created on the fly"}, userJane.Get(posts).(RecordSet).Collection().Records()[0].Get(title))
				assert.Contains(t, []string{"Another post created on the fly", "One more post created on the fly"}, userJane.Get(posts).(RecordSet).Collection().Records()[1].Get(title))

				userJane.Set(posts, post1.Call("Union", post3).(RecordSet).Collection())
				assert.EqualValues(t, userJane.Get(posts).(RecordSet).Len(), 2)

			})
		}))
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			t.Run("Checking constraint methods enforcement", func(t *testing.T) {
				tag1 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field(Name).Equals("Trending"))
				tag1.Load()
				assert.Panics(t, func() { tag1.Set(description, "Trending") })
				tag2 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field(Name).Equals("Books"))
				assert.Panics(t, func() { tag2.Set(rate, 12) })
				assert.Panics(t, func() {
					tag2.Call("Write", FieldMap{
						"Description": "Books",
						"Rate":        -3,
					})
				})
			})
		}))
	})
	t.Run("Checking SQL Constraint enforcement", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userWill := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("will.smith@example.com"))
			userWill.Call("Write", NewModelData(userModel).Set(nums, 0).Set(isPremium, true))
		}).Error(), "pq: Premium users must have positive nums"))
	})

	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	t.Run("Testing access control list on update (write only)", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")

			t.Run("Checking that user 2 cannot update records", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := FieldMap{
					"Email": "jsmith3@example.com",
					"Nums":  13,
				}
				assert.Panics(t, func() { john.Call("Write", johnValues) })
			})
			t.Run("Adding model access rights to user 2 and check update", func(t *testing.T) {
				userModel.methods.MustGet("Write").AllowGroup(group1)
				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.EqualValues(t, john.Len(), 1)
				johnValues := NewModelData(userModel).
					Set(email, "jsmith3@example.com").
					Set(nums, 13)
				john.Call("Write", johnValues)
				john.Load()
				assert.EqualValues(t, john.Get(Name), "John Smith")
				assert.EqualValues(t, john.Get(email), "jsmith3@example.com")
				assert.EqualValues(t, john.Get(nums), 13)
			})
			t.Run("Checking that user 2 cannot update profile through UpdateCity method", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userModel.methods.MustGet("UpdateCity").AllowGroup(group1)
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane A. Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				assert.Panics(t, func() { jane.Call("UpdateCity", "London") })
			})
			t.Run("Checking that user 2 can run UpdateCity after giving permission for caller", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				profileModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("UpdateCity"))
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane A. Smith"))
				assert.EqualValues(t, jane.Len(), 1)
				assert.NotPanics(t, func() { jane.Call("UpdateCity", "London") })
			})
			t.Run("Checking record rules", func(t *testing.T) {
				userJane := env.Pool("User").SearchAll()
				assert.EqualValues(t, userJane.Len(), 3)

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: env.Pool("User").Model().Field(Name).IContains("j"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "unlinkRule",
					Group:     group1,
					Condition: env.Pool("User").Model().Field(Name).Equals("Nobody"),
					Perms:     security.Unlink,
				}
				userModel.AddRecordRule(&notUsedRule)

				userJane = env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Len(), 1)
				assert.EqualValues(t, userJane.Get(Name), "Jane A. Smith")
				userJane.Set(Name, "Jane B. Smith")
				assert.EqualValues(t, userJane.Get(Name), "Jane B. Smith")

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Will Smith"))
				assert.Panics(t, func() { userWill.Set(Name, "Will Jr. Smith") })

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("unlinkRule")
			})
		}))
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	t.Run("Checking unlink method", func(t *testing.T) {
		t.Run("Deleting user John: number of deleted record should be 1", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				num := userJohn.Call("Unlink")
				assert.EqualValues(t, num, 1)
			}))
		})
		t.Run("Deleted RecordSet should update themselves when reloading", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				userJohn2 := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith").Or().Field(Name).Equals("Jane A. Smith"))
				assert.EqualValues(t, userJohn.Len(), 1)
				assert.EqualValues(t, userJohn2.Len(), 1)
				assert.EqualValues(t, users.Len(), 2)
				userJohn.Call("Unlink")
				userJohn.ForceLoad()
				assert.EqualValues(t, userJohn.Len(), 0)
				userJohn2.ForceLoad()
				assert.EqualValues(t, userJohn2.Len(), 0)
				users.ForceLoad()
				assert.EqualValues(t, users.Len(), 1)
			}))
		})
		t.Run("Deleted RecordSet should update themselves when reloading with prefetch", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User").SearchAll()
				assert.EqualValues(t, users.Len(), 3)
				assert.EqualValues(t, users.Records()[0].Get(Name), "John Smith")
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				userJohn2 := users.Records()[0]
				users.Records()[0].Call("Unlink")
				userJohn2.ForceLoad()
				assert.True(t, userJohn2.IsEmpty())
				users.ForceLoad()
				assert.EqualValues(t, users.Len(), 2)
				assert.EqualValues(t, users.Records()[0].Get(Name), "Jane A. Smith")
				assert.EqualValues(t, users.Records()[1].Get(Name), "Will Smith")
				userJohn.ForceLoad()
				assert.EqualValues(t, userJohn.Len(), 0)
			}))
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	t.Run("Checking unlink access permissions", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")
			postModel := Registry.MustGet("Post")

			t.Run("Checking that user 2 cannot unlink records", func(t *testing.T) {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.Panics(t, func() { users.Call("Unlink") })
			})
			t.Run("Adding unlink permission to user2", func(t *testing.T) {
				userModel.methods.MustGet("Unlink").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				assert.Panics(t, func() { users.Call("Unlink") })
			})
			t.Run("Adding permissions to user2 on Profile and Post", func(t *testing.T) {
				profileModel.methods.MustGet("Load").AllowGroup(group1)
				postModel.methods.MustGet("Load").AllowGroup(group1)
				postModel.methods.MustGet("Write").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				num := users.Call("Unlink")
				assert.EqualValues(t, num, 1)
			})
			t.Run("Checking record rules", func(t *testing.T) {

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: env.Pool("User").Model().Field(Name).IContains("j"),
					Perms:     security.Unlink,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: env.Pool("User").Model().Field(Name).Equals("Nobody"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&notUsedRule)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, userJane.Len(), 1)
				assert.EqualValues(t, userJane.Call("Unlink"), 1)

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Will Smith"))
				assert.EqualValues(t, userWill.Call("Unlink"), 0)

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		}))
	})
	security.Registry.UnregisterGroup(group1)
}
