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
	"sync"
)

// A cacheRef is a key to find a record in a cache
type cacheRef struct {
	model *Model
	id    int64
}

// A cache holds records field values for caching the database to
// improve performance.
type cache struct {
	sync.RWMutex
	data map[cacheRef]FieldMap
}

// updateEntry creates or updates an entry in the cache defined by its model, id and fieldName.
// fieldName must be a simple field name (no path)
func (c *cache) updateEntry(mi *Model, id int64, fieldName string, value interface{}) {
	ref := cacheRef{model: mi, id: id}
	jsonName := jsonizePath(mi, fieldName)
	c.updateEntryByRef(ref, jsonName, value)
}

// updateEntryByRef creates or updates an entry to the cache from a RecordRef and a field json name
func (c *cache) updateEntryByRef(ref cacheRef, jsonName string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.data[ref]; !ok {
		c.data[ref] = make(FieldMap)
	}
	c.data[ref][jsonName] = value
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
			c.updateEntryByRef(ref, fName, fMap[path])
		}
	}
}

// invalidateRecord removes an entire record from the cache
func (c *cache) invalidateRecord(mi *Model, id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, cacheRef{model: mi, id: id})
}

// get returns the cache value of the given fieldName
// for the given modelName and id. fieldName may be a path
// relative to this Model (e.g. "User.Profile.Age").
func (c *cache) get(mi *Model, id int64, fieldName string) interface{} {
	ref, fName, _ := c.getRelatedRef(mi, id, fieldName)
	return c.data[ref][fName]
}

// getRecord returns the whole record specified by modelName and id
// as it is currently in cache.
func (c *cache) getRecord(model *Model, id int64) FieldMap {
	ref := cacheRef{model: model, id: id}
	return c.data[ref].Copy()
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
			if _, ok := c.data[ref][path]; !ok {
				return false
			}
		}
	}
	return true
}

// getRelatedRef returns the RecordRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
func (c *cache) getRelatedRef(mi *Model, id int64, path string) (cacheRef, string, error) {
	exprs := jsonizeExpr(mi, strings.Split(path, ExprSep))
	if len(exprs) > 1 {
		relMI := mi.getRelatedModelInfo(exprs[0])
		fk, ok := c.get(mi, id, exprs[0]).(int64)
		if !ok {
			return cacheRef{}, "", errors.New("requested value not in cache")
		}
		return c.getRelatedRef(relMI, fk, strings.Join(exprs[1:], "."))
	}
	return cacheRef{model: mi, id: id}, exprs[0], nil
}

// newCache creates a pointer to a new cache instance.
func newCache() *cache {
	res := cache{
		data: make(map[cacheRef]FieldMap),
	}
	return &res
}
