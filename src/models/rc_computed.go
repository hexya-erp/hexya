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
	"sort"

	"github.com/hexya-erp/hexya/src/tools/typesutils"
)

// A recomputePair gives a method to apply on a record collection.
type recomputePair struct {
	recs   *RecordCollection
	method string
}

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
func (rc *RecordCollection) computeFieldValues(params *FieldMap, fields ...string) {
	rc.EnsureOne()
	for _, fInfo := range rc.model.fields.getComputedFields(fields...) {
		if _, exists := (*params)[fInfo.name]; exists {
			// We already have the value we need in params
			// probably because it was computed with another field
			continue
		}
		newParams := rc.Call(fInfo.compute).(RecordData).Underlying().FieldMap
		(*params).MergeWith(newParams, rc.model)
	}
}

// processTriggers execute computed fields recomputation (for stored fields) or
// invalidation (for non stored fields) based on the data of each fields 'Depends'
// attribute.
func (rc *RecordCollection) processTriggers(keys []FieldName) {
	if rc.Env().Context().GetBool("hexya_no_recompute_stored_fields") {
		return
	}
	rc.updateStoredFields(rc.retrieveComputeData(keys))
}

// retrieveComputeData looks up fields that need to be recomputed when the given fields are modified.
//
// Returned value is an ordered slice of methods to apply on records
func (rc *RecordCollection) retrieveComputeData(fields []FieldName) []recomputePair {
	// Find record fields to update from the modified fields
	var (
		toUpdateKeys []string
		res          []recomputePair
	)
	toUpdateData := make(map[string]computeData)
	for _, fieldName := range fields {
		refFieldInfo, ok := rc.model.fields.Get(fieldName.Name())
		if !ok {
			continue
		}
		for _, dep := range refFieldInfo.dependencies {
			key := fmt.Sprintf("%s-%s-%s-%t", dep.model.name, dep.path, dep.compute, dep.stored)
			if _, exists := toUpdateData[key]; !exists {
				toUpdateKeys = append(toUpdateKeys, key)
				toUpdateData[key] = dep
			}
		}
	}
	// Order the computeData keys to have deterministic recomputation
	sort.Strings(toUpdateKeys)
	// Compute all that must be computed and store the values
	for _, key := range toUpdateKeys {
		cData := toUpdateData[key]
		recs := rc
		if cData.path != "" {
			cPath := cData.model.FieldName(cData.path)
			recs = rc.Env().Pool(cData.model.name).Search(rc.Model().Field(cPath).In(rc.Ids()))
		}
		if !cData.stored {
			// Field is not stored, just invalidating cache
			for _, id := range recs.Ids() {
				rc.env.cache.removeEntry(recs.model, id, cData.fieldName, rc.query.ctxArgsSlug())
			}
			continue
		}
		recs.Fetch()
		res = append(res, recomputePair{recs: recs, method: cData.compute})
	}
	return res
}

// updateStoredFields applies each method on each record defined by compPairs
func (rc *RecordCollection) updateStoredFields(compPairs []recomputePair) {
	for _, rp := range compPairs {
		if rp.recs.IsEmpty() {
			// recs have been fetched in retrieveComputeData
			// if it is empty now, it must be because the records have been unlinked in between
			continue
		}
		rp.recs.applyMethod(rp.method)
	}
}

// applyMethod calls the method on this recordset.
func (rc *RecordCollection) applyMethod(methodName string) {
	for _, rec := range rc.Records() {
		retVal := rec.Call(methodName)
		data := retVal.(RecordData).Underlying()
		// Check if the values actually changed
		var doUpdate bool
		for f, v := range data.FieldMap {
			if f == "write_date" {
				continue
			}
			if rs, isRS := rec.Get(rec.model.FieldName(f)).(RecordSet); isRS {
				if !rs.Collection().Equals(v.(RecordSet).Collection()) {
					doUpdate = true
					break
				}
				continue
			}
			if rec.Get(rec.model.FieldName(f)) != v {
				doUpdate = true
				break
			}
		}
		if doUpdate {
			rec.WithContext("hexya_force_compute_write", true).Call("Write", data)
		}
	}
}

// processInverseMethods executes inverse methods of fields in the given
// FieldMap if it exists. It returns a new FieldMap to be used by Create/Write
// instead of the original one.
func (rc *RecordCollection) processInverseMethods(data RecordData) {
	md := NewModelDataFromRS(rc, data.Underlying().FieldMap)
	for _, fieldName := range md.Underlying().Keys() {
		fName := rc.model.FieldName(fieldName)
		fi := rc.model.getRelatedFieldInfo(fName)
		if !fi.isComputedField() || rc.Env().Context().GetBool("hexya_force_compute_write") {
			continue
		}
		if !md.Has(fName) {
			continue
		}
		val := md.Get(fName)
		if fi.inverse == "" {
			if typesutils.IsZero(val) || rc.Env().Context().GetBool("hexya_ignore_computed_fields") {
				continue
			}
			log.Panic("Trying to write a computed field without inverse method", "model", rc.model.name, "field", fieldName)
		}
		rc.Call(fi.inverse, val)
	}
}
