/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDomains(t *testing.T) {
	Convey("Testing Domains", t, func() {
		env = NewEnvironment(dORM, 1)
		Convey("Creating an extra user", func() {
			user := User_WithID{
				UserName: "Martin Weston",
				Email:    "mweston@example.com",
				Profile: &Profile_WithID{
					Age: 45,
				},
			}
			env.Pool("Profile").Call("Create", user.Profile)
			env.Pool("User").Call("Create", &user)
			var userRes User_Simple
			env.Pool("User").Filter("UserName", "Martin Weston").ReadOne(&userRes)
			So(userRes.Age, ShouldEqual, 45)
		})
		Convey("Testing simple [(A), (B)] domain", func() {
			dom1 := []interface{}{
				0: []interface{}{"UserName", "like", "Smith"},
				1: []interface{}{"Age", "=", 24},
			}
			var dom1Users []*User_WithDecoratedName
			env.Pool("User").SetCond(ParseDomain(dom1)).ReadAll(&dom1Users)
			So(len(dom1Users), ShouldEqual, 1)
			So(dom1Users[0].UserName, ShouldEqual, "Jane A. Smith")
		})
		Convey("Testing ['|', (A), (B)] domain", func() {
			dom2 := []interface{}{
				0: "|",
				1: []interface{}{"UserName", "like", "Will"},
				2: []interface{}{"Age", "<", 25},
			}
			var dom2Users []*User_WithDecoratedName
			env.Pool("User").SetCond(ParseDomain(dom2)).OrderBy("UserName").ReadAll(&dom2Users)
			So(len(dom2Users), ShouldEqual, 2)
			So(dom2Users[0].UserName, ShouldEqual, "Jane A. Smith")
			So(dom2Users[1].UserName, ShouldEqual, "Will Smith")
		})
		Convey("Testing ['|', (A), (B), (C)] domain", func() {
			dom3 := []interface{}{
				0: "|",
				1: []interface{}{"UserName", "like", "Will"},
				2: []interface{}{"Age", "<", 25},
				3: []interface{}{"Email", "not like", "will.smith"},
			}
			var dom3Users []*User_WithDecoratedName
			env.Pool("User").SetCond(ParseDomain(dom3)).OrderBy("UserName").ReadAll(&dom3Users)
			So(len(dom3Users), ShouldEqual, 1)
			So(dom3Users[0].UserName, ShouldEqual, "Jane A. Smith")
		})
	})

}
