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
	"sync"

	"github.com/hexya-erp/hexya/src/models/security"
)

// A RecordRule allow to grant a group some permissions
// on a selection of records.
// - If Global is true, then the RecordRule applies to all groups
// - Condition is the filter to apply on the model to retrieve
// the records on which to allow the Perms permission.
type RecordRule struct {
	Name      string
	Global    bool
	Group     *security.Group
	Condition *Condition
	Perms     security.Permission
}

// A RecordRuleRegistry keeps a list of RecordRule. It is meant
// to be attached to a model.
type recordRuleRegistry struct {
	sync.RWMutex
	rulesByName  map[string]*RecordRule
	rulesByGroup map[string][]*RecordRule
	globalRules  map[string]*RecordRule
}

// AddRule registers the given RecordRule to the registry with the given name.
func (rrr *recordRuleRegistry) addRule(rule *RecordRule) {
	rrr.Lock()
	defer rrr.Unlock()
	rrr.rulesByName[rule.Name] = rule
	if rule.Global {
		rrr.globalRules[rule.Name] = rule
	} else {
		rrr.rulesByGroup[rule.Group.Name] = append(rrr.rulesByGroup[rule.Group.Name], rule)
	}
}

// RemoveRule removes the RecordRule with the given name
// from the rule registry.
func (rrr *recordRuleRegistry) removeRule(name string) {
	rrr.Lock()
	defer rrr.Unlock()
	rule, exists := rrr.rulesByName[name]
	if !exists {
		log.Warn("Trying to remove non-existent record rule", "name", name)
		return
	}
	delete(rrr.rulesByName, name)
	if rule.Global {
		delete(rrr.globalRules, name)
	} else {
		newRuleSlice := make([]*RecordRule, len(rrr.rulesByGroup[rule.Group.Name])-1)
		i := 0
		for _, r := range rrr.rulesByGroup[rule.Group.Name] {
			if r.Name == rule.Name {
				continue
			}
			newRuleSlice[i] = r
			i++
		}
		rrr.rulesByGroup[rule.Group.Name] = newRuleSlice
	}
}

// newRecordRuleRegistry returns a pointer to a new RecordRuleRegistry instance
func newRecordRuleRegistry() *recordRuleRegistry {
	return &recordRuleRegistry{
		rulesByName:  make(map[string]*RecordRule),
		rulesByGroup: make(map[string][]*RecordRule),
		globalRules:  make(map[string]*RecordRule),
	}
}

// AddRecordRule registers the given RecordRule to the registry for
// the given model with the given name.
func (m *Model) AddRecordRule(rule *RecordRule) {
	m.rulesRegistry.addRule(rule)
}

// RemoveRecordRule removes the Record Rule with the given name
// from the rule registry of the given model.
func (m *Model) RemoveRecordRule(name string) {
	m.rulesRegistry.removeRule(name)
}
