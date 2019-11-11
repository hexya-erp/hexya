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

	"github.com/hexya-erp/hexya/src/models/fieldtype"
)

// A cache holds records field values for caching the database to
// improve performance. cache is not safe for concurrent access.
type cache struct {
	sync.RWMutex
	data       map[string]map[int64]FieldMap                    // cache data values by model and id
	x2mRelated map[string]map[int64]map[string]map[string]int64 // o2m and r2m relations by model, id, field, context
	m2mLinks   map[string]map[[2]int64]bool                     // many2many relations by relation model and ids
}

// notInCacheError is returned when a request in cache returns no entry
type notInCacheError struct{}

// Error method fo the notInCacheError
func (nice notInCacheError) Error() string {
	return "requested value not in cache"
}

// nonExistentPathError is returned when a request path in cache leads
// nowhere (i.e. one FK of the path is 0).
type nonExistentPathError struct{}

func (nepe nonExistentPathError) Error() string {
	return "requested path is broken"
}

// updateEntry creates or updates an entry in the cache defined by its model, id and fieldName.
// fieldName can be a path
func (c *cache) updateEntry(mi *Model, id int64, fieldName string, value interface{}, ctxSlug string) error {
	if id == 0 {
		return errors.New("skipped entry with id = 0")
	}
	mi, id, fName, err := c.getRelatedRef(mi, id, fieldName, ctxSlug)
	if err != nil {
		return err
	}
	c.updateEntryByRef(mi, id, fName, value, ctxSlug)
	return nil
}

// updateEntryByRef creates or updates an entry to the cache from a cacheRef
// and a field json name (no path).
func (c *cache) updateEntryByRef(mi *Model, id int64, jsonName string, value interface{}, ctxSlug string) {
	fi := mi.fields.MustGet(jsonName)
	switch fi.fieldType {
	case fieldtype.One2Many:
		ids := value.([]int64)
		for _, relID := range ids {
			c.updateEntry(fi.relatedModel, relID, fi.jsonReverseFK, id, ctxSlug)
		}
		if len(ids) == 1 {
			// We have only one ID.
			// We arbitrarily decide it is a related field through O2M and we do not have the complete O2M set.
			c.setX2MValue(mi.name, id, jsonName, ids[0], ctxSlug)
			break
		}
		c.setDataValue(mi.name, id, jsonName, true)

	case fieldtype.Rev2One:
		relID := value.(int64)
		c.updateEntry(fi.relatedModel, relID, fi.jsonReverseFK, id, ctxSlug)
		c.setDataValue(mi.name, id, jsonName, true)

	case fieldtype.Many2Many:
		ids := value.([]int64)
		if len(ids) == 1 {
			// We have only one ID.
			// We arbitrarily decide it is a related field through M2M and we do not have the complete M2M set
			c.addM2MLink(fi, id, ids)
			c.setX2MValue(mi.name, id, jsonName, ids[0], ctxSlug)
			break
		}
		c.removeM2MLinks(fi, id)
		c.addM2MLink(fi, id, ids)
		c.setDataValue(mi.name, id, jsonName, true)
	default:
		c.setDataValue(mi.name, id, jsonName, value)
	}
}

// setDataValue sets the value for the jsonName field of record ref to value
func (c *cache) setDataValue(model string, id int64, jsonName string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.data[model]; !ok {
		c.data[model] = make(map[int64]FieldMap)
	}
	if _, ok := c.data[model][id]; !ok {
		c.data[model][id] = make(FieldMap)
		c.data[model][id]["id"] = id
	}
	c.data[model][id][jsonName] = value
}

// setX2MValue sets the id for the jsonName field of record ref in the x2mRelation map
func (c *cache) setX2MValue(model string, id int64, jsonName string, relID int64, ctxSlug string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.x2mRelated[model]; !ok {
		c.x2mRelated[model] = make(map[int64]map[string]map[string]int64)
	}
	if _, ok := c.x2mRelated[model][id]; !ok {
		c.x2mRelated[model][id] = make(map[string]map[string]int64)
	}
	if _, ok := c.x2mRelated[model][id][jsonName]; !ok {
		c.x2mRelated[model][id][jsonName] = make(map[string]int64)
	}
	if relID == c.x2mRelated[model][id][jsonName][""] {
		// We don't add the value if the id is the same as the default context
		return
	}
	if ctxSlug == "" {
		// We are setting the default value, so we remove any other slug with the same target id
		for k, v := range c.x2mRelated[model][id][jsonName] {
			if k != "" && v == relID {
				delete(c.x2mRelated[model][id][jsonName], k)
			}
		}
	}
	c.x2mRelated[model][id][jsonName][ctxSlug] = relID
}

// getX2MValue return the X2MValue or the default value if the given ctxSlug does not exist in cache
//
// 2nd returned value is true if a suitable value has been found
//
// 3rd returned value is true if the returned value is the value for the default context
func (c *cache) getX2MValue(model string, id int64, jsonName string, ctxSlug string) (int64, bool, bool) {
	var defaultVal bool
	res, ok := c.x2mRelated[model][id][jsonName][ctxSlug]
	if !ok {
		res, ok = c.x2mRelated[model][id][jsonName][""]
		defaultVal = true
	}
	return res, ok, defaultVal
}

// deleteFieldData removes the cache entry for the jsonName field of record ref
func (c *cache) deleteFieldData(model string, id int64, jsonName string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data[model][id], jsonName)
	if _, exists := c.x2mRelated[model][id]; exists {
		delete(c.x2mRelated[model][id], jsonName)
	}
}

// deleteData removes the cache entry for the whole record ref
func (c *cache) deleteData(model string, id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.data[model], id)
	delete(c.x2mRelated[model], id)
}

// removeM2MLinks removes all M2M links associated with the record with
// the given id on the given field
func (c *cache) removeM2MLinks(fi *Field, id int64) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.m2mLinks[fi.m2mRelModel.name]; !exists {
		return
	}
	index := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	for link := range c.m2mLinks[fi.m2mRelModel.name] {
		if link[index] == id {
			delete(c.m2mLinks[fi.m2mRelModel.name], link)
		}
	}
}

// addM2MLink adds an M2M link between this record with its given ID
// and the records given by values on the given field.
func (c *cache) addM2MLink(fi *Field, id int64, values []int64) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.m2mLinks[fi.m2mRelModel.name]; !exists {
		c.m2mLinks[fi.m2mRelModel.name] = make(map[[2]int64]bool)
	}
	ourIndex := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	theirIndex := (ourIndex + 1) % 2
	for _, val := range values {
		var newLink [2]int64
		newLink[ourIndex] = id
		newLink[theirIndex] = val
		c.m2mLinks[fi.m2mRelModel.name][newLink] = true
	}
}

// getM2MLinks returns the linked ids to this id through the given field.
func (c *cache) getM2MLinks(fi *Field, id int64) []int64 {
	if _, exists := c.m2mLinks[fi.m2mRelModel.name]; !exists {
		return []int64{}
	}
	var res []int64
	ourIndex := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	theirIndex := (ourIndex + 1) % 2
	for link := range c.m2mLinks[fi.m2mRelModel.name] {
		if link[ourIndex] == id {
			res = append(res, link[theirIndex])
		}
	}
	return res
}

// addRecord successively adds each entry of the given FieldMap to the cache.
// fMap keys may be a paths relative to this Model (e.g. "User.Profile.Age").
func (c *cache) addRecord(mi *Model, id int64, fMap FieldMap, ctxSlug string) {
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
			c.updateEntry(mi, id, path, fMap[path], ctxSlug)
		}
	}
}

// invalidateRecord removes an entire record from the cache
//
// WARNING: Reload the record as soon as possible after calling
// this method, since this will bring discrepancies in the other
// records references (One2Many and Many2Many fields).
func (c *cache) invalidateRecord(mi *Model, id int64) {
	c.deleteData(mi.name, id)
	for _, fi := range mi.fields.registryByJSON {
		if fi.fieldType == fieldtype.Many2Many {
			c.removeM2MLinks(fi, id)
		}
	}
}

// removeEntry removes the given entry from cache
func (c *cache) removeEntry(mi *Model, id int64, fieldName, ctxSlug string) {
	if !c.checkIfInCache(mi, []int64{id}, []string{fieldName}, ctxSlug, true) {
		return
	}
	c.deleteFieldData(mi.name, id, fieldName)
	fi := mi.fields.MustGet(fieldName)
	if fi.fieldType == fieldtype.Many2Many {
		c.removeM2MLinks(fi, id)
	}
}

// get returns the cache value of the given fieldName
// for the given modelName and id. fieldName may be a path
// relative to this Model (e.g. "User.Profile.Age").
//
// If the requested value cannot be found, get returns nil
func (c *cache) get(mi *Model, id int64, fieldName string, ctxSlug string) interface{} {
	mi, id, fName, err := c.getRelatedRef(mi, id, fieldName, ctxSlug)
	if err != nil {
		return nil
	}
	fi := mi.fields.MustGet(fName)
	switch fi.fieldType {
	case fieldtype.One2Many:
		if _, ok := c.data[fi.relatedModelName]; !ok {
			return nil
		}
		var relIds []int64
		for cID, cVal := range c.data[fi.relatedModelName] {
			if cVal[fi.jsonReverseFK] != id {
				continue
			}
			relIds = append(relIds, cID)
		}
		return relIds
	case fieldtype.Rev2One:
		if _, ok := c.data[fi.relatedModelName]; !ok {
			return nil
		}
		for cID, cVal := range c.data[fi.relatedModelName] {
			if cVal[fi.jsonReverseFK] != id {
				continue
			}
			return cID
		}
		return nil
	case fieldtype.Many2Many:
		return c.getM2MLinks(fi, id)
	default:
		return c.data[mi.name][id][fName]
	}
}

// checkIfInCache returns true if all fields given by fieldNames are available
// in cache for all the records with the given ids in the given model.
func (c *cache) checkIfInCache(mi *Model, ids []int64, fieldNames []string, ctxSlug string, strict bool) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		for _, fName := range fieldNames {
			if !c.isInCache(mi, id, fName, ctxSlug, strict) {
				return false
			}
		}
	}
	return true
}

// isInCache returns true if the related record through path and ctxSlug strictly exists
// (i.e. no default value for context)
func (c *cache) isInCache(mi *Model, id int64, path string, ctxSlug string, strict bool) bool {
	mi, id, path, err := c.getRelatedRefCommon(mi, id, path, ctxSlug, strict)
	if err != nil {
		switch err.(type) {
		case nonExistentPathError:
			return true
		case notInCacheError:
			return false
		}
		return false
	}
	if _, ok := c.data[mi.name][id][path]; !ok {
		return false
	}
	return true
}

// getRelatedRef returns the cacheRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
func (c *cache) getRelatedRef(mi *Model, id int64, path string, ctxSlug string) (*Model, int64, string, error) {
	return c.getRelatedRefCommon(mi, id, path, ctxSlug, false)
}

// getStrictRelatedRef returns the cacheRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
//
// This method returns an error when the value for the given ctxSlug cannot be found.
func (c *cache) getStrictRelatedRef(mi *Model, id int64, path string, ctxSlug string) (*Model, int64, string, error) {
	return c.getRelatedRefCommon(mi, id, path, ctxSlug, true)
}

// getRelatedRefCommon is the common implementation of getRelatedRef and getStrictRelatedRef.
func (c *cache) getRelatedRefCommon(mi *Model, id int64, path string, ctxSlug string, strict bool) (*Model, int64, string, error) {
	if id == 0 {
		return nil, 0, "", errors.New("requested value on RecordSet with ID=0")
	}
	exprs := jsonizeExpr(mi, strings.Split(path, ExprSep))
	if len(exprs) > 1 {
		fkID, ok := c.get(mi, id, exprs[0], ctxSlug).(int64)
		if !ok {
			var defVal bool
			fkID, ok, defVal = c.getX2MValue(mi.name, id, exprs[0], ctxSlug)
			if !ok || (strict && defVal) {
				return nil, 0, "", notInCacheError{}
			}
		}
		relMI := mi.getRelatedModelInfo(mi.FieldName(exprs[0]))
		if fkID == 0 {
			return nil, 0, "", nonExistentPathError{}
		}
		return c.getRelatedRefCommon(relMI, fkID, strings.Join(exprs[1:], ExprSep), ctxSlug, strict)
	}
	return mi, id, exprs[0], nil
}

// newCache creates a pointer to a new cache instance.
func newCache() *cache {
	res := cache{
		data:       make(map[string]map[int64]FieldMap),
		x2mRelated: make(map[string]map[int64]map[string]map[string]int64),
		m2mLinks:   make(map[string]map[[2]int64]bool),
	}
	return &res
}
