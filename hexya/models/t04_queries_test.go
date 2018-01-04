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
	"fmt"
	"testing"

	"github.com/hexya-erp/hexya/hexya/models/security"
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
					rs = env.Pool("User").Search(rs.Model().FilteredOn("profile_id", env.Pool("Profile").Model().Field("best_post_id.title").Equals("foo"))).OrderBy("ID")
					sql, args := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "T2".title AS profile_id__best_post_id__title, "user".id AS id FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE ("T2".title = ? )  ORDER BY "user".id  `)
					So(args, ShouldContain, "foo")
				})
				Convey("Simple query with struct field names", func() {
					fields := []string{"Name", "Profile.BestPost.Title"}
					sql, args := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE ("T2".title = ? )   `)
					So(args, ShouldContain, "foo")
				})
				Convey("Simple query with args inflation", func() {
					getUserID := func(rc *RecordCollection) int64 {
						return rc.Env().Uid()
					}
					rs2 := env.Pool("User").Search(rs.Model().Field("Nums").Equals(getUserID))
					fields := []string{"Name"}
					sql, args := rs2.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"  WHERE ("user".nums = ? )   `)
					So(len(args), ShouldEqual, 1)
					So(args, ShouldContain, security.SuperUserID)
				})
				Convey("true/false query", func() {
					rs3 := env.Pool("User").Search(rs.Model().Field("IsStaff").Equals(true))
					fields := []string{"Name"}
					sql, args := rs3.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"  WHERE ("user".is_staff = ? )   `)
					So(len(args), ShouldEqual, 1)
					So(args, ShouldContain, true)
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
					c2 := rs.Model().Field("name").Contains("jane").Or().Field("Profile.Money").Lower(1234.56)
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user__profile__post".title = ? ) AND ("user__profile".age >= ? ) AND ("user".name LIKE ? OR "user__profile".money < ? ) `)
					So(args, ShouldContain, "%jane%")
					So(args, ShouldContain, 1234.56)
					sql, _ = rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE ("T2".title = ? ) AND ("T1".age >= ? ) AND ("user".name LIKE ? OR "T1".money < ? )   `)
				})
				Convey("Check multi-join queries", func() {
					rs = rs.Search(rs.Model().Field("Profile.Age").GreaterOrEqual(12))
					c2 := rs.Model().Field("name").Contains("jane").Or().Field("Resume.Education").Contains("MIT")
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user__profile__post".title = ? ) AND ("user__profile".age >= ? ) AND ("user".name LIKE ? OR "user__resume".education LIKE ? ) `)
					So(args, ShouldContain, "%jane%")
					So(args, ShouldContain, "%MIT%")
					sql, _ = rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id INNER JOIN "resume" "T3" ON "user".resume_id="T3".id  WHERE ("T2".title = ? ) AND ("T1".age >= ? ) AND ("user".name LIKE ? OR "T3".education LIKE ? )   `)
				})
				Convey("Testing query without WHERE clause", func() {
					rs = env.Pool("User").Load()
					fields := []string{"name"}
					sql, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"    `)
				})
				Convey("Testing query with LIMIT clause", func() {
					rs = env.Pool("User").Search(rs.Model().Field("email").IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Load()
					fields := []string{"name"}
					sql, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user".id AS id FROM "user" "user"  WHERE ("user".email ILIKE ? )  ORDER BY "user".id  LIMIT 1 `)
				})
				Convey("Testing query with LIMIT and OFFSET clauses", func() {
					rs = env.Pool("User").Search(rs.Model().Field("email").IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Call("Offset", 2).(RecordSet).Collection().Load()
					fields := []string{"name"}
					sql, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user".id AS id FROM "user" "user"  WHERE ("user".email ILIKE ? )  ORDER BY "user".id  LIMIT 1 OFFSET 2`)
				})
				Convey("Testing query with ORDER BY clauses", func() {
					rs = env.Pool("User").Search(rs.Model().Field("email").IContains("jane.smith@example.com")).Call("OrderBy", []string{"Email", "ID"}).(RecordSet).Collection().Load()
					fields := []string{"name"}
					sql, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name, "user".email AS email, "user".id AS id FROM "user" "user"  WHERE ("user".email ILIKE ? )  ORDER BY "user".email , "user".id  `)
				})
				Convey("Testing complex conditions", func() {
					rs = env.Pool("User").Search(rs.Model().Field("Profile.Age").GreaterOrEqual(12).
						AndNot().Field("Name").IContains("Jane").
						OrNot().FilteredOn("Profile", env.Pool("Profile").Model().Field("Age").Equals(20)))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user__profile".age >= ? AND NOT "user".name ILIKE ? OR NOT "user__profile".age = ? ) `)
					So(args, ShouldContain, 12)
					So(args, ShouldContain, "%Jane%")
					So(args, ShouldContain, 20)
					cond1 := env.Pool("User").Model().Field("Name").IContains("Jane")
					cond2 := env.Pool("User").Model().Field("Name").IContains("John")
					rs := env.Pool("User").Search(
						env.Pool("User").Model().Field("Age").GreaterOrEqual(30).
							AndNotCond(cond1).
							OrNotCond(cond2))
					sql, args = rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".age >= ? AND NOT ("user".name ILIKE ? ) OR NOT ("user".name ILIKE ? ) ) `)
					So(args, ShouldContain, 30)
					So(args, ShouldContain, "%Jane%")
					So(args, ShouldContain, "%John%")
				})
			})
		}
	})
	Convey("Testing predicate operators", t, func() {
		if dbArgs.Driver == "postgres" {
			SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				rs := env.Pool("User")
				Convey("Equals", func() {
					rs = rs.Search(rs.Model().Field("Name").Equals("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name = ? ) `)
					So(args, ShouldContain, "John")
				})
				Convey("NotEquals", func() {
					rs = rs.Search(rs.Model().Field("Name").NotEquals("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name != ? ) `)
					So(args, ShouldContain, "John")
				})
				Convey("Greater", func() {
					rs = rs.Search(rs.Model().Field("Nums").Greater(12))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".nums > ? ) `)
					So(args, ShouldContain, 12)
				})
				Convey("GreaterOrEqual", func() {
					rs = rs.Search(rs.Model().Field("Nums").GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".nums >= ? ) `)
					So(args, ShouldContain, 12)
				})
				Convey("Lower", func() {
					rs = rs.Search(rs.Model().Field("Nums").Lower(12))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".nums < ? ) `)
					So(args, ShouldContain, 12)
				})
				Convey("LowerOrEqual", func() {
					rs = rs.Search(rs.Model().Field("Nums").LowerOrEqual(12))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".nums <= ? ) `)
					So(args, ShouldContain, 12)
				})
				Convey("Contains", func() {
					rs = rs.Search(rs.Model().Field("Name").Contains("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name LIKE ? ) `)
					So(args, ShouldContain, "%John%")
				})
				Convey("Not Contains", func() {
					rs = rs.Search(rs.Model().Field("Name").NotContains("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name NOT LIKE ? ) `)
					So(args, ShouldContain, "%John%")
				})
				Convey("IContains", func() {
					rs = rs.Search(rs.Model().Field("Name").IContains("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name ILIKE ? ) `)
					So(args, ShouldContain, "%John%")
				})
				Convey("Not IContains", func() {
					rs = rs.Search(rs.Model().Field("Name").NotIContains("John"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name NOT ILIKE ? ) `)
					So(args, ShouldContain, "%John%")
				})
				Convey("Contains pattern", func() {
					rs = rs.Search(rs.Model().Field("Name").Like("John%"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name LIKE ? ) `)
					So(args, ShouldContain, "John%")
				})
				Convey("IContains pattern", func() {
					rs = rs.Search(rs.Model().Field("Name").ILike("John%"))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".name ILIKE ? ) `)
					So(args, ShouldContain, "John%")
				})
				Convey("In", func() {
					rs = rs.Search(rs.Model().Field("ID").In([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".id IN (?) ) `)
					So(args, ShouldContain, []int64{23, 31})
				})
				Convey("Not In", func() {
					rs = rs.Search(rs.Model().Field("ID").NotIn([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause()
					So(sql, ShouldEqual, `WHERE ("user".id NOT IN (?) ) `)
					So(args, ShouldContain, []int64{23, 31})
				})
				Convey("Child Of without parent field", func() {
					rs = rs.Search(rs.Model().Field("ID").ChildOf(101))
					sql, args := rs.query.selectQuery([]string{"Name"})
					So(sql, ShouldEqual, `SELECT DISTINCT "user".name AS name FROM "user" "user"  WHERE ("user".id = ? )   `)
					So(args, ShouldContain, 101)
				})
			})
		}
	})
}

func TestConditionSerialization(t *testing.T) {
	Convey("Testing condition serialization", t, func() {
		Convey("Testing simple A AND B condition", func() {
			cond := newCondition().And().Field("Name").IContains("John").And().Field("Age").Greater(18)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[& [Name ilike John] [Age > 18]]")
		})
		Convey("Testing simple A OR B condition", func() {
			cond := newCondition().And().Field("Name").IContains("John").Or().Field("Age").Greater(18)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [Age > 18] [Name ilike John]]")
		})
		Convey("Testing A AND B OR C condition", func() {
			cond := newCondition().And().Field("Name").IContains("John").And().Field("Age").Greater(18).Or().Field("IsStaff").Equals(true)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [IsStaff = true] & [Name ilike John] [Age > 18]]")
		})
		Convey("Testing (A OR B) AND (C OR D) OR F condition", func() {
			aOrB := newCondition().And().Field("A").Equals("A Value").Or().Field("B").Equals("B Value")
			cOrD := newCondition().And().Field("C").Equals("C Value").Or().Field("D").Equals("D Value")
			cond := newCondition().AndCond(aOrB).AndCond(cOrD).Or().Field("F").Equals("F Value")
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [F = F Value] & | [B = B Value] [A = A Value] | [D = D Value] [C = C Value]]")
		})
	})
}
