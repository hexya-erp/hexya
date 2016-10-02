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

import "fmt"

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
func (rs RecordCollection) computeFieldValues(params *FieldMap, fields ...string) {
	for _, fInfo := range rs.mi.fields.getComputedFields(fields...) {
		if _, exists := (*params)[fInfo.name]; exists {
			// We already have the value we need in params
			// probably because it was computed with another field
			continue
		}
		newParams := rs.Call(fInfo.compute).(FieldMap)
		for k, v := range newParams {
			key, _ := rs.mi.fields.get(k)
			(*params)[key.json] = v
		}
	}
}

/*
updateStoredFields updates all dependent fields of rs that are included in the given FieldMap.
*/
func (rs RecordCollection) updateStoredFields(fMap FieldMap) {
	fieldNames := fMap.Keys()
	var toUpdate []computeData
	for _, fieldName := range fieldNames {
		refFieldInfo, ok := rs.mi.fields.get(fieldName)
		if !ok {
			continue
		}
		toUpdate = append(toUpdate, refFieldInfo.dependencies...)
	}
	// Compute all that must be computed and store the values
	computed := make(map[string]bool)
	rSet := rs.LazyLoad()
	for _, cData := range toUpdate {
		methUID := fmt.Sprintf("%s.%s", cData.modelInfo.tableName, cData.compute)
		if _, ok := computed[methUID]; ok {
			continue
		}
		recs := rSet.env.Pool(cData.modelInfo.name)
		if cData.path != "" {
			recs = recs.Filter(cData.path, "in", rSet.Ids())
		} else {
			recs = rSet
		}
		recs.LazyLoad()
		for _, rec := range recs.Records() {
			vals := rec.Call(cData.compute)
			if len(vals.(FieldMap)) > 0 {
				rec.Call("Write", vals.(FieldMap))
			}
		}
	}
}
