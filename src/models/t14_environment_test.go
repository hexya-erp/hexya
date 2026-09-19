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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/lib/pq"
)

func TestEnvironment(t *testing.T) {
	t.Run("Testing Environment Modifications", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			env.context = types.NewContext().WithKey("key", "context value")
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			t.Run("Checking WithEnv", func(t *testing.T) {
				env2 := newEnvironment(2)
				userJane1 := userJane.Call("WithEnv", env2).(RecordSet).Collection()
				assert.EqualValues(t, userJane1.Env().Uid(), 2)
				assert.EqualValues(t, userJane.Env().Uid(), 1)
				assert.True(t, userJane.Env().Context().HasKey("key"))
				assert.True(t, userJane1.Env().Context().IsEmpty())
				env2.rollback()
			})
			t.Run("Checking WithContext", func(t *testing.T) {
				userJane1 := userJane.Call("WithContext", "newKey", "This is a different key").(RecordSet).Collection()
				assert.True(t, userJane1.Env().Context().HasKey("key"))
				assert.True(t, userJane1.Env().Context().HasKey("newKey"))
				assert.EqualValues(t, userJane1.Env().Context().Get("key"), "context value")
				assert.EqualValues(t, userJane1.Env().Context().Get("newKey"), "This is a different key")
				assert.EqualValues(t, userJane1.Env().Uid(), security.SuperUserID)
				assert.True(t, userJane.Env().Context().HasKey("key"))
				assert.False(t, userJane.Env().Context().HasKey("newKey"))
				assert.EqualValues(t, userJane.Env().Context().Get("key"), "context value")
				assert.EqualValues(t, userJane.Env().Uid(), security.SuperUserID)
			})
			t.Run("Checking WithNewContext", func(t *testing.T) {
				newCtx := types.NewContext().WithKey("newKey", "This is a different key")
				userJane1 := userJane.Call("WithNewContext", newCtx).(RecordSet).Collection()
				assert.False(t, userJane1.Env().Context().HasKey("key"))
				assert.True(t, userJane1.Env().Context().HasKey("newKey"))
				assert.EqualValues(t, userJane1.Env().Context().Get("newKey"), "This is a different key")
				assert.EqualValues(t, userJane1.Env().Uid(), security.SuperUserID)
				assert.True(t, userJane.Env().Context().HasKey("key"))
				assert.False(t, userJane.Env().Context().HasKey("newKey"))
				assert.EqualValues(t, userJane.Env().Context().Get("key"), "context value")
				assert.EqualValues(t, userJane.Env().Uid(), security.SuperUserID)
			})
			t.Run("Checking Sudo", func(t *testing.T) {
				userJane1 := userJane.Sudo(2)
				userJane2 := userJane1.Call("Sudo").(RecordSet).Collection()
				assert.EqualValues(t, userJane1.Env().Uid(), 2)
				assert.EqualValues(t, userJane.Env().Uid(), security.SuperUserID)
				assert.EqualValues(t, userJane2.Env().Uid(), security.SuperUserID)
			})
			t.Run("Checking combined modifications", func(t *testing.T) {
				userJane1 := userJane.Sudo(2)
				userJane2 := userJane1.Sudo()
				userJane = userJane.WithContext("key", "modified value")
				assert.EqualValues(t, userJane.Env().Context().Get("key"), "modified value")
				assert.EqualValues(t, userJane1.Env().Context().Get("key"), "context value")
				assert.EqualValues(t, userJane1.Env().Uid(), 2)
				assert.EqualValues(t, userJane2.Env().Context().Get("key"), "context value")
				assert.EqualValues(t, userJane2.Env().Uid(), security.SuperUserID)
			})
			t.Run("Checking overridden WithContext", func(t *testing.T) {
				allPosts := env.Pool("Post").SearchAll()
				posts1 := allPosts.WithContext("foo", "bar")
				assert.True(t, posts1.Env().Context().HasKey("foo"))
				assert.EqualValues(t, posts1.Env().Context().GetString("foo"), "bar")
				assert.False(t, allPosts.Env().Context().HasKey("foo"))
			})
		}))
	})
	t.Run("Testing cache operation", func(t *testing.T) {
		t.Run("Cache should be empty at startup", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				assert.Empty(t, env.cache.data)
				assert.Empty(t, env.cache.m2mLinks)
			}))
		})
		t.Run("Loading a RecordSet should populate the cache", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Load()
				assert.Empty(t, env.cache.m2mLinks)
				assert.Len(t, env.cache.data, 4)
				assert.Contains(t, env.cache.data, users.model.name)
				assert.Contains(t, env.cache.data[users.model.name], userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				assert.Contains(t, janeEntry, "id")
				assert.EqualValues(t, janeEntry["id"], userJane.ids[0])
				assert.Contains(t, janeEntry, "name")
				assert.EqualValues(t, janeEntry["name"], "Jane A. Smith")
				assert.Contains(t, janeEntry, "email")
				assert.EqualValues(t, janeEntry["email"], "jane.smith@example.com")
				assert.True(t, env.cache.checkIfInCache(users.model, userJane.ids, []string{"id", "name", "email"}, "", false))
			}))
		})
		t.Run("Calling values already in cache should not call the DB", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Load()
				id, dbCalled := userJane.get(ID, true)
				assert.False(t, dbCalled)
				assert.EqualValues(t, id, userJane.ids[0])
				name, dbCalled := userJane.get(Name, true)
				assert.False(t, dbCalled)
				assert.EqualValues(t, name, "Jane A. Smith")
				mail, dbCalled := userJane.get(email, true)
				assert.False(t, dbCalled)
				assert.EqualValues(t, mail, "jane.smith@example.com")
			}))
		})
		t.Run("Testing O2M fields in cache", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Load(posts)
				postModel := env.Pool("Post").Model()
				post1 := env.Pool("Post").Search(postModel.Field(title).Equals("1st Post"))
				post3 := env.Pool("Post").Search(postModel.Field(title).Equals("3rd Post"))
				posts13 := post1.Union(post3)
				assert.Contains(t, env.cache.data, users.model.name)
				assert.Contains(t, env.cache.data[users.model.name], userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				assert.Contains(t, janeEntry, "posts_ids")
				assert.EqualValues(t, janeEntry["posts_ids"], true)
				assert.Len(t, env.cache.get(userJane.model, userJane.ids[0], "posts_ids", ""), posts13.Len())
				for _, id := range posts13.ids {
					assert.Contains(t, env.cache.get(userJane.model, userJane.ids[0], "posts_ids", ""), id)
				}
				assert.EqualValues(t, userJane.Get(posts).(RecordSet).Collection().Len(), 2)
			}))
		})
		t.Run("Creating an extra post should update jane's posts", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Load(posts)
				assert.EqualValues(t, userJane.Get(posts).(RecordSet).Collection().Len(), 2)
				env.Pool("Post").Call("Create", NewModelData(Registry.MustGet("Post")).
					Set(title, "Extra Post").
					Set(user, userJane))
				assert.EqualValues(t, userJane.Get(posts).(RecordSet).Collection().Len(), 3)
			}))
		})
		t.Run("Reading M2M fields should work both ways", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				postModel := env.Pool("Post").Model()
				tagModel := env.Pool("Tag").Model()
				post2 := env.Pool("Post").Search(postModel.Field(title).Equals("2nd Post"))
				assert.Empty(t, env.cache.m2mLinks)
				post2.Load()
				assert.EqualValues(t, post2.Len(), 1)
				assert.Len(t, env.cache.data, 2)
				assert.Len(t, env.cache.data["Post"], 1)
				bjTags := env.Pool("Tag").Search(tagModel.Field(Name).In([]string{"Books", "Jane's"}))
				bjTags.Fetch()
				assert.EqualValues(t, bjTags.Len(), 2)
				post2.Load(tags)
				assert.Len(t, env.cache.m2mLinks, 1)
				assert.Contains(t, env.cache.m2mLinks, "PostTagRel")
				assert.Contains(t, env.cache.m2mLinks["PostTagRel"], [2]int64{post2.ids[0], bjTags.ids[0]})
				assert.Contains(t, env.cache.m2mLinks["PostTagRel"], [2]int64{post2.ids[0], bjTags.ids[1]})
				assert.Len(t, env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), 2)
				assert.Contains(t, env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), bjTags.ids[0])
				assert.Contains(t, env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), bjTags.ids[1])
				assert.Len(t, env.cache.get(tagModel, bjTags.ids[0], posts.JSON(), ""), 1)
				assert.Contains(t, env.cache.get(tagModel, bjTags.ids[0], posts.JSON(), ""), post2.ids[0])
				assert.Len(t, env.cache.get(tagModel, bjTags.ids[1], posts.JSON(), ""), 1)
				assert.Contains(t, env.cache.get(tagModel, bjTags.ids[1], posts.JSON(), ""), post2.ids[0])
				assert.True(t, env.cache.checkIfInCache(postModel, post2.ids, []string{tags.JSON()}, "", false))
				assert.Len(t, post2.Get(tags).(RecordSet).Collection().Ids(), 2)
				assert.Contains(t, post2.Get(tags).(RecordSet).Collection().Ids(), bjTags.ids[0])
				assert.Contains(t, post2.Get(tags).(RecordSet).Collection().Ids(), bjTags.ids[1])
				assert.Len(t, bjTags.Records()[0].Get(posts).(RecordSet).Collection().Ids(), 1)
				assert.Contains(t, bjTags.Records()[0].Get(posts).(RecordSet).Collection().Ids(), post2.ids[0])
				assert.Len(t, bjTags.Records()[1].Get(posts).(RecordSet).Collection().Ids(), 2)
				assert.Contains(t, bjTags.Records()[1].Get(posts).(RecordSet).Collection().Ids(), post2.ids[0])
			}))
		})
		t.Run("Check that computed fields are not stored in cache", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				userJane.Load()
				assert.Contains(t, env.cache.data, users.model.name)
				assert.Contains(t, env.cache.data[users.model.name], userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				assert.Contains(t, janeEntry, "id")
				assert.Contains(t, janeEntry, "name")
				assert.NotContains(t, janeEntry, "decorated_name")
				userJane.Get(decoratedName)
				assert.NotContains(t, janeEntry, "decorated_name")
			}))
		})
		t.Run("Checking cache dump for debug", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
				assert.EqualValues(t, env.DumpCache(), `Data
====

M2M Links
=========

X2M Links
=========
`)
				userJane.Load()
				userJane.Load(postsTags)
				assert.Greater(t, len(env.DumpCache()), 1360)
			}))
		})
		t.Run("Check that new works correctly", func(t *testing.T) {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				users := env.Pool("User")
				userMattData := NewModelData(users.Model()).
					Set(Name, "Matt Smith").
					Set(email, "msmith@example.com").
					Set(isStaff, true).
					Set(nums, 4).
					Set(age, 23)
				userMatt := users.new(userMattData)
				assert.EqualValues(t, userMatt.Get(Name), "Matt Smith")
				assert.EqualValues(t, userMatt.Get(email), "msmith@example.com")
				assert.EqualValues(t, userMatt.Get(isStaff), true)
				assert.EqualValues(t, userMatt.Get(isPremium), false)
				assert.EqualValues(t, userMatt.Get(nums), 4)
				assert.EqualValues(t, userMatt.Get(age), 23)
			}))
		})
	})
	t.Run("Testing prefetch", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userSet := users.Search(users.Model().Field(email).Equals("jane.smith@example.com").
				Or().Field(Name).Equals("John Smith"))
			assert.False(t, userSet.fetched)
			assert.False(t, env.cache.checkIfInCache(users.Model(), userSet.ids, []string{Name.JSON()}, "", false))
			t.Run("Loading one record should load all of its original record", func(t *testing.T) {
				records := userSet.SortedByField(ID, false).Records()
				assert.True(t, userSet.fetched)
				assert.Len(t, records, 2)
				assert.True(t, records[0].fetched)
				assert.True(t, records[1].fetched)
				assert.False(t, env.cache.checkIfInCache(users.Model(), userSet.ids, []string{Name.JSON()}, "", false))
				name, fetched := records[0].get(Name, false)
				assert.EqualValues(t, name, "John Smith")
				assert.True(t, fetched)
				name2, fetched2 := records[1].get(Name, false)
				assert.EqualValues(t, name2, "Jane A. Smith")
				assert.False(t, fetched2)
			})
			t.Run("Returned recordset by Load should be the right one", func(t *testing.T) {
				records := userSet.SortedByField(ID, false).Records()
				rc := records[0].Load()
				assert.True(t, rc.Equals(records[0]))
			})
			t.Run("Nested records", func(t *testing.T) {
				records := userSet.SortedByField(ID, false).Records()
				postsJohn := records[0].Get(posts).(*RecordCollection).Records()
				postsJane := records[1].Get(posts).(*RecordCollection).Records()
				assert.Len(t, postsJohn, 0)
				assert.Len(t, postsJane, 2)
			})
		}))
	})
	t.Run("Checking error types", func(t *testing.T) {
		nice := new(notInCacheError)
		assert.EqualValues(t, nice.Error(), "requested value not in cache")
		nepe := new(nonExistentPathError)
		assert.EqualValues(t, nepe.Error(), "requested path is broken")
	})
	t.Run("Testing db error retries", func(t *testing.T) {
		t.Run("ExecuteInNewEnvironment should retry db errors up to max retries", func(t *testing.T) {
			var retries uint8
			assert.NotNil(t, doExecuteInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				panic(&pq.Error{Code: "40001"})
			}))
			assert.EqualValues(t, retries, DBSerializationMaxRetries)
		})
		t.Run("ExecuteInNewEnvironment should retry db errors and stop when ok", func(t *testing.T) {
			var retries uint8
			assert.Nil(t, doExecuteInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				if retries < 3 {
					panic(&pq.Error{Code: "40001"})
				}
			}))
			assert.EqualValues(t, retries, 3)
		})
		t.Run("SimulateInNewEnvironment should retry db errors up to max retries", func(t *testing.T) {
			var retries uint8
			assert.NotNil(t, doSimulateInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				panic(&pq.Error{Code: "40001"})
			}))
			assert.EqualValues(t, retries, DBSerializationMaxRetries)
		})
		t.Run("SimulateInNewEnvironment should retry db errors and stop when ok", func(t *testing.T) {
			var retries uint8
			assert.Nil(t, doSimulateInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				if retries < 3 {
					panic(&pq.Error{Code: "40001"})
				}
			}))
			assert.EqualValues(t, retries, 3)
		})
	})
}
