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

// An AccessControlList defines the permissions for given groups.
// It is meant to be a property of a field.
type AccessControlList struct {
	sync.RWMutex
	perms map[*Group]Permission
}

// AddPermission adds the given permission to the given group, keeping
// existing permissions untouched.
func (acl *AccessControlList) AddPermission(group *Group, perm Permission) {
	acl.Lock()
	defer acl.Unlock()
	acl.perms[group] |= perm
}

// RemovePermission removes the given permission from the given group, keeping
// other permissions untouched.
func (acl *AccessControlList) RemovePermission(group *Group, perm Permission) {
	acl.Lock()
	defer acl.Unlock()
	acl.perms[group] &^= perm
}

// ReplacePermission replaces the current permission of the given group, by
// the given perm. It overrides any existing permission.
func (acl *AccessControlList) ReplacePermission(group *Group, perm Permission) {
	acl.Lock()
	defer acl.Unlock()
	acl.perms[group] = perm
}

// CheckPermission returns true if the given group has the given permission,
// either directly granted to it or granted to one of its inherited groups.
func (acl *AccessControlList) CheckPermission(group *Group, perm Permission) bool {
	if perm == 0 {
		log.Panic("Trying to check nil permission for group", "group", group.Name)
	}
	if acl.perms[group]&perm == perm {
		return true
	}
	for _, inhGroup := range group.Inherits {
		return acl.CheckPermission(inhGroup, perm)
	}
	return false
}

// Permissions returns the list of all permissions of this ACL
func (acl *AccessControlList) Permissions() map[*Group]Permission {
	res := make(map[*Group]Permission)
	for k, v := range acl.perms {
		res[k] = v
	}
	return res
}

// NewAccessControlList returns a pointer to a new empty AccessControlList
func NewAccessControlList() *AccessControlList {
	acl := AccessControlList{
		perms: make(map[*Group]Permission),
	}
	acl.perms[GroupEveryone] = All
	return &acl
}
