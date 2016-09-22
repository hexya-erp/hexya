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
	"strconv"
	"strings"

	"github.com/npiganeau/yep/yep/tools"
)

/*
computeData holds data to recompute another field.
- modelInfo is a pointer to the modelInfo instance to recompute
- compute is the name of the function to call on modelInfo
- path is the search string that will be used to find records to update
(e.g. path = "Profile.BestPost").
*/
type computeData struct {
	modelInfo *modelInfo
	compute   string
	path      string
}

// fieldsCollection is a collection of fieldInfo instances in a model.
type fieldsCollection struct {
	registryByName       map[string]*fieldInfo
	registryByJSON       map[string]*fieldInfo
	computedFields       []*fieldInfo
	computedStoredFields []*fieldInfo
	relatedFields        []*fieldInfo
	bootstrapped         bool
}

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

// storedFieldNames returns a slice with the names of all the stored fields
// If fields are given, return only names in the list
func (fc *fieldsCollection) storedFieldNames(fieldNames ...string) []string {
	var res []string
	for fName, fi := range fc.registryByName {
		var keepField bool
		if len(fieldNames) == 0 {
			keepField = true
		} else {
			for _, f := range fieldNames {
				if fName == f {
					keepField = true
					break
				}
			}
		}
		if fi.isStored() && keepField {
			res = append(res, fName)
		}
	}
	return res
}

// nonRelatedFieldJSONNames returns a slice with the JSON names of all the fields that
// are not relations.
func (fc *fieldsCollection) nonRelatedFieldJSONNames() []string {
	var res []string
	for fName, fi := range fc.registryByJSON {
		if fi.relatedModel == nil {
			res = append(res, fName)
		}
	}
	return res
}

/*
getComputedFields returns the slice of fieldInfo of the computed, but not
stored fields of the given modelName.
If fields are given, return only fieldInfo instances in the list
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

// newFieldsCollection returns a pointer to a new empty fieldsCollection with
// all maps initialized.
func newFieldsCollection() *fieldsCollection {
	return &fieldsCollection{
		registryByName: make(map[string]*fieldInfo),
		registryByJSON: make(map[string]*fieldInfo),
	}
}

/*
add adds the given fieldInfo to the fieldsCollection.
*/
func (fc *fieldsCollection) add(fInfo *fieldInfo) {
	name := fInfo.name
	jsonName := fInfo.json
	fc.registryByName[name] = fInfo
	fc.registryByJSON[jsonName] = fInfo
	if fInfo.computed() {
		if fInfo.stored {
			fc.computedStoredFields = append(fc.computedStoredFields, fInfo)
		} else {
			fc.computedFields = append(fc.computedFields, fInfo)
		}
	}
	if fInfo.related() {
		fc.relatedFields = append(fc.relatedFields, fInfo)
	}
}

// fieldInfo holds the meta information about a field
type fieldInfo struct {
	mi            *modelInfo
	name          string
	json          string
	description   string
	help          string
	stored        bool
	required      bool
	unique        bool
	index         bool
	compute       string
	depends       []string
	html          bool
	relatedModel  *modelInfo
	reverseFK     string
	selection     Selection
	fieldType     tools.FieldType
	groupOperator string
	size          int
	digits        tools.Digits
	structField   reflect.StructField
	relatedPath   string
	dependencies  []computeData
	inherits      bool
	noCopy        bool
}

// computed returns true if this field is computed
func (fi *fieldInfo) computed() bool {
	return fi.compute != ""
}

// related returns true if this field is related
func (fi *fieldInfo) related() bool {
	return fi.relatedPath != ""
}

// isStored returns true if this field is stored in database
func (fi *fieldInfo) isStored() bool {
	if fi.fieldType == tools.ONE2MANY ||
		fi.fieldType == tools.MANY2MANY ||
		fi.fieldType == tools.REV2ONE {
		// reverse fields are not stored
		return false
	}
	if (fi.computed() || fi.related()) && !fi.stored {
		// Computed and related non stored fields are not stored
		return false
	}
	return true
}

// createFieldInfo creates and returns a new fieldInfo pointer from the given
// StructField and modelInfo.
func createFieldInfo(sf reflect.StructField, mi *modelInfo) *fieldInfo {
	var (
		attrs map[string]bool
		tags  map[string]string
	)
	parseStructTag(sf.Tag.Get(defaultStructTagName), &attrs, &tags)

	_, stored := attrs["store"]
	_, required := attrs["required"]
	_, unique := attrs["unique"]
	_, index := attrs["index"]
	_, inherits := attrs["inherits"]
	_, noCopy := attrs["nocopy"]

	computeName := tags["compute"]
	relatedPath := tags["related"]
	sStr := tags["size"]
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

	fk, ok := tags["fk"]
	if typ == tools.ONE2MANY && !ok {
		tools.LogAndPanic(log, "'one2many' fields must define an 'fk' tag", "model", mi.name, "field", sf.Name, "type", typ)
	}

	if inherits && typ != tools.MANY2ONE && typ != tools.ONE2ONE {
		log.Warn("'inherits' should be set only on many2one or one2one fields", "model", mi.name, "field", sf.Name, "type", typ)
		inherits = false
	}

	sels, ok := tags["selection"]
	var selection Selection
	if ok {
		if sf.Type.Kind() == reflect.String {
			typ = tools.SELECTION
			selection = make(Selection)
			for _, sel := range strings.Split(sels, defaultTagDataDelim) {
				selParts := strings.Split(sel, "|")
				code := strings.TrimSpace(selParts[0])
				value := strings.TrimSpace(selParts[1])
				selection[code] = value
			}
		} else {
			log.Warn("'selection' should be set only on string type fields", "model", mi.name, "field", sf.Name, "type", typ)
			selection = nil
		}
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
		mi:            mi,
		compute:       computeName,
		stored:        stored,
		required:      required,
		unique:        unique,
		index:         index,
		depends:       depends,
		description:   desc,
		help:          tags["help"],
		fieldType:     typ,
		groupOperator: groupOp,
		structField:   sf,
		size:          size,
		digits:        digits,
		relatedPath:   relatedPath,
		inherits:      inherits,
		noCopy:        noCopy,
		reverseFK:     fk,
		selection:     selection,
	}
	return &fInfo
}

/*
processDepends populates the dependencies of each fieldInfo from the depends strings of
each fieldInfo instances.
*/
func processDepends() {
	for _, mi := range modelRegistry.registryByTableName {
		for _, fInfo := range mi.fields.registryByJSON {
			var refName string
			for _, depString := range fInfo.depends {
				if depString != "" {
					tokens := jsonizeExpr(mi, strings.Split(depString, ExprSep))
					refName = tokens[len(tokens)-1]
					path := strings.Join(tokens[:len(tokens)-1], ExprSep)
					targetComputeData := computeData{
						modelInfo: mi,
						compute:   fInfo.compute,
						path:      path,
					}
					refModelInfo := mi.getRelatedModelInfo(path)
					refField, ok := refModelInfo.fields.get(refName)
					if !ok {
						tools.LogAndPanic(log, "Unknown field in model", "model", refModelInfo.name, "field", refField)
					}
					refField.dependencies = append(refField.dependencies, targetComputeData)
				}
			}
		}
	}
}
