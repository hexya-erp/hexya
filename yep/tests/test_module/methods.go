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

package test_module

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
}

func PrefixedUser(rs pool.Test__UserSet, prefix string) []string {
	var res []string
	for _, u := range rs.Records() {
		res = append(res, fmt.Sprintf("%s: %s", prefix, u.UserName()))
	}
	return res
}

//models.ExtendMethod("Test__User", "PrefixedUser",
//	func(rs pool.Test__UserSet, prefix string) []string {
//		res := rs.Super(prefix).([]string)
//		for i, u := range rs.Records() {
//			res[i] = fmt.Sprintf("%s %s", res[i], rs.DecorateEmail(u.Email()))
//		}
//		return res
//	})

func DecorateEmail(rs pool.Test__UserSet, email string) string {
	return fmt.Sprintf("<%s>", email)
}

func DecorateEmailExtension(rs pool.Test__UserSet, email string) string {
	res := rs.Super(email).(string)
	return fmt.Sprintf("[%s]", res)
}

//models.CreateMethod("Test__User", "computeDecoratedName",
//	func(rs pool.Test__UserSet) models.FieldMap {
//		res := make(models.FieldMap)
//		res["DecoratedName"] = rs.PrefixedUser("User").([]string)[0]
//		return res
//	})

func ComputeAge(rs pool.Test__UserSet) models.FieldMap {
	res := make(models.FieldMap)
	res["Age"] = rs.Profile().Age()
	return res
}
