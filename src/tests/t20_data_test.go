// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func TestDataLoading(t *testing.T) {
	t.Run("Testing CSV data loading into database", func(t *testing.T) {
		t.Run("Simple import of users - no update", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/User.csv")
				users := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, users.Len(), 5)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				assert.EqualValues(t, userPeter.Nums(), 1)
				assert.EqualValues(t, userPeter.IsStaff(), true)
				assert.EqualValues(t, userPeter.Size(), 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary"))
				assert.EqualValues(t, userMary.Nums(), 3)
				assert.EqualValues(t, userMary.IsStaff(), false)
				assert.EqualValues(t, userMary.Size(), 1.59)

				assert.Panics(t, func() { models.LoadCSVDataFile("testdata/001User.csv") })
				assert.Panics(t, func() { models.LoadCSVDataFile("testdata/011User.csv") })
				assert.Panics(t, func() { models.LoadCSVDataFile("testdata/012User.csv") })
			}))
		})
		t.Run("Check that no update does not update existing records", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter")).Fetch()
				userPeter.SetName("Peter Modified")
				userPeter.Load()
				assert.EqualValues(t, userPeter.Name(), "Peter Modified")
				models.LoadCSVDataFile("testdata/User.csv")
				userPeter.Load()
				assert.EqualValues(t, userPeter.Name(), "Peter Modified")
			}))
		})
		t.Run("Check that import with update updates even existing", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/200User_update.csv")
				users := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, users.Len(), 6)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				assert.EqualValues(t, userPeter.Nums(), 2)
				assert.EqualValues(t, userPeter.IsStaff(), true)
				assert.EqualValues(t, userPeter.Size(), 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary"))
				assert.EqualValues(t, userMary.Nums(), 5)
				assert.EqualValues(t, userMary.IsStaff(), false)
				assert.EqualValues(t, userMary.Size(), 1.59)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				assert.EqualValues(t, userNick.Nums(), 8)
				assert.EqualValues(t, userNick.IsStaff(), true)
				assert.EqualValues(t, userNick.Size(), 1.85)
			}))
		})
		t.Run("Checking import with future version", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/User_12.csv")
				users := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, users.Len(), 7)
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				assert.EqualValues(t, userPeter.HexyaVersion(), 0)
				assert.EqualValues(t, userPeter.Nums(), 2)
				assert.EqualValues(t, userPeter.IsStaff(), true)
				assert.EqualValues(t, userPeter.Size(), 1.78)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary modified"))
				assert.EqualValues(t, userMary.HexyaVersion(), 12)
				assert.EqualValues(t, userMary.Nums(), 5)
				assert.EqualValues(t, userMary.IsStaff(), false)
				assert.EqualValues(t, userMary.Size(), 1.58)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				assert.EqualValues(t, userNick.HexyaVersion(), 0)
				assert.EqualValues(t, userNick.Nums(), 8)
				assert.EqualValues(t, userNick.IsStaff(), true)
				assert.EqualValues(t, userNick.Size(), 1.85)
				userRob := h.User().Search(env, q.User().Name().Equals("Rob"))
				assert.EqualValues(t, userRob.HexyaVersion(), 12)
				assert.EqualValues(t, userRob.Nums(), 14)
				assert.EqualValues(t, userRob.IsStaff(), false)
				assert.EqualValues(t, userRob.Size(), 1.81)
			}))
		})
		t.Run("Checking import with past version", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/User_2.csv")
				users := h.User().NewSet(env).SearchAll()
				assert.EqualValues(t, users.Len(), 8)
				userMary := h.User().Search(env, q.User().Name().Equals("Mary modified"))
				assert.EqualValues(t, userMary.HexyaVersion(), 12)
				assert.EqualValues(t, userMary.Nums(), 5)
				assert.EqualValues(t, userMary.IsStaff(), false)
				assert.EqualValues(t, userMary.Size(), 1.58)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				assert.EqualValues(t, userNick.HexyaVersion(), 2)
				assert.EqualValues(t, userNick.Nums(), 54)
				assert.EqualValues(t, userNick.IsStaff(), true)
				assert.EqualValues(t, userNick.Size(), 1.86)
				userKen := h.User().Search(env, q.User().Name().Equals("Ken"))
				assert.EqualValues(t, userKen.HexyaVersion(), 2)
				assert.EqualValues(t, userKen.Nums(), 10)
				assert.EqualValues(t, userKen.IsStaff(), false)
				assert.EqualValues(t, userKen.Size(), 1.76)
			}))
		})
		t.Run("Test with contexted on embedded field", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/013User.csv")
				userPete := h.User().Search(env, q.User().Email().Equals("peter@hexya.io"))
				assert.EqualValues(t, userPete.Education(), "Hexya University")
			}))
		})
		t.Run("Checking imports with foreign keys", func(t *testing.T) {
			assert.Nil(t, models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				models.LoadCSVDataFile("testdata/010-Tag.csv")
				models.LoadCSVDataFile("testdata/Post.csv")
				userPeter := h.User().Search(env, q.User().Name().Equals("Peter"))
				assert.EqualValues(t, userPeter.Posts().Len(), 1)
				peterPost := userPeter.Posts()
				assert.EqualValues(t, peterPost.Title(), "Peter's Post")
				assert.EqualValues(t, peterPost.Content(), "This is peter's post content")
				assert.EqualValues(t, peterPost.Tags().Len(), 2)
				userNick := h.User().Search(env, q.User().Name().Equals("Nick"))
				assert.EqualValues(t, userNick.Posts().Len(), 1)
				nickPost := userNick.Posts()
				assert.EqualValues(t, nickPost.Title(), "Nick's Post")
				assert.EqualValues(t, nickPost.Content(), "No content")
				assert.EqualValues(t, nickPost.Tags().Len(), 3)

				assert.Panics(t, func() { models.LoadCSVDataFile("testdata/001Post.csv") })
				assert.Panics(t, func() { models.LoadCSVDataFile("testdata/002Post.csv") })
			}))
		})
	})
}
