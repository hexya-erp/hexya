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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models/security"
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
	hexyaExternalID          = fieldName{name: "HexyaExternalID", json: "hexya_external_id"}
)

func TestConditions(t *testing.T) {
	t.Run("Testing SQL building for queries", func(t *testing.T) {
		if dbArgs.Driver == "postgres" {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				var (
					rs     *RecordCollection
					fields []FieldName
				)
				reset := func() {
					rs = env.Pool("User")
					rs = rs.Search(rs.Model().FilteredOn(profile, env.Pool("Profile").Model().FilteredOn(bestPost, env.Pool("Post").Model().Field(title).Equals("foo"))))
					fields = []FieldName{Name, fieldName{name: "Profile.BestPost.Title", json: "profile_id.best_post_id.title"}}
				}
				reset()
				t.Run("Simple query with database field names", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Search(rs.Model().FilteredOn(profile, env.Pool("Profile").Model().Field(bestPostTitle).Equals("foo"))).OrderBy("ID")
					sql, args, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title, "user".id AS id FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE "T2".title = ? ORDER BY "user".id ) foo ORDER BY id `)
					assert.Contains(t, args, "foo")
				})
				t.Run("Simple query with struct field names", func(t *testing.T) {
					reset()
					fields = []FieldName{Name, fieldName{name: "Profile.BestPost.Title", json: "profile_id.best_post_id.title"}}
					sql, args, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE "T2".title = ? ORDER BY "user".id ) foo  `)
					assert.Contains(t, args, "foo")
				})
				t.Run("Query with one2many relations", func(t *testing.T) {
					reset()
					rso2m := env.Pool("User").Search(rs.Model().Field(postsTitle).Equals("1st post"))
					fields = []FieldName{Name}
					sql, args, _ := rso2m.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user" LEFT JOIN "post" "T1" ON "user".id="T1".user_id  WHERE "T1".title = ? ORDER BY "user".id ) foo  `)
					assert.Contains(t, args, "1st post")
				})
				t.Run("Simple query with args inflation", func(t *testing.T) {
					reset()
					getUserID := func(rc *RecordCollection) int64 {
						return rc.Env().Uid()
					}
					rs2 := env.Pool("User").Search(rs.Model().Field(nums).Equals(getUserID))
					fields = []FieldName{Name}
					sql, args, _ := rs2.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".nums = ? ORDER BY "user".id ) foo  `)
					assert.EqualValues(t, len(args), 1)
					assert.Contains(t, args, security.SuperUserID)
				})
				t.Run("true/false query", func(t *testing.T) {
					reset()
					rs3 := env.Pool("User").Search(rs.Model().Field(isStaff).Equals(true))
					fields = []FieldName{Name}
					sql, args, _ := rs3.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".is_staff = ? ORDER BY "user".id ) foo  `)
					assert.EqualValues(t, len(args), 1)
					assert.Contains(t, args, true)
				})
				t.Run("Check WHERE clause with additionnal filter", func(t *testing.T) {
					reset()
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user__profile__post".title = ?) AND ("user__profile".age >= ?)`)
					assert.Contains(t, args, 12)
					assert.Contains(t, args, "foo")
				})
				t.Run("Check full query with all conditions", func(t *testing.T) {
					reset()
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					c2 := rs.Model().Field(Name).Contains("jane").Or().Field(profileMoney).Lower(1234.56)
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE (("user__profile__post".title = ?) AND ("user__profile".age >= ?)) AND ("user".name LIKE ? OR "user__profile".money < ?)`)
					assert.Contains(t, args, "%jane%")
					assert.Contains(t, args, 1234.56)
					sql, _, _ = rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id  WHERE (("T2".title = ?) AND ("T1".age >= ?)) AND ("user".name LIKE ? OR "T1".money < ?) ORDER BY "user".id ) foo  `)
				})
				t.Run("Check multi-join queries", func(t *testing.T) {
					reset()
					rs = rs.Search(rs.Model().Field(profileAge).GreaterOrEqual(12))
					c2 := rs.Model().Field(Name).Contains("jane").Or().Field(resumeEducation).Contains("MIT")
					rs = rs.Search(c2)
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE (("user__profile__post".title = ?) AND ("user__profile".age >= ?)) AND ("user".name LIKE ? OR "user__resume".education LIKE ?)`)
					assert.Contains(t, args, "%jane%")
					assert.Contains(t, args, "%MIT%")
					sql, _, _ = rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "T2".title AS profile_id__best_post_id__title FROM "user" "user" LEFT JOIN "profile" "T1" ON "user".profile_id="T1".id LEFT JOIN "post" "T2" ON "T1".best_post_id="T2".id LEFT JOIN "resume" "T3" ON "user".resume_id="T3".id  WHERE (("T2".title = ?) AND ("T1".age >= ?)) AND ("user".name LIKE ? OR "T3".education LIKE ?) ORDER BY "user".id ) foo  `)
				})
				t.Run("Testing query without WHERE clause", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"   ORDER BY "user".id ) foo  `)
				})
				t.Run("Testing query with LIMIT clause", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY id LIMIT 1 `)
				})
				t.Run("Testing query with LIMIT and OFFSET clauses", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("Limit", 1).(RecordSet).Collection().Call("Offset", 2).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY id LIMIT 1 OFFSET 2`)
				})
				t.Run("Testing query with ORDER BY clauses", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Search(rs.Model().Field(email).IContains("jane.smith@example.com")).Call("OrderBy", []string{"Email", "ID"}).(RecordSet).Collection().Load()
					fields = []FieldName{Name}
					sql, _, _ := rs.query.selectQuery(fields)
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name, "user".email AS email, "user".id AS id FROM "user" "user"  WHERE "user".email ILIKE ? ORDER BY "user".id ) foo ORDER BY email, id `)
				})
				t.Run("Testing complex conditions", func(t *testing.T) {
					reset()
					rs = env.Pool("User").Search(rs.Model().Field(profileAge).GreaterOrEqual(12).
						AndNot().Field(Name).IContains("Jane").
						OrNot().FilteredOn(profile, env.Pool("Profile").Model().Field(age).Equals(20)))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user__profile".age >= ? AND NOT "user".name ILIKE ? OR NOT "user__profile".age = ?`)
					assert.Contains(t, args, 12)
					assert.Contains(t, args, "%Jane%")
					assert.Contains(t, args, 20)
					cond1 := env.Pool("User").Model().Field(Name).IContains("Jane")
					cond2 := env.Pool("User").Model().Field(Name).IContains("John")
					rs = env.Pool("User").Search(
						env.Pool("User").Model().Field(age).GreaterOrEqual(30).
							AndNotCond(cond1).
							OrNotCond(cond2))
					sql, args = rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE (("user".age >= ?) AND NOT ("user".name ILIKE ?)) OR NOT ("user".name ILIKE ?)`)
					assert.Contains(t, args, 30)
					assert.Contains(t, args, "%Jane%")
					assert.Contains(t, args, "%John%")
				})
			}))
		}
	})
	t.Run("Testing predicate operators", func(t *testing.T) {
		if dbArgs.Driver == "postgres" {
			assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
				t.Run("Equals", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).Equals("John"))
					cond := rs.Condition()
					res := rs.CallMulti("SQLFromCondition", cond)
					sql := res[0]
					args := res[1]
					assert.EqualValues(t, sql, `"user".name = ?`)
					assert.Contains(t, args, "John")
				})
				t.Run("NotEquals", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).NotEquals("John"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NULL OR "user".name != ?)`)
					assert.Contains(t, args, "John")
				})
				t.Run("Greater", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(nums).Greater(12))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".nums > ?`)
					assert.Contains(t, args, 12)
				})
				t.Run("GreaterOrEqual", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(nums).GreaterOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".nums >= ?`)
					assert.Contains(t, args, 12)
				})
				t.Run("Lower", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(nums).Lower(12))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".nums < ?`)
					assert.Contains(t, args, 12)
				})
				t.Run("LowerOrEqual", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(nums).LowerOrEqual(12))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".nums <= ?`)
					assert.Contains(t, args, 12)
				})
				t.Run("Contains", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).Contains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".name LIKE ?`)
					assert.Contains(t, args, "%John%")
				})
				t.Run("Not Contains", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).NotContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NULL OR "user".name NOT LIKE ?)`)
					assert.Contains(t, args, "%John%")
				})
				t.Run("IContains", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).IContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".name ILIKE ?`)
					assert.Contains(t, args, "%John%")
				})
				t.Run("Not IContains", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).NotIContains("John"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NULL OR "user".name NOT ILIKE ?)`)
					assert.Contains(t, args, "%John%")
				})
				t.Run("Contains pattern", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).Like("John%"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".name LIKE ?`)
					assert.Contains(t, args, "John%")
				})
				t.Run("IContains pattern", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).ILike("John%"))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".name ILIKE ?`)
					assert.Contains(t, args, "John%")
				})
				t.Run("In", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(ID).In([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE "user".id IN (?)`)
					assert.Contains(t, args, []int64{23, 31})
				})
				t.Run("Not In", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(ID).NotIn([]int64{23, 31}))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".id IS NULL OR "user".id NOT IN (?))`)
					assert.Contains(t, args, []int64{23, 31})
				})
				t.Run("Is Null", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).IsNull())
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NULL OR "user".name = ?)`)
					assert.Contains(t, args, "")
				})
				t.Run("Is Not Null", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).IsNotNull())
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NOT NULL AND "user".name != ?)`)
					assert.Contains(t, args, "")
				})
				t.Run("Empty string", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(Name).Equals(""))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".name IS NULL OR "user".name = ?)`)
					assert.Contains(t, args, "")
				})
				t.Run("False bool", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(isStaff).Equals(false))
					sql, args := rs.query.sqlWhereClause(true)
					assert.EqualValues(t, sql, `WHERE ("user".is_staff IS NULL OR "user".is_staff = ?)`)
					assert.Contains(t, args, false)
				})
				t.Run("Child Of without parent field", func(t *testing.T) {
					rs := env.Pool("User")
					rs = rs.Search(rs.Model().Field(ID).ChildOf(101))
					sql, args, _ := rs.query.selectQuery([]FieldName{Name})
					assert.EqualValues(t, sql, `SELECT * FROM (SELECT DISTINCT ON ("user".id) "user".name AS name FROM "user" "user"  WHERE "user".id = ? ORDER BY "user".id ) foo  `)
					assert.Contains(t, args, 101)
				})
			}))
		}
	})
	t.Run("Testing Condition Methods", func(t *testing.T) {
		assert.Nil(t, SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			cond := env.Pool("User").Model().Field(Name).IContains("Jane")
			cond2 := env.Pool("User").Model().Field(ID).NotIn([]int64{23, 31})
			t.Run("HasField", func(t *testing.T) {
				assert.True(t, cond.HasField(Registry.MustGet("User").Fields().MustGet("Name")))
				assert.False(t, cond.HasField(Registry.MustGet("User").Fields().MustGet("Status")))
				assert.True(t, cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("Name")))
				assert.True(t, cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("ID")))
				assert.False(t, cond.AndCond(cond2).HasField(Registry.MustGet("User").Fields().MustGet("Status")))
			})
			t.Run("String", func(t *testing.T) {
				assert.EqualValues(t, cond.OrNotCond(cond2).String(), `AND Name ilike Jane
OR NOT (
AND ID not in [23 31]

)
`)
			})
		}))
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
	t.Run("Testing condition serialization", func(t *testing.T) {
		t.Run("Testing simple A AND B condition", func(t *testing.T) {
			cond := newCondition().And().Field(Name).IContains("John").And().Field(age).Greater(18)
			dom := cond.Serialize()
			assert.EqualValues(t, fmt.Sprint(dom), "[& [name ilike John] [age > 18]]")
		})
		t.Run("Testing simple A OR B condition", func(t *testing.T) {
			cond := newCondition().And().Field(Name).IContains("John").Or().Field(age).Greater(18)
			dom := cond.Serialize()
			assert.EqualValues(t, fmt.Sprint(dom), "[| [age > 18] [name ilike John]]")
		})
		t.Run("Testing A AND B OR C condition", func(t *testing.T) {
			cond := newCondition().And().Field(Name).IContains("John").And().Field(age).Greater(18).Or().Field(isStaff).Equals(true)
			dom := cond.Serialize()
			assert.EqualValues(t, fmt.Sprint(dom), "[| [is_staff = true] & [name ilike John] [age > 18]]")
		})
		t.Run("Testing (A OR B) AND (C OR D) OR F condition", func(t *testing.T) {
			aOrB := newCondition().And().Field(a).Equals("A Value").Or().Field(b).Equals("B Value")
			cOrD := newCondition().And().Field(c).Equals("C Value").Or().Field(d).Equals("D Value")
			cond := newCondition().AndCond(aOrB).AndCond(cOrD).Or().Field(f).Equals("F Value")
			dom := cond.Serialize()
			assert.EqualValues(t, fmt.Sprint(dom), "[| [F = F Value] & | [B = B Value] [A = A Value] | [D = D Value] [C = C Value]]")
		})
		t.Run("Testing (A OR B OR C) AND (D) condition", func(t *testing.T) {
			aOrBOrC := newCondition().And().Field(a).Equals("A Value").Or().Field(b).Equals("B Value").Or().Field(c).Equals("C Value")
			D := newCondition().And().Field(d).Equals("D Value")
			cond := newCondition().AndCond(aOrBOrC).AndCond(D)
			dom := cond.Serialize()
			assert.EqualValues(t, fmt.Sprint(dom), "[& | [C = C Value] | [B = B Value] [A = A Value] [D = D Value]]")
		})
	})
}
