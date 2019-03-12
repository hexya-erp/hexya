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
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := FieldMap{
					"Name":    "John Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				}
				users := env.Pool("User").Call("Create", userJohnData).(RecordSet).Collection()
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID"), ShouldBeGreaterThan, 0)
				So(users.Get("Resume").(RecordSet).IsEmpty(), ShouldBeFalse)
			})
			Convey("Creating user Jane with related Profile and Posts and Tags and Comments", func() {
				userJaneProfileData := NewModelData(Registry.MustGet("Profile")).
					Set("Age", 23).
					Set("Money", 12345).
					Set("Street", "165 5th Avenue").
					Set("City", "New York").
					Set("Zip", "0305").
					Set("Country", "USA")
				profile := env.Pool("Profile").Call("Create", userJaneProfileData).(RecordSet).Collection()
				So(profile.Len(), ShouldEqual, 1)
				So(profile.Get("UserName"), ShouldBeBlank)

				tag1 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Trending",
				}).(RecordSet).Collection()
				tag2 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Books",
				}).(RecordSet).Collection()
				tag3 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Jane's",
				}).(RecordSet).Collection()
				So(tag1.Len(), ShouldEqual, 1)
				So(tag2.Len(), ShouldEqual, 1)
				So(tag3.Len(), ShouldEqual, 1)

				post1Data := FieldMap{
					"Title":   "1st Post",
					"Content": "Content of first post",
					"Tags":    tag1.Union(tag3),
				}
				post1 := env.Pool("Post").Call("Create", post1Data).(RecordSet).Collection()
				So(post1.Len(), ShouldEqual, 1)
				post2Data := FieldMap{
					"Title":   "2nd Post",
					"Content": "Content of second post",
				}
				post2 := env.Pool("Post").Call("Create", post2Data).(RecordSet).Collection()
				So(post2.Len(), ShouldEqual, 1)
				posts := post1.Union(post2)
				profile.Set("BestPost", post1)
				userJaneData := FieldMap{
					"Name":    "Jane Smith",
					"Email":   "jane.smith@example.com",
					"Profile": profile,
					"Posts":   posts,
					"Nums":    2,
				}
				userJane := env.Pool("User").Call("Create", userJaneData).(RecordSet).Collection()
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("ID"), ShouldEqual, profile.Get("ID"))
				So(profile.Get("UserName"), ShouldEqual, "Jane Smith")

				So(post1.Get("User").(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post2.Get("User").(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("ID"))
				janePosts := userJane.Get("Posts").(RecordSet).Collection()
				So(janePosts.Len(), ShouldEqual, 2)

				So(post2.Get("LastTagName"), ShouldBeBlank)
				post2.Set("Tags", tag2.Union(tag3))
				So(post1.Get("LastTagName"), ShouldEqual, "Jane's")
				post1Tags := post1.Get("Tags").(RecordSet).Collection()
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Get("Name"), ShouldBeIn, "Trending", "Jane's")
				So(post1Tags.Records()[1].Get("Name"), ShouldBeIn, "Trending", "Jane's")
				post2Tags := post2.Get("Tags").(RecordSet).Collection()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Get("Name"), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Get("Name"), ShouldBeIn, "Books", "Jane's")

				So(post1.Get("LastCommentText").(string), ShouldBeBlank)
				env.Pool("Comment").Call("Create", FieldMap{
					"Post": post1,
					"Text": "First Comment",
				})
				env.Pool("Comment").Call("Create", FieldMap{
					"Post": post1,
					"Text": "Another Comment",
				})
				env.Pool("Comment").Call("Create", FieldMap{
					"Post": post1,
					"Text": "Third Comment",
				})
				So(post1.Get("LastCommentText").(string), ShouldEqual, "Third Comment")
				So(post1.Get("Comments").(RecordSet).Len(), ShouldEqual, 3)
			})
			Convey("Creating a user Will Smith", func() {
				userWillData := FieldMap{
					"Name":    "Will Smith",
					"Email":   "will.smith@example.com",
					"IsStaff": true,
					"Nums":    3,
				}
				userWill := env.Pool("User").Call("Create", userWillData).(RecordSet).Collection()
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.Get("ID"), ShouldBeGreaterThan, 0)
			})
		}), ShouldBeNil)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Checking constraint methods enforcement", func() {
				tag1Data := FieldMap{
					"Name":        "Tag1",
					"Description": "Tag1",
				}
				So(func() { env.Pool("Tag").Call("Create", tag1Data) }, ShouldPanic)
				tag2Data := FieldMap{
					"Name": "Tag2",
					"Rate": 12,
				}
				So(func() { env.Pool("Tag").Call("Create", tag2Data) }, ShouldPanic)
				tag3Data := FieldMap{
					"Name":        "Tag2",
					"Description": "Tag2",
					"Rate":        -3,
				}
				So(func() { env.Pool("Tag").Call("Create", tag3Data) }, ShouldPanic)
			})
			Convey("Checking that we can't create two users with the same name", func() {
				user1Data := FieldMap{
					"Name": "User1",
				}
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldPanic)
			})
			Convey("Checking that we can't create two users with a empty string name", func() {
				user1Data := FieldMap{
					"Name": "",
				}
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user1Data).(RecordSet).Collection() }, ShouldPanic)
			})
			Convey("Checking that we can create as many users with a NULL name", func() {
				user2Data := FieldMap{
					"Email": "user2@example.com",
				}
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
				So(func() { env.Pool("User").Call("Create", user2Data).(RecordSet).Collection() }, ShouldNotPanic)
			})
		}), ShouldBeNil)
	})
	Convey("Checking SQL Constraint enforcement", t, func() {
		err := SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userRobData := FieldMap{
				"Name":      "Rob Smith",
				"IsPremium": true,
			}
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
			resumeModel := Registry.MustGet("Resume")

			Convey("Checking that user 2 cannot create records", func() {
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				userModel.methods.MustGet("Create").AllowGroup(group1)
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Write"))
				userTomData := FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				}
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for resume and it works", func() {
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				resumeModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("Create"))
				userTomData := FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(func() { userTom.Get("Name") }, ShouldPanic)
			})
			Convey("Revoking model access rights to user 2 for resume and it doesn't works", func() {
				resumeModel.methods.MustGet("Create").RevokeGroup(group1)
				userTomData := FieldMap{
					"Name":       "Tom Smith",
					"Email":      "tsmith@example.com",
					"Experience": "10 year of Hexya development",
				}
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Regranting model access rights to user 2 for posts and it works", func() {
				resumeModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(func() { userTom.Get("Name") }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordSet).Collection()
				So(userTom.Get("Name"), ShouldEqual, "Tom Smith")
				So(userTom.Get("Email"), ShouldEqual, "tsmith@example.com")
			})
			Convey("Checking that we can create tags", func() {
				tagData := FieldMap{
					"Name": "My Tag",
				}
				So(func() { env.Pool("Tag").Call("Create", tagData) }, ShouldNotPanic)
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID      int64
			Name    string
			Email   string
			Profile *RecordCollection
		}
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Searching User Jane", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(userJane.Len(), ShouldEqual, 1)
				Convey("Reading Jane with Get", func() {
					So(userJane.Get("Name").(string), ShouldEqual, "Jane Smith")
					So(userJane.Get("Email"), ShouldEqual, "jane.smith@example.com")
					So(userJane.Get("Profile").(RecordSet).Collection().Get("Age"), ShouldEqual, 23)
					So(userJane.Get("Profile").(RecordSet).Collection().Get("Money"), ShouldEqual, 12345)
					So(userJane.Get("Profile").(RecordSet).Collection().Get("Country"), ShouldEqual, "USA")
					So(userJane.Get("Profile").(RecordSet).Collection().Get("Zip"), ShouldEqual, "0305")
					recs := userJane.Get("Posts").(RecordSet).Collection().Records()
					So(recs, ShouldHaveLength, 2)
					So(recs[0].Get("Title"), ShouldEqual, "1st Post")
					So(recs[1].Get("Title"), ShouldEqual, "2nd Post")
				})
				Convey("Reading Jane with ReadFirst", func() {
					ujData := userJane.First()
					name, ok := ujData.Get("Name")
					So(name, ShouldEqual, "Jane Smith")
					So(ok, ShouldBeTrue)
					email, ok := ujData.Get("Email")
					So(email, ShouldEqual, "jane.smith@example.com")
					So(ok, ShouldBeTrue)
					id, ok := ujData.Get("ID")
					So(id, ShouldEqual, userJane.Get("ID").(int64))
					So(ok, ShouldBeTrue)
					profile, ok := ujData.Get("Profile")
					So(profile.(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("Profile").(RecordSet).Collection().Get("ID"))
					So(ok, ShouldBeTrue)
				})
			})

			Convey("Testing search all users", func() {
				usersAll := env.Pool("User").Call("SearchAll").(RecordSet).Collection()
				So(usersAll.Len(), ShouldEqual, 3)
				usersAll = env.Pool("User").OrderBy("Name")
				So(usersAll.Len(), ShouldEqual, 3)
				Convey("Reading first user with Get", func() {
					So(usersAll.Get("Name"), ShouldEqual, "Jane Smith")
					So(usersAll.Get("Email"), ShouldEqual, "jane.smith@example.com")
				})
				Convey("Reading all users with Records and Get", func() {
					recs := usersAll.Records()
					So(len(recs), ShouldEqual, 3)
					So(recs[0].Get("Email"), ShouldEqual, "jane.smith@example.com")
					So(recs[1].Get("Email"), ShouldEqual, "jsmith@example.com")
					So(recs[2].Get("Email"), ShouldEqual, "will.smith@example.com")
				})
				Convey("Reading all users with ReadAll()", func() {
					usersData := usersAll.All()
					email0, ok := usersData[0].Get("Email")
					So(email0, ShouldEqual, "jane.smith@example.com")
					So(ok, ShouldBeTrue)
					email1, ok := usersData[1].Get("Email")
					So(email1, ShouldEqual, "jsmith@example.com")
					So(ok, ShouldBeTrue)
					email2, ok := usersData[2].Get("Email")
					So(email2, ShouldEqual, "will.smith@example.com")
					So(ok, ShouldBeTrue)
				})
			})
			Convey("Testing search on manual model", func() {
				userViews := env.Pool("UserView").SearchAll()
				So(userViews.Len(), ShouldEqual, 3)
				userViews = env.Pool("UserView").OrderBy("Name")
				So(userViews.Len(), ShouldEqual, 3)
				recs := userViews.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Get("Name"), ShouldEqual, "Jane Smith")
				So(recs[1].Get("Name"), ShouldEqual, "John Smith")
				So(recs[2].Get("Name"), ShouldEqual, "Will Smith")
				So(recs[0].Get("City"), ShouldEqual, "New York")
				So(recs[1].Get("City"), ShouldEqual, "")
				So(recs[2].Get("City"), ShouldEqual, "")
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Testing access control list while searching", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")
			Convey("Checking that user 2 cannot access records", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Get("Name").(string), ShouldEqual, "Jane Smith")
				So(userJane.Get("Email").(string), ShouldEqual, "jane.smith@example.com")
				So(userJane.Get("Age"), ShouldEqual, 23)
				So(func() { userJane.Get("Profile").(RecordSet).Collection().Get("Age") }, ShouldPanic)
			})
			Convey("Revoking model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").RevokeGroup(group1)
				userJohn := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(func() { userJohn.Load() }, ShouldPanic)
			})
			Convey("Regranting model access rights to user 2 and checking access", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(func() { userJane.Get("Profile").(RecordSet).Collection().Get("Age") }, ShouldPanic)
			})
			Convey("Checking record rules", func() {
				users := env.Pool("User").SearchAll()
				So(users.Len(), ShouldEqual, 3)

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: users.Model().Field("Name").IContains("j"),
					Perms:     security.Read,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: users.Model().Field("Name").Equals("Nobody"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&notUsedRule)

				users = env.Pool("User").SearchAll()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Get("Name"), ShouldBeIn, []string{"Jane Smith", "John Smith"})
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
			jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields with ids", func() {
				profileID := jane.Get("Profile").(RecordSet).Collection().Get("ID").(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").Equals(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Condition on m2o relation fields with recordset", func() {
				profile := jane.Get("Profile").(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").Equals(profile))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Empty recordset", func() {
				profile := env.Pool("Profile")
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").Equals(profile))
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Empty recordset with IsNull", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").IsNull())
				So(users.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2o relation fields with IN operator and ids", func() {
				profileID := jane.Get("Profile").(RecordSet).Collection().Get("ID").(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Condition on m2o relation fields with IN operator and recordset", func() {
				profile := jane.Get("Profile").(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profile))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Empty recordset with IN operator", func() {
				profile := env.Pool("Profile")
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profile))
				So(users.Len(), ShouldEqual, 0)
			})
			Convey("M2O chain", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field("profile_id.best_post_id.title").Equals("1st Post"))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on o2m relation with slice of ids", func() {
				postID := jane.Get("Posts").(RecordSet).Collection().Ids()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").Equals(postID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with recordset", func() {
				post := jane.Get("Posts").(RecordSet).Collection().Records()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").Equals(post))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with null", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").IsNull())
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].Get("Name"), ShouldEqual, "John Smith")
				So(userRecs[1].Get("Name"), ShouldEqual, "Will Smith")
			})
			Convey("Condition on o2m relation with IN operator and slice of ids", func() {
				postIds := jane.Get("Posts").(RecordSet).Collection().Ids()
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").In(postIds))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with IN operator and recordset", func() {
				posts := jane.Get("Posts").(RecordSet).Collection()
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").In(posts))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("O2M Chain", func() {
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts.Title").Equals("1st Post"))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries on M2M relations", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			post1 := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Equals("1st Post"))
			So(post1.Len(), ShouldEqual, 1)
			post2 := env.Pool("Post").Search(env.Pool("Post").Model().Field("Title").Equals("2nd Post"))
			So(post2.Len(), ShouldEqual, 1)
			tag1 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field("Name").Equals("Trending"))
			tag2 := env.Pool("Tag").Search(env.Pool("Tag").Model().Field("Name").Equals("Books"))
			So(tag1.Len(), ShouldEqual, 1)
			Convey("Condition on m2m relation with slice of ids", func() {
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags").Equals(tag1.Get("ID")))
				So(posts.Len(), ShouldEqual, 1)
				So(posts.Get("ID").(int64), ShouldEqual, post1.Get("ID").(int64))
			})
			Convey("Condition on m2m relation with recordset", func() {
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags").Equals(tag1))
				So(posts.Len(), ShouldEqual, 1)
				So(posts.Get("ID").(int64), ShouldEqual, post1.Get("ID").(int64))
			})
			Convey("Condition on m2m relation with null", func() {
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags").IsNull())
				So(posts.Len(), ShouldEqual, 0)
			})
			Convey("Condition on m2m relation with IN operator and ids", func() {
				tags := tag1.Union(tag2)
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags").In(tags.Ids()))
				So(posts.Len(), ShouldEqual, 2)
			})
			Convey("Condition on m2m relation with IN operator and recordset", func() {
				tags := tag1.Union(tag2)
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags").In(tags))
				So(posts.Len(), ShouldEqual, 2)
			})
			Convey("M2M Chain", func() {
				posts := env.Pool("Post").Search(env.Pool("Post").Model().Field("Tags.Name").Equals("Trending"))
				So(posts.Len(), ShouldEqual, 1)
				So(posts.Get("ID").(int64), ShouldEqual, post1.Get("ID").(int64))
			})
		}), ShouldBeNil)
	})
	Convey("Testing advanced queries with multiple joins", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			jane := users.Search(users.Model().Field("Name").Equals("Jane Smith"))
			john := users.Search(users.Model().Field("Name").Equals("John Smith"))
			So(jane.Len(), ShouldEqual, 1)
			So(john.Len(), ShouldEqual, 1)
			Convey("Testing M2O-M2O-M2O", func() {
				So(jane.Get("Profile.BestPost.User").(RecordSet).Collection().Equals(jane), ShouldBeTrue)
				So(john.Get("Profile.BestPost.User").(RecordSet).Collection().IsEmpty(), ShouldBeTrue)
				johnProfile := env.Pool("Profile").Call("Create", NewModelData(users.Model()))
				john.Set("Profile", johnProfile)
				So(john.Get("Profile.BestPost.User").(RecordSet).Collection().Get("Email"), ShouldBeEmpty)
			})
		}), ShouldBeNil)
	})
}

func TestGroupedQueries(t *testing.T) {
	Convey("Testing grouped queries", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Simple grouped query on the whole table", func() {
				groupedUsers := env.Pool("User").SearchAll().Call("GroupBy", []FieldNamer{FieldName("IsStaff")}).(RecordSet).Collection().Call("Aggregates", []FieldNamer{FieldName("IsStaff"), FieldName("Nums")}).([]GroupAggregateRow)
				So(len(groupedUsers), ShouldEqual, 2)
				is, ok1 := groupedUsers[0].Values.Get("IsStaff")
				So(ok1, ShouldBeTrue)
				So(is, ShouldBeFalse)
				n, ok2 := groupedUsers[0].Values.Get("Nums")
				So(ok2, ShouldBeTrue)
				So(n, ShouldEqual, 2)
				So(groupedUsers[0].Count, ShouldEqual, 1)
				is1, ok3 := groupedUsers[1].Values.Get("IsStaff")
				So(ok3, ShouldBeTrue)
				So(is1, ShouldBeTrue)
				n1, ok4 := groupedUsers[1].Values.Get("Nums")
				So(ok4, ShouldBeTrue)
				So(n1, ShouldEqual, 4)
				So(groupedUsers[1].Count, ShouldEqual, 2)
			})
		}), ShouldBeNil)
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		So(ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Update on users Jane and John with Write and Set", func() {
				jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(jane.Len(), ShouldEqual, 1)
				jane.Set("Name", "Jane A. Smith")
				jane.Load()
				So(jane.Get("Name"), ShouldEqual, "Jane A. Smith")
				So(jane.Get("Email"), ShouldEqual, "jane.smith@example.com")

				john := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := FieldMap{
					"Email":   "jsmith2@example.com",
					"Nums":    13,
					"IsStaff": false,
				}
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get("Name"), ShouldEqual, "John Smith")
				So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
				So(john.Get("Nums"), ShouldEqual, 13)
				So(john.Get("IsStaff"), ShouldBeFalse)
				john.Set("IsStaff", true)
				So(john.Get("IsStaff"), ShouldBeTrue)
			})
			Convey("Multiple updates at once on users", func() {
				cond := env.Pool("User").Model().Field("Name").Equals("Jane A. Smith").Or().Field("Name").Equals("John Smith")
				users := env.Pool("User").Search(cond).Load()
				So(users.Len(), ShouldEqual, 2)
				userRecs := users.Records()
				So(userRecs[0].Get("IsStaff").(bool), ShouldBeTrue)
				So(userRecs[1].Get("IsStaff").(bool), ShouldBeFalse)
				So(userRecs[0].Get("IsActive").(bool), ShouldBeFalse)
				So(userRecs[1].Get("IsActive").(bool), ShouldBeFalse)

				users.Set("IsStaff", true)
				users.Load()
				So(userRecs[0].Get("IsStaff").(bool), ShouldBeTrue)
				So(userRecs[1].Get("IsStaff").(bool), ShouldBeTrue)

				fMap := FieldMap{
					"IsStaff":  false,
					"IsActive": true,
				}
				users.Call("Write", fMap)
				users.Load()
				So(userRecs[0].Get("IsStaff").(bool), ShouldBeFalse)
				So(userRecs[1].Get("IsStaff").(bool), ShouldBeFalse)
				So(userRecs[0].Get("IsActive").(bool), ShouldBeTrue)
				So(userRecs[1].Get("IsActive").(bool), ShouldBeTrue)
			})
			Convey("Updating many2one fields", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("jane.smith@example.com"))
				profile := userJane.Get("Profile").(RecordSet).Collection()
				userJane.Set("Profile", nil)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("ID"), ShouldEqual, 0)
				userJane.Set("Profile", profile.Get("ID"))
				So(userJane.Get("Profile").(RecordSet).Collection().Get("ID"), ShouldEqual, profile.ids[0])
				userJane.Set("Profile", env.Pool("Profile"))
				So(userJane.Get("Profile").(RecordSet).Collection().Get("ID"), ShouldEqual, 0)
				userJane.Set("Profile", profile)
				So(userJane.Get("Profile").(RecordSet).Collection().Get("ID"), ShouldEqual, profile.ids[0])
			})
			Convey("Updating many2many fields", func() {
				posts := env.Pool("Post")
				post1 := posts.Search(posts.Model().Field("title").Equals("1st Post"))
				tagBooks := env.Pool("Tag").Search(env.Pool("Tag").Model().Field("name").Equals("Books"))
				post1.Set("Tags", tagBooks)

				post1Tags := post1.Get("Tags").(RecordSet).Collection()
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Get("Name"), ShouldEqual, "Books")
				post2Tags := posts.Search(posts.Model().Field("title").Equals("2nd Post")).Get("Tags").(RecordSet).Collection()
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Get("Name"), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Get("Name"), ShouldBeIn, "Books", "Jane's")
			})
			Convey("Updating One2many fields", func() {
				posts := env.Pool("Post")
				post1 := posts.Search(posts.Model().Field("title").Equals("1st Post"))
				post2 := posts.Search(posts.Model().Field("title").Equals("2nd Post"))
				post3 := posts.Call("Create", FieldMap{
					"Title":   "3rd Post",
					"Content": "Content of third post",
				}).(RecordSet).Collection()
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("jane.smith@example.com"))
				userJane.Set("Posts", post1.Call("Union", post3).(RecordSet).Collection())
				So(post1.Get("User").(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post3.Get("User").(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post2.Get("User").(RecordSet).Collection().Get("ID"), ShouldEqual, 0)
			})
		}), ShouldBeNil)
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Checking constraint methods enforcement", func() {
				tag1 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field("Name").Equals("Trending"))
				tag1.Load()
				So(func() { tag1.Set("Description", "Trending") }, ShouldPanic)
				tag2 := env.Pool("Tag").Search(Registry.MustGet("Tag").Field("Name").Equals("Books"))
				So(func() { tag2.Set("Rate", 12) }, ShouldPanic)
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
			userWill := env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("will.smith@example.com"))
			userWill.Call("Write", FieldMap{"Nums": 0, "IsPremium": true})
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
				john := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := FieldMap{
					"Email": "jsmith3@example.com",
					"Nums":  13,
				}
				So(func() { john.Call("Write", johnValues) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check update", func() {
				userModel.methods.MustGet("Write").AllowGroup(group1)
				john := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := FieldMap{
					"Email": "jsmith3@example.com",
					"Nums":  13,
				}
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get("Name"), ShouldEqual, "John Smith")
				So(john.Get("Email"), ShouldEqual, "jsmith3@example.com")
				So(john.Get("Nums"), ShouldEqual, 13)
			})
			Convey("Checking that user 2 cannot update profile through UpdateCity method", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userModel.methods.MustGet("UpdateCity").AllowGroup(group1)
				jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.Call("UpdateCity", "London") }, ShouldPanic)
			})
			Convey("Checking that user 2 can run UpdateCity after giving permission for caller", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				profileModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("UpdateCity"))
				jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.Call("UpdateCity", "London") }, ShouldNotPanic)
			})
			Convey("Checking record rules", func() {
				userJane := env.Pool("User").SearchAll()
				So(userJane.Len(), ShouldEqual, 3)

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: env.Pool("User").Model().Field("Name").IContains("j"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "unlinkRule",
					Group:     group1,
					Condition: env.Pool("User").Model().Field("Name").Equals("Nobody"),
					Perms:     security.Unlink,
				}
				userModel.AddRecordRule(&notUsedRule)

				userJane = env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Get("Name"), ShouldEqual, "Jane A. Smith")
				userJane.Set("Name", "Jane B. Smith")
				So(userJane.Get("Name"), ShouldEqual, "Jane B. Smith")

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Will Smith"))
				So(func() { userWill.Set("Name", "Will Jr. Smith") }, ShouldPanic)

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("unlinkRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
			num := users.Call("Unlink")
			Convey("Number of deleted record should be 1", func() {
				So(num, ShouldEqual, 1)
			})
		}), ShouldBeNil)
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Checking unlink access permissions", t, func() {
		So(SimulateInNewEnvironment(2, func(env Environment) {
			userModel := Registry.MustGet("User")

			Convey("Checking that user 2 cannot unlink records", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(func() { users.Call("Unlink") }, ShouldPanic)
			})
			Convey("Adding unlink permission to user2", func() {
				userModel.methods.MustGet("Unlink").AllowGroup(group1)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				num := users.Call("Unlink")
				So(num, ShouldEqual, 1)
			})
			Convey("Checking record rules", func() {

				rule := RecordRule{
					Name:      "jOnly",
					Group:     group1,
					Condition: env.Pool("User").Model().Field("Name").IContains("j"),
					Perms:     security.Unlink,
				}
				userModel.AddRecordRule(&rule)

				notUsedRule := RecordRule{
					Name:      "writeRule",
					Group:     group1,
					Condition: env.Pool("User").Model().Field("Name").Equals("Nobody"),
					Perms:     security.Write,
				}
				userModel.AddRecordRule(&notUsedRule)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("jane.smith@example.com"))
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Call("Unlink"), ShouldEqual, 1)

				userWill := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Will Smith"))
				So(userWill.Call("Unlink"), ShouldEqual, 0)

				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		}), ShouldBeNil)
	})
	security.Registry.UnregisterGroup(group1)
}
