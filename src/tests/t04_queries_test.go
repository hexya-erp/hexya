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

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConditions(t *testing.T) {
	Convey("Testing SQL building for queries", t, func() {
		if driver == "postgres" {
			So(models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				rs := h.User().NewSet(env)
				rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().BestPostFilteredOn(q.Post().Title().Equals("foo"))))
				Convey("Simple query", func() {
					So(func() { rs.Load() }, ShouldNotPanic)
				})
				Convey("Simple query with args inflation", func() {
					getUserID := func(rs models.RecordSet) int {
						return int(rs.Env().Uid())
					}
					rs2 := h.User().Search(env, q.User().Nums().EqualsFunc(getUserID))
					So(func() { rs2.Load() }, ShouldNotPanic)
				})
				Convey("Check WHERE clause with additionnal filter", func() {
					rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().Age().GreaterOrEqual(12)))
					So(func() { rs.Load() }, ShouldNotPanic)
				})
				Convey("Check full query with all conditions", func() {
					rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().Age().GreaterOrEqual(12)).Or().Name().ILike("John"))
					c2 := q.User().Name().Like("jane").Or().ProfileFilteredOn(q.Profile().Money().Lower(1234.56))
					rs = rs.Search(c2)
					rs.Load()
					So(func() { rs.Load() }, ShouldNotPanic)
				})
			}), ShouldBeNil)
		}
	})
}
