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

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/models"
	"github.com/hexya-erp/hexya/src/models/security"
	"github.com/hexya-erp/pool/h"
	"github.com/hexya-erp/pool/q"
)

func TestConditions(t *testing.T) {
	t.Run("Testing SQL building for queries", func(t *testing.T) {
		if driver == "postgres" {
			assert.Nil(t, models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				rs := h.User().NewSet(env)
				rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().BestPostFilteredOn(q.Post().Title().Equals("foo"))))
				t.Run("Simple query", func(t *testing.T) {
					assert.NotPanics(t, func() { rs.Load() })
				})
				t.Run("Simple query with args inflation", func(t *testing.T) {
					getUserID := func(rs models.RecordSet) int {
						return int(rs.Env().Uid())
					}
					rs2 := h.User().Search(env, q.User().Nums().EqualsFunc(getUserID))
					assert.NotPanics(t, func() { rs2.Load() })
				})
				t.Run("Check WHERE clause with additionnal filter", func(t *testing.T) {
					rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().Age().GreaterOrEqual(12)))
					assert.NotPanics(t, func() { rs.Load() })
				})
				t.Run("Check full query with all conditions", func(t *testing.T) {
					rs = rs.Search(q.User().ProfileFilteredOn(q.Profile().Age().GreaterOrEqual(12)).Or().Name().ILike("John"))
					c2 := q.User().Name().Like("jane").Or().ProfileFilteredOn(q.Profile().Money().Lower(1234.56))
					rs = rs.Search(c2)
					rs.Load()
					assert.NotPanics(t, func() { rs.Load() })
				})
			}))
		}
	})
}
