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
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/tools/logging"
)

// AllowModelAccess grants the given permission to the given group for the given model.
// This also sets the permissions for all fields of the model.
func AllowModelAccess(model ModelName, group *security.Group, perm security.Permission) {
	mi := modelRegistry.mustGet(string(model))
	// Add perm to the model's acl
	mi.acl.AddPermission(group, perm)
	// Add perm to all fields of the model
	for fName := range mi.fields.registryByName {
		AllowFieldAccess(model, FieldName(fName), group, perm)
	}
}

// AllowFieldAccess grants the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func AllowFieldAccess(model ModelName, field FieldName, group *security.Group, perm security.Permission) {
	perm = perm & (security.Read | security.Write | security.Create)
	mi := modelRegistry.mustGet(string(model))
	if !mi.acl.CheckPermission(group, security.Read) {
		log.Warn("Trying to add permission on field, but model is not readable", "model", model, "field", field, "perm", perm)
	}
	fi := mi.fields.mustGet(string(field))
	fi.acl.AddPermission(group, perm)
}

// DenyModelAccess denies the given permission to the given group for the given model.
// This also unsets the permissions for all fields of the model.
func DenyModelAccess(model ModelName, group *security.Group, perm security.Permission) {
	mi := modelRegistry.mustGet(string(model))
	// Remove perm to the model's acl
	mi.acl.RemovePermission(group, perm)
	// Remove perm to all fields of the model
	for fName := range mi.fields.registryByName {
		DenyFieldAccess(model, FieldName(fName), group, perm)
	}
}

// RemoveFieldAccess denies the given perm to the given group on the given field of model.
// Only security.Read and security.Write permissions are taken into account by
// this function, others are discarded.
func DenyFieldAccess(model ModelName, field FieldName, group *security.Group, perm security.Permission) {
	perm = perm & (security.Read | security.Write | security.Create)
	mi := modelRegistry.mustGet(string(model))
	fi := mi.fields.mustGet(string(field))
	fi.acl.RemovePermission(group, perm)
}

// checkModelPermission checks if the given uid has the given perm on the given model info.
// It panics if the user does not have the required permission and does nothing otherwise.
func mustCheckModelPermission(mi *modelInfo, uid int64, perm security.Permission) {
	userGroups := security.AuthenticationRegistry.UserGroups(uid)
	for _, group := range userGroups {
		if mi.acl.CheckPermission(group, perm) {
			return
		}
	}
	logging.LogAndPanic(log, "User does not have required permission on model", "uid", uid, "model", mi.name, "permission", perm)
}

// checkFieldPermission checks if the given uid has the given perm on the given field info.
func checkFieldPermission(fi *fieldInfo, uid int64, perm security.Permission) bool {
	userGroups := security.AuthenticationRegistry.UserGroups(uid)
	for _, group := range userGroups {
		if fi.acl.CheckPermission(group, perm) {
			return true
		}
	}
	return false
}

// filterOnAuthorizedFields returns the fields slice with only the fields on
// which the current user has the given permission.
func filterOnAuthorizedFields(mi *modelInfo, uid int64, fields []string, perm security.Permission) []string {
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
func filterMapOnAuthorizedFields(mi *modelInfo, fMap FieldMap, uid int64, perm security.Permission) FieldMap {
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
			newFMap[field] = value
		}
	}
	return newFMap
}
