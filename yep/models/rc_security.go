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

import "github.com/npiganeau/yep/yep/models/security"

// addRecordRuleConditions adds the RecordRule conditions on the query of this
// RecordSet for the user with the given uid and for the given perm Permission.
func (rc RecordCollection) addRecordRuleConditions(uid int64, perm security.Permission) RecordCollection {
	if rc.filtered {
		return rc
	}
	rSet := rc
	// Add global rules
	for _, rule := range rSet.model.rulesRegistry.globalRules {
		if perm&rule.Perms > 0 {
			rSet = rSet.Search(rule.Condition)
		}
	}
	// Add groups rules
	userGroups := security.AuthenticationRegistry.UserGroups(uid)
	groupCondition := newCondition()
	for _, group := range userGroups {
		for _, rule := range rSet.model.rulesRegistry.rulesByGroup[group.Name] {
			if perm&rule.Perms > 0 {
				groupCondition = groupCondition.OrCond(rule.Condition)
			}
		}
	}
	if !groupCondition.IsEmpty() {
		rSet = rSet.Search(groupCondition)
	}
	rSet.filtered = true
	return rSet
}
