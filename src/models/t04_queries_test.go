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

	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	profile                  = fieldName{name: "profile", json: "profile_id"}
	nums                     = fieldName{name: "Nums", json: "nums"}
	age                      = fieldName{name: "Age", json: "age"}
	email                    = fieldName{name: "Email", json: "email"}
	email2                   = fieldName{name: "Email2", json: "email2"}
	bestPost                 = fieldName{name: "BestPost", json: "best_post_id"}
	title                    = fieldName{name: "Title", json: "title"}
	isStaff                  = fieldName{name: "IsStaff", json: "is_staff"}
	resume                   = fieldName{name: "Resume", json: "resume_id"}
	coolType                 = fieldName{name: "CoolType", json: "cool_type"}
	mana                     = fieldName{name: "Mana", json: "mana"}
	other                    = fieldName{name: "Other", json: "other"}
	money                    = fieldName{name: "Money", json: "money"}
	active                   = fieldName{name: "Active", json: "active"}
	isActive                 = fieldName{name: "IsActive", json: "is_active"}
	isPremium                = fieldName{name: "IsPremium", json: "is_premium"}
	decoratedName            = fieldName{name: "DecoratedName", json: "decorated_name"}
	displayName              = fieldName{name: "DisplayName", json: "display_name"}
	writerAge                = fieldName{name: "WriterAge", json: "writer_age"}
	writerMoney              = fieldName{name: "WriterMoney", json: "writer_money"}
	postWriter               = fieldName{name: "PostWriter", json: "post_writer_id"}
	pMoney                   = fieldName{name: "PMoney", json: "p_money"}
	street                   = fieldName{name: "Street", json: "street"}
	city                     = fieldName{name: "City", json: "city"}
	zip                      = fieldName{name: "Zip", json: "zip"}
	country                  = fieldName{name: "Country", json: "country"}
	user                     = fieldName{name: "User", json: "user_id"}
	text                     = fieldName{name: "Text", json: "text"}
	record                   = fieldName{name: "Record", json: "record_id"}
	lang                     = fieldName{name: "Lang", json: "lang"}
	userName                 = fieldName{name: "UserName", json: "user_name"}
	profileAge               = fieldName{name: "Profile.Age", json: "profile_id.age"}
	profileMoney             = fieldName{name: "Profile.Money", json: "profile_id.money"}
	posts                    = fieldName{name: "Posts", json: "posts_ids"}
	content                  = fieldName{name: "Content", json: "content"}
	tags                     = fieldName{name: "Tags", json: "tags_ids"}
	tagsName                 = fieldName{name: "Tags.Name", json: "tags_ids.name"}
	description              = fieldName{name: "Description", json: "description"}
	rate                     = fieldName{name: "Rate", json: "rate"}
	comments                 = fieldName{name: "Comments", json: "comments_ids"}
	experience               = fieldName{name: "Experience", json: "experience"}
	leisure                  = fieldName{name: "Leisure", json: "leisure"}
	education                = fieldName{name: "Education", json: "education"}
	lastPost                 = fieldName{name: "LastPost", json: "last_post_id"}
	lastTagName              = fieldName{name: "LastTagName", json: "last_tag_name"}
	lastCommentText          = fieldName{name: "LastCommentText", json: "last_comment_text"}
	postsTitle               = fieldName{name: "Posts.Title", json: "posts_ids.title"}
	postsTags                = fieldName{name: "Posts.Tags", json: "posts_ids.tags_ids"}
	bestPostTitle            = fieldName{name: "BestPost.Title", json: "best_post_id.title"}
	profileBestPostTitle     = fieldName{name: "Profile.BestPost.Title", json: "profile_id.best_post_id.title"}
	profileBestPostUser      = fieldName{name: "Profile.BestPost.User", json: "profile_id.best_post_id.user_id"}
	resumeEducation          = fieldName{name: "Resume.Education", json: "resume_id.education"}
	descriptionHexyaContexts = fieldName{name: "DescriptionHexyaContexts", json: "description_hexya_contexts"}
	lastupdate               = fieldName{name: "LastUpdate", json: "__last_update"}
	createDate               = fieldName{name: "CreateDate", json: "create_date"}
	writeDate                = fieldName{name: "WriteDate", json: "write_date"}
	parent                   = fieldName{name: "Parent", json: "parent_id"}
	value                    = fieldName{name: "Value", json: "value"}
	password                 = fieldName{name: "Password", json: "password"}
	size                     = fieldName{name: "Size", json: "size"}
	hexyaVersion             = fieldName{name: "HexyaVersion", json: "hexya_version"}
)

func TestConditions(t *testing.T) {
	Convey("Testing SQL building for queries", t, func() {
		if dbArgs.Driver == "postgres" {
			So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				rs := env.Pool("User")
				rs = rs.Search(rs.Model().FilteredOn(profile, env.Pool("Profile").Model().FilteredOn(bestPost, env.Pool("Post").Model().Field(title).Equals("foo"))))
				fields := []FieldName{Name, fieldName{name: "Profile.BestPost.Title", json: "profile_id.best_post_id.title"}}
				Convey("Simple query with database field names", func() {
					rs = env.Pool("User").Search(rs.Model().FilteredOn(profile, env.Pool("Profile").Model().Field(bestPostTitle).Equals("foo"))).OrderBy("ID")
					sql, args, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title, "user".id AS id FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE "T2".title = ? ORDER BY "user".id ) foo ORDER BY id `)
					So(args, ShouldContain, "foo")
				})
				Convey("Simple query with struct field names", func() {
					fields = []FieldName{Name, fieldName{name: "Profile.BestPost.Title", json: "profile_id.best_post_id.title"}}
					sql, args, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE "T2".title = ? ORDER BY "user".id ) foo  `)
					So(args, ShouldContain, "foo")
				})
				Convey("Query with one2many relations", func() {
					rso2m := env.Pool("User").Search(rs.Model().Field(postsTitle).Equals("1st post"))
					fields = []FieldName{Name}
					sql, args, _ := rso2m.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user" LEFT JOIN "post" "T1" ON "user".id="T1".user_id  WHERE "T1".title = ? ORDER BY "user".id ) foo  `)
					So(args, ShouldContain, "1st post")
				})
				Convey("Simple query with args inflation", func() {
					getUserID := func(rc *RecordCollection) int64 {
						return rc.Env().Uid()
					}
					rs2 := env.Pool("User").Search(rs.Model().Field(nums).Equals(getUserID))
					fields = []FieldName{Name}
					sql, args, _ := rs2.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".nums = ? ORDER BY "user".id ) foo  `)
					So(len(args), ShouldEqual, 1)
					So(args, ShouldContain, security.SuperUserID)
				})
				Convey("true/false query", func() {
					rs3 := env.Pool("User").Search(rs.Model().Field(isStaff).Equals(true))
					fields = []FieldName{Name}
					sql, args, _ := rs3.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".is_staff = ? ORDER BY "user".id ) foo  `)
					So(len(args), ShouldEqual, 1)
					So(args, ShouldContain, true)
				})
				Convey("Check WHERE clause with additionnal filter", func() {
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user__profile__post".title = ?) AND ("user__profile".age >= ?)`)
					So(args, ShouldContain, 12)
					So(args, ShouldContain, "foo")
				})
				Convey("Check full query with all conditions", func() {
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					c2 := rs.Model().Field(Name).Contains("jane").Or().Field(profileMoney).Lower(1234.56)
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE (("user__profile__post".title = ?) AND ("user__profile".age >= ?)) AND ("user".name LIKE ? OR "user__profile".money < ?)`)
					So(args, ShouldContain, "%jane%")
					So(args, ShouldContain, 1234.56)
					sql, _, _ = rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE (("T2".title = ?) AND ("T1".age >= ?)) AND ("user".name LIKE ? OR "T1".money < ?) ORDER BY "user".id ) foo  `)
				})
				Convey("Check multi-join queries", func() {
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					c2 := rs.Model().Field(Name).Contains("jane").Or().Field(resumeEducation).Contains("MIT")
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE (("user__profile__post".title = ?) AND ("user__profile".age >= ?)) AND ("user".name LIKE ? OR "user__resume".education LIKE ?)`)
					So(args, ShouldContain, "%jane%")
					So(args, ShouldContain, "%MIT%")
					sql, _, _ = rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id LEFT JOIN "resume" "T3" ON "user".resume_id="T3".id  WHERE (("T2".title = ?) AND ("T1".age >= ?)) AND ("user".name LIKE ? OR "T3".education LIKE ?) ORDER BY "user".id ) foo  `)
				})
				Convey("Testing query without WHERE clause", func() {
					rs = env.Pool("User").Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"   ORDER BY "user".id ) foo  `)
				})
				Convey("Testing query with LIMIT clause", func() {
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY id LIMIT 1 `)
				})
				Convey("Testing query with LIMIT and OFFSET clauses", func() {
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Call("Offset", 2).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY id LIMIT 1 OFFSET 2`)
				})
				Convey("Testing query with ORDER BY clauses", func() {
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("OrderBy", []string{"Email", "ID"}).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".email AS email, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY email, id `)
				})
				Convey("Testing complex conditions", func() {
					rs = env.Pool("User").Search(rs.Model().Field(profileAge).GreaterOrEqual(12).
						AndNot().Field(Name).IContains("Jane").
						OrNot().FilteredOn(profile, env.Pool("Profile").Model().Field(age).Equals(20)))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user__profile".age >= ? AND NOT "user".name ILIKE ? OR NOT "user__profile".age = ?`)
					So(args, ShouldContain, 12)
					So(args, ShouldContain, "%Jane%")
					So(args, ShouldContain, 20)
					cond1 := env.Pool("User").Model().Field(Name).IContains("Jane")
					cond2 := env.Pool("User").Model().Field(Name).IContains("John")
					rs = env.Pool("User").Search(
						env.Pool("User").Model().Field(age).GreaterOrEqual(30).
							AndNotCond(cond1).
							OrNotCond(cond2))
					sql, args = rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE (("user".age >= ?) AND NOT ("user".name ILIKE ?)) OR NOT ("user".name ILIKE ?)`)
					So(args, ShouldContain, 30)
					So(args, ShouldContain, "%Jane%")
					So(args, ShouldContain, "%John%")
				})
			}), ShouldBeNil)
		}
	})
	Convey("Testing predicate operators", t, func() {
		if dbArgs.Driver == "postgres" {
			So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				rs := env.Pool("User")
				Convey("Equals", func() {
					rs = rs.Search(rs.Model().Field(Name).Equals("John"))
					cond := rs.Condition()
					res := rs.CallMulti("SQLFromCondition", cond)
					sql := res[0]
					args := res[1]
					So(sql, ShouldEqual, `"user".name = ?`)
					So(args, ShouldContain, "John")
				})
				Convey("NotEquals", func() {
					rs = rs.Search(rs.Model().Field(Name).NotEquals("John"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NULL OR "user".name != ?)`)
					So(args, ShouldContain, "John")
				})
				Convey("Greater", func() {
					rs = rs.Search(rs.Model().Field(nums).Greater(12))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".nums > ?`)
					So(args, ShouldContain, 12)
				})
				Convey("GreaterOrEqual", func() {
					rs = rs.Search(rs.Model().Field(nums).GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".nums >= ?`)
					So(args, ShouldContain, 12)
				})
				Convey("Lower", func() {
					rs = rs.Search(rs.Model().Field(nums).Lower(12))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".nums < ?`)
					So(args, ShouldContain, 12)
				})
				Convey("LowerOrEqual", func() {
					rs = rs.Search(rs.Model().Field(nums).LowerOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".nums <= ?`)
					So(args, ShouldContain, 12)
				})
				Convey("Contains", func() {
					rs = rs.Search(rs.Model().Field(Name).Contains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".name LIKE ?`)
					So(args, ShouldContain, "%John%")
				})
				Convey("Not Contains", func() {
					rs = rs.Search(rs.Model().Field(Name).NotContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NULL OR "user".name NOT LIKE ?)`)
					So(args, ShouldContain, "%John%")
				})
				Convey("IContains", func() {
					rs = rs.Search(rs.Model().Field(Name).IContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".name ILIKE ?`)
					So(args, ShouldContain, "%John%")
				})
				Convey("Not IContains", func() {
					rs = rs.Search(rs.Model().Field(Name).NotIContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NULL OR "user".name NOT ILIKE ?)`)
					So(args, ShouldContain, "%John%")
				})
				Convey("Contains pattern", func() {
					rs = rs.Search(rs.Model().Field(Name).Like("John%"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".name LIKE ?`)
					So(args, ShouldContain, "John%")
				})
				Convey("IContains pattern", func() {
					rs = rs.Search(rs.Model().Field(Name).ILike("John%"))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".name ILIKE ?`)
					So(args, ShouldContain, "John%")
				})
				Convey("In", func() {
					rs = rs.Search(rs.Model().Field(ID).In([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE "user".id IN (?)`)
					So(args, ShouldContain, []int64{23, 31})
				})
				Convey("Not In", func() {
					rs = rs.Search(rs.Model().Field(ID).NotIn([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".id IS NULL OR "user".id NOT IN (?))`)
					So(args, ShouldContain, []int64{23, 31})
				})
				Convey("Is Null", func() {
					rs = rs.Search(rs.Model().Field(Name).IsNull())
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NULL OR "user".name = ?)`)
					So(args, ShouldContain, "")
				})
				Convey("Is Not Null", func() {
					rs = rs.Search(rs.Model().Field(Name).IsNotNull())
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NOT NULL AND "user".name != ?)`)
					So(args, ShouldContain, "")
				})
				Convey("Empty string", func() {
					rs = rs.Search(rs.Model().Field(Name).Equals(""))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".name IS NULL OR "user".name = ?)`)
					So(args, ShouldContain, "")
				})
				Convey("False bool", func() {
					rs = rs.Search(rs.Model().Field(isStaff).Equals(false))
					sql, args := rs.query.sqlWhereClause(true)
					So(sql, ShouldEqual, `WHERE ("user".is_staff IS NULL OR "user".is_staff = ?)`)
					So(args, ShouldContain, false)
				})
				Convey("Child Of without parent field", func() {
					rs = rs.Search(rs.Model().Field(ID).ChildOf(101))
					sql, args, _ := rs.query.selectQuery([]FieldName{Name})
					So(sql, ShouldEqual, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".id = ? ORDER BY "user".id ) foo  `)
					So(args, ShouldContain, 101)
				})
			}), ShouldBeNil)
		}
	})
	Convey("Testing Condition Methods", t, func() {
		So(SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			cond := env.Pool("User").Model().Field(Name).IContains("Jane")
			cond2 := env.Pool("User").Model().Field(ID).NotIn([]int64{23, 31})
			Convey("HasField", func() {
				So(cond.HasField(Registry.MustGet("User").Fields().MustGet("Name")), ShouldBeTrue)
				So(cond.HasField(Registry.MustGet("User").Fields().MustGet("Status")), ShouldBeFalse)
				So(cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("Name")), ShouldBeTrue)
				So(cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("ID")), ShouldBeTrue)
				So(cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("Status")), ShouldBeFalse)
			})
			Convey("String", func() {
				So(cond.OrNotCond(cond2).String(), ShouldEqual, `AND Name ilike Jane
OR NOT (
AND ID not in [23 31]

)
`)
			})
		}), ShouldBeNil)
	})
}

func TestConditionSerialization(t *testing.T) {
	var (
		a = fieldName{name: "A", json: "A"}
		b = fieldName{name: "B", json: "B"}
		c = fieldName{name: "C", json: "C"}
		d = fieldName{name: "D", json: "D"}
		f = fieldName{name: "F", json: "F"}
	)
	Convey("Testing condition serialization", t, func() {
		Convey("Testing simple A AND B condition", func() {
			cond := newCondition().And().Field(Name).IContains("John").And().Field(age).Greater(18)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[& [name ilike John] [age > 18]]")
		})
		Convey("Testing simple A OR B condition", func() {
			cond := newCondition().And().Field(Name).IContains("John").Or().Field(age).Greater(18)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [age > 18] [name ilike John]]")
		})
		Convey("Testing A AND B OR C condition", func() {
			cond := newCondition().And().Field(Name).IContains("John").And().Field(age).Greater(18).Or().Field(isStaff).Equals(true)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [is_staff = true] & [name ilike John] [age > 18]]")
		})
		Convey("Testing (A OR B) AND (C OR D) OR F condition", func() {
			aOrB := newCondition().And().Field(a).Equals("A Value").Or().Field(b).Equals("B Value")
			cOrD := newCondition().And().Field(c).Equals("C Value").Or().Field(d).Equals("D Value")
			cond := newCondition().AndCond(aOrB).AndCond(cOrD).Or().Field(f).Equals("F Value")
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[| [F = F Value] & | [B = B Value] [A = A Value] | [D = D Value] [C = C Value]]")
		})
		Convey("Testing (A OR B OR C) AND (D) condition", func() {
			aOrBOrC := newCondition().And().Field(a).Equals("A Value").Or().Field(b).Equals("B Value").Or().Field(c).Equals("C Value")
			D := newCondition().And().Field(d).Equals("D Value")
			cond := newCondition().AndCond(aOrBOrC).AndCond(D)
			dom := cond.Serialize()
			So(fmt.Sprint(dom), ShouldEqual, "[& | [C = C Value] | [B = B Value] [A = A Value] [D = D Value]]")
		})
	})
}
