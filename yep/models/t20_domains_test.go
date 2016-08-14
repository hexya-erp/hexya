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

	. "github.com/smartystreets/goconvey/convey"
)

func TestDomains(t *testing.T) {
	Convey("Testing Domains", t, func() {
		env := NewEnvironment(1)
		Convey("Creating an extra user", func() {
			user := User_WithID{
				UserName: "Martin Weston",
				Email:    "mweston@example.com",
				Profile: &Profile_WithID{
					Age: 45,
				},
			}
			env.Pool("Profile").Create(user.Profile)
			So(user.Profile.ID, ShouldNotEqual, 0)
			env.Pool("User").Create(&user)
			var userRes User_Simple
			env.Pool("User").Filter("UserName", "=", "Martin Weston").RelatedDepth(1).ReadOne(&userRes)
			So(userRes.Profile.Age, ShouldEqual, 45)
		})
		Convey("Testing simple [(A), (B)] domain", func() {
			dom1 := []interface{}{
				0: []interface{}{"UserName", "like", "Smith"},
				1: []interface{}{"Age", "=", 24},
			}
			var dom1Users []*User_WithDecoratedName
			env.Pool("User").Condition(ParseDomain(dom1)).ReadAll(&dom1Users)
			So(len(dom1Users), ShouldEqual, 1)
			So(dom1Users[0].UserName, ShouldEqual, "Jane A. Smith")
		})
		Convey("Testing ['|', (A), (B)] domain", func() {
			dom2 := []interface{}{
				0: "|",
				1: []interface{}{"UserName", "like", "Will"},
				2: []interface{}{"Email", "ilike", "Jane.Smith"},
			}
			var dom2Users []*User_WithDecoratedName
			env.Pool("User").Condition(ParseDomain(dom2)).OrderBy("UserName").ReadAll(&dom2Users)
			So(len(dom2Users), ShouldEqual, 2)
			So(dom2Users[0].UserName, ShouldEqual, "Jane A. Smith")
			So(dom2Users[1].UserName, ShouldEqual, "Will Smith")
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
			var dom3Users []*User_WithDecoratedName
			env.Pool("User").Condition(ParseDomain(dom3)).OrderBy("UserName").ReadAll(&dom3Users)
			So(len(dom3Users), ShouldEqual, 1)
			So(dom3Users[0].UserName, ShouldEqual, "Jane A. Smith")
		})
		env.cr.Rollback()
	})

}
