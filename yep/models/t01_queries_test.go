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
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

var DBARGS = struct {
	Driver string
	Source string
	Debug  string
}{
	os.Getenv("ORM_DRIVER"),
	os.Getenv("ORM_SOURCE"),
	os.Getenv("ORM_DEBUG"),
}

func TestConditions(t *testing.T) {
	Convey("Testing SQL building for conditions", t, func() {
		DB = sqlx.MustConnect(DBARGS.Driver, DBARGS.Source)
		query := Query{
			cond: NewCondition(),
		}
		if DBARGS.Driver == "postgres" {
			query.cond = query.cond.And("user_id.profile_id.name", "=", "foo")
			sql, args := query.sqlWhereClause()
			So(sql, ShouldEqual, "WHERE user_id__profile_id.name = ? ")
			So(args, ShouldContain, "foo")

			query.cond = query.cond.And("age", ">=", 12)
			sql, args = query.sqlWhereClause()
			So(sql, ShouldEqual, "WHERE user_id__profile_id.name = ? AND age >= ? ")
			So(args, ShouldContain, 12)
			So(args, ShouldContain, "foo")

			c2 := NewCondition().And("user_id.name", "like", "jane").Or("user_id.money", "<", 1234.56)
			query.cond = query.cond.OrCond(c2)
			sql, args = query.sqlWhereClause()
			So(sql, ShouldEqual, "WHERE user_id__profile_id.name = ? AND age >= ? OR (user_id.name LIKE %?% OR user_id.money < ? ) ")
			So(args, ShouldContain, "jane")
			So(args, ShouldContain, 1234.56)
		}
	})

}
