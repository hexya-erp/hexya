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
	models.CreateMethod("Test__User", "PrefixedUser", PrefixedUser)
	models.CreateMethod("Test__User", "DecorateEmail", DecorateEmail)
	models.ExtendMethod("Test__User", "DecorateEmail", DecorateEmailExtension)
	models.CreateMethod("Test__User", "computeAge", ComputeAge)

	models.ExtendMethod("Test__User", "PrefixedUser",
		func(rs pool.Test__UserSet, prefix string) []string {
			res := rs.Super(prefix).([]string)
			for i, u := range rs.Records() {
				res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
			}
			return res
		})

	models.CreateMethod("Test__User", "computeDecoratedName",
		func(rs pool.Test__UserSet) (*pool.Test__User, []models.FieldName) {
			res := pool.Test__User{
				DecoratedName: rs.PrefixedUser("User")[0],
			}
			return &res, []models.FieldName{pool.Test__User_DecoratedName}
		})

	models.CreateMethod("Test__AddressMixIn", "SayHello", SayHello)
	models.CreateMethod("Test__AddressMixIn", "PrintAddress", PrintAddressMixIn)
	models.CreateMethod("Test__Profile", "PrintAddress", PrintAddress)

	models.ExtendMethod("Test__AddressMixIn", "PrintAddress", func(rs pool.Test__AddressMixInSet) string {
		res := rs.Super()
		return fmt.Sprintf("<%s>", res)
	})

	models.ExtendMethod("Test__Profile", "PrintAddress", func(rs pool.Test__ProfileSet) string {
		res := rs.Super()
		return fmt.Sprintf("[%s]", res)
	})

	models.CreateMethod("Test__ActiveMixIn", "IsActivated", func(rs pool.Test__ActiveMixInSet) bool {
		return rs.Active()
	})
}

// PrefixedUser is a sample method layer for testing
func PrefixedUser(rs pool.Test__UserSet, prefix string) []string {
	var res []string
	for _, u := range rs.Records() {
		res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName()))
	}
	return res
}

// DecorateEmail is a sample method layer for testing
func DecorateEmail(rs pool.Test__UserSet, email string) string {
	return fmt.Sprintf("<%s>", email)
}

// DecorateEmailExtension is a sample method layer for testing
func DecorateEmailExtension(rs pool.Test__UserSet, email string) string {
	res := rs.Super(email).(string)
	return fmt.Sprintf("[%s]", res)
}

// ComputeAge is a sample method layer for testing
func ComputeAge(rs pool.Test__UserSet) (*pool.Test__User, []models.FieldName) {
	res := pool.Test__User{
		Age: rs.Profile().Age(),
	}
	return &res, []models.FieldName{pool.Test__User_Age}
}

// PrintAddress is a sample method layer for testing
func PrintAddress(rs pool.Test__ProfileSet) string {
	res := rs.Super()
	return fmt.Sprintf("%s, %s", res, rs.Country())
}

// PrintAddressMixIn is a sample method layer for testing
func PrintAddressMixIn(rs pool.Test__AddressMixInSet) string {
	return fmt.Sprintf("%s, %s %s", rs.Street(), rs.Zip(), rs.City())
}

// SayHello is a sample method layer for testing
func SayHello(rs pool.Test__AddressMixInSet) string {
	return "Hello !"
}
