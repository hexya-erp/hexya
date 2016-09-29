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

// substituteRelatedFields returns a copy of the given fields slice with
// related fields substituted by their related field path. It also returns
// the list of substitutions to be given to resetRelatedFields.
func (rs *RecordCollection) substituteRelatedFields(fields []string) ([]string, []KeySubstitution) {
	// We create a map to check if the substituted field already exists
	duplMap := make(map[string]bool, len(fields))
	for _, field := range fields {
		duplMap[jsonizePath(rs.mi, field)] = true
	}
	// Now we go for the substitution
	res := make([]string, len(fields))
	var substs []KeySubstitution
	for i, field := range fields {
		fi, ok := rs.mi.fields.get(field)
		if ok && fi.related() {
			res[i] = fi.relatedPath
			relatedJSONPath := jsonizePath(rs.mi, fi.relatedPath)
			substs = append(substs, KeySubstitution{
				Orig: relatedJSONPath,
				New:  fi.json,
				Keep: duplMap[relatedJSONPath],
			})
			continue
		}
		res[i] = field
	}
	return res, substs
}
