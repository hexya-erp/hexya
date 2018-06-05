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

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
)

// A cacheRef is a key to find a record in a cache
type cacheRef struct {
	model *Model
	id    int64
}

// A cache holds records field values for caching the database to
// improve performance. cache is not safe for concurrent access.
type cache struct {
	sync.RWMutex
	data       map[cacheRef]FieldMap
	x2mRelated map[cacheRef]map[string]map[string]int64
	m2mLinks   map[*Model]map[[2]int64]bool
}

// updateEntry creates or updates an entry in the cache defined by its model, id and fieldName.
// fieldName can be a path
func (c *cache) updateEntry(mi *Model, id int64, fieldName string, value interface{}, ctxSlug string) error {
	if id == 0 {
		return errors.New("skipped entry with id = 0")
	}
	ref, fName, err := c.getRelatedRef(mi, id, fieldName, ctxSlug)
	if err != nil {
		return err
	}
	c.updateEntryByRef(ref, fName, value, ctxSlug)
	return nil
}

// updateEntryByRef creates or updates an entry to the cache from a cacheRef
// and a field json name (no path).
func (c *cache) updateEntryByRef(ref cacheRef, jsonName string, value interface{}, ctxSlug string) {
	if _, ok := c.data[ref]; !ok {
		c.data[ref] = make(FieldMap)
		c.data[ref]["id"] = ref.id
	}
	fi := ref.model.fields.MustGet(jsonName)
	switch fi.fieldType {
	case fieldtype.One2Many:
		switch ids := value.(type) {
		case int64:
			// Related field through O2M, we do not have the complete O2M set.
			// ids is a single int64 here.
			c.updateEntry(fi.relatedModel, ids, fi.jsonReverseFK, ref.id, ctxSlug)
			c.setX2MValue(ref, jsonName, ids, ctxSlug)
		case []int64:
			for _, id := range ids {
				c.updateEntry(fi.relatedModel, id, fi.jsonReverseFK, ref.id, ctxSlug)
			}
			c.setDataValue(ref, jsonName, true)
		}

	case fieldtype.Rev2One:
		id := value.(int64)
		c.updateEntry(fi.relatedModel, id, fi.jsonReverseFK, ref.id, ctxSlug)
		c.setDataValue(ref, jsonName, true)

	case fieldtype.Many2Many:
		switch ids := value.(type) {
		case int64:
			// Related field through O2M, we do not have the complete M2M set
			// ids is a single int64 here.
			c.addM2MLink(fi, ref.id, []int64{ids})
			c.setX2MValue(ref, jsonName, ids, ctxSlug)
		case []int64:
			c.removeM2MLinks(fi, ref.id)
			c.addM2MLink(fi, ref.id, ids)
			c.setDataValue(ref, jsonName, true)
		}

	default:
		c.setDataValue(ref, jsonName, value)
	}
}

// setDataValue sets the value for the jsonName field of record ref to value
func (c *cache) setDataValue(ref cacheRef, jsonName string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.data[ref]; !ok {
		c.data[ref] = make(FieldMap)
		c.data[ref]["id"] = ref.id
	}
	c.data[ref][jsonName] = value
}

// setX2MValue sets the id for the jsonName field of record ref in the x2mRelation map
func (c *cache) setX2MValue(ref cacheRef, jsonName string, id int64, ctxSlug string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.x2mRelated[ref]; !ok {
		c.x2mRelated[ref] = make(map[string]map[string]int64)
	}
	if _, ok := c.x2mRelated[ref][jsonName]; !ok {
		c.x2mRelated[ref][jsonName] = make(map[string]int64)
	}
	if id == c.x2mRelated[ref][jsonName][""] {
		// We don't add the value if the id is the same as the default context
		return
	}
	c.x2mRelated[ref][jsonName][ctxSlug] = id
}

// getX2MValue return the X2MValue or the default value if the given ctxSlug does not exist in cache
//
// 2nd returned value is true if a suitable value has been found
//
// 3rd returned value is true if the returned value is the value for the default context
func (c *cache) getX2MValue(ref cacheRef, jsonName string, ctxSlug string) (int64, bool, bool) {
	var defaultVal bool
	res, ok := c.x2mRelated[ref][jsonName][ctxSlug]
	if !ok {
		res, ok = c.x2mRelated[ref][jsonName][""]
		defaultVal = true
	}
	return res, ok, defaultVal
}

// deleteFieldData removes the cache entry for the jsonName field of record ref
func (c *cache) deleteFieldData(ref cacheRef, jsonName string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data[ref], jsonName)
	if _, exists := c.x2mRelated[ref]; exists {
		delete(c.x2mRelated[ref], jsonName)
	}
}

// deleteData removes the cache entry for the whole record ref
func (c *cache) deleteData(ref cacheRef) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, ref)
	delete(c.x2mRelated, ref)
}

// removeM2MLinks removes all M2M links associated with the record with
// the given id on the given field
func (c *cache) removeM2MLinks(fi *Field, id int64) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.m2mLinks[fi.m2mRelModel]; !exists {
		return
	}
	index := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	for link := range c.m2mLinks[fi.m2mRelModel] {
		if link[index] == id {
			delete(c.m2mLinks[fi.m2mRelModel], link)
		}
	}
}

// addM2MLink adds an M2M link between this record with its given ID
// and the records given by values on the given field.
func (c *cache) addM2MLink(fi *Field, id int64, values []int64) {
	c.Lock()
	defer c.Unlock()
	if _, exists := c.m2mLinks[fi.m2mRelModel]; !exists {
		c.m2mLinks[fi.m2mRelModel] = make(map[[2]int64]bool)
	}
	ourIndex := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	theirIndex := (ourIndex + 1) % 2
	for _, val := range values {
		var newLink [2]int64
		newLink[ourIndex] = id
		newLink[theirIndex] = val
		c.m2mLinks[fi.m2mRelModel][newLink] = true
	}
}

// getM2MLinks returns the linked ids to this id through the given field.
func (c *cache) getM2MLinks(fi *Field, id int64) []int64 {
	if _, exists := c.m2mLinks[fi.m2mRelModel]; !exists {
		return []int64{}
	}
	var res []int64
	ourIndex := (strings.Compare(fi.m2mOurField.name, fi.m2mTheirField.name) + 1) / 2
	theirIndex := (ourIndex + 1) % 2
	for link := range c.m2mLinks[fi.m2mRelModel] {
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
	c.deleteData(cacheRef{model: mi, id: id})
	for _, fi := range mi.fields.registryByJSON {
		if fi.fieldType == fieldtype.Many2Many {
			c.removeM2MLinks(fi, id)
		}
	}
}

// removeEntry removes the given entry from cache
func (c *cache) removeEntry(mi *Model, id int64, fieldName, ctxSlug string) {
	if !c.checkIfInCache(mi, []int64{id}, []string{fieldName}, ctxSlug) {
		return
	}
	c.deleteFieldData(cacheRef{model: mi, id: id}, fieldName)
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
	ref, fName, err := c.getRelatedRef(mi, id, fieldName, ctxSlug)
	if err != nil {
		return nil
	}
	fi := ref.model.fields.MustGet(fName)
	switch fi.fieldType {
	case fieldtype.One2Many:
		var relIds []int64
		for cRef, cVal := range c.data {
			if cRef.model != fi.relatedModel {
				continue
			}
			if cVal[fi.jsonReverseFK] != ref.id {
				continue
			}
			relIds = append(relIds, cRef.id)
		}
		return relIds
	case fieldtype.Rev2One:
		for cRef, cVal := range c.data {
			if cRef.model != fi.relatedModel {
				continue
			}
			if cVal[fi.jsonReverseFK] != ref.id {
				continue
			}
			return cRef.id
		}
		return nil
	case fieldtype.Many2Many:
		return c.getM2MLinks(fi, ref.id)
	default:
		return c.data[ref][fName]
	}
}

// getRecord returns the whole record specified by modelName and id
// as it is currently in cache.
func (c *cache) getRecord(model *Model, id int64, ctxSlug string) FieldMap {
	res := make(FieldMap)
	ref := cacheRef{model: model, id: id}
	for _, fName := range c.data[ref].Keys() {
		res[fName] = c.get(model, id, fName, ctxSlug)
	}
	return res
}

// checkIfInCache returns true if all fields given by fieldNames are available
// in cache for all the records with the given ids in the given model.
func (c *cache) checkIfInCache(mi *Model, ids []int64, fieldNames []string, ctxSlug string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		for _, fName := range fieldNames {
			if !c.isInCache(mi, id, fName, ctxSlug) {
				return false
			}
		}
	}
	return true
}

// isInCache returns true if the related record through path and ctxSlug strictly exists
// (i.e. no default value for context)
func (c *cache) isInCache(mi *Model, id int64, path string, ctxSlug string) bool {
	ref, path, err := c.getStrictRelatedRef(mi, id, path, ctxSlug)
	if err != nil {
		return false
	}
	if _, ok := c.data[ref][path]; !ok {
		return false
	}
	return true
}

// getRelatedRef returns the cacheRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
func (c *cache) getRelatedRef(mi *Model, id int64, path string, ctxSlug string) (cacheRef, string, error) {
	return c.getRelatedRefCommon(mi, id, path, ctxSlug, false)
}

// getStrictRelatedRef returns the cacheRef and field name of the field that is
// defined by path when walking from the given model with the given ID.
//
// This method returns an error when the value for the given ctxSlug cannot be found.
func (c *cache) getStrictRelatedRef(mi *Model, id int64, path string, ctxSlug string) (cacheRef, string, error) {
	return c.getRelatedRefCommon(mi, id, path, ctxSlug, true)
}

// getRelatedRefCommon is the common implementation of getRelatedRef and getStrictRelatedRef.
func (c *cache) getRelatedRefCommon(mi *Model, id int64, path string, ctxSlug string, strict bool) (cacheRef, string, error) {
	if id == 0 {
		return cacheRef{}, "", errors.New("requested value on RecordSet with ID=0")
	}
	exprs := jsonizeExpr(mi, strings.Split(path, ExprSep))
	if len(exprs) > 1 {
		fkID, ok := c.get(mi, id, exprs[0], ctxSlug).(int64)
		if !ok {
			var defVal bool
			fkID, ok, defVal = c.getX2MValue(cacheRef{model: mi, id: id}, exprs[0], ctxSlug)
			if !ok || (strict && defVal) {
				return cacheRef{}, "", errors.New("requested value not in cache")
			}
		}
		relMI := mi.getRelatedModelInfo(exprs[0])
		return c.getRelatedRefCommon(relMI, fkID, strings.Join(exprs[1:], ExprSep), ctxSlug, strict)
	}
	return cacheRef{model: mi, id: id}, exprs[0], nil
}

// newCache creates a pointer to a new cache instance.
func newCache() *cache {
	res := cache{
		data:       make(map[cacheRef]FieldMap),
		x2mRelated: make(map[cacheRef]map[string]map[string]int64),
		m2mLinks:   make(map[*Model]map[[2]int64]bool),
	}
	return &res
}
