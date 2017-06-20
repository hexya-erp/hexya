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

	"github.com/hexya-erp/hexya/hexya/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Creating simple user John with no relations and checking ID", func() {
				userJohnData := FieldMap{
					"Name":    "John Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				}
				users := env.Pool("User").Call("Create", userJohnData).(RecordCollection)
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID"), ShouldBeGreaterThan, 0)
			})
			Convey("Creating user Jane with related Profile and Posts and Tags", func() {
				userJaneProfileData := FieldMap{
					"Age":     23,
					"Money":   12345,
					"Street":  "165 5th Avenue",
					"City":    "New York",
					"Zip":     "0305",
					"Country": "USA",
				}
				profile := env.Pool("Profile").Call("Create", userJaneProfileData).(RecordCollection)
				So(profile.Len(), ShouldEqual, 1)
				post1Data := FieldMap{
					"Title":   "1st Post",
					"Content": "Content of first post",
				}
				post1 := env.Pool("Post").Call("Create", post1Data).(RecordCollection)
				So(post1.Len(), ShouldEqual, 1)
				post2Data := FieldMap{
					"Title":   "2nd Post",
					"Content": "Content of second post",
				}
				post2 := env.Pool("Post").Call("Create", post2Data).(RecordCollection)
				So(post2.Len(), ShouldEqual, 1)
				posts := post1.Union(post2)
				userJaneData := FieldMap{
					"Name":    "Jane Smith",
					"Email":   "jane.smith@example.com",
					"Profile": profile,
					"Posts":   posts,
					"Nums":    2,
				}
				userJane := env.Pool("User").Call("Create", userJaneData).(RecordCollection)
				So(userJane.Len(), ShouldEqual, 1)
				So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, profile.Get("ID"))
				So(post1.Get("User").(RecordCollection).Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post2.Get("User").(RecordCollection).Get("ID"), ShouldEqual, userJane.Get("ID"))
				janePosts := userJane.Get("Posts").(RecordCollection)
				So(janePosts.Len(), ShouldEqual, 2)

				tag1 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Trending",
				}).(RecordCollection)
				tag2 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Books",
				}).(RecordCollection)
				tag3 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Jane's",
				}).(RecordCollection)
				post1.Set("Tags", tag1.Union(tag3))
				post2.Set("Tags", tag2.Union(tag3))
				post1Tags := post1.Get("Tags").(RecordCollection)
				So(post1Tags.Len(), ShouldEqual, 2)
				So(post1Tags.Records()[0].Get("Name"), ShouldBeIn, "Trending", "Jane's")
				So(post1Tags.Records()[1].Get("Name"), ShouldBeIn, "Trending", "Jane's")
				post2Tags := post2.Get("Tags").(RecordCollection)
				So(post2Tags.Len(), ShouldEqual, 2)
				So(post2Tags.Records()[0].Get("Name"), ShouldBeIn, "Books", "Jane's")
				So(post2Tags.Records()[1].Get("Name"), ShouldBeIn, "Books", "Jane's")
			})
			Convey("Creating a user Will Smith", func() {
				userWillData := FieldMap{
					"Name":    "Will Smith",
					"Email":   "will.smith@example.com",
					"IsStaff": true,
					"Nums":    3,
				}
				userWill := env.Pool("User").Call("Create", userWillData).(RecordCollection)
				So(userWill.Len(), ShouldEqual, 1)
				So(userWill.Get("ID"), ShouldBeGreaterThan, 0)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	Convey("Testing access control list on creation (create only)", t, func() {
		SimulateInNewEnvironment(2, func(env Environment) {
			security.Registry.AddMembership(2, group1)
			userModel := Registry.MustGet("User")
			postModel := Registry.MustGet("Post")

			Convey("Checking that user 2 cannot create records", func() {
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 and check failure again", func() {
				userModel.methods.MustGet("Create").AllowGroup(group1)
				postModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Write"))
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
			})
			Convey("Adding model access rights to user 2 for posts and it works", func() {
				postModel.methods.MustGet("Create").AllowGroup(group1, userModel.methods.MustGet("Create"))
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
				So(func() { userTom.Get("Name") }, ShouldPanic)
			})
			Convey("Checking creation again with read rights too", func() {
				userModel.methods.MustGet("Load").AllowGroup(group1)
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
				So(userTom.Get("Name"), ShouldEqual, "Tom Smith")
				So(userTom.Get("Email"), ShouldEqual, "tsmith@example.com")
			})
			Convey("Removing Create right on Email field", func() {
				userModel.fields.MustGet("Email").RevokeAccess(security.GroupEveryone, security.Write)
				userTomData := FieldMap{
					"Name":  "Tom Smith",
					"Email": "tsmith@example.com",
				}
				userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
				So(userTom.Get("Name"), ShouldEqual, "Tom Smith")
				So(userTom.Get("Email").(string), ShouldBeBlank)

				userModel.fields.MustGet("Email").GrantAccess(security.GroupEveryone, security.Write)
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestSearchRecordSet(t *testing.T) {
	Convey("Testing search through RecordSets", t, func() {
		type UserStruct struct {
			ID      int64
			Name    string
			Email   string
			Profile RecordCollection
		}
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Searching User Jane", func() {
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(userJane.Len(), ShouldEqual, 1)
				Convey("Reading Jane with Get", func() {
					So(userJane.Get("Name").(string), ShouldEqual, "Jane Smith")
					So(userJane.Get("Email"), ShouldEqual, "jane.smith@example.com")
					So(userJane.Get("Profile").(RecordCollection).Get("Age"), ShouldEqual, 23)
					So(userJane.Get("Profile").(RecordCollection).Get("Money"), ShouldEqual, 12345)
					So(userJane.Get("Profile").(RecordCollection).Get("Country"), ShouldEqual, "USA")
					So(userJane.Get("Profile").(RecordCollection).Get("Zip"), ShouldEqual, "0305")
					recs := userJane.Get("Posts").(RecordCollection).Records()
					So(recs[0].Get("Title"), ShouldEqual, "1st Post")
					So(recs[1].Get("Title"), ShouldEqual, "2nd Post")
				})
				Convey("Reading Jane with ReadFirst", func() {
					var userJaneStruct UserStruct
					userJane.First(&userJaneStruct)
					So(userJaneStruct.Name, ShouldEqual, "Jane Smith")
					So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
					So(userJaneStruct.ID, ShouldEqual, userJane.Get("ID").(int64))
					So(userJaneStruct.Profile.Get("ID"), ShouldEqual, userJane.Get("Profile").(RecordCollection).Get("ID"))
				})
			})

			Convey("Testing search all users", func() {
				usersAll := env.Pool("User").Call("FetchAll").(RecordCollection)
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
					var userStructs []*UserStruct
					usersAll.All(&userStructs)
					So(userStructs[0].Email, ShouldEqual, "jane.smith@example.com")
					So(userStructs[1].Email, ShouldEqual, "jsmith@example.com")
					So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
				})
			})

			Convey("Testing search on manual model", func() {
				userViews := env.Pool("UserView").FetchAll()
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
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Testing access control list while searching", t, func() {
		SimulateInNewEnvironment(2, func(env Environment) {
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
				So(func() { userJane.Get("Profile").(RecordCollection).Get("Age") }, ShouldPanic)
			})
			Convey("Adding field access rights to user 2 and checking access", func() {
				userModel.fields.MustGet("Email").RevokeAccess(security.GroupEveryone, security.Read)
				userModel.fields.MustGet("Age").RevokeAccess(security.GroupEveryone, security.Read)

				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
				So(func() { userJane.Load() }, ShouldNotPanic)
				So(userJane.Get("Name").(string), ShouldEqual, "Jane Smith")
				So(userJane.Get("Email").(string), ShouldBeBlank)
				So(userJane.Get("Age"), ShouldEqual, 0)

				userModel.fields.MustGet("Email").GrantAccess(security.GroupEveryone, security.Read)
				userModel.fields.MustGet("Age").GrantAccess(security.GroupEveryone, security.Read)
			})
			Convey("Checking record rules", func() {
				users := env.Pool("User").FetchAll()
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

				users = env.Pool("User").FetchAll()
				So(users.Len(), ShouldEqual, 2)
				So(users.Records()[0].Get("Name"), ShouldBeIn, []string{"Jane Smith", "John Smith"})
				userModel.RemoveRecordRule("jOnly")
				userModel.RemoveRecordRule("writeRule")
			})
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestAdvancedQueries(t *testing.T) {
	Convey("Testing advanced queries on M2O relations", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on m2o relation fields with ids", func() {
				profileID := jane.Get("Profile").(RecordCollection).Get("ID").(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").Equals(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Condition on m2o relation fields with recordset", func() {
				profile := jane.Get("Profile").(RecordCollection)
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
				profileID := jane.Get("Profile").(RecordCollection).Get("ID").(int64)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profileID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Condition on m2o relation fields with IN operator and recordset", func() {
				profile := jane.Get("Profile").(RecordCollection)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profile))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Empty recordset with IN operator", func() {
				profile := env.Pool("Profile")
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Profile").In(profile))
				So(users.Len(), ShouldEqual, 0)
			})
		})
	})
	Convey("Testing advanced queries on O2M relations", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane Smith"))
			So(jane.Len(), ShouldEqual, 1)
			Convey("Condition on o2m relation with slice of ids", func() {
				postID := jane.Get("Posts").(RecordCollection).Ids()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").Equals(postID))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with recordset", func() {
				post := jane.Get("Posts").(RecordCollection).Records()[0]
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").Equals(post))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Condition on o2m relation with IN operator and slice of ids", func() {
				postIds := jane.Get("Posts").(RecordCollection).Ids()
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").In(postIds))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
			Convey("Conditions on o2m relation with IN operator and recordset", func() {
				posts := jane.Get("Posts").(RecordCollection)
				users := env.Pool("User").Search(env.Pool("User").Model().Field("Posts").In(posts))
				So(users.Len(), ShouldEqual, 1)
				So(users.Get("ID").(int64), ShouldEqual, jane.Get("ID").(int64))
			})
		})
	})
}

func TestGroupedQueries(t *testing.T) {
	Convey("Testing grouped queries", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			Convey("Simple grouped query on the whole table", func() {
				groupedUsers := env.Pool("User").Call("GroupBy", []FieldNamer{FieldName("IsStaff")}).(RecordCollection).Call("Aggregates", []FieldNamer{FieldName("IsStaff"), FieldName("Nums")}).([]GroupAggregateRow)
				So(len(groupedUsers), ShouldEqual, 2)
				So(groupedUsers[0].Values, ShouldContainKey, "is_staff")
				So(groupedUsers[0].Values, ShouldContainKey, "nums")
				So(groupedUsers[1].Values, ShouldContainKey, "is_staff")
				So(groupedUsers[1].Values, ShouldContainKey, "nums")
				So(groupedUsers[0].Values["is_staff"], ShouldBeFalse)
				So(groupedUsers[0].Values["nums"], ShouldEqual, 2)
				So(groupedUsers[0].Count, ShouldEqual, 1)
				So(groupedUsers[1].Values["is_staff"], ShouldBeTrue)
				So(groupedUsers[1].Values["nums"], ShouldEqual, 4)
				So(groupedUsers[1].Count, ShouldEqual, 2)
			})
		})
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
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
					"Email": "jsmith2@example.com",
					"Nums":  13,
				}
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get("Name"), ShouldEqual, "John Smith")
				So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
				So(john.Get("Nums"), ShouldEqual, 13)
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
				profile := userJane.Get("Profile").(RecordCollection)
				userJane.Set("Profile", nil)
				So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, 0)
				userJane.Set("Profile", profile.Get("ID"))
				So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, profile.ids[0])
				userJane.Set("Profile", env.Pool("Profile"))
				So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, 0)
				userJane.Set("Profile", profile)
				So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, profile.ids[0])
			})
			Convey("Updating many2many fields", func() {
				posts := env.Pool("Post")
				post1 := posts.Search(posts.Model().Field("title").Equals("1st Post"))
				tagBooks := env.Pool("Tag").Search(env.Pool("Tag").Model().Field("name").Equals("Books"))
				post1.Set("Tags", tagBooks)

				post1Tags := post1.Get("Tags").(RecordCollection)
				So(post1Tags.Len(), ShouldEqual, 1)
				So(post1Tags.Get("Name"), ShouldEqual, "Books")
				post2Tags := posts.Search(posts.Model().Field("title").Equals("2nd Post")).Get("Tags").(RecordCollection)
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
				}).(RecordCollection)
				userJane := env.Pool("User").Search(env.Pool("User").Model().Field("Email").Equals("jane.smith@example.com"))
				userJane.Set("Posts", post1.Call("Union", post3).(RecordCollection))
				So(post1.Get("User").(RecordCollection).Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post3.Get("User").(RecordCollection).Get("ID"), ShouldEqual, userJane.Get("ID"))
				So(post2.Get("User").(RecordCollection).Get("ID"), ShouldEqual, 0)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Testing access control list on update (write only)", t, func() {
		SimulateInNewEnvironment(2, func(env Environment) {
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
				profileModel.methods.MustGet("Load").AllowGroup(group1, userModel.methods.MustGet("UpdateCity"))
				profileModel.methods.MustGet("Write").AllowGroup(group1, userModel.methods.MustGet("UpdateCity"))
				jane := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("Jane A. Smith"))
				So(jane.Len(), ShouldEqual, 1)
				So(func() { jane.Call("UpdateCity", "London") }, ShouldNotPanic)
			})
			Convey("Removing Update right on Email field", func() {
				userModel.fields.MustGet("Email").RevokeAccess(security.GroupEveryone, security.Write)
				john := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
				So(john.Len(), ShouldEqual, 1)
				johnValues := FieldMap{
					"Email": "jsmith3@example.com",
					"Nums":  13,
				}
				john.Call("Write", johnValues)
				john.Load()
				So(john.Get("Name"), ShouldEqual, "John Smith")
				So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
				So(john.Get("Nums"), ShouldEqual, 13)
				userModel.fields.MustGet("Email").GrantAccess(security.GroupEveryone, security.Write)
			})
			Convey("Checking record rules", func() {
				userJane := env.Pool("User").FetchAll()
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
		})
	})
	security.Registry.UnregisterGroup(group1)
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User").Search(env.Pool("User").Model().Field("Name").Equals("John Smith"))
			num := users.Call("Unlink")
			Convey("Number of deleted record should be 1", func() {
				So(num, ShouldEqual, 1)
			})
		})
	})
	group1 := security.Registry.NewGroup("group1", "Group 1")
	security.Registry.AddMembership(2, group1)
	Convey("Checking unlink access permissions", t, func() {
		SimulateInNewEnvironment(2, func(env Environment) {
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
		})
	})
	security.Registry.UnregisterGroup(group1)
}
