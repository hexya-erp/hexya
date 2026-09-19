// Copyright 2017 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
)

func TestDataLoading(t *testing.T) {
	t.Run("Testing CSV data loading into database", func(t *testing.T) {
		t.Run("Simple import of users - no update", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/User.csv")
				users := userObj.SearchAll()
				assert.EqualValues(t, users.Len(), 5)
				userPeter := userObj.Search(userObj.Model().Field(Name).Equals("Peter"))
				assert.EqualValues(t, userPeter.Get(nums).(int), 1)
				assert.EqualValues(t, userPeter.Get(isStaff).(bool), true)
				assert.EqualValues(t, userPeter.Get(size).(float64), 1.78)
				userMary := userObj.Search(userObj.Model().Field(Name).Equals("Mary"))
				assert.EqualValues(t, userMary.Get(nums).(int), 3)
				assert.EqualValues(t, userMary.Get(isStaff).(bool), false)
				assert.EqualValues(t, userMary.Get(size).(float64), 1.59)

				assert.Panics(t, func() { LoadCSVDataFile("testdata/001User.csv") })
				assert.Panics(t, func() { LoadCSVDataFile("testdata/011User.csv") })
				assert.Panics(t, func() { LoadCSVDataFile("testdata/012User.csv") })
			}))
		})
		t.Run("Check that no update does not update existing records", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				userPeter := userObj.Search(userObj.Model().Field(Name).Equals("Peter")).Fetch()
				userPeter.Set(Name, "Peter Modified")
				userPeter.Load()
				assert.EqualValues(t, userPeter.Get(Name), "Peter Modified")
				LoadCSVDataFile("testdata/User.csv")
				userPeter.Load()
				assert.EqualValues(t, userPeter.Get(Name), "Peter Modified")
			}))
		})
		t.Run("Check that import with update updates even existing", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/200User_update.csv")
				users := userObj.SearchAll()
				assert.EqualValues(t, users.Len(), 6)
				userPeter := userObj.Search(userObj.Model().Field(Name).Equals("Peter"))
				assert.EqualValues(t, userPeter.Get(nums).(int), 2)
				assert.EqualValues(t, userPeter.Get(isStaff).(bool), true)
				assert.EqualValues(t, userPeter.Get(size).(float64), 1.78)
				userMary := userObj.Search(userObj.Model().Field(Name).Equals("Mary"))
				assert.EqualValues(t, userMary.Get(nums).(int), 5)
				assert.EqualValues(t, userMary.Get(isStaff).(bool), false)
				assert.EqualValues(t, userMary.Get(size).(float64), 1.59)
				userNick := userObj.Search(userObj.Model().Field(Name).Equals("Nick"))
				assert.EqualValues(t, userNick.Get(nums).(int), 8)
				assert.EqualValues(t, userNick.Get(isStaff).(bool), true)
				assert.EqualValues(t, userNick.Get(size).(float64), 1.85)
			}))
		})
		t.Run("Checking import with future version", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/User_12.csv")
				users := userObj.SearchAll()
				assert.EqualValues(t, users.Len(), 7)
				userPeter := userObj.Search(userObj.Model().Field(Name).Equals("Peter"))
				assert.EqualValues(t, userPeter.Get(hexyaVersion).(int), 0)
				assert.EqualValues(t, userPeter.Get(nums).(int), 2)
				assert.EqualValues(t, userPeter.Get(isStaff).(bool), true)
				assert.EqualValues(t, userPeter.Get(size).(float64), 1.78)
				userMary := userObj.Search(userObj.Model().Field(Name).Equals("Mary modified"))
				assert.EqualValues(t, userMary.Get(hexyaVersion).(int), 12)
				assert.EqualValues(t, userMary.Get(nums).(int), 5)
				assert.EqualValues(t, userMary.Get(isStaff).(bool), false)
				assert.EqualValues(t, userMary.Get(size).(float64), 1.58)
				userNick := userObj.Search(userObj.Model().Field(Name).Equals("Nick"))
				assert.EqualValues(t, userNick.Get(hexyaVersion).(int), 0)
				assert.EqualValues(t, userNick.Get(nums).(int), 8)
				assert.EqualValues(t, userNick.Get(isStaff).(bool), true)
				assert.EqualValues(t, userNick.Get(size).(float64), 1.85)
				userRob := userObj.Search(userObj.Model().Field(Name).Equals("Rob"))
				assert.EqualValues(t, userRob.Get(hexyaVersion).(int), 12)
				assert.EqualValues(t, userRob.Get(nums).(int), 14)
				assert.EqualValues(t, userRob.Get(isStaff).(bool), false)
				assert.EqualValues(t, userRob.Get(size).(float64), 1.81)
			}))
		})
		t.Run("Checking import with past version", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/User_2.csv")
				users := userObj.SearchAll()
				assert.EqualValues(t, users.Len(), 8)
				userMary := userObj.Search(userObj.Model().Field(Name).Equals("Mary modified"))
				assert.EqualValues(t, userMary.Get(hexyaVersion).(int), 12)
				assert.EqualValues(t, userMary.Get(nums).(int), 5)
				assert.EqualValues(t, userMary.Get(isStaff).(bool), false)
				assert.EqualValues(t, userMary.Get(size).(float64), 1.58)
				userNick := userObj.Search(userObj.Model().Field(Name).Equals("Nick"))
				assert.EqualValues(t, userNick.Get(hexyaVersion).(int), 2)
				assert.EqualValues(t, userNick.Get(nums).(int), 54)
				assert.EqualValues(t, userNick.Get(isStaff).(bool), true)
				assert.EqualValues(t, userNick.Get(size).(float64), 1.86)
				userKen := userObj.Search(userObj.Model().Field(Name).Equals("Ken"))
				assert.EqualValues(t, userKen.Get(hexyaVersion).(int), 2)
				assert.EqualValues(t, userKen.Get(nums).(int), 10)
				assert.EqualValues(t, userKen.Get(isStaff).(bool), false)
				assert.EqualValues(t, userKen.Get(size).(float64), 1.76)
			}))
		})
		t.Run("Test with contexted on embedded field", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/013User.csv")
				peter := userObj.Search(userObj.Model().Field(email).Equals("peter@hexya.io"))
				assert.EqualValues(t, peter.Get(education), "Hexya University")
			}))
		})
		t.Run("Checking imports with foreign keys", func(t *testing.T) {
			assert.Nil(t, ExecuteInNewEnvironment(security.SuperUserID, func(env Environment) {
				userObj := env.Pool("User")
				LoadCSVDataFile("testdata/010-Tag.csv")
				LoadCSVDataFile("testdata/Post.csv")
				userPeter := userObj.Search(userObj.Model().Field(Name).Equals("Peter"))
				assert.EqualValues(t, userPeter.Get(posts).(RecordSet).Collection().Len(), 1)
				peterPost := userPeter.Get(posts).(RecordSet).Collection()
				assert.EqualValues(t, peterPost.Get(title), "Peter's Post")
				assert.EqualValues(t, peterPost.Get(content), "This is peter's post content")
				assert.EqualValues(t, peterPost.Get(tags).(RecordSet).Collection().Len(), 2)
				userNick := userObj.Search(userObj.Model().Field(Name).Equals("Nick"))
				assert.EqualValues(t, userNick.Get(posts).(RecordSet).Collection().Len(), 1)
				nickPost := userNick.Get(posts).(RecordSet).Collection()
				assert.EqualValues(t, nickPost.Get(title), "Nick's Post")
				assert.EqualValues(t, nickPost.Get(content), "No content")
				assert.EqualValues(t, nickPost.Get(tags).(RecordSet).Collection().Len(), 3)

				assert.Panics(t, func() { LoadCSVDataFile("testdata/001Post.csv") })
				assert.Panics(t, func() { LoadCSVDataFile("testdata/002Post.csv") })
			}))
		})
	})
}
