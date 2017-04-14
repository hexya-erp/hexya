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

func TestConditions(t *testing.T) {
	Convey("Testing SQL building for queries", t, func() {
		if dbArgs.Driver == "postgres" {
			SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				rs := env.Pool("User")
				rs = rs.Search(rs.Model().FilteredOn("Profile", env.Pool("Profile").Model().FilteredOn("BestPost", env.Pool("Post").Model().Field("Title").Equals("foo"))))
				fields := []string{"name", "profile_id.best_post_id.title"}
				Convey("Simple query with database field names", func() {
					rs = env.Pool("User").Search(rs.Model().FilteredOn("profile_id", env.Pool("Profile").Model().Field("best_post_id.title").Equals("foo")))
					sql, args := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user__profile__post".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "user__profile" ON "user".profile_id="user__profile".id LEFT JOIN "post" "user__profile__post" ON "user__profile".best_post_id="user__profile__post".id  WHERE ("user__profile__post".title = ? ) ORDER BY id `)
					So(args, ShouldContain, "foo")
				})
				Convey("Simple query with struct field names", func() {
					fields := []string{"Name", "Profile.BestPost.Title"}
					sql, args := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user__profile__post".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "user__profile" ON "user".profile_id="user__profile".id LEFT JOIN "post" "user__profile__post" ON "user__profile".best_post_id="user__profile__post".id  WHERE ("user__profile__post".title = ? ) ORDER BY id `)
					So(args, ShouldContain, "foo")
				})
				Convey("Simple query with args inflation", func() {
					getUserID := func(rc RecordCollection) int64 {
						return rc.Env().Uid()
					}
					rs2 := env.Pool("User").Search(rs.Model().Field("Nums").Equals(getUserID))
					fields := []string{"Name"}
					sql, args := rs2.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"  WHERE ("user".nums = ? ) ORDER BY id `)
					So(len(args), ShouldEqual, 1)
					So(args, ShouldContain, security.SuperUserID)
				})
				Convey("Check WHERE clause with additionnal filter", func() {
					rs = rs.Search(rs.Model().Field("Profile.Age").GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user__profile__post".title = ? ) AND ("user__profile".age >= ? ) `)
					So(args, ShouldContain, 12)
					So(args, ShouldContain, "foo")
				})
				Convey("Check full query with all conditions", func() {
					rs = rs.Search(rs.Model().Field("Profile.Age").GreaterOrEqual(12))
					c2 := rs.Model().Field("name").Like("jane").Or().Field("Profile.Money").Lower(1234.56)
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user__profile__post".title = ? ) AND ("user__profile".age >= ? ) AND ("user".name LIKE ? OR "user__profile".money < ? ) `)
					So(args, ShouldContain, "%jane%")
					So(args, ShouldContain, 1234.56)
					sql, _ = rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user__profile__post".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "user__profile" ON "user".profile_id="user__profile".id LEFT JOIN "post" "user__profile__post" ON "user__profile".best_post_id="user__profile__post".id  WHERE ("user__profile__post".title = ? ) AND ("user__profile".age >= ? ) AND ("user".name LIKE ? OR "user__profile".money < ? ) ORDER BY id `)
				})
				Convey("Testing query without WHERE clause", func() {
					rs = env.Pool("User").Load()
					fields := []string{"name"}
					sql, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"  ORDER BY id `)
				})
			})
		}
	})
}
