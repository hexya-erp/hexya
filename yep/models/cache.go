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

// addEntry to the cache. fieldName may be a path relative to this modelInfo.
// (e.g. "User.Profile.Age").
func (c *cache) addEntry(mi *modelInfo, ID int64, fieldName string, value interface{}) {
	relMI := mi.getRelatedFieldInfo(fieldName).mi
	_, ok := (*c)[RecordRef{ModelName: relMI.name, ID: ID}]
	if !ok {
		(*c)[RecordRef{ModelName: relMI.name, ID: ID}] = make(FieldMap)
	}
	jsonFName := jsonizePath(mi, fieldName)
	(*c)[RecordRef{ModelName: relMI.name, ID: ID}][jsonFName] = value
}

// addRecord successively adds each entry of the given FieldMap to the cache.
// fMap keys may be a paths relative to this modelInfo.
// (e.g. "User.Profile.Age").
func (c *cache) addRecord(mi *modelInfo, ID int64, fMap FieldMap) {
	for k, v := range fMap {
		c.addEntry(mi, ID, k, v)
	}
}

// invalidateRecord removes an entire record from the cache
func (c *cache) invalidateRecord(mi *modelInfo, ID int64) {
	delete((*c), RecordRef{ModelName: mi.name, ID: ID})
}

// get returns the cache value of the given fieldName
// for the given modelName and ID. fieldName may be a path
// relative to this modelInfo (e.g. "User.Profile.Age").
func (c *cache) get(mi *modelInfo, ID int64, fieldName string) interface{} {
	jsonFName := jsonizePath(mi, fieldName)
	relMI := mi.getRelatedFieldInfo(jsonFName).mi
	res := (*c)[RecordRef{ModelName: relMI.name, ID: ID}][jsonFName]
	return res
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
			relMI := mi.getRelatedFieldInfo(fName).mi
			ref := RecordRef{ModelName: relMI.name, ID: id}
			jsonFName := jsonizePath(mi, fName)
			if _, ok := (*c)[ref][jsonFName]; !ok {
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
