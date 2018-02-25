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
	"sort"
	"testing"

	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnvironment(t *testing.T) {
	Convey("Testing Environment Modifications", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			env.context = types.NewContext().WithKey("key", "context value")
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
			Convey("Checking WithEnv", func() {
				env2 := newEnvironment(2)
				userJane1 := userJane.Call("WithEnv", env2).(RecordSet).Collection()
				So(userJane1.Env().Uid(), ShouldEqual, 2)
				So(userJane.Env().Uid(), ShouldEqual, 1)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane1.Env().Context().IsEmpty(), ShouldBeTrue)
				So(userJane.Env().callStack, ShouldBeEmpty)
				So(userJane1.Env().callStack, ShouldBeEmpty)
				env2.rollback()
			})
			Convey("Checking WithContext", func() {
				userJane1 := userJane.Call("WithContext", "newKey", "This is a different key").(RecordSet).Collection()
				So(userJane1.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
				So(userJane1.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
				So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane1.Env().callStack, ShouldBeEmpty)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
				So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane.Env().callStack, ShouldBeEmpty)
			})
			Convey("Checking WithNewContext", func() {
				newCtx := types.NewContext().WithKey("newKey", "This is a different key")
				userJane1 := userJane.Call("WithNewContext", newCtx).(RecordSet).Collection()
				So(userJane1.Env().Context().HasKey("key"), ShouldBeFalse)
				So(userJane1.Env().Context().HasKey("newKey"), ShouldBeTrue)
				So(userJane1.Env().Context().Get("newKey"), ShouldEqual, "This is a different key")
				So(userJane1.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane1.Env().callStack, ShouldBeEmpty)
				So(userJane.Env().Context().HasKey("key"), ShouldBeTrue)
				So(userJane.Env().Context().HasKey("newKey"), ShouldBeFalse)
				So(userJane.Env().Context().Get("key"), ShouldEqual, "context value")
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane.Env().callStack, ShouldBeEmpty)
			})
			Convey("Checking Sudo", func() {
				userJane1 := userJane.Sudo(2)
				userJane2 := userJane1.Call("Sudo").(RecordSet).Collection()
				So(userJane1.Env().Uid(), ShouldEqual, 2)
				So(userJane1.Env().callStack, ShouldBeEmpty)
				So(userJane.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane.Env().callStack, ShouldBeEmpty)
				So(userJane2.Env().Uid(), ShouldEqual, security.SuperUserID)
				So(userJane2.Env().callStack, ShouldBeEmpty)
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
				posts := env.Pool("Post").SearchAll()
				posts1 := posts.WithContext("foo", "bar")
				So(posts1.Env().Context().HasKey("foo"), ShouldBeTrue)
				So(posts1.Env().Context().GetString("foo"), ShouldEqual, "bar")
				So(posts1.Env().callStack, ShouldBeEmpty)
				So(posts.Env().Context().HasKey("foo"), ShouldBeFalse)
				So(posts.Env().callStack, ShouldBeEmpty)
			})
		})
	})
	Convey("Testing cache operation", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userJane := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com"))
			Convey("Cache should be empty at startup", func() {
				So(env.cache.data, ShouldBeEmpty)
				So(env.cache.m2mLinks, ShouldBeEmpty)
			})
			Convey("Loading a RecordSet should populate the cache", func() {
				userJane.Load()
				So(env.cache.m2mLinks, ShouldBeEmpty)
				So(env.cache.data, ShouldHaveLength, 1)
				janeCacheRef := cacheRef{model: users.model, id: userJane.ids[0]}
				So(env.cache.data, ShouldContainKey, janeCacheRef)
				So(env.cache.data[janeCacheRef], ShouldContainKey, "id")
				So(env.cache.data[janeCacheRef]["id"], ShouldEqual, userJane.ids[0])
				So(env.cache.data[janeCacheRef], ShouldContainKey, "name")
				So(env.cache.data[janeCacheRef]["name"], ShouldEqual, "Jane A. Smith")
				So(env.cache.data[janeCacheRef], ShouldContainKey, "email")
				So(env.cache.data[janeCacheRef]["email"], ShouldEqual, "jane.smith@example.com")
				So(env.cache.checkIfInCache(users.model, userJane.ids, []string{"id", "name", "email"}), ShouldBeTrue)
			})
			Convey("Calling values already in cache should not call the DB", func() {
				userJane.Load()
				id, dbCalled := userJane.get("id", true)
				So(dbCalled, ShouldBeFalse)
				So(id, ShouldEqual, userJane.ids[0])
				name, dbCalled := userJane.get("name", true)
				So(dbCalled, ShouldBeFalse)
				So(name, ShouldEqual, "Jane A. Smith")
				email, dbCalled := userJane.get("email", true)
				So(dbCalled, ShouldBeFalse)
				So(email, ShouldEqual, "jane.smith@example.com")
			})
			Convey("Testing O2M fields in cache", func() {
				userJane.Load("Posts")
				postModel := env.Pool("Post").Model()
				post1 := env.Pool("Post").Search(postModel.Field("Title").Equals("1st Post"))
				post3 := env.Pool("Post").Search(postModel.Field("Title").Equals("3rd Post"))
				posts := post1.Union(post3)
				janeCacheRef := cacheRef{model: users.model, id: userJane.ids[0]}
				So(env.cache.data, ShouldContainKey, janeCacheRef)
				So(env.cache.data[janeCacheRef], ShouldContainKey, "posts_ids")
				So(env.cache.data[janeCacheRef]["posts_ids"], ShouldEqual, true)
				So(env.cache.get(userJane.model, userJane.ids[0], "posts_ids"), ShouldHaveLength, posts.Len())
				for _, id := range posts.ids {
					So(env.cache.get(userJane.model, userJane.ids[0], "posts_ids"), ShouldContain, id)
				}
				So(userJane.Get("Posts").(RecordSet).Collection().Len(), ShouldEqual, 2)
			})
			Convey("Creating an extra post should update jane's posts", func() {
				userJane.Load("Posts")
				So(userJane.Get("Posts").(RecordSet).Collection().Len(), ShouldEqual, 2)
				env.Pool("Post").Call("Create", FieldMap{
					"Title": "Extra Post",
					"User":  userJane,
				})
				So(userJane.Get("Posts").(RecordSet).Collection().Len(), ShouldEqual, 3)
			})
			Convey("Reading M2M fields should work both ways", func() {
				postModel := env.Pool("Post").Model()
				tagModel := env.Pool("Tag").Model()
				post2 := env.Pool("Post").Search(postModel.Field("Title").Equals("2nd Post"))
				post2.Load()
				So(post2.Len(), ShouldEqual, 1)
				So(env.cache.m2mLinks, ShouldBeEmpty)
				So(env.cache.data, ShouldHaveLength, 1)
				tags := env.Pool("Tag").Search(tagModel.Field("Name").In([]string{"Books", "Jane's"}))
				tags.Fetch()
				So(tags.Len(), ShouldEqual, 2)
				post2.Load("Tags")
				So(env.cache.m2mLinks, ShouldHaveLength, 1)
				linkModel := env.Pool("PostTagRel").Model()
				So(env.cache.m2mLinks, ShouldContainKey, linkModel)
				So(env.cache.m2mLinks[linkModel], ShouldContainKey, [2]int64{post2.ids[0], tags.ids[0]})
				So(env.cache.m2mLinks[linkModel], ShouldContainKey, [2]int64{post2.ids[0], tags.ids[1]})
				So(env.cache.get(postModel, post2.ids[0], "Tags"), ShouldHaveLength, 2)
				So(env.cache.get(postModel, post2.ids[0], "Tags"), ShouldContain, tags.ids[0])
				So(env.cache.get(postModel, post2.ids[0], "Tags"), ShouldContain, tags.ids[1])
				So(env.cache.get(tagModel, tags.ids[0], "Posts"), ShouldHaveLength, 1)
				So(env.cache.get(tagModel, tags.ids[0], "Posts"), ShouldContain, post2.ids[0])
				So(env.cache.get(tagModel, tags.ids[1], "Posts"), ShouldHaveLength, 1)
				So(env.cache.get(tagModel, tags.ids[1], "Posts"), ShouldContain, post2.ids[0])
				So(env.cache.checkIfInCache(postModel, post2.ids, []string{"Tags"}), ShouldBeTrue)
				So(post2.Get("Tags").(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(post2.Get("Tags").(RecordSet).Collection().Ids(), ShouldContain, tags.ids[0])
				So(post2.Get("Tags").(RecordSet).Collection().Ids(), ShouldContain, tags.ids[1])
				So(tags.Records()[0].Get("Posts").(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(tags.Records()[0].Get("Posts").(RecordSet).Collection().Ids(), ShouldContain, post2.ids[0])
				So(tags.Records()[1].Get("Posts").(RecordSet).Collection().Ids(), ShouldHaveLength, 1)
				So(tags.Records()[1].Get("Posts").(RecordSet).Collection().Ids(), ShouldContain, post2.ids[0])
			})
			Convey("Check that computed fields are stored and read in cache", func() {
				userJane.Load()
				janeCacheRef := cacheRef{model: users.model, id: userJane.ids[0]}
				So(env.cache.data, ShouldContainKey, janeCacheRef)
				So(env.cache.data[janeCacheRef], ShouldContainKey, "id")
				So(env.cache.data[janeCacheRef], ShouldContainKey, "name")
				So(env.cache.data[janeCacheRef], ShouldNotContainKey, "decorated_name")
				decoratedName := userJane.Get("DecoratedName")
				So(env.cache.data[janeCacheRef], ShouldContainKey, "decorated_name")
				So(env.cache.data[janeCacheRef]["decorated_name"], ShouldEqual, decoratedName)
			})
		})
	})
	Convey("Testing prefetch", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			users := env.Pool("User")
			userSet := users.Search(users.Model().Field("Email").Equals("jane.smith@example.com").
				Or().Field("name").Equals("John Smith"))
			So(userSet.fetched, ShouldBeFalse)
			So(env.cache.checkIfInCache(users.Model(), userSet.ids, []string{"Name"}), ShouldBeFalse)
			Convey("Loading one record should load all of its original record", func() {
				records := userSet.Records()
				So(userSet.fetched, ShouldBeTrue)
				So(records, ShouldHaveLength, 2)
				So(records[0].fetched, ShouldBeTrue)
				So(records[1].fetched, ShouldBeTrue)
				sort.Slice(records, func(i, j int) bool {
					return records[i].Get("ID").(int64) < records[j].Get("ID").(int64)
				})
				So(env.cache.checkIfInCache(users.Model(), userSet.ids, []string{"Name"}), ShouldBeFalse)
				name, fetched := records[0].get("Name", false)
				So(name, ShouldEqual, "John Smith")
				So(fetched, ShouldBeTrue)
				name2, fetched2 := records[1].get("Name", false)
				So(name2, ShouldEqual, "Jane A. Smith")
				So(fetched2, ShouldBeFalse)
			})
			Convey("Returned recordset by Load should be the right one", func() {
				records := userSet.Records()
				sort.Slice(records, func(i, j int) bool {
					return records[i].Get("ID").(int64) < records[j].Get("ID").(int64)
				})
				rc := records[0].Load()
				So(rc.Equals(records[0]), ShouldBeTrue)
			})
			Convey("Nested records", func() {
				records := userSet.Records()
				sort.Slice(records, func(i, j int) bool {
					return records[i].Get("ID").(int64) < records[j].Get("ID").(int64)
				})
				postsJohn := records[0].Get("Posts").(*RecordCollection).Records()
				postsJane := records[1].Get("Posts").(*RecordCollection).Records()
				So(postsJohn, ShouldHaveLength, 0)
				So(postsJane, ShouldHaveLength, 2)
			})
		})
	})
}
