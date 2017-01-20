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
	"errors"
	"strings"
)

// A cache holds records field values for caching the database to
// improve performance.
type cache map[RecordRef]FieldMap

// addEntry to the cache. fieldName must be a simple field name (no path)
func (c *cache) addEntry(mi *Model, ID int64, fieldName string, value interface{}) {
	ref := RecordRef{ModelName: mi.name, ID: ID}
	jsonName := jsonizePath(mi, fieldName)
	c.addEntryByRef(ref, jsonName, value)
}

// addEntryByRef adds an entry to the cache from a RecordRef and a field json name
func (c *cache) addEntryByRef(ref RecordRef, jsonName string, value interface{}) {
	if _, ok := (*c)[ref]; !ok {
		(*c)[ref] = make(FieldMap)
	}
	(*c)[ref][jsonName] = value
}

// addRecord successively adds each entry of the given FieldMap to the cache.
// fMap keys may be a paths relative to this Model (e.g. "User.Profile.Age").
func (c *cache) addRecord(mi *Model, ID int64, fMap FieldMap) {
	paths := make(map[int][]string)
	var maxLen int
	// We create our exprsMap with the length of the path as key
	for _, path := range fMap.Keys() {
		exprs := strings.Split(path, ExprSep)
		paths[len(exprs)] = append(paths[len(exprs)], path)
		if len(exprs) > maxLen {
			maxLen = len(exprs)
		}
	}
	// We add entries into the cache, starting from the smallest paths
	for i := 0; i <= maxLen; i++ {
		for _, path := range paths[i] {
			ref, fName, _ := c.getRelatedRef(mi, ID, path)
			c.addEntryByRef(ref, fName, fMap[path])
		}
	}
}

// invalidateRecord removes an entire record from the cache
func (c *cache) invalidateRecord(mi *Model, ID int64) {
	delete((*c), RecordRef{ModelName: mi.name, ID: ID})
}

// get returns the cache value of the given fieldName
// for the given modelName and ID. fieldName may be a path
// relative to this Model (e.g. "User.Profile.Age").
func (c *cache) get(mi *Model, ID int64, fieldName string) interface{} {
	ref, fName, _ := c.getRelatedRef(mi, ID, fieldName)
	res := (*c)[ref][fName]
	return res
}

// getRecord returns the whole record specified by modelName and ID
// as it is currently in cache.
func (c *cache) getRecord(modelName string, ID int64) FieldMap {
	return (*c)[RecordRef{ModelName: modelName, ID: ID}]
}

// checkIfInCache returns true if all fields given by fieldNames are available
// in cache for all the records with the given ids in the given model.
func (c *cache) checkIfInCache(mi *Model, ids []int64, fieldNames []string) bool {
	for _, id := range ids {
		for _, fName := range fieldNames {
			ref, path, err := c.getRelatedRef(mi, id, fName)
			if err != nil {
				return false
			}
			if _, ok := (*c)[ref][path]; !ok {
				return false
			}
		}
	}
	return true
}

// getRelatedRef returns the RecordRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
func (c *cache) getRelatedRef(mi *Model, ID int64, path string) (RecordRef, string, error) {
	exprs := jsonizeExpr(mi, strings.Split(path, ExprSep))
	if len(exprs) > 1 {
		relMI := mi.getRelatedModelInfo(exprs[0])
		fk, ok := c.get(mi, ID, exprs[0]).(int64)
		if !ok {
			return RecordRef{}, "", errors.New("Requested value not in cache")
		}
		return c.getRelatedRef(relMI, fk, strings.Join(exprs[1:], "."))
	}
	return RecordRef{ModelName: mi.name, ID: ID}, exprs[0], nil
}

// newCache creates a pointer to a new cache instance.
func newCache() *cache {
	res := make(cache)
	return &res
}
