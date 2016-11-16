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

package security

import "sync"

const (
	// SuperUserID is the uid of the administrator
	SuperUserID int64 = 1
	// AdminGroupName is the name of the group with all permissions
	AdminGroupName string = "Admin"
)

// GroupRegistry of the application
var GroupRegistry *GroupCollection

// AdminGroup which has all permissions
var AdminGroup *Group

// A Group defines a role which can be granted or denied permissions.
// - Groups can inherit from other groups and get access to these groups
// permissions.
// - A user can belong to one or several groups, and thus inherit from the
// permissions of the groups.
type Group struct {
	Name     string
	Inherits []*Group
}

// NewGroup returns a pointer to a new Group with the given name
// and inheriting from the given inherits groups.
func NewGroup(name string, inherits ...*Group) *Group {
	return &Group{
		Name:     name,
		Inherits: inherits,
	}
}

// A GroupCollection keeps a list of groups
type GroupCollection struct {
	sync.RWMutex
	groups map[string]*Group
}

// RegisterGroup adds the given group to this GroupCollection
// If group with the same name exists, it is replaced by this one.
func (gc *GroupCollection) RegisterGroup(group *Group) {
	gc.Lock()
	defer gc.Unlock()
	gc.groups[group.Name] = group
}

// UnregisterGroup removes the given group to this GroupCollection
func (gc *GroupCollection) UnregisterGroup(group *Group) {
	gc.Lock()
	defer gc.Unlock()
	delete(gc.groups, group.Name)
}

// Get returns the group with the given groupName or nil if
// not found
func (gc *GroupCollection) Get(groupName string) *Group {
	return gc.groups[groupName]
}

// NewGroupCollection returns a pointer to a new empty GroupCollection
func NewGroupCollection() *GroupCollection {
	gc := GroupCollection{
		groups: make(map[string]*Group),
	}
	return &gc
}
