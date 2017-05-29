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

import "github.com/hexya-erp/hexya/hexya/models/security"

// GrantAccess grants the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func (f *Field) GrantAccess(group *security.Group, perm security.Permission) *Field {
	perm = perm & (security.Read | security.Write)
	f.acl.AddPermission(group, perm)
	return f
}

// RevokeAccess denies the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func (f *Field) RevokeAccess(group *security.Group, perm security.Permission) *Field {
	perm = perm & (security.Read | security.Write)
	f.acl.RemovePermission(group, perm)
	return f
}

// checkFieldPermission checks if the given uid has the given perm on the given field info.
func checkFieldPermission(f *Field, uid int64, perm security.Permission) bool {
	userGroups := security.Registry.UserGroups(uid)
	for group := range userGroups {
		if f.acl.CheckPermission(group, perm) {
			return true
		}
	}
	return false
}

// filterOnAuthorizedFields returns the fields slice with only the fields on
// which the current user has the given permission.
func filterOnAuthorizedFields(m *Model, uid int64, fields []string, perm security.Permission) []string {
	perm = perm & (security.Read | security.Write)
	if perm == 0 {
		// We are trying to check perms that are not read or write which
		// means they don't apply to fields, so we return the whole slice
		return fields
	}
	var res []string

	for _, field := range fields {
		f := m.getRelatedFieldInfo(field)
		if checkFieldPermission(f, uid, perm) {
			res = append(res, field)
		}
	}
	return res
}

// filterMapOnAuthorizedFields returns a new FieldMap from fMap
// with only the fields on which the given uid user has access.
// All field names are JSONized.
func filterMapOnAuthorizedFields(m *Model, fMap FieldMap, uid int64, perm security.Permission) FieldMap {
	perm = perm & (security.Read | security.Write)
	if perm == 0 {
		// We are trying to check perms that are not read or write which
		// means they don't apply to fields, so we return the whole map
		return fMap
	}
	newFMap := make(FieldMap)
	for field, value := range fMap {
		f := m.getRelatedFieldInfo(field)
		if checkFieldPermission(f, uid, perm) {
			newFMap[f.json] = value
		}
	}
	return newFMap
}
