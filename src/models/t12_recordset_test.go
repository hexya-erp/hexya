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

	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")
			tagModel := Registry.MustGet("Tag")
			postModel := Registry.MustGet("Post")
			commentModel := Registry.MustGet("Comment")
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := NewModelData(userModel).
					Set(Name, "John Smith").
					Set(email, "jsmith@example.com").
					Set(isStaff, true).
					Set(nums, 1)
				users := env.Pool("User").Call("Create", userJohnData).(RecordSet).Collection()
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID), ShouldBeGreaterThan, 0)
				So(users.Get(resume).(RecordSet).IsEmpty(), ShouldBeFalse)
			})
			Convey("Creating user Jane with related Profile and Posts and Tags and Comments", func() {
				tag1 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Trending",
				})).(RecordSet).Collection()
				tag2 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Books",
				})).(RecordSet).Collection()
				tag3 := env.Pool("Tag").Call("Create", NewModelData(tagModel, FieldMap{
					"Name": "Jane's",
				})).(RecordSet).Collection()
				So(tag1.Len(), ShouldEqual, 1)
				So(tag2.Len(), ShouldEqual, 1)
				So(tag3.Len(), ShouldEqual, 1)

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
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Get(profile).(RecordSet).Collection().Get(ID), ShouldNotEqual, 0)
				So(userJane.Get(profile).(RecordSet).Collection().Get(userName), ShouldEqual, "Jane Smith")

				post1 := env.Pool("Post").Search(postModel.Field(title).Equals("1st Post"))
				post2 := env.Pool("Post").Search(postModel.Field(title).Equals("2nd Post"))
				So(post1.Len(), ShouldEqual, 1)
				So(post2.Len(), ShouldEqual, 1)
				So(post1.Get(user).(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(ID))
				So(post2.Get(user).(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(ID))
				janePosts := userJane.Get(posts).(RecordSet).Collection()
				So(janePosts.Len(), ShouldEqual, 2)

				userJane.Get(profile).(RecordSet).Collection().Set(bestPost, post1)

				So(post2.Get(lastTagName), ShouldBeBlank)
				post2.Set(tags, tag2.Union(tag3))
				So(post1.Get(lastTagName), ShouldEqual, "Jane's")
				post1Tags := post1.Get(tags).(RecordSet).Collection()
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Get(Name), ShouldBeIn, "Trending", "Jane's")
				So(post1Tags.Records()[1].Get(Name), ShouldBeIn, "Trending", "Jane's")
				post2Tags := post2.Get(tags).(RecordSet).Collection()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Get(Name), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Get(Name), ShouldBeIn, "Books", "Jane's")

				So(post1.Get(lastCommentText).(string), ShouldBeBlank)
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
				So(post1.Get(lastCommentText).(string), ShouldEqual, "First Comment")
				So(post1.Get(comments).(RecordSet).Len(), ShouldEqual, 3)
			})
			Convey("Creating a user Will Smith", func() {
				userWillData := NewModelData(userModel, FieldMap{
					"Name":    "Will Smith",
					"Email":   "will.smith@example.com",
					"IsStaff": true,
					"Nums":    3,
				})
				userWill := env.Pool("User").Call("Create", userWillData).(RecordSet).Collection()
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.Get(ID), ShouldBeGreaterThan, 0)
			})
		}), ShouldBeNil)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			tagModel := Registry.MustGet("Tag")
			userModel := Registry.MustGet("User")
			Convey("Checking constraint methods enforcement", func() {
				tag1Data := NewModelData(tagModel, FieldMap{
					"Name":        "Tag1",
					"Description": "Tag1",
				})
				So(func() { env.Pool("Tag").Call("Create", tag1Data) }, ShouldPanic)
				tag2Data := NewModelData(tagModel, FieldMap{
					"Name": "Tag2",
					"Rate": 12,
				})
				So(func() { env.Pool("Tag").Call("Create", tag2Data) }, ShouldPanic)
				tag3Data := NewModelData(tagModel, FieldMap{
					"Name":        "Tag2",
					"Description": "Tag2",
					"Rate":        -3,
				})
				So(func() { env.Pool("Tag").Call("Create", tag3Data) }, ShouldPanic)
			})
			Convey("Checking that we can't create two users with the same name", func() {
				user1Data := NewModelData(userModel, FieldMap{
					"Name": "User1",
				})
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldPanic)
			})
			Convey("Checking that we can't create two users with a empty string name", func() {
				user1Data := NewModelData(userModel, FieldMap{
					"Name": "",
				})
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldPanic)
			})
			Convey("Checking that we can create as many users with a NULL name", func() {
				user2Data := NewModelData(userModel, FieldMap{
					"Email": "user2@example.com",
				})
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
			})
		}), ShouldBeNil)
	})
	Convey("Checking SQL Constraint enforcement", t, func() {
		err := SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userRobData := NewModelData(userModel, FieldMap{
				"Name":      "Rob Smith",
				"IsPremium": true,
			})
			env.Pool("User").Call("Create", userRobData)
		})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldStartWith, "pq: Premium users must have positive nums")
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on creation (create only)", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			security.Registry.AddMembership(2, group1)
			userModel := Registry.MustGet("User")
			tagModel := Registry.MustGet("Tag")
			resumeModel := Registry.MustGet("Resume")

			Convey("Checking that user 2 cannot create records", func() {
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				userModel.methods.MustGet("Create").AllowGroup(group1)
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Write"))
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for resume and it works", func() {
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				resumeModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("Create"))
				updateContextModelsSecurity()
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(func() { userTom.Get(Name) }, ShouldPanic)
			})
			Convey("Revoking model access rights to user 2 for resume and it doesn't works", func() {
				resumeModel.methods.MustGet("Create").RevokeGroup(group1)
				userTomData := NewModelData(userModel, FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				})
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Regranting model access rights to user 2 for posts and it works", func() {
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(func() { userTom.Get(Name) }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userTomData := NewModelData(userModel, FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				})
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(userTom.Get(Name), ShouldEqual, "Tom Smith")
				So(userTom.Get(email), ShouldEqual, "tsmith@example.com")
			})
			Convey("Checking that we can create tags", func() {
				tagData := NewModelData(tagModel, FieldMap{
					"Name": "My Tag",
				})
				env.Pool("Tag").Call("Create", tagData)
				So(func() {}, ShouldNotPanic)
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Searching User Jane", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				So(userJane.Len(), ShouldEqual, 1)
				Convey("Reading Jane with Get", func() {
					So(userJane.Get(Name).(string), ShouldEqual, "Jane Smith")
					So(userJane.Get(email), ShouldEqual, "jane.smith@example.com")
					So(userJane.Get(profile).(RecordSet).Collection().Get(age), ShouldEqual, 23)
					So(userJane.Get(profile).(RecordSet).Collection().Get(money), ShouldEqual, 12345)
					So(userJane.Get(profile).(RecordSet).Collection().Get(country), ShouldEqual, "USA")
					So(userJane.Get(profile).(RecordSet).Collection().Get(zip), ShouldEqual, "0305")
					recs := userJane.Get(posts).(RecordSet).Collection().Records()
					So(recs, ShouldHaveLength, 2)
					So(recs[0].Get(title), ShouldEqual, "1st Post")
					So(recs[1].Get(title), ShouldEqual, "2nd Post")
				})
				Convey("Reading Jane with ReadFirst", func() {
					ujData := userJane.First()
					So(ujData.Get(Name), ShouldEqual, "Jane Smith")
					So(ujData.Has(Name), ShouldBeTrue)
					So(ujData.Get(email), ShouldEqual, "jane.smith@example.com")
					So(ujData.Has(email), ShouldBeTrue)
					So(ujData.Get(ID), ShouldEqual, userJane.Get(ID).(int64))
					So(ujData.Has(ID), ShouldBeTrue)
					So(ujData.Get(profile).(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(profile).(RecordSet).Collection().Get(ID))
					So(ujData.Has(profile), ShouldBeTrue)
				})
				Convey("Reading an empty RecordSet should return zero value", func() {
					empty := env.Pool("User")
					So(empty.Get(Name), ShouldEqual, "")
				})
				Convey("Reading an invalid RecordSet should return zero value", func() {
					empty := &RecordCollection{model: Registry.MustGet("User")}
					So(empty.Get(Name), ShouldEqual, "")
				})
			})

			Convey("Testing search all users", func() {
				usersAll := env.Pool("User").Call("SearchAll").(RecordSet).Collection()
				So(usersAll.Len(), ShouldEqual, 3)
				usersAll = env.Pool("User").OrderBy("Name")
				So(usersAll.Len(), ShouldEqual, 3)
				Convey("Reading first user with Get", func() {
					So(usersAll.Get(Name), ShouldEqual, "Jane Smith")
					So(usersAll.Get(email), ShouldEqual, "jane.smith@example.com")
				})
				Convey("Reading all users with Records and Get", func() {
					recs := usersAll.Records()
					So(len(recs), ShouldEqual, 3)
					So(recs[0].Get(email), ShouldEqual, "jane.smith@example.com")
					So(recs[1].Get(email), ShouldEqual, "jsmith@example.com")
					So(recs[2].Get(email), ShouldEqual, "will.smith@example.com")
				})
				Convey("Reading all users with ReadAll()", func() {
					usersData := usersAll.All()
					So(usersData[0].Get(email), ShouldEqual, "jane.smith@example.com")
					So(usersData[0].Has(email), ShouldBeTrue)
					So(usersData[1].Get(email), ShouldEqual, "jsmith@example.com")
					So(usersData[1].Has(email), ShouldBeTrue)
					So(usersData[2].Get(email), ShouldEqual, "will.smith@example.com")
					So(usersData[2].Has(email), ShouldBeTrue)
				})
			})
			Convey("Testing search on manual model", func() {
				userViews := env.Pool("UserView").SearchAll()
				So(userViews.Len(), ShouldEqual, 3)
				userViews = env.Pool("UserView").OrderBy("Name")
				So(userViews.Len(), ShouldEqual, 3)
				recs := userViews.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Get(Name), ShouldEqual, "Jane Smith")
				So(recs[1].Get(Name), ShouldEqual, "John Smith")
				So(recs[2].Get(Name), ShouldEqual, "Will Smith")
				So(recs[0].Get(city), ShouldEqual, "New York")
				So(recs[1].Get(city), ShouldEqual, "")
				So(recs[2].Get(city), ShouldEqual, "")
			})
			Convey("Testing browse with empty ids", func() {
				var ids []int64
				users := env.Pool("User").Model().Browse(env, ids)
				So(users.Len(), ShouldEqual, 0)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Testing access control list while searching", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			Convey("Checking that user 2 cannot access records", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Get(Name).(string), ShouldEqual, "Jane Smith")
				So(userJane.Get(email).(string), ShouldEqual, "jane.smith@example.com")
				So(userJane.Get(age), ShouldEqual, 23)
				So(func() { userJane.Get(profile).(RecordSet).Collection().Get(age) }, ShouldPanic)
			})
			Convey("Revoking model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").RevokeGroup(group1)
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(func() { userJohn.Load() }, ShouldPanic)
			})
			Convey("Regranting model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(func() { userJane.Get(profile).(RecordSet).Collection().Get(age) }, ShouldPanic)
			})
			Convey("Checking record rules", func() {
				users := env.Pool("User").SearchAll()
				So(users.Len(), ShouldEqual, 3)

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
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Get(Name), ShouldBeIn, []string{"Jane Smith", "John Smith"})
				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	Convey("Testing advanced queries on M2O relations", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields with ids", func() {
				profileID := jane.Get(profile).(RecordSet).Collection().Get(ID).(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Condition on m2o relation fields with recordset", func() {
				janeProfile := jane.Get(profile).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(janeProfile))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Empty recordset", func() {
				emptyProfile := env.Pool("Profile")
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).Equals(emptyProfile))
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Empty recordset with IsNull", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).IsNull())
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2o relation fields with IN operator and ids", func() {
				profileID := jane.Get(profile).(RecordSet).Collection().Get(ID).(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).In(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Condition on m2o relation fields with IN operator and recordset", func() {
				janeProfile := jane.Get(profile).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profile).In(janeProfile))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Empty recordset with IN operator", func() {
				emptyProfile := env.Pool("Profile")
				users := env.Pool("User").Search(
					env.Pool("User").Model().Field(profile).In(emptyProfile))
				So(users.Len(), ShouldEqual, 0)
				users = env.Pool("User").Search(
					env.Pool("User").Model().Field(profile).In(emptyProfile).
						And().Field(isStaff).Equals(false))
				So(users.Len(), ShouldEqual, 0)
			})
			Convey("M2O chain", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(profileBestPostTitle).Equals("1st Post"))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on o2m relation with slice of ids", func() {
				postID := jane.Get(posts).(RecordSet).Collection().Ids()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).Equals(postID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Conditions on o2m relation with recordset", func() {
				post := jane.Get(posts).(RecordSet).Collection().Records()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).Equals(post))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Conditions on o2m relation with null", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).IsNull())
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].Get(Name), ShouldEqual, "John Smith")
				So(userRecs[1].Get(Name), ShouldEqual, "Will Smith")
			})
			Convey("Condition on o2m relation with IN operator and slice of ids", func() {
				postIds := jane.Get(posts).(RecordSet).Collection().Ids()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).In(postIds))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("Conditions on o2m relation with IN operator and recordset", func() {
				janePosts := jane.Get(posts).(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field(posts).In(janePosts))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
			Convey("O2M Chain", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field(postsTitle).Equals("1st Post"))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get(ID).(int64), ShouldEqual, jane.Get(ID).(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on M2M relations", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			post1 := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Equals("1st Post"))
			So(post1.Len(), ShouldEqual, 1)
			post2 := env.Pool("Post").Search(env.Pool("Post").Model().Field(title).Equals("2nd Post"))
			So(post2.Len(), ShouldEqual, 1)
			tag1 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Trending"))
			tag2 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Books"))
			So(tag1.Len(), ShouldEqual, 1)
			Convey("Condition on m2m relation with slice of ids", func() {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).Equals(tag1.Get(ID)))
				So(rPosts.Len(), ShouldEqual, 1)
				So(rPosts.Get(ID).(int64), ShouldEqual, post1.Get(ID).(int64))
			})
			Convey("Condition on m2m relation with recordset", func() {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).Equals(tag1))
				So(rPosts.Len(), ShouldEqual, 1)
				So(rPosts.Get(ID).(int64), ShouldEqual, post1.Get(ID).(int64))
			})
			Convey("Condition on m2m relation with null", func() {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).IsNull())
				So(rPosts.Len(), ShouldEqual, 0)
			})
			Convey("Condition on m2m relation with IN operator and ids", func() {
				tags12 := tag1.Union(tag2)
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tags12.Ids()))
				So(rPosts.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2m relation with IN operator and empty ids", func() {
				var tagIds []int64
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tagIds))
				So(rPosts.Len(), ShouldEqual, 0)
			})
			Convey("Condition on m2m relation with IN operator and recordset", func() {
				tags12 := tag1.Union(tag2)
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(tags12))
				So(rPosts.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2m relation with IN operator and empty recordset", func() {
				emptyTags := env.Pool("Tag")
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tags).In(emptyTags))
				So(rPosts.Len(), ShouldEqual, 0)
			})
			Convey("M2M Chain", func() {
				rPosts := env.Pool("Post").Search(env.Pool("Post").Model().Field(tagsName).Equals("Trending"))
				So(rPosts.Len(), ShouldEqual, 1)
				So(rPosts.Get(ID).(int64), ShouldEqual, post1.Get(ID).(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries with multiple joins", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			jane := users.Search(users.Model().Field(Name).Equals("Jane Smith"))
			john := users.Search(users.Model().Field(Name).Equals("John Smith"))
			So(jane.Len(), ShouldEqual, 1)
			So(john.Len(), ShouldEqual, 1)
			Convey("Testing M2O-M2O-M2O", func() {
				So(jane.Get(profileBestPostUser).(RecordSet).Collection().Equals(jane), ShouldBeTrue)
				So(john.Get(profileBestPostUser).(RecordSet).Collection().IsEmpty(), ShouldBeTrue)
				johnProfile := env.Pool("Profile").Call("Create", NewModelData(Registry.MustGet("Profile")))
				john.Set(profile, johnProfile)
				So(john.Get(profileBestPostUser).(RecordSet).Collection().Get(email), ShouldBeEmpty)
			})
		}), ShouldBeNil)
	})
}

func TestGroupedQueries(t *testing.T) {
	Convey("Testing grouped queries", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Simple grouped query on the whole table", func() {
				groupedUsers := env.Pool("User").SearchAll().Call("GroupBy", []FieldName{isStaff}).(RecordSet).Collection().Aggregates(isStaff, nums)
				So(len(groupedUsers), ShouldEqual, 2)
				So(groupedUsers[0].Values.Has(isStaff), ShouldBeTrue)
				So(groupedUsers[0].Values.Get(isStaff), ShouldBeFalse)
				So(groupedUsers[0].Values.Has(nums), ShouldBeTrue)
				So(groupedUsers[0].Values.Get(nums), ShouldEqual, 2)
				So(groupedUsers[0].Count, ShouldEqual, 1)
				So(groupedUsers[1].Values.Has(isStaff), ShouldBeTrue)
				So(groupedUsers[1].Values.Get(isStaff), ShouldBeTrue)
				So(groupedUsers[1].Values.Has(nums), ShouldBeTrue)
				So(groupedUsers[1].Values.Get(nums), ShouldEqual, 4)
				So(groupedUsers[1].Count, ShouldEqual, 2)
			})
		}), ShouldBeNil)
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			postModel := Registry.MustGet("Post")
			tagModel := Registry.MustGet("Tag")
			Convey("Update on users Jane and John with Write and Set", func() {
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set(Name, "Jane A. Smith")
				jane.Load()
				So(jane.Get(Name), ShouldEqual, "Jane A. Smith")
				So(jane.Get(email), ShouldEqual, "jane.smith@example.com")

				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := NewModelData(userModel).
					Set(email, "jsmith2@example.com").
					Set(nums, 13).
					Set(isStaff, false)
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get(Name), ShouldEqual, "John Smith")
				So(john.Get(email), ShouldEqual, "jsmith2@example.com")
				So(john.Get(nums), ShouldEqual, 13)
				So(john.Get(isStaff), ShouldBeFalse)
				john.Set(isStaff, true)
				So(john.Get(isStaff), ShouldBeTrue)
			})
			Convey("Updating an empty RecordSet should do nothing", func() {
				empty := env.Pool("User")
				So(func() { empty.Set(Name, "Foo") }, ShouldNotPanic)
				So(func() {
					empty.Call("Write", NewModelData(userModel).
						Set(Name, "Bar"))
				}, ShouldNotPanic)
			})
			Convey("Multiple updates at once on users", func() {
				cond := env.Pool("User").Model().Field(Name).Equals("Jane A. Smith").Or().Field(Name).Equals("John Smith")
				users := env.Pool("User").Search(cond).Load()
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].Get(isStaff).(bool), ShouldBeTrue)
				So(userRecs[1].Get(isStaff).(bool), ShouldBeFalse)
				So(userRecs[0].Get(isActive).(bool), ShouldBeFalse)
				So(userRecs[1].Get(isActive).(bool), ShouldBeFalse)

				users.Set(isStaff, true)
				users.Load()
				So(userRecs[0].Get(isStaff).(bool), ShouldBeTrue)
				So(userRecs[1].Get(isStaff).(bool), ShouldBeTrue)

				data := NewModelData(userModel).
					Set(isStaff, false).
					Set(isActive, true)
				users.Call("Write", data)
				users.Load()
				So(userRecs[0].Get(isStaff).(bool), ShouldBeFalse)
				So(userRecs[1].Get(isStaff).(bool), ShouldBeFalse)
				So(userRecs[0].Get(isActive).(bool), ShouldBeTrue)
				So(userRecs[1].Get(isActive).(bool), ShouldBeTrue)
			})
			Convey("Updating many2one fields", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				janeProfile := userJane.Get(profile).(RecordSet).Collection()
				userJane.Set(profile, nil)
				So(userJane.Get(profile).(RecordSet).Collection().Get(ID), ShouldEqual, 0)
				userJane.Set(profile, janeProfile.Get(ID))
				So(userJane.Get(profile).(RecordSet).Collection().Get(ID), ShouldEqual, janeProfile.ids[0])
				userJane.Set(profile, env.Pool("Profile"))
				So(userJane.Get(profile).(RecordSet).Collection().Get(ID), ShouldEqual, 0)
				userJane.Set(profile, janeProfile)
				So(userJane.Get(profile).(RecordSet).Collection().Get(ID), ShouldEqual, janeProfile.ids[0])

				post1 := janeProfile.Get(bestPost)
				janeProfile.Call("Write", NewModelData(janeProfile.model).
					Create(bestPost, NewModelData(postModel).
						Set(title, "Post created on the Fly")))
				So(janeProfile.Get(bestPost).(RecordSet).Collection().Get(title), ShouldEqual, "Post created on the Fly")
				janeProfile.Set(bestPost, post1)
			})
			Convey("Updating many2many fields", func() {
				emptyPosts := env.Pool("Post")
				post1 := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("1st Post"))
				post1.Call("Write", NewModelData(postModel).
					Create(tags, NewModelData(tagModel).
						Set(Name, "Tag created on the fly")).
					Create(tags, NewModelData(tagModel).
						Set(Name, "Second Tag on the fly")))
				post1Tags := post1.Get(tags).(RecordSet).Collection()
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Get(Name), ShouldBeIn, []string{"Tag created on the fly", "Second Tag on the fly"})
				So(post1Tags.Records()[1].Get(Name), ShouldBeIn, []string{"Tag created on the fly", "Second Tag on the fly"})

				tagBooks := env.Pool("Tag").Search(env.Pool("Tag").Model().Field(Name).Equals("Books"))
				post1.Set(tags, tagBooks)
				post1Tags = post1.Get(tags).(RecordSet).Collection()
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Get(Name), ShouldEqual, "Books")

				post2Tags := emptyPosts.Search(emptyPosts.Model().Field(title).Equals("2nd Post")).Get(tags).(RecordSet).Collection()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Get(Name), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Get(Name), ShouldBeIn, "Books", "Jane's")
			})
			Convey("Updating One2many fields", func() {
				mPosts := env.Pool("Post")
				post1 := mPosts.Search(mPosts.Model().Field(title).Equals("1st Post"))
				post2 := mPosts.Search(mPosts.Model().Field(title).Equals("2nd Post"))
				post3 := mPosts.Call("Create", NewModelData(postModel, FieldMap{
					"Title":   "3rd Post",
					"Content": "Content of third post",
				})).(RecordSet).Collection()
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Set(posts, post1.Call("Union", post3).(RecordSet).Collection())
				So(post1.Get(user).(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(ID))
				So(post3.Get(user).(RecordSet).Collection().Get(ID), ShouldEqual, userJane.Get(ID))
				So(post2.Get(user).(RecordSet).Collection().Get(ID), ShouldEqual, 0)

				userJane.Set(posts, nil)
				userJane.Call("Write", NewModelData(userModel).
					Create(posts, NewModelData(postModel).
						Set(title, "Another post created on the fly")).
					Create(posts, NewModelData(postModel).
						Set(title, "One more post created on the fly")))
				So(userJane.Get(posts).(RecordSet).Len(), ShouldEqual, 2)
				So(userJane.Get(posts).(RecordSet).Collection().Records()[0].Get(title),
					ShouldBeIn, []string{"Another post created on the fly", "One more post created on the fly"})
				So(userJane.Get(posts).(RecordSet).Collection().Records()[1].Get(title),
					ShouldBeIn, []string{"Another post created on the fly", "One more post created on the fly"})

				userJane.Set(posts, post1.Call("Union", post3).(RecordSet).Collection())
				So(userJane.Get(posts).(RecordSet).Len(), ShouldEqual, 2)

			})
		}), ShouldBeNil)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Checking constraint methods enforcement", func() {
				tag1 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field(Name).Equals("Trending"))
				tag1.Load()
				So(func() { tag1.Set(description, "Trending") }, ShouldPanic)
				tag2 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field(Name).Equals("Books"))
				So(func() { tag2.Set(rate, 12) }, ShouldPanic)
				So(func() {
					tag2.Call("Write", FieldMap{
						"Description": "Books",
						"Rate":        -3,
					})
				}, ShouldPanic)
			})
		}), ShouldBeNil)
	})
	Convey("Checking SQL Constraint enforcement", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userWill := env.Pool("User").Search(env.Pool("User").Model().Field(email).Equals("will.smith@example.com"))
			userWill.Call("Write", NewModelData(userModel).Set(nums, 0).Set(isPremium, true))
		}).Error(), ShouldStartWith, "pq: Premium users must have positive nums")
	})

	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Testing access control list on update (write only)", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")

			Convey("Checking that user 2 cannot update records", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := FieldMap{
					"Email": "jsmith3@example.com",
					"Nums":  13,
				}
				So(func() { john.Call("Write", johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				userModel.methods.MustGet("Write").AllowGroup(group1)
				john := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := NewModelData(userModel).
					Set(email, "jsmith3@example.com").
					Set(nums, 13)
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get(Name), ShouldEqual, "John Smith")
				So(john.Get(email), ShouldEqual, "jsmith3@example.com")
				So(john.Get(nums), ShouldEqual, 13)
			})
			Convey("Checking that user 2 cannot update profile through UpdateCity method", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userModel.methods.MustGet("UpdateCity").AllowGroup(group1)
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.Call("UpdateCity", "London") }, ShouldPanic)
			})
			Convey("Checking that user 2 can run UpdateCity after giving permission for caller", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				profileModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("UpdateCity"))
				jane := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.Call("UpdateCity", "London") }, ShouldNotPanic)
			})
			Convey("Checking record rules", func() {
				userJane := env.Pool("User").SearchAll()
				So(userJane.Len(), ShouldEqual, 3)

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
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Get(Name), ShouldEqual, "Jane A. Smith")
				userJane.Set(Name, "Jane B. Smith")
				So(userJane.Get(Name), ShouldEqual, "Jane B. Smith")

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Will Smith"))
				So(func() { userWill.Set(Name, "Will Jr. Smith") }, ShouldPanic)

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("unlinkRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Checking unlink method", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Deleting user John: number of deleted record should be 1", func() {
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				num := userJohn.Call("Unlink")
				So(num, ShouldEqual, 1)
			})
			Convey("Deleted RecordSet should update themselves when reloading", func() {
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				userJohn2 := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith").Or().Field(Name).Equals("Jane A. Smith"))
				So(userJohn.Len(), ShouldEqual, 1)
				So(userJohn2.Len(), ShouldEqual, 1)
				So(users.Len(), ShouldEqual, 2)
				userJohn.Call("Unlink")
				userJohn.ForceLoad()
				So(userJohn.Len(), ShouldEqual, 0)
				userJohn2.ForceLoad()
				So(userJohn2.Len(), ShouldEqual, 0)
				users.ForceLoad()
				So(users.Len(), ShouldEqual, 1)
			})
			Convey("Deleted RecordSet should update themselves when reloading with prefetch", func() {
				users := env.Pool("User").SearchAll()
				So(users.Len(), ShouldEqual, 3)
				So(users.Records()[0].Get(Name), ShouldEqual, "John Smith")
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				userJohn2 := users.Records()[0]
				users.Records()[0].Call("Unlink")
				userJohn2.ForceLoad()
				So(userJohn2.IsEmpty(), ShouldBeTrue)
				users.ForceLoad()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Get(Name), ShouldEqual, "Jane A. Smith")
				So(users.Records()[1].Get(Name), ShouldEqual, "Will Smith")
				userJohn.ForceLoad()
				So(userJohn.Len(), ShouldEqual, 0)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Checking unlink access permissions", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			profileModel := Registry.MustGet("Profile")
			postModel := Registry.MustGet("Post")

			Convey("Checking that user 2 cannot unlink records", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(func() { users.Call("Unlink") }, ShouldPanic)
			})
			Convey("Adding unlink permission to user2", func() {
				userModel.methods.MustGet("Unlink").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				So(func() { users.Call("Unlink") }, ShouldPanic)
			})
			Convey("Adding permissions to user2 on Profile and Post", func() {
				profileModel.methods.MustGet("Load").AllowGroup(group1)
				postModel.methods.MustGet("Load").AllowGroup(group1)
				postModel.methods.MustGet("Write").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("John Smith"))
				num := users.Call("Unlink")
				So(num, ShouldEqual, 1)
			})
			Convey("Checking record rules", func() {

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
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Call("Unlink"), ShouldEqual, 1)

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field(Name).Equals("Will Smith"))
				So(userWill.Call("Unlink"), ShouldEqual, 0)

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}
