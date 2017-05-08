// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataLoading(t *testing.T) {
	Convey("Testing CSV data loading into database", t, func() {
		ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
			userObj := env.Pool("User")
			Convey("Simple import of users - no update", func() {
				LoadCSVDataFile("testdata/User.csv")
				users := userObj.FetchAll()
				So(users.Len(), ShouldEqual, 5)
				userPeter := userObj.Search(userObj.Model().Field("Name").Equals("Peter"))
				So(userPeter.Get("Nums").(int), ShouldEqual, 1)
				So(userPeter.Get("IsStaff").(bool), ShouldEqual, true)
				So(userPeter.Get("Size").(float64), ShouldEqual, 1.78)
				userMary := userObj.Search(userObj.Model().Field("Name").Equals("Mary"))
				So(userMary.Get("Nums").(int), ShouldEqual, 3)
				So(userMary.Get("IsStaff").(bool), ShouldEqual, false)
				So(userMary.Get("Size").(float64), ShouldEqual, 1.59)
			})
			Convey("Check that no update does not update existing records", func() {
				userPeter := userObj.Search(userObj.Model().Field("Name").Equals("Peter")).Fetch()
				userPeter.Set("Name", "Peter Modified")
				userPeter.Load()
				So(userPeter.Get("Name"), ShouldEqual, "Peter Modified")
				LoadCSVDataFile("testdata/User.csv")
				userPeter.Load()
				So(userPeter.Get("Name"), ShouldEqual, "Peter Modified")
			})
			Convey("Check that import with update updates even existing", func() {
				LoadCSVDataFile("testdata/User_update.csv")
				users := userObj.FetchAll()
				So(users.Len(), ShouldEqual, 6)
				userPeter := userObj.Search(userObj.Model().Field("Name").Equals("Peter"))
				So(userPeter.Get("Nums").(int), ShouldEqual, 2)
				So(userPeter.Get("IsStaff").(bool), ShouldEqual, true)
				So(userPeter.Get("Size").(float64), ShouldEqual, 1.78)
				userMary := userObj.Search(userObj.Model().Field("Name").Equals("Mary"))
				So(userMary.Get("Nums").(int), ShouldEqual, 5)
				So(userMary.Get("IsStaff").(bool), ShouldEqual, false)
				So(userMary.Get("Size").(float64), ShouldEqual, 1.59)
				userNick := userObj.Search(userObj.Model().Field("Name").Equals("Nick"))
				So(userNick.Get("Nums").(int), ShouldEqual, 8)
				So(userNick.Get("IsStaff").(bool), ShouldEqual, true)
				So(userNick.Get("Size").(float64), ShouldEqual, 1.85)
			})
			Convey("Checking imports with foreign keys", func() {
				LoadCSVDataFile("testdata/Tag.csv")
				LoadCSVDataFile("testdata/Post.csv")
				userPeter := userObj.Search(userObj.Model().Field("Name").Equals("Peter"))
				So(userPeter.Get("Posts").(RecordCollection).Len(), ShouldEqual, 1)
				peterPost := userPeter.Get("Posts").(RecordCollection)
				So(peterPost.Get("Title"), ShouldEqual, "Peter's Post")
				So(peterPost.Get("Content"), ShouldEqual, "This is peter's post content")
				So(peterPost.Get("Tags").(RecordCollection).Len(), ShouldEqual, 2)
				userNick := userObj.Search(userObj.Model().Field("Name").Equals("Nick"))
				So(userNick.Get("Posts").(RecordCollection).Len(), ShouldEqual, 1)
				nickPost := userNick.Get("Posts").(RecordCollection)
				So(nickPost.Get("Title"), ShouldEqual, "Nick's Post")
				So(nickPost.Get("Content"), ShouldEqual, "No content")
				So(nickPost.Get("Tags").(RecordCollection).Len(), ShouldEqual, 3)
			})
		})
	})
}
