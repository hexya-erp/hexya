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

// A cache holds records field values for caching the database to
// improve performance.
type cache map[RecordRef]FieldMap

// addEntry to the cache
func (c *cache) addEntry(modelName string, ID int64, values FieldMap) {
	(*c)[RecordRef{ModelName: modelName, ID: ID}] = values
}

// invalidateEntry removes an entry from the cache
func (c *cache) invalidateEntry(modelName string, ID int64) {
	delete((*c), RecordRef{ModelName: modelName, ID: ID})
}

// get returns the cache value of the given fieldName
// for the given modelName and ID.
func (c *cache) get(modelName string, ID int64, fieldName string) interface{} {
	return (*c)[RecordRef{ModelName: modelName, ID: ID}][fieldName]
}

// getRecord returns the whole record specified by modelName and ID
// as it is currently in cache.
func (c *cache) getRecord(modelName string, ID int64) FieldMap {
	return (*c)[RecordRef{ModelName: modelName, ID: ID}]
}

// checkIfInCache returns true if all fields given by fieldNames are available
// in cache for all the records with the given ids in the given model.
func (c *cache) checkIfInCache(mi *modelInfo, ids []int64, fieldNames []string) bool {
	for _, id := range ids {
		for _, fName := range fieldNames {
			relMI := mi.getRelatedModelInfo(fName)
			ref := RecordRef{ModelName: relMI.name, ID: id}
			if _, ok := (*c)[ref][fName]; !ok {
				return false
			}
		}
	}
	return true
}

// newCache creates a pointer to a new cache instance.
func newCache() *cache {
	res := make(cache)
	return &res
}
