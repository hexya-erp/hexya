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
	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnvironment(t *testing.T) {
	Convey("Testing Environment Modifications", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			env.context = types.NewContext().WithKey("key", "context value")
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			Convey("Checking WithEnv", func() {
				env2 := newEnvironment(2)
				userJane1 := userJane.Call("WithEnv", env2).(RecordSet).Collection()
				So(userJane1.Env().Uid(), ShouldEqual, 2)
				So(userJane.Env().Uid(), ShouldEqual, 1)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane1.Env().Context().IsEmpty(), ShouldBeTrue)
				env2.rollback()
			})
			Convey("Checking WithContext", func() {
				userJane1 := userJane.Call("WithContext", "newKey", "This is a different key").(RecordSet).Collection()
				So(userJane1.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
				So(userJane1.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
				So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
				So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
			})
			Convey("Checking WithNewContext", func() {
				newCtx := types.NewContext().WithKey("newKey", "This is a different key")
				userJane1 := userJane.Call("WithNewContext", newCtx).(RecordSet).Collection()
				So(userJane1.Env().Context().HasKey("key"), ShouldBeFalse)
				So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
				So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
				So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
				So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
			})
			Convey("Checking Sudo", func() {
				userJane1 := userJane.Sudo(2)
				userJane2 := userJane1.Call("Sudo").(RecordSet).Collection()
				So(userJane1.Env().Uid(), ShouldEqual, 2)
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane2.Env().Uid(), ShouldEqual, security.SuperUserID)
			})
			Convey("Checking combined modifications", func() {
				userJane1 := userJane.Sudo(2)
				userJane2 := userJane1.Sudo()
				userJane = userJane.WithContext("key", "modified value")
				So(userJane.Env().Context().Get("key"), ShouldEqual, "modified value")
				So(userJane1.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane1.Env().Uid(), ShouldEqual, 2)
				So(userJane2.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane2.Env().Uid(), ShouldEqual, security.SuperUserID)
			})
			Convey("Checking overridden WithContext", func() {
				allPosts := env.Pool("Post").SearchAll()
				posts1 := allPosts.WithContext("foo", "bar")
				So(posts1.Env().Context().HasKey("foo"), ShouldBeTrue)
				So(posts1.Env().Context().GetString("foo"), ShouldEqual, "bar")
				So(allPosts.Env().Context().HasKey("foo"), ShouldBeFalse)
			})
		}), ShouldBeNil)
	})
	Convey("Testing cache operation", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field(email).Equals("jane.smith@example.com"))
			Convey("Cache should be empty at startup", func() {
				So(env.cache.data, ShouldBeEmpty)
				So(env.cache.m2mLinks, ShouldBeEmpty)
			})
			Convey("Loading a RecordSet should populate the cache", func() {
				userJane.Load()
				So(env.cache.m2mLinks, ShouldBeEmpty)
				So(env.cache.data, ShouldHaveLength, 4)
				So(env.cache.data, ShouldContainKey, users.model.name)
				So(env.cache.data[users.model.name], ShouldContainKey, userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				So(janeEntry, ShouldContainKey, "id")
				So(janeEntry["id"], ShouldEqual, userJane.ids[0])
				So(janeEntry, ShouldContainKey, "name")
				So(janeEntry["name"], ShouldEqual, "Jane A. Smith")
				So(janeEntry, ShouldContainKey, "email")
				So(janeEntry["email"], ShouldEqual, "jane.smith@example.com")
				So(env.cache.checkIfInCache(users.model, userJane.ids, []string{"id", "name", "email"}, "", false), ShouldBeTrue)
			})
			Convey("Calling values already in cache should not call the DB", func() {
				userJane.Load()
				id, dbCalled := userJane.get(ID, true)
				So(dbCalled, ShouldBeFalse)
				So(id, ShouldEqual, userJane.ids[0])
				name, dbCalled := userJane.get(Name, true)
				So(dbCalled, ShouldBeFalse)
				So(name, ShouldEqual, "Jane A. Smith")
				mail, dbCalled := userJane.get(email, true)
				So(dbCalled, ShouldBeFalse)
				So(mail, ShouldEqual, "jane.smith@example.com")
			})
			Convey("Testing O2M fields in cache", func() {
				userJane.Load(posts)
				postModel := env.Pool("Post").Model()
				post1 := env.Pool("Post").Search(postModel.Field(title).Equals("1st Post"))
				post3 := env.Pool("Post").Search(postModel.Field(title).Equals("3rd Post"))
				posts13 := post1.Union(post3)
				So(env.cache.data, ShouldContainKey, users.model.name)
				So(env.cache.data[users.model.name], ShouldContainKey, userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				So(janeEntry, ShouldContainKey, "posts_ids")
				So(janeEntry["posts_ids"], ShouldEqual, true)
				So(env.cache.get(userJane.model, userJane.ids[0], "posts_ids", ""), ShouldHaveLength, posts13.Len())
				for _, id := range posts13.ids {
					So(env.cache.get(userJane.model, userJane.ids[0], "posts_ids", ""), ShouldContain, id)
				}
				So(userJane.Get(posts).(RecordSet).Collection().Len(), ShouldEqual, 2)
			})
			Convey("Creating an extra post should update jane's posts", func() {
				userJane.Load(posts)
				So(userJane.Get(posts).(RecordSet).Collection().Len(), ShouldEqual, 2)
				env.Pool("Post").Call("Create", NewModelData(Registry.MustGet("Post")).
					Set(title, "Extra Post").
					Set(user, userJane))
				So(userJane.Get(posts).(RecordSet).Collection().Len(), ShouldEqual, 3)
			})
			Convey("Reading M2M fields should work both ways", func() {
				postModel := env.Pool("Post").Model()
				tagModel := env.Pool("Tag").Model()
				post2 := env.Pool("Post").Search(postModel.Field(title).Equals("2nd Post"))
				So(env.cache.m2mLinks, ShouldBeEmpty)
				post2.Load()
				So(post2.Len(), ShouldEqual, 1)
				So(env.cache.data, ShouldHaveLength, 2)
				So(env.cache.data["Post"], ShouldHaveLength, 1)
				bjTags := env.Pool("Tag").Search(tagModel.Field(Name).In([]string{"Books", "Jane's"}))
				bjTags.Fetch()
				So(bjTags.Len(), ShouldEqual, 2)
				post2.Load(tags)
				So(env.cache.m2mLinks, ShouldHaveLength, 1)
				So(env.cache.m2mLinks, ShouldContainKey, "PostTagRel")
				So(env.cache.m2mLinks["PostTagRel"], ShouldContainKey, [2]int64{post2.ids[0], bjTags.ids[0]})
				So(env.cache.m2mLinks["PostTagRel"], ShouldContainKey, [2]int64{post2.ids[0], bjTags.ids[1]})
				So(env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), ShouldHaveLength, 2)
				So(env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), ShouldContain, bjTags.ids[0])
				So(env.cache.get(postModel, post2.ids[0], tags.JSON(), ""), ShouldContain, bjTags.ids[1])
				So(env.cache.get(tagModel, bjTags.ids[0], posts.JSON(), ""), ShouldHaveLength, 1)
				So(env.cache.get(tagModel, bjTags.ids[0], posts.JSON(), ""), ShouldContain, post2.ids[0])
				So(env.cache.get(tagModel, bjTags.ids[1], posts.JSON(), ""), ShouldHaveLength, 1)
				So(env.cache.get(tagModel, bjTags.ids[1], posts.JSON(), ""), ShouldContain, post2.ids[0])
				So(env.cache.checkIfInCache(postModel, post2.ids, []string{tags.JSON()}, "", false), ShouldBeTrue)
				So(post2.Get(tags).(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(post2.Get(tags).(RecordSet).Collection().Ids(), ShouldContain, bjTags.ids[0])
				So(post2.Get(tags).(RecordSet).Collection().Ids(), ShouldContain, bjTags.ids[1])
				So(bjTags.Records()[0].Get(posts).(RecordSet).Collection().Ids(), ShouldHaveLength, 1)
				So(bjTags.Records()[0].Get(posts).(RecordSet).Collection().Ids(), ShouldContain, post2.ids[0])
				So(bjTags.Records()[1].Get(posts).(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(bjTags.Records()[1].Get(posts).(RecordSet).Collection().Ids(), ShouldContain, post2.ids[0])
			})
			Convey("Check that computed fields are not stored in cache", func() {
				userJane.Load()
				So(env.cache.data, ShouldContainKey, users.model.name)
				So(env.cache.data[users.model.name], ShouldContainKey, userJane.ids[0])
				janeEntry := env.cache.data[users.model.name][userJane.ids[0]]
				So(janeEntry, ShouldContainKey, "id")
				So(janeEntry, ShouldContainKey, "name")
				So(janeEntry, ShouldNotContainKey, "decorated_name")
				userJane.Get(decoratedName)
				So(janeEntry, ShouldNotContainKey, "decorated_name")
			})
			Convey("Checking cache dump for debug", func() {
				So(env.DumpCache(), ShouldEqual, `Data
====

M2M Links
=========

X2M Links
=========
`)
				userJane.Load()
				userJane.Load(postsTags)
				So(len(env.DumpCache()), ShouldBeGreaterThan, 1360)
			})
		}), ShouldBeNil)
	})
	Convey("Testing prefetch", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userSet := users.Search(users.Model().Field(email).Equals("jane.smith@example.com").
				Or().Field(Name).Equals("John Smith"))
			So(userSet.fetched, ShouldBeFalse)
			So(env.cache.checkIfInCache(users.Model(), userSet.ids, []string{Name.JSON()}, "", false), ShouldBeFalse)
			Convey("Loading one record should load all of its original record", func() {
				records := userSet.SortedByField(ID, false).Records()
				So(userSet.fetched, ShouldBeTrue)
				So(records, ShouldHaveLength, 2)
				So(records[0].fetched, ShouldBeTrue)
				So(records[1].fetched, ShouldBeTrue)
				So(env.cache.checkIfInCache(users.Model(), userSet.ids, []string{Name.JSON()}, "", false), ShouldBeFalse)
				name, fetched := records[0].get(Name, false)
				So(name, ShouldEqual, "John Smith")
				So(fetched, ShouldBeTrue)
				name2, fetched2 := records[1].get(Name, false)
				So(name2, ShouldEqual, "Jane A. Smith")
				So(fetched2, ShouldBeFalse)
			})
			Convey("Returned recordset by Load should be the right one", func() {
				records := userSet.SortedByField(ID, false).Records()
				rc := records[0].Load()
				So(rc.Equals(records[0]), ShouldBeTrue)
			})
			Convey("Nested records", func() {
				records := userSet.SortedByField(ID, false).Records()
				postsJohn := records[0].Get(posts).(*RecordCollection).Records()
				postsJane := records[1].Get(posts).(*RecordCollection).Records()
				So(postsJohn, ShouldHaveLength, 0)
				So(postsJane, ShouldHaveLength, 2)
			})
		}), ShouldBeNil)
	})
	Convey("Checking error types", t, func() {
		nice := new(notInCacheError)
		So(nice.Error(), ShouldEqual, "requested value not in cache")
		nepe := new(nonExistentPathError)
		So(nepe.Error(), ShouldEqual, "requested path is broken")
	})
	Convey("Testing db error retries", t, func() {
		Convey("ExecuteInNewEnvironment should retry db errors up to max retries", func() {
			var retries uint8
			So(doExecuteInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				panic(&pq.Error{Code: "40001"})
			}), ShouldNotBeNil)
			So(retries, ShouldEqual, DBSerializationMaxRetries)
		})
		Convey("ExecuteInNewEnvironment should retry db errors and stop when ok", func() {
			var retries uint8
			So(doExecuteInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				if retries < 3 {
					panic(&pq.Error{Code: "40001"})
				}
			}), ShouldBeNil)
			So(retries, ShouldEqual, 3)
		})
		Convey("SimulateInNewEnvironment should retry db errors up to max retries", func() {
			var retries uint8
			So(doSimulateInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				panic(&pq.Error{Code: "40001"})
			}), ShouldNotBeNil)
			So(retries, ShouldEqual, DBSerializationMaxRetries)
		})
		Convey("SimulateInNewEnvironment should retry db errors and stop when ok", func() {
			var retries uint8
			So(doSimulateInNewEnvironment(security.SuperUserID, 0, func(env Environment) {
				retries++
				if retries < 3 {
					panic(&pq.Error{Code: "40001"})
				}
			}), ShouldBeNil)
			So(retries, ShouldEqual, 3)
		})
	})
}
