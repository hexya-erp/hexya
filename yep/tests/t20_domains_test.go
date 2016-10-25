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

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDomains(t *testing.T) {
	Convey("Testing Domains", t, func() {
		env := models.NewEnvironment(1)
		Convey("Creating an extra user", func() {
			profile := pool.NewTest__ProfileSet(env).Create(&pool.Test__Profile{Age: 45})
			userData := pool.Test__User{
				UserName: "Martin Weston",
				Email:    "mweston@example.com",
				Profile:  profile,
			}
			user := pool.NewTest__UserSet(env).Create(&userData)
			So(user.Profile().Age(), ShouldEqual, 45)
		})
		Convey("Testing simple [(A), (B)] domain", func() {
			dom1 := []interface{}{
				0: []interface{}{"UserName", "like", "Smith"},
				1: []interface{}{"Age", "=", 24},
			}
			dom1Users := pool.NewTest__UserSet(env).Search(models.ParseDomain(dom1))
			So(dom1Users.Len(), ShouldEqual, 1)
			So(dom1Users.UserName(), ShouldEqual, "Jane A. Smith")
		})
		Convey("Testing ['|', (A), (B)] domain", func() {
			dom2 := []interface{}{
				0: "|",
				1: []interface{}{"UserName", "like", "Will"},
				2: []interface{}{"Email", "ilike", "Jane.Smith"},
			}
			dom2Users := pool.NewTest__UserSet(env).Search(models.ParseDomain(dom2)).OrderBy("UserName")
			So(dom2Users.Len(), ShouldEqual, 2)
			userRecs := dom2Users.Records()
			So(userRecs[0].UserName(), ShouldEqual, "Jane A. Smith")
			So(userRecs[1].UserName(), ShouldEqual, "Will Smith")
		})
		Convey("Testing ['|', (A), '&' , (B), (C), (D)] domain", func() {
			dom3 := []interface{}{
				0: "|",
				1: []interface{}{"UserName", "like", "Will"},
				2: "&",
				3: []interface{}{"Age", ">", 0},
				4: []interface{}{"Age", "<", 25},
				5: []interface{}{"Email", "not like", "will.smith"},
			}
			dom3Users := pool.NewTest__UserSet(env).Search(models.ParseDomain(dom3)).OrderBy("UserName")
			So(dom3Users.Len(), ShouldEqual, 1)
			So(dom3Users.UserName(), ShouldEqual, "Jane A. Smith")
		})
		env.Rollback()
	})
}
