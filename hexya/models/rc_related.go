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
func (rc *RecordCollection) substituteRelatedFields(fields []string) ([]string, *RecordCollection) {
	// Create a keys map with our fields
	keys := make(map[string]bool)
	for _, field := range fields {
		// Inflate our related fields
		inflatedPath := jsonizePath(rc.model, rc.substituteRelatedInPath(field))
		keys[inflatedPath] = true
		// Add intermediate records to our map
		exprs := strings.Split(inflatedPath, ExprSep)
		if len(exprs) == 1 {
			continue
		}
		var curPath string
		for _, expr := range exprs {
			curPath = strings.TrimLeft(fmt.Sprintf("%s.%s", curPath, expr), ".")
			keys[curPath] = true
		}
	}
	// Extract keys from our map to res
	res := make([]string, len(keys))
	var i int
	for key := range keys {
		res[i] = key
		i++
	}

	// Substitute in RecordCollection query
	substs := make(map[string][]string)
	queryExprs := rc.query.cond.getAllExpressions(rc.model)
	for _, exprs := range queryExprs {
		if len(exprs) == 0 {
			continue
		}
		var curPath string
		var resExprs []string
		for _, expr := range exprs {
			resExprs = append(resExprs, expr)
			curPath = strings.Join(resExprs, ExprSep)
			fi := rc.model.getRelatedFieldInfo(curPath)
			curFI := fi
			for curFI.isRelatedField() {
				// We loop because target field may be related itself
				reLen := len(resExprs)
				jsonPath := jsonizePath(curFI.model, curFI.relatedPath)
				resExprs = append(resExprs[:reLen-1], strings.Split(jsonPath, ExprSep)...)
				curFI = rc.model.getRelatedFieldInfo(strings.Join(resExprs, ExprSep))
			}
		}
		substs[strings.Join(exprs, ExprSep)] = resExprs
	}
	rc.query.substituteConditionExprs(substs)

	return res, rc
}

// substituteRelatedInPath recursively substitutes path for its related value.
// If path is not a related field, it is returned as is.
func (rc *RecordCollection) substituteRelatedInPath(path string) string {
	fi := rc.model.getRelatedFieldInfo(path)
	if !fi.isRelatedField() {
		return path
	}
	return rc.substituteRelatedInPath(fi.relatedPath)
}
