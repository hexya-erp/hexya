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
	"sort"
	"strings"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
)

// substituteRelatedFields returns a copy of the given fields slice with related fields substituted by their related
// field path. It also adds the fk and pk fields of all records in the related paths.
//
// This method removes duplicates and change all field names to their json names.
func (rc *RecordCollection) substituteRelatedFields(fields []string) []string {
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
			curPath = strings.TrimLeft(curPath+ExprSep+expr, ExprSep)
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
	return res
}

// substituteRelatedFieldsInMap returns a copy of the given FieldMap with related fields
// substituted by their related field path.
func (rc *RecordCollection) substituteRelatedFieldsInMap(fMap FieldMap) FieldMap {
	res := make(FieldMap)
	for field, value := range fMap {
		// Inflate our related fields
		inflatedPath := jsonizePath(rc.model, rc.substituteRelatedInPath(field))
		res[inflatedPath] = value
	}
	return res
}

// substituteRelatedInQuery returns a new RecordCollection with related fields
// substituted in the query.
func (rc *RecordCollection) substituteRelatedInQuery() *RecordCollection {
	// Substitute in RecordCollection query
	substs := make(map[string][]string)
	queryExprs := rc.query.getAllExpressions()
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

	return rc
}

// substituteRelatedInPath recursively substitutes path for its related value.
// If path is not a related field, it is returned as is.
func (rc *RecordCollection) substituteRelatedInPath(path string) string {
	fi := rc.model.getRelatedFieldInfo(path)
	if !fi.isRelatedField() {
		return path
	}
	exprs := strings.Split(path, ExprSep)
	newPath := strings.Join(exprs[:len(exprs)-1], ExprSep) + ExprSep + fi.relatedPath
	newPath = strings.TrimLeft(newPath, ExprSep)
	return rc.substituteRelatedInPath(newPath)
}

// createRelatedRecords creates the records of related fields set in fMap
// if they do not exist.
func (rc *RecordCollection) createRelatedRecords(fMap FieldMap) {
	for _, rec := range rc.Records() {
		rec.applyContexts()
		// 1. We substitute our related fields everywhere
		allFields := rec.substituteRelatedFields(fMap.Keys())

		// 2. We create a new slice with only paths so as to exit early if we have only simple fields
		sort.Strings(allFields)
		var fields []string
		for _, f := range allFields {
			exprs := strings.Split(f, ExprSep)
			if len(exprs) <= 1 {
				// Don't include simple fields
				continue
			}
			fields = append(fields, f)
		}

		// 3. We create a list of paths to records to create, by path length
		// We do not call "Load" directly to have the caller set in the callstack for permissions
		rec.Call("Load", fields)
		var (
			maxLen       int
			toInvalidate bool
		)
		paths := make(map[int]map[string]bool)
		for _, field := range fields {
			if rec.env.cache.isInCache(rec.model, rec.ids[0], field, rec.query.ctxArgsSlug()) {
				// Record exists
				continue
			}
			toInvalidate = true
			exprs := strings.Split(field, ExprSep)
			if paths[len(exprs)] == nil {
				paths[len(exprs)] = make(map[string]bool)
			}
			paths[len(exprs)][strings.Join(exprs[:len(exprs)-1], ExprSep)] = true
			if len(exprs) > maxLen {
				maxLen = len(exprs)
			}
		}
		if toInvalidate {
			// invalidate cache in case we have a contexted field set with a new context
			rec.InvalidateCache()
		}
		// 4. We create our records starting by smallest paths
		for i := 0; i <= maxLen; i++ {
			rec.createRelatedRecordForPaths(paths[i])
		}
	}
}

// createRelatedRecordForPaths creates Records at the given paths, starting from this recordset.
// This method does not check whether such a records already exists or not.
func (rc *RecordCollection) createRelatedRecordForPaths(paths map[string]bool) {
	rc.EnsureOne()
	rsPaths := map[string]*RecordCollection{"": rc}
	for path := range paths {
		fi := rc.model.getRelatedFieldInfo(path)
		switch fi.fieldType {
		case fieldtype.Many2One, fieldtype.One2One, fieldtype.Many2Many:
			// We do not call "create" directly to have the caller set in the callstack for permissions
			res := rc.env.Pool(fi.relatedModel.name).Call("Create", FieldMap{})
			if resRS, ok := res.(RecordSet); ok {
				rc.env.Pool(fi.model.name).Call("Set", fi.name, resRS.Collection())
				rsPaths[path] = resRS.Collection()
			}
		case fieldtype.One2Many, fieldtype.Rev2One:
			exprs := strings.Split(path, ExprSep)
			// We do not call "create" directly to have the caller set in the callstack for permissions
			res := rc.env.Pool(fi.relatedModel.name).Call("Create", FieldMap{
				fi.jsonReverseFK: rsPaths[strings.Join(exprs[:len(exprs)-1], ExprSep)]})
			if resRS, ok := res.(RecordSet); ok {
				rsPaths[path] = resRS.Collection()
			}
		}
	}
}
