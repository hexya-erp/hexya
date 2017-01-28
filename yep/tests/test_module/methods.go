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

package testmodule

import (
	"fmt"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

func declareMethods() {
	user := models.Registry.MustGet("User")
	user.CreateMethod("PrefixedUser",
		`PrefixedUser is a sample method layer for testing`,
		func(rs pool.UserSet, prefix string) []string {
			var res []string
			for _, u := range rs.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName()))
			}
			return res
		})

	user.CreateMethod("DecorateEmail",
		`DecorateEmail is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			return fmt.Sprintf("<%s>", email)
		})

	user.ExtendMethod("DecorateEmail",
		`DecorateEmailExtension is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			res := rs.Super(email).(string)
			return fmt.Sprintf("[%s]", res)
		})

	user.CreateMethod("computeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			res := pool.UserData{
				Age: rs.Profile().Age(),
			}
			return &res, []models.FieldNamer{pool.User().Age()}
		})

	user.ExtendMethod("PrefixedUser", "",
		func(rs pool.UserSet, prefix string) []string {
			res := rs.Super(prefix).([]string)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	user.CreateMethod("computeDecoratedName", "",
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			res := pool.UserData{
				DecoratedName: rs.PrefixedUser("User")[0],
			}
			return &res, []models.FieldNamer{pool.User().DecoratedName()}
		})

	addressMI := pool.AddressMixIn()
	addressMI.CreateMethod("SayHello",
		`SayHello is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return "Hello !"
		})

	addressMI.CreateMethod("PrintAddress",
		`PrintAddressMixIn is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
		})

	pool.Profile().CreateMethod("PrintAddress",
		`PrintAddress is a sample method layer for testing`,
		func(rs pool.ProfileSet) string {
			res := rs.Super()
			return fmt.Sprintf("%s, %s", res, rs.Country())
		})

	addressMI.ExtendMethod("PrintAddress", "",
		func(rs pool.AddressMixInSet) string {
			res := rs.Super()
			return fmt.Sprintf("<%s>", res)
		})

	pool.Profile().ExtendMethod("PrintAddress", "",
		func(rs pool.ProfileSet) string {
			res := rs.Super()
			return fmt.Sprintf("[%s]", res)
		})

	// Chained declaration
	activeMI1 := pool.ActiveMixIn()
	activeMI := activeMI1
	activeMI.CreateMethod("IsActivated",
		`IsACtivated is a sample method of ActiveMixIn"`,
		func(rs pool.ActiveMixInSet) bool {
			return rs.Active()
		})
}
