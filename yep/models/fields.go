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
	"reflect"
	"strings"
	"sync"

	"github.com/npiganeau/yep/yep/tools"
	"strconv"
)

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

// fieldsCollection is a collection of fieldInfo instances in a model.
type fieldsCollection struct {
	sync.RWMutex
	registryByName       map[string]*fieldInfo
	registryByJSON       map[string]*fieldInfo
	computedFields       []*fieldInfo
	computedStoredFields []*fieldInfo
	bootstrapped         bool
	//dependencyMap        map[fieldRef][]computeData
}

// newFieldsCollection returns a pointer to a new empty fieldsCollection with
// all maps initialized.
func newFieldsCollection() *fieldsCollection {
	return &fieldsCollection{
		registryByName: make(map[string]*fieldInfo),
		registryByJSON: make(map[string]*fieldInfo),
		//dependencyMap:        make(map[fieldRef][]computeData),
	}
}

// fieldInfo holds the meta information about a field
type fieldInfo struct {
	modelInfo     *modelInfo
	name          string
	json          string
	description   string
	help          string
	computed      bool
	stored        bool
	required      bool
	unique        bool
	index         bool
	compute       string
	depends       []string
	html          bool
	relatedModel  *modelInfo
	fieldType     tools.FieldType
	groupOperator string
	size          int
	digits        tools.Digits
	structField   reflect.StructField
}

// isStored returns true if this field is stored in database
func (fi *fieldInfo) isStored() bool {
	if fi.fieldType == tools.ONE2MANY ||
		fi.fieldType == tools.MANY2MANY ||
		fi.fieldType == tools.REV2ONE {
		// reverse fields are not stored
		return false
	}
	if fi.computed && !fi.stored {
		// Computed non stored fields are not stored
		return false
	}
	return true
}

///*
//field is the key to find a field in the fieldsCache
//*/
//type fieldRef struct {
//	modelName string
//	name      string
//}
//
//// ConvertToName converts the given field ref to a fieldRef
//// of type [modelName, fieldName].
//func (fr *fieldRef) ConvertToName() {
//	fi, ok := fieldsCache.get(*fr)
//	if !ok {
//		panic(fmt.Errorf("unknown fieldRef `%s`", *fr))
//	}
//	fr.name = fi.name
//}

/*
get returns the fieldInfo of the field with the given name.
name can be either the name of the field or its JSON name.
*/
func (fc *fieldsCollection) get(name string) (fi *fieldInfo, ok bool) {
	fi, ok = fc.registryByName[name]
	if !ok {
		fi, ok = fc.registryByJSON[name]
	}
	return
}

/*
getComputedFields returns the slice of fieldInfo of the computed, but not
stored fields of the given modelName.
If fields are given, return only fieldInfo in the list
*/
func (fc *fieldsCollection) getComputedFields(fields ...string) (fil []*fieldInfo) {
	fInfos := fc.computedFields
	if len(fields) > 0 {
		for _, f := range fields {
			for _, fInfo := range fInfos {
				if fInfo.name == tools.ConvertMethodName(f) {
					fil = append(fil, fInfo)
					continue
				}
			}
		}
	} else {
		fil = fInfos
	}
	return
}

/*
getComputedStoredFields returns the slice of fieldInfo of the computed and stored
fields of the given modelName.
*/
func (fc *fieldsCollection) getComputedStoredFields() (fil []*fieldInfo) {
	fil = fc.computedStoredFields
	return
}

///*
//getDependentFields return the fields that must be recomputed when ref is modified.
//*/
//func (fc *_fieldsCache) getDependentFields(ref fieldRef) (target []computeData, ok bool) {
//	target, ok = fc.dependencyMap[ref]
//	return
//}

/*
add adds the given fieldInfo to the fieldsCollection.
*/
func (fc *fieldsCollection) add(fInfo *fieldInfo) {
	name := fInfo.name
	jsonName := fInfo.json
	fc.registryByName[name] = fInfo
	fc.registryByJSON[jsonName] = fInfo
	if fInfo.computed {
		if fInfo.stored {
			fc.computedStoredFields = append(fc.computedStoredFields, fInfo)
		} else {
			fc.computedFields = append(fc.computedFields, fInfo)
		}
	}
}

///*
//setDependency adds a dependency in the dependencyMap.
//target field depends on ref field, i.e. when ref field is modified,
//target field must be recomputed.
//*/
//func (fc *_fieldsCache) setDependency(ref fieldRef, target computeData) {
//	fc.dependencyMap[ref] = append(fc.dependencyMap[ref], target)
//}

// createFieldInfo creates and returns a new fieldInfo pointer from the given
// StructField and modelInfo.
func createFieldInfo(sf reflect.StructField, mi *modelInfo) *fieldInfo {
	var (
		attrs map[string]bool
		tags  map[string]string
	)
	parseStructTag(sf.Tag.Get(defaultStructTagName), &attrs, &tags)

	_, stored := attrs["store"]
	_, html := attrs["html"]
	_, required := attrs["required"]
	_, unique := attrs["unique"]
	_, index := attrs["index"]

	computeName, computed := tags["compute"]
	sStr, _ := tags["size"]
	size, _ := strconv.Atoi(sStr)

	var depends []string
	if depTag, ok := tags["depends"]; ok {
		depends = strings.Split(depTag, defaultTagDataDelim)
	}

	var digits tools.Digits
	if dTag, ok := tags["digits"]; ok {
		dSlice := strings.Split(dTag, defaultTagDataDelim)
		d0, _ := strconv.Atoi(dSlice[0])
		d1, _ := strconv.Atoi(dSlice[1])
		digits = tools.Digits{0: d0, 1: d1}
	}

	desc, ok := tags["string"]
	if !ok {
		desc = sf.Name
	}

	typStr, ok := tags["type"]
	typ := tools.FieldType(typStr)
	if !ok {
		typ = getFieldType(sf.Type)
	}

	json, ok := tags["json"]
	if !ok {
		json = tools.SnakeCaseString(sf.Name)
		if typ == tools.MANY2ONE || typ == tools.ONE2ONE {
			json += "_id"
		} else if typ == tools.ONE2MANY || typ == tools.MANY2MANY {
			json += "_ids"
		}
	}

	groupOp, ok := tags["group_operator"]
	if !ok {
		groupOp = "sum"
	}

	fInfo := fieldInfo{
		name:          sf.Name,
		json:          json,
		modelInfo:     mi,
		compute:       computeName,
		computed:      computed,
		stored:        stored,
		required:      required,
		unique:        unique,
		index:         index,
		depends:       depends,
		description:   desc,
		help:          tags["help"],
		html:          html,
		fieldType:     typ,
		groupOperator: groupOp,
		structField:   sf,
		size:          size,
		digits:        digits,
	}
	return &fInfo
}

///*
//processDepends populates the dependsMap of the fieldsCache from the depends strings of
//each fieldInfo instance.
//*/
//func processDepends() {
//	for targetField, fInfo := range fieldsCache.cache {
//		var (
//			refName string
//		)
//		for _, depString := range fInfo.depends {
//			if depString != "" {
//				tokens := strings.Split(depString, orm.ExprSep)
//				refName = tokens[len(tokens)-1]
//				refModelName := getRelatedModelName(targetField.modelName, depString)
//				refField := fieldRef{
//					modelName: refModelName,
//					name:      refName,
//				}
//				path := strings.Join(tokens[:len(tokens)-1], orm.ExprSep)
//				targetComputeData := computeData{
//					modelName: fInfo.modelName,
//					compute:   fInfo.compute,
//					path:      path,
//				}
//				fieldsCache.setDependency(refField, targetComputeData)
//			}
//		}
//	}
//}

///*
//getRelatedModelName returns the model name of the field given by path calculated from the origin model.
//path is a query string as used in RecordSet.Filter()
//*/
//func getRelatedModelName(origin, path string) string {
//	qs := orm.NewOrm().QueryTable(origin)
//	modelName, _ := qs.TargetModelField(path)
//	return modelName
//}
