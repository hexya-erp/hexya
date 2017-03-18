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

	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/logging"
)

/*
computeData holds data to recompute another field.
- Model is a pointer to the Model instance to recompute
- compute is the name of the function to call on Model
- path is the search string that will be used to find records to update
(e.g. path = "Profile.BestPost").
*/
type computeData struct {
	modelInfo *Model
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
	relatedFields        []*fieldInfo
	bootstrapped         bool
}

// get returns the fieldInfo of the field with the given name.
// name can be either the name of the field or its JSON name.
func (fc *fieldsCollection) get(name string) (fi *fieldInfo, ok bool) {
	fi, ok = fc.registryByName[name]
	if !ok {
		fi, ok = fc.registryByJSON[name]
	}
	return
}

// MustGet returns the fieldInfo of the field with the given name or panics
// name can be either the name of the field or its JSON name.
func (fc *fieldsCollection) mustGet(name string) *fieldInfo {
	fi, ok := fc.get(name)
	if !ok {
		var model string
		for _, f := range fc.registryByName {
			model = f.model.name
			break
		}
		logging.LogAndPanic(log, "Unknown field in model", "model", model, "field", name)
	}
	return fi
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

// relatedNonStoredFieldNames returns a slice with all the related
// non-stored field names.
func (fc *fieldsCollection) relatedNonStoredFieldNames() []string {
	var res []string
	for _, fi := range fc.relatedFields {
		if !fi.stored {
			res = append(res, fi.name)
		}
	}
	return res
}

// nonRelatedFieldJSONNames returns a slice with the JSON names of all the fields that
// are not relations.
func (fc *fieldsCollection) nonRelationFieldJSONNames() []string {
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

// add the given fieldInfo to the fieldsCollection.
func (fc *fieldsCollection) add(fInfo *fieldInfo) {
	if _, exists := fc.registryByName[fInfo.name]; exists {
		logging.LogAndPanic(log, "Trying to add already existing field", "model", fInfo.model.name, "field", fInfo.name)
	}
	fc.register(fInfo)
}

// override a fieldInfo in the collection.
// Mapping is done on the fInfo name.
func (fc *fieldsCollection) override(fInfo *fieldInfo) {
	if _, exists := fc.registryByName[fInfo.name]; !exists {
		logging.LogAndPanic(log, "Trying to override a non-existant field", "model", fInfo.model.name, "field", fInfo.name)
	}
	fc.register(fInfo)
}

// register adds or override the given fInfo in the collection.
func (fc *fieldsCollection) register(fInfo *fieldInfo) {
	fc.Lock()
	defer fc.Unlock()

	checkFieldInfo(fInfo)
	name := fInfo.name
	jsonName := fInfo.json
	fc.registryByName[name] = fInfo
	fc.registryByJSON[jsonName] = fInfo
	if fInfo.isComputedField() {
		if fInfo.stored {
			fc.computedStoredFields = append(fc.computedStoredFields, fInfo)
		} else {
			fc.computedFields = append(fc.computedFields, fInfo)
		}
	}
	if fInfo.isRelatedField() {
		fc.relatedFields = append(fc.relatedFields, fInfo)
	}
}

// fieldInfo holds the meta information about a field
type fieldInfo struct {
	model            *Model
	acl              *security.AccessControlList
	name             string
	json             string
	description      string
	help             string
	stored           bool
	required         bool
	unique           bool
	index            bool
	compute          string
	depends          []string
	relatedModelName string
	relatedModel     *Model
	reverseFK        string
	m2mRelModel      *Model
	m2mOurField      *fieldInfo
	m2mTheirField    *fieldInfo
	selection        Selection
	fieldType        types.FieldType
	groupOperator    string
	size             int
	digits           types.Digits
	structField      reflect.StructField
	relatedPath      string
	dependencies     []computeData
	embed            bool
	noCopy           bool
}

// isComputedField returns true if this field is computed
func (fi *fieldInfo) isComputedField() bool {
	return fi.compute != ""
}

// isComputedField returns true if this field is related
func (fi *fieldInfo) isRelatedField() bool {
	return fi.relatedPath != ""
}

// isRelationField returns true if this field points to another model
func (fi *fieldInfo) isRelationField() bool {
	// We check on relatedModelName and not relatedModel to be able
	// to use this method even if the models have not been bootstrapped yet.
	return fi.relatedModelName != ""
}

// isStored returns true if this field is stored in database
func (fi *fieldInfo) isStored() bool {
	if fi.fieldType.IsNonStoredRelationType() {
		// reverse fields are not stored
		return false
	}
	if (fi.isComputedField() || fi.isRelatedField()) && !fi.stored {
		// Computed and related non stored fields are not stored
		return false
	}
	return true
}

// checkFieldInfo makes sanity checks on the given fieldInfo.
// It panics in case of severe error and logs recoverable errors.
func checkFieldInfo(fi *fieldInfo) {
	if fi.fieldType.IsReverseRelationType() && fi.reverseFK == "" {
		logging.LogAndPanic(log, "'one2many' and 'rev2one' fields must define an 'fk' tag", "model",
			fi.model.name, "field", fi.name, "type", fi.fieldType)
	}

	if fi.embed && !fi.fieldType.IsStoredRelationType() {
		log.Warn("'embed' should be set only on many2one or one2one fields", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
		fi.embed = false
	}

	if fi.structField.Type == reflect.TypeOf(RecordCollection{}) && fi.relatedModel.name == "" {
		logging.LogAndPanic(log, "Undefined comodel on related field", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
	}

	if fi.stored && !fi.isComputedField() {
		log.Warn("'store' should be set only on computed fields", "model", fi.model.name, "field", fi.name,
			"type", fi.fieldType)
		fi.stored = false
	}

	if fi.selection != nil && fi.structField.Type.Kind() != reflect.String {
		logging.LogAndPanic(log, "'selection' tag can only be set on string types", "model", fi.model.name, "field", fi.name)
	}
}

// jsonizeFieldName returns a snake cased field name, adding '_id' on x2one
// relation fields and '_ids' to x2many relation fields.
func snakeCaseFieldName(fName string, typ types.FieldType) string {
	res := tools.SnakeCaseString(fName)
	if typ.Is2OneRelationType() {
		res += "_id"
	} else if typ.Is2ManyRelationType() {
		res += "_ids"
	}
	return res
}

// createM2MRelModelInfo creates a Model relModelName (if it does not exist)
// for the m2m relation defined between model1 and model2.
// It returns the Model of the intermediate model, the fieldInfo of that model
// pointing to our model, and the fieldInfo pointing to the other model.
func createM2MRelModelInfo(relModelName, model1, model2 string) (*Model, *fieldInfo, *fieldInfo) {
	if relMI, exists := Registry.get(relModelName); exists {
		var m1, m2 *fieldInfo
		for fName, fi := range relMI.fields.registryByName {
			if fName == model1 {
				m1 = fi
			} else if fName == model2 {
				m2 = fi
			}
		}
		return relMI, m1, m2
	}

	newMI := &Model{
		name:      relModelName,
		acl:       security.NewAccessControlList(),
		tableName: tools.SnakeCaseString(relModelName),
		fields:    newFieldsCollection(),
		methods:   newMethodsCollection(),
		options:   Many2ManyLinkModel,
	}
	ourField := &fieldInfo{
		name:             model1,
		json:             tools.SnakeCaseString(model1) + "_id",
		acl:              security.NewAccessControlList(),
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        types.Many2One,
		relatedModelName: model1,
		index:            true,
		structField: reflect.StructField{
			Name: model1,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(ourField)

	theirField := &fieldInfo{
		name:             model2,
		json:             tools.SnakeCaseString(model2) + "_id",
		acl:              security.NewAccessControlList(),
		model:            newMI,
		required:         true,
		noCopy:           true,
		fieldType:        types.Many2One,
		relatedModelName: model2,
		index:            true,
		structField: reflect.StructField{
			Name: model2,
			Type: reflect.TypeOf(int64(0)),
		},
	}
	newMI.fields.add(theirField)
	Registry.add(newMI)
	return newMI, ourField, theirField
}

// processDepends populates the dependencies of each fieldInfo from the depends strings of
// each fieldInfo instances.
func processDepends() {
	for _, mi := range Registry.registryByTableName {
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
					refField := refModelInfo.fields.mustGet(refName)
					refField.dependencies = append(refField.dependencies, targetComputeData)
				}
			}
		}
	}
}

// checkComputeMethodsSignature checks all methods used in computed
// fields and check their signature. It panics if it is not the case.
func checkComputeMethodsSignature() {
	checkMethType := func(method *methodInfo, stored bool) {
		methType := method.methodType
		var msg string
		switch {
		case methType.NumIn() != 1:
			msg = "Compute methods should have no arguments"
		case methType.NumOut() == 0:
			msg = "Compute methods should return a value"
		case methType.NumOut() == 1 && stored:
			msg = "Compute methods for stored field must return fields to unset as second value"
		case methType.NumOut() == 2 && methType.Out(1) != reflect.TypeOf([]FieldNamer{}):
			msg = "Second return value of compute methods must be []models.FieldNamer"
		case methType.NumOut() > 2:
			msg = "Too many return values for compute method"
		}
		if msg != "" {
			logging.LogAndPanic(log, msg, "model", method.mi.name, "method", method.name)
		}
	}
	for _, mi := range Registry.registryByName {
		for _, fi := range mi.fields.computedFields {
			method := mi.methods.mustGet(fi.compute)
			checkMethType(method, false)
		}
		for _, fi := range mi.fields.computedStoredFields {
			method := mi.methods.mustGet(fi.compute)
			checkMethType(method, true)
		}
	}
}
