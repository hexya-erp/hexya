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
	models.CreateMethod("User", "PrefixedUser",
		`PrefixedUser is a sample method layer for testing`,
		func(rs pool.UserSet, prefix string) []string {
			var res []string
			for _, u := range rs.Records() {
				res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName()))
			}
			return res
		})

	models.CreateMethod("User", "DecorateEmail",
		`DecorateEmail is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			return fmt.Sprintf("<%s>", email)
		})

	models.ExtendMethod("User", "DecorateEmail",
		`DecorateEmailExtension is a sample method layer for testing`,
		func(rs pool.UserSet, email string) string {
			res := rs.Super(email).(string)
			return fmt.Sprintf("[%s]", res)
		})

	models.CreateMethod("User", "computeAge",
		`ComputeAge is a sample method layer for testing`,
		func(rs pool.UserSet) (*pool.User, []models.FieldName) {
			res := pool.User{
				Age: rs.Profile().Age(),
			}
			return &res, []models.FieldName{pool.User_Age}
		})

	models.ExtendMethod("User", "PrefixedUser", "",
		func(rs pool.UserSet, prefix string) []string {
			res := rs.Super(prefix).([]string)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	models.CreateMethod("User", "computeDecoratedName", "",
		func(rs pool.UserSet) (*pool.User, []models.FieldName) {
			res := pool.User{
				DecoratedName: rs.PrefixedUser("User")[0],
			}
			return &res, []models.FieldName{pool.User_DecoratedName}
		})

	models.CreateMethod("AddressMixIn", "SayHello",
		`SayHello is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return "Hello !"
		})

	models.CreateMethod("AddressMixIn", "PrintAddress",
		`PrintAddressMixIn is a sample method layer for testing`,
		func(rs pool.AddressMixInSet) string {
			return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
		})

	models.CreateMethod("Profile", "PrintAddress",
		`PrintAddress is a sample method layer for testing`,
		func(rs pool.ProfileSet) string {
			res := rs.Super()
			return fmt.Sprintf("%s, %s", res, rs.Country())
		})

	models.ExtendMethod("AddressMixIn", "PrintAddress", "",
		func(rs pool.AddressMixInSet) string {
			res := rs.Super()
			return fmt.Sprintf("<%s>", res)
		})

	models.ExtendMethod("Profile", "PrintAddress", "",
		func(rs pool.ProfileSet) string {
			res := rs.Super()
			return fmt.Sprintf("[%s]", res)
		})

	models.CreateMethod("ActiveMixIn", "IsActivated",
		`IsACtivated is a sample method of ActiveMixIn"`,
		func(rs pool.ActiveMixInSet) bool {
			return rs.Active()
		})
}
