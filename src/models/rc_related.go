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

	"github.com/hexya-erp/hexya/src/models/fieldtype"
	"github.com/hexya-erp/hexya/src/tools/strutils"
)

// substituteRelatedFields returns a copy of the given fields slice with related fields substituted by their related
// field path. It also adds the fk and pk fields of all records in the related paths.
//
// The second returned value is a map the keys of which are the related field paths, and the values are the
// corresponding original fields if they exist.
//
// This method removes duplicates and change all field names to their json names.
func (rc *RecordCollection) substituteRelatedFields(fields []string) ([]string, map[string]string) {
	res := make(map[string]string)
	for i, field := range fields {
		relPath := jsonizePath(rc.model, rc.substituteRelatedInPath(field))
		res[relPath] = field
		fields[i] = relPath
	}
	fields = rc.addIntermediatePaths(fields)
	return fields, res
}

// addIntermediatePaths adds the paths that compose fields and returns a new slice.
//
// e.g. given [User.Address.Country Note Partner.Age] will return
// [User User.Address User.address.country Note Partner Partner.Age]
//
// This method removes duplicates
func (rc *RecordCollection) addIntermediatePaths(fields []string) []string {
	// Create a keys map with our fields to avoid duplicates
	keys := make(map[string]bool)
	// Add intermediate records to our map
	for _, field := range fields {
		jsonField := jsonizePath(rc.model, field)
		keys[jsonField] = true
		exprs := strings.Split(jsonField, ExprSep)
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
//
// This method substitute the first level only (to work with data structs)
func (rc *RecordCollection) substituteRelatedFieldsInMap(fMap FieldMap) FieldMap {
	res := make(FieldMap)
	for field, value := range fMap {
		// Inflate our related fields
		fi := rc.model.getRelatedFieldInfo(field)
		path := field
		if fi.relatedPath != "" {
			path = fi.relatedPath
		}
		inflatedPath := jsonizePath(rc.model, path)
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

// createRelatedRecord creates Records at the given path, starting from this recordset.
// This method does not check whether such a records already exists or not.
func (rc *RecordCollection) createRelatedRecord(path string, vals RecordData) *RecordCollection {
	log.Debug("Creating related record", "recordset", rc, "path", path, "vals", vals)
	rc.EnsureOne()
	fi := rc.model.getRelatedFieldInfo(path)
	exprs := strings.Split(path, ExprSep)
	switch fi.fieldType {
	case fieldtype.Many2One, fieldtype.One2One, fieldtype.Many2Many:
		resRS := rc.createRelatedFKRecord(fi, vals)
		rc.Set(path, resRS.Collection())
		return resRS.Collection()
	case fieldtype.One2Many, fieldtype.Rev2One:
		target := rc
		if len(exprs) > 1 {
			target = rc.Get(strings.Join(exprs[:len(exprs)-1], ExprSep)).(RecordSet).Collection()
			if target.IsEmpty() {
				log.Panic("Target record does not exist", "recordset", rc, "path", strings.Join(exprs[:len(exprs)-1], ExprSep))
			}
			target = target.Records()[0]
		}
		vals.Underlying().Set(fi.jsonReverseFK, target)
		return rc.env.Pool(fi.relatedModel.name).Call("Create", vals).(RecordSet).Collection()
	}
	return rc.env.Pool(rc.ModelName())
}

// createRelatedFKRecord creates a single related record for the given FK field
func (rc *RecordCollection) createRelatedFKRecord(fi *Field, data RecordData) *RecordCollection {
	rSet := rc.env.Pool(fi.relatedModel.name)
	if fi.embed {
		rSet = rSet.WithContext("default_hexya_external_id", fmt.Sprintf("%s_%s", rc.Get("HexyaExternalID"), strutils.SnakeCase(fi.relatedModel.name)))
	}
	res := rSet.Call("Create", data)
	return res.(RecordSet).Collection()
}
