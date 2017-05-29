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

	"github.com/hexya-erp/hexya-base/web/domains"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDomains(t *testing.T) {
	Convey("Testing Domains", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating an extra user", func() {
				profile := pool.Profile().Create(env, &pool.ProfileData{Age: 45})
				userData := pool.UserData{
					Name:    "Martin Weston",
					Email:   "mweston@example.com",
					Profile: profile,
				}
				user := pool.User().Create(env, &userData)
				So(user.Profile().Age(), ShouldEqual, 45)
			})
			Convey("Testing simple [(A), (B)] domain", func() {
				dom1 := []interface{}{
					0: []interface{}{"Name", "like", "Smith"},
					1: []interface{}{"Age", "=", 24},
				}
				dom1Users := pool.User().Search(env, pool.UserCondition{Condition: domains.ParseDomain(dom1)})
				So(dom1Users.Len(), ShouldEqual, 1)
				So(dom1Users.Name(), ShouldEqual, "Jane A. Smith")
			})
			Convey("Testing ['|', (A), (B)] domain", func() {
				dom2 := []interface{}{
					0: "|",
					1: []interface{}{"Name", "like", "Will"},
					2: []interface{}{"Email", "ilike", "Jane.Smith"},
				}
				dom2Users := pool.User().Search(env, pool.UserCondition{Condition: domains.ParseDomain(dom2)}).OrderBy("Name")
				So(dom2Users.Len(), ShouldEqual, 2)
				userRecs := dom2Users.Records()
				So(userRecs[0].Name(), ShouldEqual, "Jane A. Smith")
				So(userRecs[1].Name(), ShouldEqual, "Will Smith")
			})
			Convey("Testing ['|', (A), '&' , (B), (C), (D)] domain", func() {
				dom3 := []interface{}{
					0: "|",
					1: []interface{}{"Name", "like", "Will"},
					2: "&",
					3: []interface{}{"Age", ">", 0},
					4: []interface{}{"Age", "<", 25},
					5: []interface{}{"Email", "not like", "will.smith"},
				}
				dom3Users := pool.User().Search(env, pool.UserCondition{Condition: domains.ParseDomain(dom3)}).OrderBy("Name")
				So(dom3Users.Len(), ShouldEqual, 1)
				So(dom3Users.Name(), ShouldEqual, "Jane A. Smith")
			})
		})
	})
}
