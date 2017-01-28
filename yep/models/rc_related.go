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
	"fmt"
	"strings"
)

// substituteRelatedFields returns :
// - a copy of the given fields slice with related fields substituted by their related
// field path. It also adds the fk and pk fields of all records in the related paths.
// - a new RecordCollection with substitution of related fields in the query.
//
// This function removes duplicates and change all field names to their json names.
func (rc RecordCollection) substituteRelatedFields(fields []string) ([]string, RecordCollection) {
	// Create a keys map with our fields
	keys := make(map[string]bool)
	for _, field := range fields {
		fi := rc.model.getRelatedFieldInfo(field)
		if fi.isRelatedField() {
			keys[fi.relatedPath] = true
			var curPath string
			for _, expr := range strings.Split(fi.relatedPath, ExprSep) {
				curPath = strings.TrimLeft(fmt.Sprintf("%s.%s", curPath, expr), ".")
				keys[curPath] = true
			}
			continue
		}
		keys[jsonizePath(rc.model, field)] = true
	}
	// extract keys from our map to res
	res := make([]string, len(keys))
	var i int
	for key := range keys {
		res[i] = key
		i++
	}

	// Substitute in RecordCollection query
	substs := make(map[string][]string)
	for _, fi := range rc.model.fields.relatedFields {
		substs[fi.json] = strings.Split(fi.relatedPath, ExprSep)
	}
	rc.query.substituteConditionExprs(substs)

	return res, rc
}
