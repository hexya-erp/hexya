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
	"github.com/npiganeau/yep/yep/orm"
	"reflect"
	"strings"
	"sync"
)

var fieldsCache = &_fieldsCache{
	cache:                make(map[field]*fieldInfo),
	computedFields:       make(map[string][]*fieldInfo),
	computedStoredFields: make(map[string][]*fieldInfo),
	dependencyMap:        make(map[field][]computeData),
}

/*
field is the key to find a field in the fieldsCache
*/
type field struct {
	modelName string
	name      string
}

/*
computeData holds data to recompute another field.
- modelName is the name of the model to recompute
- compute is the name of the function to call on modelName
- path is the search string that will be used to find records to update.
The path should take an ID as argument (e.g. path = "Profile__BestPost").
*/
type computeData struct {
	modelName string
	compute   string
	path      string
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
	depends     []string
}

// fieldsCache is the fieldInfo collection
type _fieldsCache struct {
	sync.RWMutex
	cache                map[field]*fieldInfo
	computedFields       map[string][]*fieldInfo
	computedStoredFields map[string][]*fieldInfo
	dependencyMap        map[field][]computeData
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
getDependentFields return the fields that must be recomputed when ref is modified.
*/
func (fc *_fieldsCache) getDependentFields(ref field) (target []computeData, ok bool) {
	target, ok = fc.dependencyMap[ref]
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
setDependency adds a dependency in the dependencyMap.
target field depends on ref field, i.e. when ref field is modified,
target field must be recomputed.
*/
func (fc *_fieldsCache) setDependency(ref field, target computeData) {
	fc.dependencyMap[ref] = append(fc.dependencyMap[ref], target)
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
		_, stored := attrs["store"]
		depends := strings.Split(tags["depends"], defaultDependsTagDelim)
		fInfo := fieldInfo{
			name:        sf.Name,
			modelName:   name,
			compute:     computeName,
			computed:    computed,
			stored:      stored,
			depends:     depends,
			description: desc,
			help:        tags["help"],
		}
		fieldsCache.set(field{name: fInfo.name, modelName: fInfo.modelName}, &fInfo)
	}
}

/*
processDepends populates the dependsMap of the fieldsCache from the depends strings of
each fieldInfo instance.
*/
func processDepends() {
	for targetField, fInfo := range fieldsCache.cache {
		var (
			refName string
		)
		for _, depString := range fInfo.depends {
			if depString != "" {
				tokens := strings.Split(depString, orm.ExprSep)
				refName = tokens[len(tokens)-1]
				refModelName := getRelatedModelName(targetField.modelName, depString)
				refField := field{
					modelName: refModelName,
					name:      refName,
				}
				path := strings.Join(tokens[:len(tokens)-1], orm.ExprSep)
				targetComputeData := computeData{
					modelName: fInfo.modelName,
					compute:   fInfo.compute,
					path:      path,
				}
				fieldsCache.setDependency(refField, targetComputeData)
			}
		}
	}
}

/*
getRelatedModelName returns the model name of the field given by path calculated from the origin model.
path is a query string as used in RecordSet.Filter()
*/
func getRelatedModelName(origin, path string) string {
	qs := orm.NewOrm().QueryTable(origin)
	modelName, _ := qs.TargetModelField(path)
	return modelName
}
