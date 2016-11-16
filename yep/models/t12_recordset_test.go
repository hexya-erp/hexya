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

	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRecordSet(t *testing.T) {
	Convey("Test record creation", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Creating simple user John with no relations and checking ID", func() {
			userJohnData := FieldMap{
				"UserName": "John Smith",
				"Email":    "jsmith@example.com",
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
			userJaneData := FieldMap{
				"UserName": "Jane Smith",
				"Email":    "jane.smith@example.com",
				"Profile":  profile,
			}
			userJane := env.Pool("User").Call("Create", userJaneData).(RecordCollection)
			So(userJane.Len(), ShouldEqual, 1)
			So(userJane.Get("Profile").(RecordCollection).Get("ID"), ShouldEqual, profile.Get("ID"))
			post1Data := FieldMap{
				"User":    userJane,
				"Title":   "1st Post",
				"Content": "Content of first post",
			}
			post1 := env.Pool("Post").Call("Create", post1Data).(RecordCollection)
			So(post1.Len(), ShouldEqual, 1)
			post2Data := FieldMap{
				"User":    userJane,
				"Title":   "2nd Post",
				"Content": "Content of second post",
			}
			post2 := env.Pool("Post").Call("Create", post2Data).(RecordCollection)
			So(post2.Len(), ShouldEqual, 1)
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
				"UserName": "Will Smith",
				"Email":    "will.smith@example.com",
			}
			userWill := env.Pool("User").Call("Create", userWillData).(RecordCollection)
			So(userWill.Len(), ShouldEqual, 1)
			So(userWill.Get("ID"), ShouldBeGreaterThan, 0)
		})
		env.Commit()
	})
	Convey("Testing access control list on creation (create only)", t, func() {
		env := NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot create records", func() {
			userTomData := FieldMap{
				"UserName": "Tom Smith",
				"Email":    "tsmith@example.com",
			}
			So(func() { env.Pool("User").Call("Create", userTomData) }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and check creation", func() {
			AllowModelAccess(ModelName("User"), group1, security.Create)
			userTomData := FieldMap{
				"UserName": "Tom Smith",
				"Email":    "tsmith@example.com",
			}
			userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
			So(func() { userTom.Get("UserName") }, ShouldPanic)
		})
		Convey("Checking creation again with read rights too", func() {
			AllowModelAccess(ModelName("User"), group1, security.Create|security.Read)
			userTomData := FieldMap{
				"UserName": "Tom Smith",
				"Email":    "tsmith@example.com",
			}
			userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
			So(userTom.Get("UserName"), ShouldEqual, "Tom Smith")
			So(userTom.Get("Email"), ShouldEqual, "tsmith@example.com")
		})
		Convey("Removing Create right on Email field", func() {
			AllowModelAccess(ModelName("User"), group1, security.Create|security.Read)
			DenyFieldAccess(ModelName("User"), FieldName("Email"), group1, security.Create)
			userTomData := FieldMap{
				"UserName": "Tom Smith",
				"Email":    "tsmith@example.com",
			}
			userTom := env.Pool("User").Call("Create", userTomData).(RecordCollection)
			So(userTom.Get("UserName"), ShouldEqual, "Tom Smith")
			So(userTom.Get("Email").(string), ShouldBeBlank)
			DenyModelAccess(ModelName("User"), group1, security.Create|security.Read)
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
		env := NewEnvironment(security.SuperUserID)
		Convey("Searching User Jane", func() {
			userJane := env.Pool("User").Filter("UserName", "=", "Jane Smith")
			So(userJane.Len(), ShouldEqual, 1)
			Convey("Reading Jane with Get", func() {
				So(userJane.Get("UserName").(string), ShouldEqual, "Jane Smith")
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
				So(userJaneStruct.UserName, ShouldEqual, "Jane Smith")
				So(userJaneStruct.Email, ShouldEqual, "jane.smith@example.com")
				So(userJaneStruct.ID, ShouldEqual, userJane.Get("ID").(int64))
			})
		})

		Convey("Testing search all users", func() {
			usersAll := env.Pool("User").Load()
			So(usersAll.Len(), ShouldEqual, 3)
			Convey("Reading first user with Get", func() {
				So(usersAll.Get("UserName"), ShouldEqual, "John Smith")
				So(usersAll.Get("Email"), ShouldEqual, "jsmith@example.com")
			})
			Convey("Reading all users with Records and Get", func() {
				recs := usersAll.Records()
				So(len(recs), ShouldEqual, 3)
				So(recs[0].Get("Email"), ShouldEqual, "jsmith@example.com")
				So(recs[1].Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(recs[2].Get("Email"), ShouldEqual, "will.smith@example.com")
			})
			Convey("Reading all users with ReadAll()", func() {
				var userStructs []*UserStruct
				usersAll.All(&userStructs)
				So(userStructs[0].Email, ShouldEqual, "jsmith@example.com")
				So(userStructs[1].Email, ShouldEqual, "jane.smith@example.com")
				So(userStructs[2].Email, ShouldEqual, "will.smith@example.com")
			})
		})
		env.Rollback()
	})
	Convey("Testing access control list while searching", t, func() {
		env := NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot access records", func() {
			userJane := env.Pool("User").Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and checking access", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read)

			userJane := env.Pool("User").Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldNotPanic)
			So(userJane.Get("UserName").(string), ShouldEqual, "Jane Smith")
			So(userJane.Get("Email").(string), ShouldEqual, "jane.smith@example.com")
			So(userJane.Get("Age"), ShouldEqual, 23)
			So(func() { userJane.Get("Profile").(RecordCollection).Get("Age") }, ShouldPanic)
		})
		Convey("Adding field access rights to user 2 and checking access", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read)
			DenyFieldAccess(ModelName("User"), FieldName("Email"), group1, security.Read)
			DenyFieldAccess(ModelName("User"), FieldName("Age"), group1, security.Read)

			userJane := env.Pool("User").Filter("UserName", "=", "Jane Smith")
			So(func() { userJane.Load() }, ShouldNotPanic)
			So(userJane.Get("UserName").(string), ShouldEqual, "Jane Smith")
			So(userJane.Get("Email").(string), ShouldBeBlank)
			So(userJane.Get("Age"), ShouldEqual, 0)
		})
		env.Rollback()
	})
}

func TestUpdateRecordSet(t *testing.T) {
	Convey("Testing updates through RecordSets", t, func() {
		env := NewEnvironment(security.SuperUserID)
		Convey("Update on users Jane and John with Write and Set", func() {
			jane := env.Pool("User").Filter("UserName", "=", "Jane Smith")
			So(jane.Len(), ShouldEqual, 1)
			jane.Set("UserName", "Jane A. Smith")
			jane.Load()
			So(jane.Get("UserName"), ShouldEqual, "Jane A. Smith")
			So(jane.Get("Email"), ShouldEqual, "jane.smith@example.com")

			john := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := FieldMap{
				"Email": "jsmith2@example.com",
				"Nums":  13,
			}
			john.Call("Write", johnValues)
			john.Load()
			So(john.Get("UserName"), ShouldEqual, "John Smith")
			So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
			So(john.Get("Nums"), ShouldEqual, 13)
		})
		Convey("Multiple updates at once on users", func() {
			cond := NewCondition().And("UserName", "=", "Jane A. Smith").Or("UserName", "=", "John Smith")
			users := env.Pool("User").Search(cond).Load()
			So(users.Len(), ShouldEqual, 2)
			userRecs := users.Records()
			So(userRecs[0].Get("IsStaff").(bool), ShouldBeFalse)
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
		Convey("Updating many2many fields", func() {
			post1 := env.Pool("Post").Filter("title", "=", "1st Post")
			tagBooks := env.Pool("Tag").Filter("name", "=", "Books")
			post1.Set("Tags", tagBooks)

			post1Tags := post1.Get("Tags").(RecordCollection)
			So(post1Tags.Len(), ShouldEqual, 1)
			So(post1Tags.Get("Name"), ShouldEqual, "Books")
			post2Tags := env.Pool("Post").Filter("title", "=", "2nd Post").Get("Tags").(RecordCollection)
			So(post2Tags.Len(), ShouldEqual, 2)
			So(post2Tags.Records()[0].Get("Name"), ShouldBeIn, "Books", "Jane's")
			So(post2Tags.Records()[1].Get("Name"), ShouldBeIn, "Books", "Jane's")
		})
		env.Commit()
	})
	Convey("Testing access control list on update (write only)", t, func() {
		env := NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot update records", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read)
			john := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := FieldMap{
				"Email": "jsmith3@example.com",
				"Nums":  13,
			}
			So(func() { john.Call("Write", johnValues) }, ShouldPanic)
		})
		Convey("Adding model access rights to user 2 and check update", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read|security.Write)
			john := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := FieldMap{
				"Email": "jsmith3@example.com",
				"Nums":  13,
			}
			john.Call("Write", johnValues)
			john.Load()
			So(john.Get("UserName"), ShouldEqual, "John Smith")
			So(john.Get("Email"), ShouldEqual, "jsmith3@example.com")
			So(john.Get("Nums"), ShouldEqual, 13)
		})
		Convey("Removing Update right on Email field", func() {
			AllowModelAccess(ModelName("User"), group1, security.Write|security.Read)
			DenyFieldAccess(ModelName("User"), FieldName("Email"), group1, security.Write)
			john := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(john.Len(), ShouldEqual, 1)
			johnValues := FieldMap{
				"Email": "jsmith3@example.com",
				"Nums":  13,
			}
			john.Call("Write", johnValues)
			john.Load()
			So(john.Get("UserName"), ShouldEqual, "John Smith")
			So(john.Get("Email"), ShouldEqual, "jsmith2@example.com")
			So(john.Get("Nums"), ShouldEqual, 13)
		})
		env.Rollback()
	})
}

func TestDeleteRecordSet(t *testing.T) {
	Convey("Delete user John Smith", t, func() {
		env := NewEnvironment(security.SuperUserID)
		users := env.Pool("User").Filter("UserName", "=", "John Smith")
		num := users.Call("Unlink")
		Convey("Number of deleted record should be 1", func() {
			So(num, ShouldEqual, 1)
		})
		env.Rollback()
	})
	Convey("Checking unlink access permissions", t, func() {
		env := NewEnvironment(2)
		group1 := security.NewGroup("Group1")
		gmBackend := make(security.GroupMapBackend)
		security.AuthenticationRegistry.RegisterBackend(gmBackend)
		gmBackend[2] = []*security.Group{group1}

		Convey("Checking that user 2 cannot delete records", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read)
			users := env.Pool("User").Filter("UserName", "=", "John Smith")
			So(func() { users.Call("Unlink") }, ShouldPanic)
		})
		Convey("Adding unlink permission to user2", func() {
			AllowModelAccess(ModelName("User"), group1, security.Read|security.Unlink)
			users := env.Pool("User").Filter("UserName", "=", "John Smith")
			num := users.Call("Unlink")
			So(num, ShouldEqual, 1)
		})
		env.Rollback()
	})
}
