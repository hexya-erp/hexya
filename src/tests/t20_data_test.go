// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDataLoading(t *testing.T) {
	Convey("Testing CSV data loading into database", t, func() {
		So(models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Simple import of users - no update", func() {
				models.LoadCSVDataFile("testdata/User.csv")
				users := h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 5)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				So(userPeter.Nums(), ShouldEqual, 1)
				So(userPeter.IsStaff(), ShouldEqual, true)
				So(userPeter.Size(), ShouldEqual, 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary"))
				So(userMary.Nums(), ShouldEqual, 3)
				So(userMary.IsStaff(), ShouldEqual, false)
				So(userMary.Size(), ShouldEqual, 1.59)

				So(func() { models.LoadCSVDataFile("testdata/001User.csv") }, ShouldPanic)
				So(func() { models.LoadCSVDataFile("testdata/011User.csv") }, ShouldPanic)
				So(func() { models.LoadCSVDataFile("testdata/012User.csv") }, ShouldPanic)
			})
			Convey("Check that no update does not update existing records", func() {
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter")).Fetch()
				userPeter.SetName("Peter Modified")
				userPeter.Load()
				So(userPeter.Name(), ShouldEqual, "Peter Modified")
				models.LoadCSVDataFile("testdata/User.csv")
				userPeter.Load()
				So(userPeter.Name(), ShouldEqual, "Peter Modified")
			})
			Convey("Check that import with update updates even existing", func() {
				models.LoadCSVDataFile("testdata/200User_update.csv")
				users := h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 6)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				So(userPeter.Nums(), ShouldEqual, 2)
				So(userPeter.IsStaff(), ShouldEqual, true)
				So(userPeter.Size(), ShouldEqual, 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary"))
				So(userMary.Nums(), ShouldEqual, 5)
				So(userMary.IsStaff(), ShouldEqual, false)
				So(userMary.Size(), ShouldEqual, 1.59)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				So(userNick.Nums(), ShouldEqual, 8)
				So(userNick.IsStaff(), ShouldEqual, true)
				So(userNick.Size(), ShouldEqual, 1.85)
			})
			Convey("Checking import with future version", func() {
				models.LoadCSVDataFile("testdata/User_12.csv")
				users := h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 7)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				So(userPeter.HexyaVersion(), ShouldEqual, 0)
				So(userPeter.Nums(), ShouldEqual, 2)
				So(userPeter.IsStaff(), ShouldEqual, true)
				So(userPeter.Size(), ShouldEqual, 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary modified"))
				So(userMary.HexyaVersion(), ShouldEqual, 12)
				So(userMary.Nums(), ShouldEqual, 5)
				So(userMary.IsStaff(), ShouldEqual, false)
				So(userMary.Size(), ShouldEqual, 1.58)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				So(userNick.HexyaVersion(), ShouldEqual, 0)
				So(userNick.Nums(), ShouldEqual, 8)
				So(userNick.IsStaff(), ShouldEqual, true)
				So(userNick.Size(), ShouldEqual, 1.85)
				userRob := h.User().Search(env, q.User().Name().Equals("Rob"))
				So(userRob.HexyaVersion(), ShouldEqual, 12)
				So(userRob.Nums(), ShouldEqual, 14)
				So(userRob.IsStaff(), ShouldEqual, false)
				So(userRob.Size(), ShouldEqual, 1.81)
			})
			Convey("Checking import with past version", func() {
				models.LoadCSVDataFile("testdata/User_2.csv")
				users := h.User().NewSet(env).SearchAll()
				So(users.Len(), ShouldEqual, 8)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary modified"))
				So(userMary.HexyaVersion(), ShouldEqual, 12)
				So(userMary.Nums(), ShouldEqual, 5)
				So(userMary.IsStaff(), ShouldEqual, false)
				So(userMary.Size(), ShouldEqual, 1.58)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				So(userNick.HexyaVersion(), ShouldEqual, 2)
				So(userNick.Nums(), ShouldEqual, 54)
				So(userNick.IsStaff(), ShouldEqual, true)
				So(userNick.Size(), ShouldEqual, 1.86)
				userKen := h.User().Search(env, q.User().Name().Equals("Ken"))
				So(userKen.HexyaVersion(), ShouldEqual, 2)
				So(userKen.Nums(), ShouldEqual, 10)
				So(userKen.IsStaff(), ShouldEqual, false)
				So(userKen.Size(), ShouldEqual, 1.76)
			})
			Convey("Test with contexted on embedded field", func() {
				models.LoadCSVDataFile("testdata/013User.csv")
				userPete := h.User().Search(env, q.User().Email().Equals("peter@hexya.io"))
				So(userPete.Education(), ShouldEqual, "Hexya University")
			})
			Convey("Checking imports with foreign keys", func() {
				models.LoadCSVDataFile("testdata/010-Tag.csv")
				models.LoadCSVDataFile("testdata/Post.csv")
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				So(userPeter.Posts().Len(), ShouldEqual, 1)
				peterPost := userPeter.Posts()
				So(peterPost.Title(), ShouldEqual, "Peter's Post")
				So(peterPost.Content(), ShouldEqual, "This is peter's post content")
				So(peterPost.Tags().Len(), ShouldEqual, 2)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				So(userNick.Posts().Len(), ShouldEqual, 1)
				nickPost := userNick.Posts()
				So(nickPost.Title(), ShouldEqual, "Nick's Post")
				So(nickPost.Content(), ShouldEqual, "No content")
				So(nickPost.Tags().Len(), ShouldEqual, 3)

				So(func() { models.LoadCSVDataFile("testdata/001Post.csv") }, ShouldPanic)
				So(func() { models.LoadCSVDataFile("testdata/002Post.csv") }, ShouldPanic)
			})
		}), ShouldBeNil)
	})
}
