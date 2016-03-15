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
	"fmt"
	"reflect"
	"sync"
)

var fieldsCache = &_fieldsCache{
	cache:                make(map[field]*fieldInfo),
	computedFields:       make(map[string][]*fieldInfo),
	computedStoredFields: make(map[string][]*fieldInfo),
}

/*
field is the key to find a field in the fieldsCache
*/
type field struct {
	modelName string
	name      string
}

/*
fieldInfo holds the meta information about a field
*/
type fieldInfo struct {
	modelName   string
	name        string
	description string
	help        string
	computed    bool
	stored      bool
	compute     string
}

// fieldsCache is the fieldInfo collection
type _fieldsCache struct {
	sync.RWMutex
	cache                map[field]*fieldInfo
	computedFields       map[string][]*fieldInfo
	computedStoredFields map[string][]*fieldInfo
	done                 bool
}

/*
get returns the fieldInfo of the given method.
*/
func (fc *_fieldsCache) get(ref field) (fi *fieldInfo, ok bool) {
	fi, ok = fc.cache[ref]
	return
}

/*
getComputedFields returns the slice of fieldInfo of the computed, but not
stored fields of the given modelName.
*/
func (fc *_fieldsCache) getComputedFields(modelName string) (fil []*fieldInfo, ok bool) {
	fil, ok = fc.computedFields[modelName]
	return
}

/*
getComputedStoredFields returns the slice of fieldInfo of the computed and stored
fields of the given modelName.
*/
func (fc *_fieldsCache) getComputedStoredFields(modelName string) (fil []*fieldInfo, ok bool) {
	fil, ok = fc.computedStoredFields[modelName]
	return
}

/*
set adds the given fieldInfo to the fieldsCache.
*/
func (fc *_fieldsCache) set(ref field, fInfo *fieldInfo) {
	fc.cache[ref] = fInfo
	if fInfo.computed {
		if fInfo.stored {
			fc.computedStoredFields[fInfo.modelName] = append(fc.computedStoredFields[fInfo.modelName], fInfo)
		} else {
			fc.computedFields[fInfo.modelName] = append(fc.computedFields[fInfo.modelName], fInfo)
		}
	}
}

/*
registerModelFields populates the fieldsCache with the given structPtr fields
*/
func registerModelFields(name string, structPtr interface{}) {
	var (
		attrs map[string]bool
		tags  map[string]string
	)

	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)
	typ := ind.Type()

	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		panic(fmt.Errorf("<models.registerModelFields> cannot use non-ptr model struct `%s`", getModelName(typ)))
	}

	for i := 0; i < ind.NumField(); i++ {
		sf := ind.Type().Field(i)
		if sf.PkgPath != "" {
			continue
		}
		parseStructTag(sf.Tag.Get(defaultStructTagName), &attrs, &tags)
		desc, ok := tags["string"]
		if !ok {
			desc = sf.Name
		}
		computeName, computed := tags["compute"]
		_, stored := attrs["stored"]
		fInfo := fieldInfo{
			name:        sf.Name,
			modelName:   name,
			compute:     computeName,
			computed:    computed,
			stored:      stored,
			description: desc,
			help:        tags["help"],
		}
		fieldsCache.set(field{name: fInfo.name, modelName: fInfo.modelName}, &fInfo)
	}
}
