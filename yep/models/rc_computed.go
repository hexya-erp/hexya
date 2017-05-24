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

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
func (rc RecordCollection) computeFieldValues(params *FieldMap, fields ...string) {
	for _, fInfo := range rc.model.fields.getComputedFields(fields...) {
		if !checkFieldPermission(fInfo, rc.env.uid, security.Read) {
			// We do not have the access rights on this field, so we skip it.
			continue
		}
		if _, exists := (*params)[fInfo.name]; exists {
			// We already have the value we need in params
			// probably because it was computed with another field
			continue
		}
		newParams := rc.Call(fInfo.compute).(FieldMapper).FieldMap()
		for k, v := range newParams {
			key, _ := rc.model.fields.get(k)
			(*params)[key.json] = v
		}
	}
}

//updateStoredFields updates all dependent fields of rc that are included in the given FieldMap.
func (rc RecordCollection) updateStoredFields(fMap FieldMap) {
	fieldNames := fMap.Keys()
	var toUpdate []computeData
	for _, fieldName := range fieldNames {
		refFieldInfo, ok := rc.model.fields.get(fieldName)
		if !ok {
			continue
		}
		toUpdate = append(toUpdate, refFieldInfo.dependencies...)
	}
	// Compute all that must be computed and store the values
	rSet := rc.Fetch()
	for _, cData := range toUpdate {
		recs := rSet.env.Pool(cData.modelInfo.name)
		if cData.path != "" {
			recs = recs.Search(rSet.Model().Field(cData.path).In(rSet.Ids()))
		} else {
			recs = rSet
		}
		recs = recs.Fetch()
		for _, rec := range recs.Records() {
			retVal := rec.CallMulti(cData.compute)
			vals := retVal[0].(FieldMapper).FieldMap()
			toUnset := retVal[1].([]FieldNamer)
			rec.Call("Write", vals, toUnset)
		}
	}
}
