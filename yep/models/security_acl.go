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

// AllowModelAccess grants the given permission to the given group for the given model.
// This also sets the permissions for all fields of the model.
func (m *Model) AllowModelAccess(group *security.Group, perm security.Permission) {
	// Add perm to the model's acl
	m.acl.AddPermission(group, perm)
	// Add perm to all fields of the model
	for fName := range m.fields.registryByName {
		m.AllowFieldAccess(FieldName(fName), group, perm)
	}
}

// AllowFieldAccess grants the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func (m *Model) AllowFieldAccess(field FieldNamer, group *security.Group, perm security.Permission) {
	perm = perm & (security.Read | security.Write | security.Create)
	if !m.acl.CheckPermission(group, security.Read) {
		log.Warn("Trying to add permission on field, but model is not readable", "model", m, "field", field, "perm", perm)
	}
	fi := m.fields.mustGet(string(field.FieldName()))
	fi.acl.AddPermission(group, perm)
}

// DenyModelAccess denies the given permission to the given group for the given model.
// This also unsets the permissions for all fields of the model.
func (m *Model) DenyModelAccess(group *security.Group, perm security.Permission) {
	// Remove perm to the model's acl
	m.acl.RemovePermission(group, perm)
	// Remove perm to all fields of the model
	for fName := range m.fields.registryByName {
		m.DenyFieldAccess(FieldName(fName), group, perm)
	}
}

// DenyFieldAccess denies the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func (m *Model) DenyFieldAccess(field FieldNamer, group *security.Group, perm security.Permission) {
	perm = perm & (security.Read | security.Write | security.Create)
	fi := m.fields.mustGet(string(field.FieldName()))
	fi.acl.RemovePermission(group, perm)
}

// mustCheckModelPermission checks if the given uid has the given perm on the given model info.
// It panics if the user does not have the required permission and does nothing otherwise.
func mustCheckModelPermission(mi *Model, uid int64, perm security.Permission) {
	userGroups := security.Registry.UserGroups(uid)
	for group := range userGroups {
		if mi.acl.CheckPermission(group, perm) {
			return
		}
	}
	log.Panic("User does not have required permission on model", "uid", uid, "model", mi.name, "permission", perm)
}

// checkFieldPermission checks if the given uid has the given perm on the given field info.
func checkFieldPermission(fi *fieldInfo, uid int64, perm security.Permission) bool {
	userGroups := security.Registry.UserGroups(uid)
	for group := range userGroups {
		if fi.acl.CheckPermission(group, perm) {
			return true
		}
	}
	return false
}

// filterOnAuthorizedFields returns the fields slice with only the fields on
// which the current user has the given permission.
func filterOnAuthorizedFields(mi *Model, uid int64, fields []string, perm security.Permission) []string {
	perm = perm & (security.Read | security.Write | security.Create)
	if perm == 0 {
		// We are trying to check perms that are not read or write which
		// means they don't apply to fields, so we return the whole slice
		return fields
	}
	var res []string

	for _, field := range fields {
		fi := mi.getRelatedFieldInfo(field)
		if checkFieldPermission(fi, uid, perm) {
			res = append(res, field)
		}
	}
	return res
}

// filterMapOnAuthorizedFields returns a new FieldMap from fMap
// with only the fields on which the given uid user has access.
// All field names are JSONized.
func filterMapOnAuthorizedFields(mi *Model, fMap FieldMap, uid int64, perm security.Permission) FieldMap {
	perm = perm & (security.Read | security.Write | security.Create)
	if perm == 0 {
		// We are trying to check perms that are not read or write which
		// means they don't apply to fields, so we return the whole map
		return fMap
	}
	newFMap := make(FieldMap)
	for field, value := range fMap {
		fi := mi.getRelatedFieldInfo(field)
		if checkFieldPermission(fi, uid, perm) {
			newFMap[fi.json] = value
		}
	}
	return newFMap
}
