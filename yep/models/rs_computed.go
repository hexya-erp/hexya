/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import "fmt"

// computeFieldValues updates the given params with the given computed (non stored) fields
// or all the computed fields of the model if not given.
// Returned fieldMap keys are field's JSON name
func (rs RecordSet) computeFieldValues(params *FieldMap, fields ...string) {
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
func (rs RecordSet) updateStoredFields(fMap FieldMap) {
	// First get list of fields that have been passed through structPtrOrParams
	fieldNames := fMap.Keys()
	var toUpdate []computeData
	for _, fieldName := range fieldNames {
		//refField := fieldRef{modelName: rs.ModelName(), name: fieldName}
		refFieldInfo, ok := rs.mi.fields.get(fieldName)
		if !ok {
			continue
		}
		toUpdate = append(toUpdate, refFieldInfo.dependencies...)
	}
	// Compute all that must be computed and store the values
	computed := make(map[string]bool)
	rs = *rs.Search()
	for _, cData := range toUpdate {
		methUID := fmt.Sprintf("%s.%s", cData.modelInfo.tableName, cData.compute)
		if _, ok := computed[methUID]; ok {
			continue
		}
		recs := rs.env.Pool(cData.modelInfo.name)
		if cData.path != "" {
			recs = recs.Filter(cData.path, "in", rs.Ids())
		} else {
			recs = &rs
		}
		recs.Search()
		for _, rec := range recs.Records() {
			vals := rec.Call(cData.compute)
			if len(vals.(FieldMap)) > 0 {
				rec.Write(vals.(FieldMap))
			}
		}
	}
}
